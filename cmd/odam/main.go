package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math"
	"time"

	"net/http"

	darknet "github.com/LdDl/go-darknet"
	"github.com/LdDl/gocv-blob/blob"
	"github.com/LdDl/odam"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
)

var (
	settingsFile = flag.String("settings", "conf.json", "Path to application's settings")
)

var (
	window          *gocv.Window
	stream          *mjpeg.Stream
	imagesChannel   chan *odam.FrameData
	detectedChannel chan []odam.DetectedObject
	detected        []odam.DetectedObject
	allblobies      *blob.Blobies
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}

func main() {

	// Settings
	flag.Parse()
	settings, err := odam.NewSettings(*settingsFile)
	if err != nil {
		log.Fatalln(err)
	}

	// MJPEG server
	if settings.MjpegSettings.Enable {
		stream = mjpeg.NewStream()
		go func() {
			fmt.Printf("Starting MJPEG on http://localhost:%d\n", settings.MjpegSettings.Port)
			http.Handle("/", stream)
			log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", settings.MjpegSettings.Port), nil))
		}()
	}

	// gRPC sender
	var grpcConn *grpc.ClientConn
	if settings.GrpcSettings.Enable {
		url := fmt.Sprintf("%s:%d", settings.GrpcSettings.ServerIP, settings.GrpcSettings.ServerPort)
		grpcConn, err = grpc.Dial(url, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Println(err)
			return
		}
		defer grpcConn.Close()
	}
	// Neural network
	neuralNet := darknet.YOLONetwork{
		GPUDeviceIndex:           0,
		NetworkConfigurationFile: settings.NeuralNetworkSettings.DarknetCFG,
		WeightsFile:              settings.NeuralNetworkSettings.DarknetWeights,
		Threshold:                float32(settings.NeuralNetworkSettings.ConfThreshold),
	}
	if err := neuralNet.Init(); err != nil {
		log.Fatalln(err)
	}
	defer neuralNet.Close()

	// Tracker
	allblobies = blob.NewBlobiesDefaults()
	allblobies.DrawingOptions = settings.TrackerSettings.DrawOptions

	// Video capture
	videoCapturer, err := gocv.OpenVideoCapture(settings.VideoSettings.Source)
	if err != nil {
		log.Fatalln(err)
	}
	if settings.MjpegSettings.ImshowEnable {
		fmt.Println("Press 'ESC' to stop imshow()")
		window = gocv.NewWindow("ODAM v0.1.0")
		window.ResizeWindow(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
		defer window.Close()
	}

	// Initial channels
	imagesChannel = make(chan *odam.FrameData, 1)
	detectedChannel = make(chan []odam.DetectedObject)
	img := odam.NewFrameData()

	// Read first frame
	if ok := videoCapturer.Read(&img.ImgSource); !ok {
		log.Fatalf("Error cannot read video %v\n", settings.VideoSettings.Source)
	}
	err = img.Preprocess(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
	if err != nil {
		log.Fatalln("First preprocess step:", err)
	}

	// First step of processing
	processFrame(img)
	go performDetection(&neuralNet, settings.NeuralNetworkSettings.TargetClasses)

	// Read frames in a loop
	for {
		if ok := videoCapturer.Read(&img.ImgSource); !ok {
			fmt.Println("Can't read next frame, stop grabbing")
			break
		}
		if img.ImgSource.Empty() {
			fmt.Println("Empty frame has been detected. Sleep for 400 ms")
			time.Sleep(400 * time.Millisecond)
			continue
		}
		err := img.Preprocess(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
		if err != nil {
			fmt.Println("Can't preprocess. Sleep for 400ms:", err)
			time.Sleep(400 * time.Millisecond)
			continue
		}
		select {
		case detected = <-detectedChannel:
			processFrame(img)
			if len(detected) != 0 {
				detectedObject := make([]*blob.Blobie, len(detected))
				for i := range detected {
					detectedObject[i] = blob.NewBlobie(detected[i].Rect, 10, detected[i].ClassID, detected[i].ClassName)
					detectedObject[i].SetDraw(allblobies.DrawingOptions)
				}
				allblobies.MatchToExisting(detectedObject)
				for _, vline := range settings.TrackerSettings.LinesSettings {
					for _, b := range allblobies.Objects {
						shift := 20
						// shift = b.Center.Y + b.CurrentRect.Dy()/2
						if b.IsCrossedTheLineWithShift(vline.VLine.RightPT.Y, vline.VLine.LeftPT.X, vline.VLine.RightPT.X, vline.VLine.Direction, shift) {
							minx, miny := math.Floor(float64(b.CurrentRect.Min.X)*settings.VideoSettings.ScaleX), math.Floor(float64(b.CurrentRect.Min.Y)*settings.VideoSettings.ScaleY)
							maxx, maxy := math.Floor(float64(b.CurrentRect.Max.X)*settings.VideoSettings.ScaleX), math.Floor(float64(b.CurrentRect.Max.Y)*settings.VideoSettings.ScaleY)
							cropRect := image.Rect(
								int(minx)+5,  // add a bit width to crop bigger region
								int(miny)+10, // add a bit height to crop bigger region
								int(maxx)+5,
								int(maxy)+10,
							)
							odam.FixRectForOpenCV(&cropRect, settings.VideoSettings.Width, settings.VideoSettings.Height)
							cropImage := img.ImgSource.Region(cropRect)
							copyCrop := cropImage.Clone()

							cropImageSTD, err := copyCrop.ToImage()
							if err != nil {
								fmt.Println("can't convert cropped gocv.Mat to image.Image:", err)
								cropImage.Close()
								copyCrop.Close()
								continue
							}
							cropImage.Close()
							copyCrop.Close()

							buf := new(bytes.Buffer)
							err = jpeg.Encode(buf, cropImageSTD, nil)
							sendS3 := buf.Bytes()
							sendData := odam.ObjectInformation{
								CamId:     settings.VideoSettings.CameraID,
								Timestamp: time.Now().UTC().Unix(),
								Image:     sendS3,
								Detection: &odam.Detection{
									XLeft:  0,
									YTop:   0,
									Width:  int32(cropRect.Dx()),
									Height: int32(cropRect.Dy()),
								},
								Class: &odam.ClassInfo{
									ClassId:   int32(b.GetClassID()),
									ClassName: b.GetClassName(),
								},
								VirtualLine: &odam.VirtualLineInfo{
									Id:     vline.LineID,
									LeftX:  int32(vline.VLine.SourceLeftPT.X),
									LeftY:  int32(vline.VLine.SourceLeftPT.Y),
									RightX: int32(vline.VLine.SourceRightPT.X),
									RightY: int32(vline.VLine.SourceRightPT.Y),
								},
							}

							if settings.GrpcSettings.Enable {
								go sendDataToServer(grpcConn, &sendData)
							}
							// result := gocv.IMWrite("dets/"+i.String()+".jpeg", cropImage)
							// fmt.Println("saved?", result, i)
						}
					}
				}
			}
		default:
			// show current frame without blocking, so do nothing here
		}

		if settings.MjpegSettings.ImshowEnable || settings.MjpegSettings.Enable {
			for i := range settings.TrackerSettings.LinesSettings {
				settings.TrackerSettings.LinesSettings[i].VLine.Draw(&img.ImgScaled)
			}
			for i, b := range (*allblobies).Objects {
				_ = i
				// (*b).DrawTrack(&img.ImgScaled, fmt.Sprintf("%v", i))
				(*b).DrawTrack(&img.ImgScaled, "")
			}
		}
		if settings.MjpegSettings.ImshowEnable {

			window.IMShow(img.ImgScaled)
			if window.WaitKey(1) == 27 {
				break
			}
		}
		if settings.MjpegSettings.Enable {
			buf, err := gocv.IMEncode(".jpg", img.ImgScaled)
			if err != nil {
				log.Println("Error while decoding to JPG (mjpeg)", err)
			} else {
				stream.UpdateJPEG(buf)
			}
		}

	}

	fmt.Println("Shutting down...")
	time.Sleep(2 * time.Second) // @todo temporary fix: need to wait a bit time for last call of neuralNet.Detect(...)

	// hard release memory
	img.Close()

	// pprof
	if settings.MatPPROFSettings.Enable {
		var b bytes.Buffer
		// go run -tags matprofile main.go
		// gocv.MatProfile.WriteTo(&b, 1)
		fmt.Print(b.String())
	}
}

func processFrame(fd *odam.FrameData) {
	frame := odam.NewFrameData()
	fd.ImgSource.CopyTo(&frame.ImgSource)
	fd.ImgScaled.CopyTo(&frame.ImgScaled)
	frame.ImgSTD = fd.ImgSTD
	imagesChannel <- frame
}

func performDetection(neuralNet *darknet.YOLONetwork, targetClasses []string) {
	fmt.Println("Start performDetection thread")
	for {

		frame := <-imagesChannel

		darknetImage, err := darknet.Image2Float32(frame.ImgSTD)
		if err != nil {
			log.Println("Image2Float32 error:", err)
			frame.Close()
			// Error: no handling
			time.Sleep(100 * time.Millisecond)
			continue
		}

		dr, err := neuralNet.Detect(darknetImage)
		if err != nil {
			frame.Close()
			darknetImage.Close()
			fmt.Println("Detect error:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		darknetImage.Close() // free the memory
		darknetImage = nil

		detectedRects := make([]odam.DetectedObject, 0, len(dr.Detections))
		for _, d := range dr.Detections {
			for i := range d.ClassIDs {
				if stringInSlice(&d.ClassNames[i], targetClasses) {
					bBox := d.BoundingBox
					minX, minY := float64(bBox.StartPoint.X), float64(bBox.StartPoint.Y)
					maxX, maxY := float64(bBox.EndPoint.X), float64(bBox.EndPoint.Y)
					rect := odam.DetectedObject{
						Rect:       image.Rect(odam.Round(minX), odam.Round(minY), odam.Round(maxX), odam.Round(maxY)),
						ClassName:  d.ClassNames[i],
						ClassID:    d.ClassIDs[i],
						Confidence: d.Probabilities[i],
					}
					detectedRects = append(detectedRects, rect)
				}
			}
		}

		frame.Close() // free the memory
		detectedChannel <- detectedRects
	}
}

func stringInSlice(str *string, sl []string) bool {
	for i := range sl {
		if sl[i] == *str {
			return true
		}
	}
	return false
}

func sendDataToServer(grpcConn *grpc.ClientConn, data *odam.ObjectInformation) {
	client := odam.NewServiceYOLOClient(grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := client.SendDetection(
		ctx,
		data,
	)
	if err != nil {
		log.Println("grpc send error:", err)
		return
	}

	if len(r.GetError()) != 0 {
		log.Println("grpc accepts error:", r.GetError())
		return
	}

	if len(r.GetWarning()) != 0 {
		log.Println("grpc accepts warning:", r.GetWarning())
		return
	}

	log.Println("grpc answer:", r.GetMessage())
	return
}

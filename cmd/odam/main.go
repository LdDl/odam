package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"log"
	"net/http"
	"time"

	darknet "github.com/LdDl/go-darknet"
	"github.com/LdDl/go-lpr/utils"
	"github.com/LdDl/gocv-blob/blob"
	"github.com/LdDl/odam"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
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
	neuralNet       darknet.YOLONetwork
)

func main() {

	// Settings
	flag.Parse()
	settings, err := odam.NewSettings(*settingsFile)
	if err != nil {
		log.Fatalln(err)
	}

	if settings.MjpegSettings.Enable {
		stream = mjpeg.NewStream()
		go func() {
			fmt.Printf("Starting MJPEG on http://localhost:%d\n", settings.MjpegSettings.Port)
			http.Handle("/", stream)
			log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", settings.MjpegSettings.Port), nil))
		}()
	}

	neuralNet = darknet.YOLONetwork{
		GPUDeviceIndex:           0,
		NetworkConfigurationFile: settings.NeuralNetworkSettings.DarknetCFG,
		WeightsFile:              settings.NeuralNetworkSettings.DarknetWeights,
		Threshold:                float32(settings.NeuralNetworkSettings.ConfThreshold),
	}
	if err := neuralNet.Init(); err != nil {
		log.Fatalln(err)
	}
	defer neuralNet.Close()

	allblobies = blob.NewBlobiesDefaults()

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

	processFrame(img)
	go performDetection()
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
				detectedObject := make([]image.Rectangle, len(detected))
				for i := range detected {
					detectedObject[i] = detected[i].Rect
				}
				allblobies.MatchToExisting(detectedObject)

				// for i := range detected {
				// 	if settings.MjpegSettings.ImshowEnable {
				// 		gocv.Rectangle(&img.ImgScaled, detected[i].Rect, color.RGBA{255, 255, 0, 0}, 2)
				// 	}
				// }
				// allblobies.MatchToExisting(detectedObject)
			}

		default:
			// show current frame without blocking, so do nothing here
		}

		if settings.MjpegSettings.ImshowEnable {
			for i := range settings.TrackerSettings.LinesSettings {
				settings.TrackerSettings.LinesSettings[i].VLine.Draw(&img.ImgScaled)
			}
			for i, b := range (*allblobies).Objects {
				(*b).DrawTrack(&img.ImgScaled, fmt.Sprintf("%v", i))
			}

			window.IMShow(img.ImgScaled)
			if window.WaitKey(1) == 27 {
				break
			}
		}
		if settings.MjpegSettings.Enable {
			buf, _ := gocv.IMEncode(".jpg", img.ImgScaled)
			stream.UpdateJPEG(buf)
		}

	}

	// Release memory
	img.Close()

	// pprof
	if settings.PPROFSettings.Enable {
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

func performDetection() {
	fmt.Println("Start performDetection thread")
	for {

		frame := <-imagesChannel
		defer frame.Close()

		darknetImage, err := darknet.Image2Float32(frame.ImgSTD)
		if err != nil {
			log.Println("ImageFromMemory error")
			// Error: no handling
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer darknetImage.Close()

		dr, err := neuralNet.Detect(darknetImage)
		if err != nil {
			fmt.Println(err)
			time.Sleep(100 * time.Millisecond)
		}

		detectedRects := make([]odam.DetectedObject, 0, len(dr.Detections))
		for _, d := range dr.Detections {
			for i := range d.ClassIDs {
				if d.ClassNames[i] != "car" && d.ClassNames[i] != "motorbike" && d.ClassNames[i] != "bus" && d.ClassNames[i] != "train" && d.ClassNames[i] != "truck" {
					continue
				}
				bBox := d.BoundingBox
				// minX, minY := float64(bBox.StartPoint.X)/0.33, float64(bBox.StartPoint.Y)/scaleHeight
				// maxX, maxY := float64(bBox.EndPoint.X)/0.33, float64(bBox.EndPoint.Y)/scaleHeight
				minX, minY := float64(bBox.StartPoint.X), float64(bBox.StartPoint.Y)
				maxX, maxY := float64(bBox.EndPoint.X), float64(bBox.EndPoint.Y)
				rect := odam.DetectedObject{
					Rect:       image.Rect(utils.Round(minX), utils.Round(minY), utils.Round(maxX), utils.Round(maxY)),
					Classname:  d.ClassNames[i],
					Confidence: d.Probabilities[i],
				}
				detectedRects = append(detectedRects, rect)
			}
		}

		detectedChannel <- detectedRects
	}
}

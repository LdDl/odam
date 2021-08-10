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
	blob "github.com/LdDl/gocv-blob/v2/blob"
	"github.com/LdDl/odam"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
)

var (
	settingsFile    = flag.String("settings", "conf.json", "Path to application's settings")
	window          *gocv.Window
	stream          *mjpeg.Stream
	imagesChannel   chan *odam.FrameData
	detectedChannel chan []*odam.DetectedObject
	detected        []*odam.DetectedObject
	allblobies      *blob.Blobies
)

func main() {

	/* Read settings */
	flag.Parse()
	settings, err := odam.NewSettings(*settingsFile)
	if err != nil {
		log.Println(err)
		return
	}

	/* Initialize MJPEG server if needed */
	if settings.MjpegSettings.Enable {
		stream = mjpeg.NewStream()
		go func() {
			fmt.Printf("Starting MJPEG on http://localhost:%d\n", settings.MjpegSettings.Port)
			http.Handle("/", stream)
			err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", settings.MjpegSettings.Port), nil)
			if err != nil {
				log.Fatalln(err)
			}
		}()
	}

	/* Initialize gRPC data forwarding if needed */
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

	/* Initialize neural network */
	neuralNet := darknet.YOLONetwork{
		GPUDeviceIndex:           0,
		NetworkConfigurationFile: settings.NeuralNetworkSettings.DarknetCFG,
		WeightsFile:              settings.NeuralNetworkSettings.DarknetWeights,
		Threshold:                float32(settings.NeuralNetworkSettings.ConfThreshold),
	}
	err = neuralNet.Init()
	if err != nil {
		log.Println(err)
		return
	}
	defer neuralNet.Close()

	/* Initialize objects tracker */
	allblobies = blob.NewBlobiesDefaults()
	trackerType := settings.TrackerSettings.GetTrackerType()
	fmt.Printf("Using tracker: '%s'\n", settings.TrackerSettings.TrackerType)

	/* Initialize GIS converter (for speed estimation) if needed*/
	// It just helps to figure out what does [Longitude; Latitude] pair correspond to certain pixel
	var gisConverter func(gocv.Point2f) gocv.Point2f
	if settings.TrackerSettings.SpeedEstimationSettings.Enabled {
		if len(settings.TrackerSettings.SpeedEstimationSettings.Mapper) != 4 {
			fmt.Println("[WARNING] 'mapper' field in 'speed_estimation_settings' should contain exactly 4 elements. Disabling speed estimation feature...")
			settings.TrackerSettings.SpeedEstimationSettings.Enabled = false
		} else {
			src := make([]gocv.Point2f, len(settings.TrackerSettings.SpeedEstimationSettings.Mapper))
			dst := make([]gocv.Point2f, len(settings.TrackerSettings.SpeedEstimationSettings.Mapper))
			for i := range settings.TrackerSettings.SpeedEstimationSettings.Mapper {
				ptImage := settings.TrackerSettings.SpeedEstimationSettings.Mapper[i].ImageCoordinates
				ptGIS := settings.TrackerSettings.SpeedEstimationSettings.Mapper[i].EPSG4326
				src[i] = gocv.Point2f{X: ptImage[0], Y: ptImage[1]}
				dst[i] = gocv.Point2f{X: ptGIS[0], Y: ptGIS[1]}
			}
			gisConverter = odam.GetPerspectiveTransformer(src, dst)
		}
	}

	/* Open video capturer */
	videoCapturer, err := gocv.OpenVideoCapture(settings.VideoSettings.Source)
	if err != nil {
		log.Println(err)
		return
	}
	/* Open imshow() GUI in needed */
	if settings.MjpegSettings.ImshowEnable {
		fmt.Println("Press 'ESC' to stop imshow()")
		window = gocv.NewWindow("ODAM v0.8.0")
		window.ResizeWindow(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
		defer window.Close()
	}

	/* Initialize channels */
	imagesChannel = make(chan *odam.FrameData, 1)
	detectedChannel = make(chan []*odam.DetectedObject)
	img := odam.NewFrameData()

	/* Read first frame */
	if ok := videoCapturer.Read(&img.ImgSource); !ok {
		log.Printf("Error cannot read video '%s'\n", settings.VideoSettings.Source)
		return
	}
	/* Scale first frame */
	err = img.Preprocess(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
	if err != nil {
		log.Println("Error on first preprocessing step", err)
		return
	}
	/* Process first frame */
	processFrame(img)

	/* Start goroutine for object detection purposes */
	go performDetection(&neuralNet, settings.NeuralNetworkSettings.TargetClasses)

	/* Initialize variables for evaluation of time difference between frames */
	lastMS := 0.0
	lastTime := time.Now()

	/* Start continuous frame reading */
	for {
		/* Read frame */
		if ok := videoCapturer.Read(&img.ImgSource); !ok {
			fmt.Println("Can't read next frame, stop grabbing...")
			break
		}
		/* Evaluate time difference */
		currentMS := videoCapturer.Get(gocv.VideoCapturePosMsec)
		msDiff := currentMS - lastMS
		secDiff := msDiff / 1000.0
		lastTime = lastTime.Add(time.Duration(msDiff) * time.Millisecond)
		lastMS = currentMS

		/* Skip empty frame */
		if img.ImgSource.Empty() {
			fmt.Println("Empty frame has been detected. Sleep for 400 ms")
			time.Sleep(400 * time.Millisecond)
			continue
		}

		/* Scale frame */
		err := img.Preprocess(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
		if err != nil {
			fmt.Printf("Can't preprocess. Error: %s. Sleep for 400ms\n", err.Error())
			time.Sleep(400 * time.Millisecond)
			continue
		}

		/* Read data from object detection goroutine */
		select {
		case detected = <-detectedChannel:
			processFrame(img)
			if len(detected) != 0 {
				/* Prepare 'blobs' for each detected object */
				detectedObjects := make([]blob.Blobie, len(detected))
				for i := range detected {
					commonOptions := blob.BlobOptions{
						ClassID:          detected[i].ClassID,
						ClassName:        detected[i].ClassName,
						MaxPointsInTrack: settings.TrackerSettings.MaxPointsInTrack,
						Time:             lastTime,
						TimeDeltaSeconds: secDiff,
					}
					if trackerType == odam.TRACKER_SIMPLE {
						detectedObjects[i] = blob.NewSimpleBlobie(detected[i].Rect, &commonOptions)
					} else if trackerType == odam.TRACKER_KALMAN {
						detectedObjects[i] = blob.NewKalmanBlobie(detected[i].Rect, &commonOptions)
					}
					if foundOptions, ok := settings.ClassesDrawOptions[detected[i].ClassName]; ok {
						detectedObjects[i].SetDraw(foundOptions.DrawOptions)
					}
				}
				/* Match blobs to existing ones */
				allblobies.MatchToExisting(detectedObjects)

				/* Estimate speed if needed */
				if settings.TrackerSettings.SpeedEstimationSettings.Enabled {
					for _, b := range allblobies.Objects {
						blobTrack := b.GetTrack()
						trackLen := len(blobTrack)
						if trackLen >= 2 {
							blobTimestamps := b.GetTimestamps()
							fp := odam.STDPointToGoCVPoint2F(blobTrack[0])
							lp := odam.STDPointToGoCVPoint2F(blobTrack[trackLen-1])
							spd := odam.EstimateSpeed(fp, lp, blobTimestamps[0], blobTimestamps[trackLen-1], gisConverter)
							b.SetProperty("speed", spd)
						}
					}
				}
				for _, vline := range settings.TrackerSettings.LinesSettings {
					for _, b := range allblobies.Objects {
						className := b.GetClassName()
						if stringInSlice(&className, vline.DetectClasses) { // Detect if object should be detected by virtual line (filter by classname)
							crossedLine := false
							if vline.VLine.LineType == odam.HORIZONTAL_LINE {
								crossedLine = b.IsCrossedTheLine(vline.VLine.RightPT.Y, vline.VLine.LeftPT.X, vline.VLine.RightPT.X, vline.VLine.Direction)
							} else if vline.VLine.LineType == odam.OBLIQUE_LINE {
								crossedLine = b.IsCrossedTheObliqueLine(vline.VLine.RightPT.X, vline.VLine.RightPT.Y, vline.VLine.LeftPT.X, vline.VLine.LeftPT.Y, vline.VLine.Direction)
							}
							// If object crossed the virtual line
							if crossedLine {
								catchedTimestamp := time.Now().UTC().Unix()
								b.SetTracking(false)
								// If gRPC streaming data is disabled why do we need to process all stuff? We add strict condition.
								if settings.GrpcSettings.Enable {
									blobRect := b.GetCurrentRect()
									minx, miny := math.Floor(float64(blobRect.Min.X)*settings.VideoSettings.ScaleX), math.Floor(float64(blobRect.Min.Y)*settings.VideoSettings.ScaleY)
									maxx, maxy := math.Floor(float64(blobRect.Max.X)*settings.VideoSettings.ScaleX), math.Floor(float64(blobRect.Max.Y)*settings.VideoSettings.ScaleY)
									cropRect := image.Rect(
										int(minx)+5,  // add a bit width to crop bigger region
										int(miny)+10, // add a bit height to crop bigger region
										int(maxx)+5,
										int(maxy)+10,
									)
									// Make sure to be not out of image bounds
									odam.FixRectForOpenCV(&cropRect, settings.VideoSettings.Width, settings.VideoSettings.Height)
									buf := new(bytes.Buffer)
									xtop, ytop := int32(cropRect.Min.X), int32(cropRect.Min.Y)

									// Futher buffer preparation depends on 'crop_mode' in JSON'ed configuration file
									if vline.VLine.CropObject {
										cropImage := img.ImgSource.Region(cropRect)
										copyCrop := cropImage.Clone()
										cropImageSTD, err := copyCrop.ToImage()
										if err != nil {
											fmt.Println("[WARNING] Can't convert cropped gocv.Mat to image.Image:", err)
											cropImage.Close()
											copyCrop.Close()
											continue
										}
										cropImage.Close()
										copyCrop.Close()
										err = jpeg.Encode(buf, cropImageSTD, nil)
										if err != nil {
											fmt.Println("[WARNING] Can't call jpeg.Encode() on cropped gocv.Mat to image.Image:", err)
										}
										xtop, ytop = 0, 0
									} else {
										copyImage := img.ImgSource.Clone()
										copyImageSTD, err := copyImage.ToImage()
										if err != nil {
											fmt.Println("[WARNING] Can't convert source gocv.Mat to image.Image:", err)
											copyImage.Close()
											continue
										}
										err = jpeg.Encode(buf, copyImageSTD, nil)
										if err != nil {
											fmt.Println("[WARNING] Can't call jpeg.Encode() on source gocv.Mat to image.Image:", err)
										}
									}
									bytesBuffer := buf.Bytes()
									sendData := odam.ObjectInformation{
										CamId:     settings.VideoSettings.CameraID,
										Timestamp: catchedTimestamp,
										Image:     bytesBuffer,
										Detection: &odam.Detection{
											XLeft:  xtop,
											YTop:   ytop,
											Width:  int32(cropRect.Dx()),
											Height: int32(cropRect.Dy()),
										},
										Class: &odam.ClassInfo{
											ClassId:   int32(b.GetClassID()),
											ClassName: className,
										},
										VirtualLine: &odam.VirtualLineInfo{
											Id:     vline.LineID,
											LeftX:  int32(vline.VLine.SourceLeftPT.X),
											LeftY:  int32(vline.VLine.SourceLeftPT.Y),
											RightX: int32(vline.VLine.SourceRightPT.X),
											RightY: int32(vline.VLine.SourceRightPT.Y),
										},
									}
									// If it is needed to send speed and track information
									if settings.TrackerSettings.SpeedEstimationSettings.SendGRPC {
										// Extract estimated speed information
										spd := float32(0.0)
										if spdInterface, ok := b.GetProperty("speed"); ok {
											switch spdInterface.(type) { // Want to be sure that interface is float32
											case float32:
												spd = spdInterface.(float32)
												break
											default:
												break
											}
										}
										trackPixels := b.GetTrack()
										trackUnionInfo := make([]*odam.Point, len(trackPixels))
										for i, stdPt := range trackPixels {
											cvPt := odam.STDPointToGoCVPoint2F(stdPt)
											gisPt := gisConverter(cvPt)
											trackUnionInfo[i] = &odam.Point{
												EuclideanPoint: &odam.EuclideanPoint{X: cvPt.X * float32(settings.VideoSettings.ScaleX), Y: cvPt.Y * float32(settings.VideoSettings.ScaleY)},
												Wgs84Point:     &odam.WGS84Point{Longitude: gisPt.X, Latitude: gisPt.Y},
											}
										}
										sendData.TrackInformation = &odam.TrackInfo{
											EstimatedSpeed: spd,
											Points:         trackUnionInfo,
										}
									}
									go sendDataToServer(grpcConn, &sendData)
								}
							}
						}
					}
				}
			}
			break
		default:
			// show current frame without blocking, so do nothing here
			break
		}

		/* Draw info about detected objects when either MJPEG or imshow() GUI is enabled */
		if settings.MjpegSettings.ImshowEnable || settings.MjpegSettings.Enable {
			for i := range settings.TrackerSettings.LinesSettings {
				settings.TrackerSettings.LinesSettings[i].VLine.Draw(&img.ImgScaled)
			}
			for _, b := range (*allblobies).Objects {
				spd := float32(0.0)
				if spdInterface, ok := b.GetProperty("speed"); ok {
					switch spdInterface.(type) { // Want to be sure that interface is float32
					case float32:
						spd = spdInterface.(float32)
						break
					default:
						break
					}
				}
				if foundOptions, ok := settings.ClassesDrawOptions[b.GetClassName()]; ok {
					if foundOptions.DisplayObjectID {
						b.DrawTrack(&img.ImgScaled, fmt.Sprintf("v = %.2f km/h", spd), fmt.Sprintf("%v", b.GetID()))
					} else {
						b.DrawTrack(&img.ImgScaled, fmt.Sprintf("v = %.2f km/h", spd))
					}
				}
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
				log.Printf("Error while decoding to JPG (mjpeg): %s", err.Error())
			} else {
				stream.UpdateJPEG(buf)
			}
		}
	}

	fmt.Println("Shutting down...")
	time.Sleep(2 * time.Second) // @todo temporary fix: need to wait a bit time for last call of neuralNet.Detect(...)

	// Hard release memory
	img.Close()
	neuralNet.Close()

	// pprof (for debuggin purposes)
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
			log.Printf("Can't convert image to Darknet's format due the error: %s. Sleep for 100ms", err.Error())
			frame.Close()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		dr, err := neuralNet.Detect(darknetImage)
		if err != nil {
			frame.Close()
			darknetImage.Close()
			log.Printf("Can't make detection: %s. Sleep for 100ms", err.Error())
			time.Sleep(100 * time.Millisecond)
			continue
		}
		darknetImage.Close() // free the memory
		darknetImage = nil
		detectedRects := make([]*odam.DetectedObject, 0, len(dr.Detections))
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
					detectedRects = append(detectedRects, &rect)
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

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"log"
	"math"
	"time"

	"github.com/LdDl/odam"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
)

func main() {
	settingsFile := flag.String("settings", "conf.json", "Path to application's settings")
	/* Read settings */
	flag.Parse()
	settings, err := odam.NewSettings(*settingsFile)
	if err != nil {
		log.Println(err)
		return
	}

	/* Initialize application */
	app, err := odam.NewApp(settings)
	if err != nil {
		log.Println(err)
		return
	}
	defer app.Close()

	/* Initialize MJPEG server if needed */
	var stream *mjpeg.Stream
	if settings.MjpegSettings.Enable {
		stream = app.StartMJPEGStream()
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

	/* Initialize objects tracker */
	allblobies := app.GetBlobsStorage()
	fmt.Printf("Using tracker: '%s'\n", settings.TrackerSettings.TrackerType)

	/* Initialize GIS converter (for speed estimation) if needed*/
	// It just helps to figure out what does [Longitude; Latitude] pair correspond to certain pixel
	var gisConverter func(gocv.Point2f) gocv.Point2f
	if settings.TrackerSettings.SpeedEstimationSettings.Enabled {
		gisConverter = app.GetGISConverter()
	}
	/* Open video capturer */
	videoCapturer, err := gocv.OpenVideoCapture(settings.VideoSettings.Source)
	if err != nil {
		log.Println(err)
		return
	}
	/* Open imshow() GUI in needed */
	var window *gocv.Window
	if settings.MjpegSettings.ImshowEnable {
		fmt.Println("Press 'ESC' to stop imshow()")
		window = gocv.NewWindow("ODAM v0.8.0")
		window.ResizeWindow(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
		defer window.Close()
	}

	/* Read first frame */
	img := odam.NewFrameData()
	if ok := videoCapturer.Read(&img.ImgSource); !ok {
		log.Printf("Error cannot read video '%s'\n", settings.VideoSettings.Source)
		return
	}

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

		detected := performDetectionSequential(app, img, settings.NeuralNetworkSettings.NetClasses, settings.NeuralNetworkSettings.TargetClasses)
		if len(detected) != 0 {
			/* Prepare 'blob' for each detected object */
			detectedObjects := app.PrepareBlobs(detected, lastTime, secDiff)
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
						crossedLine := vline.VLine.IsBlobCrossedLine(b)
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
								var buf *bytes.Buffer
								xtop, ytop := int32(cropRect.Min.X), int32(cropRect.Min.Y)

								// Futher buffer preparation depends on 'crop_mode' in JSON'ed configuration file
								if vline.VLine.CropObject {
									buf, err = odam.PrepareCroppedImageBuffer(&img.ImgSource, cropRect)
									if err != nil {
										fmt.Println("[WARNING] Can't prepare image buffer (with crop) due ther error:", err)
									}
									xtop, ytop = 0, 0
								} else {
									buf, err = odam.PrepareImageBuffer(&img.ImgSource)
									if err != nil {
										fmt.Println("[WARNING] Can't prepare image buffer due ther error:", err)
									}
								}
								bytesBuffer := buf.Bytes()
								sendData := odam.ObjectInformation{
									CamId:       settings.VideoSettings.CameraID,
									Timestamp:   catchedTimestamp,
									Image:       bytesBuffer,
									Detection:   odam.DetectionInfoGRPC(xtop, ytop, int32(cropRect.Dx()), int32(cropRect.Dy())),
									Class:       odam.ClassInfoGRPC(b),
									VirtualLine: odam.VirtualLineInfoGRPC(vline.LineID, vline.VLine),
								}
								// If it is needed to send speed and track information
								if settings.TrackerSettings.SpeedEstimationSettings.SendGRPC {
									sendData.TrackInformation = odam.TrackInfoInfoGRPC(b, "speed", float32(settings.VideoSettings.ScaleX), float32(settings.VideoSettings.ScaleY), gisConverter)
								}
								go sendDataToServer(grpcConn, &sendData)
							}
						}
					}
				}
			}
			for _, vpolygon := range settings.TrackerSettings.PolygonsSettings {
				for _, b := range allblobies.Objects {
					className := b.GetClassName()
					if stringInSlice(&className, vpolygon.DetectClasses) { // Detect if object should be detected by virtual polygon (filter by classname)
						// if vpolygon.VPolygon.ContainsBlob(b) {
						// 	fmt.Printf("Polygon %d contains blob %s\n", vpolygon.PolygonID, b.GetID())
						// }
						// enteredPolygon := vpolygon.VPolygon.BlobEntered(b)
						// if enteredPolygon {
						// 	fmt.Println("entered blob", b.GetID())
						// }
						// leftPolygon := vpolygon.VPolygon.BlobLeft(b)
						// if leftPolygon {
						// 	fmt.Println("left blob", b.GetID())
						// }
					}
				}
			}
		}
		/* Draw info about detected objects when either MJPEG or imshow() GUI is enabled */
		if settings.MjpegSettings.ImshowEnable || settings.MjpegSettings.Enable {
			for i := range settings.TrackerSettings.LinesSettings {
				settings.TrackerSettings.LinesSettings[i].VLine.Draw(&img.ImgScaled)
			}
			for i := range settings.TrackerSettings.PolygonsSettings {
				settings.TrackerSettings.PolygonsSettings[i].VPolygon.Draw(&img.ImgScaled)
			}

			for _, b := range allblobies.Objects {
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
				if foundOptions := settings.GetDrawOptions(b.GetClassName()); foundOptions != nil {
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
				stream.UpdateJPEG(buf.GetBytes())
			}
		}
		/* temporary */
		// time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Shutting down...")
	time.Sleep(2 * time.Second) // @todo temporary fix: need to wait a bit time for last call of neuralNet.Detect(...)

	// Hard release memory
	img.Close()
	app.Close()

	// pprof (for debuggin purposes)
	if settings.MatPPROFSettings.Enable {
		var b bytes.Buffer
		// go run -tags matprofile main.go
		// gocv.MatProfile.WriteTo(&b, 1)
		fmt.Print(b.String())
	}
}

func processFrameSequential(fd *odam.FrameData) *odam.FrameData {
	frame := odam.NewFrameData()
	fd.ImgSource.CopyTo(&frame.ImgSource)
	fd.ImgScaled.CopyTo(&frame.ImgScaled)
	fd.ImgScaledCopy.CopyTo(&frame.ImgScaledCopy)
	return frame
}

func performDetectionSequential(app *odam.Application, frame *odam.FrameData, netClasses, targetClasses []string) []*odam.DetectedObject {
	detectedRects, err := odam.DetectObjects(app, frame.ImgScaledCopy, netClasses, targetClasses...)
	if err != nil {
		log.Printf("Can't detect objects on provided image due the error: %s. Sleep for 100ms", err.Error())
		frame.ImgScaledCopy.Close()
		time.Sleep(100 * time.Millisecond)
	}
	frame.ImgScaledCopy.Close() // free the memory
	return detectedRects
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

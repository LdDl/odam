package odam

import (
	"bytes"
	context "context"
	"fmt"
	"image"
	"log"
	"math"
	"net/http"
	"time"

	blob "github.com/LdDl/gocv-blob/v2/blob"
	"github.com/hybridgroup/mjpeg"
	"github.com/pkg/errors"
	"gocv.io/x/gocv"
	grpc "google.golang.org/grpc"
)

// Application Main engine
type Application struct {
	neuralNetwork  *gocv.Net
	layersNames    []string
	blobiesStorage *blob.Blobies
	trackerType    TRACKER_TYPE
	gisConverter   *SpatialConverter

	settings   *AppSettings
	grpcConn   *grpc.ClientConn
	grpcClient ServiceYOLOClient
}

// NewApp Constructor for Application
//
// settings - pointer to AppSettings object
//
func NewApp(settings *AppSettings) (*Application, error) {
	/* Initialize neural network */
	neuralNet := gocv.ReadNet(settings.NeuralNetworkSettings.DarknetWeights, settings.NeuralNetworkSettings.DarknetCFG)
	yoloLayersIdx := neuralNet.GetUnconnectedOutLayers()
	outLayerNames := make([]string, 0, 3)
	for _, idx := range yoloLayersIdx {
		layer := neuralNet.GetLayer(idx)
		outLayerNames = append(outLayerNames, layer.GetName())
	}
	err := neuralNet.SetPreferableBackend(gocv.NetBackendCUDA)
	if err != nil {
		return nil, errors.Wrap(err, "Can't set backend CUDA")
	}
	err = neuralNet.SetPreferableTarget(gocv.NetTargetCUDA)
	if err != nil {
		return nil, errors.Wrap(err, "Can't set target CUDA")
	}
	/* Initialize GIS converter (for speed estimation) if needed*/
	// It just helps to figure out what does [Longitude; Latitude] pair correspond to certain pixel
	spatialConverter := SpatialConverter{}
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
			spatialConverter.transformMat, spatialConverter.Function = GetPerspectiveTransformer(src, dst)
		}
	}
	return &Application{
		neuralNetwork:  &neuralNet,
		layersNames:    outLayerNames,
		blobiesStorage: blob.NewBlobiesDefaults(),
		trackerType:    settings.TrackerSettings.GetTrackerType(),
		gisConverter:   &spatialConverter,
		settings:       settings,
	}, nil
}

// Close Free memory for underlying objects
func (app *Application) Close() {
	app.neuralNetwork.Close()
	app.gisConverter.Close()
	app.grpcConn.Close()
}

// GetBlobsStorage Returns pointer to blob.Blobies
func (app *Application) GetBlobsStorage() *blob.Blobies {
	return app.blobiesStorage
}

// GetGISConverter Returns anonymus function for spatial conversion
func (app *Application) GetGISConverter() func(gocv.Point2f) gocv.Point2f {
	return app.gisConverter.Function
}

// StartMJPEGStream Start MJPEG video stream in separate goroutine
func (app *Application) StartMJPEGStream() *mjpeg.Stream {
	stream := mjpeg.NewStream()
	go func() {
		fmt.Printf("Starting MJPEG on http://localhost:%d\n", app.settings.MjpegSettings.Port)
		http.Handle("/", stream)
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", app.settings.MjpegSettings.Port), nil)
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return stream
}

// PrepareBlobs Convert DetectedObjects to slice of blob.Blobie
func (app *Application) PrepareBlobs(detected DetectedObjects, lastTm time.Time, secDiff float64) []blob.Blobie {
	detectedObjects := make([]blob.Blobie, len(detected))
	for i := range detected {
		commonOptions := blob.BlobOptions{
			ClassID:          detected[i].ClassID,
			ClassName:        detected[i].ClassName,
			MaxPointsInTrack: app.settings.TrackerSettings.MaxPointsInTrack,
			Time:             lastTm,
			TimeDeltaSeconds: secDiff,
		}
		if app.trackerType == TRACKER_SIMPLE {
			detectedObjects[i] = blob.NewSimpleBlobie(detected[i].Rect, &commonOptions)
		} else if app.trackerType == TRACKER_KALMAN {
			detectedObjects[i] = blob.NewKalmanBlobie(detected[i].Rect, &commonOptions)
		}
		if foundOptions := app.settings.GetDrawOptions(detected[i].ClassName); foundOptions != nil {
			detectedObjects[i].SetDraw(foundOptions.DrawOptions)
		}
	}
	return detectedObjects
}

func (app *Application) Run() error {
	settings := app.settings

	/* Open imshow() GUI in needed */
	var window *gocv.Window
	if settings.MjpegSettings.ImshowEnable {
		fmt.Println("Press 'ESC' to stop imshow()")
		window = gocv.NewWindow("ODAM v0.9.0")
		window.ResizeWindow(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
		defer window.Close()
	}

	/* Open video capturer */
	videoCapturer, err := gocv.OpenVideoCapture(settings.VideoSettings.Source)
	if err != nil {
		return errors.Wrap(err, "Can't open video capture")
	}

	/* Prepare frame */
	img := NewFrameData()
	/* Initialize variables for evaluation of time difference between frames */
	lastMS := 0.0
	lastTime := time.Now()

	/* Initialize objects tracker */
	allblobies := app.GetBlobsStorage()
	fmt.Printf("Using tracker: '%s'\n", settings.TrackerSettings.TrackerType)

	/* Initialize GIS converter (for speed estimation) if needed*/
	// It just helps to figure out what does [Longitude; Latitude] pair correspond to certain pixel
	var gisConverter func(gocv.Point2f) gocv.Point2f
	if settings.TrackerSettings.SpeedEstimationSettings.Enabled {
		gisConverter = app.GetGISConverter()
	}

	/* Initialize MJPEG server if needed */
	var stream *mjpeg.Stream
	if settings.MjpegSettings.Enable {
		stream = app.StartMJPEGStream()
	}

	/* Initialize gRPC data forwarding if needed */
	if settings.GrpcSettings.Enable {
		url := fmt.Sprintf("%s:%d", settings.GrpcSettings.ServerIP, settings.GrpcSettings.ServerPort)
		app.grpcConn, err = grpc.Dial(url, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return errors.Wrap(err, "Can't init grpc connection")
		}
		defer app.grpcConn.Close()
		app.grpcClient = NewServiceYOLOClient(app.grpcConn)
	}

	/* Read frames in a */
	for {
		// Grab a frame
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

		detected := app.performDetectionSequential(img, settings.NeuralNetworkSettings.NetClasses, settings.NeuralNetworkSettings.TargetClasses)
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
						fp := STDPointToGoCVPoint2F(blobTrack[0])
						lp := STDPointToGoCVPoint2F(blobTrack[trackLen-1])
						spd := EstimateSpeed(fp, lp, blobTimestamps[0], blobTimestamps[trackLen-1], gisConverter)
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
							b.SetTracking(false)
							// If gRPC streaming data is disabled why do we need to process all stuff? We add strict condition.
							if settings.GrpcSettings.Enable {
								catchedTimestamp := time.Now().UTC().Unix()
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
								FixRectForOpenCV(&cropRect, settings.VideoSettings.Width, settings.VideoSettings.Height)
								var buf *bytes.Buffer
								xtop, ytop := int32(cropRect.Min.X), int32(cropRect.Min.Y)

								// Futher buffer preparation depends on 'crop_mode' in JSON'ed configuration file
								if vline.VLine.CropObject {
									buf, err = PrepareCroppedImageBuffer(&img.ImgSource, cropRect)
									if err != nil {
										fmt.Println("[WARNING] Can't prepare image buffer (with crop) due ther error:", err)
									}
									xtop, ytop = 0, 0
								} else {
									buf, err = PrepareImageBuffer(&img.ImgSource)
									if err != nil {
										fmt.Println("[WARNING] Can't prepare image buffer due ther error:", err)
									}
								}
								bytesBuffer := buf.Bytes()
								sendData := ObjectInformation{
									CamId:       settings.VideoSettings.CameraID,
									Timestamp:   catchedTimestamp,
									Image:       bytesBuffer,
									Detection:   DetectionInfoGRPC(xtop, ytop, int32(cropRect.Dx()), int32(cropRect.Dy())),
									Class:       ClassInfoGRPC(b),
									VirtualLine: VirtualLineInfoGRPC(vline.LineID, vline.VLine),
								}
								// If it is needed to send speed and track information
								if settings.TrackerSettings.SpeedEstimationSettings.SendGRPC {
									sendData.TrackInformation = TrackInfoInfoGRPC(b, "speed", float32(settings.VideoSettings.ScaleX), float32(settings.VideoSettings.ScaleY), gisConverter)
								}
								go sendDataToServer(app.grpcClient, &sendData)
							}
						}
					}
				}
			}
			// for _, vpolygon := range settings.TrackerSettings.PolygonsSettings {
			// 	for _, b := range allblobies.Objects {
			// 		className := b.GetClassName()
			// 		if stringInSlice(&className, vpolygon.DetectClasses) { // Detect if object should be detected by virtual polygon (filter by classname)
			// 			// if vpolygon.VPolygon.ContainsBlob(b) {
			// 			// 	fmt.Printf("Polygon %d contains blob %s\n", vpolygon.PolygonID, b.GetID())
			// 			// }
			// 			// enteredPolygon := vpolygon.VPolygon.BlobEntered(b)
			// 			// if enteredPolygon {
			// 			// 	fmt.Println("entered blob", b.GetID())
			// 			// }
			// 			// leftPolygon := vpolygon.VPolygon.BlobLeft(b)
			// 			// if leftPolygon {
			// 			// 	fmt.Println("left blob", b.GetID())
			// 			// }
			// 		}
			// 	}
			// }
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
	}
	// Hard release memory
	img.Close()
	app.Close()

	return nil
}

func (app *Application) performDetectionSequential(frame *FrameData, netClasses, targetClasses []string) []*DetectedObject {
	detectedRects, err := DetectObjects(app, frame.ImgScaledCopy, netClasses, targetClasses...)
	if err != nil {
		log.Printf("Can't detect objects on provided image due the error: %s. Sleep for 100ms", err.Error())
		frame.ImgScaledCopy.Close()
		time.Sleep(100 * time.Millisecond)
	}
	frame.ImgScaledCopy.Close() // free the memory
	return detectedRects
}

func sendDataToServer(grpcClient ServiceYOLOClient, data *ObjectInformation) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := grpcClient.SendDetection(
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

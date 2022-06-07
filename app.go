package odam

import (
	"fmt"
	"log"
	"net/http"
	"time"

	blob "github.com/LdDl/gocv-blob/v2/blob"
	"github.com/hybridgroup/mjpeg"
	"github.com/pkg/errors"
	"gocv.io/x/gocv"
)

// Application Main engine
type Application struct {
	neuralNetwork  *gocv.Net
	layersNames    []string
	blobiesStorage *blob.Blobies
	trackerType    TRACKER_TYPE
	gisConverter   *SpatialConverter

	settings *AppSettings
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
	_ = allblobies
	fmt.Printf("Using tracker: '%s'\n", settings.TrackerSettings.TrackerType)

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
		_ = detected
		// @todo: long copy-paste stuff from cmd/odam/main.go

	}
	// Hard release memory
	img.Close()

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

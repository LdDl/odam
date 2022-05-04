package odam

import (
	"fmt"
	"log"
	"net/http"
	"time"

	darknet "github.com/LdDl/go-darknet"
	blob "github.com/LdDl/gocv-blob/v2/blob"
	"github.com/hybridgroup/mjpeg"
	"github.com/pkg/errors"
	"gocv.io/x/gocv"
)

// Application Main engine
type Application struct {
	neuralNetwork  *darknet.YOLONetwork
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
	neuralNet := darknet.YOLONetwork{
		GPUDeviceIndex:           0,
		NetworkConfigurationFile: settings.NeuralNetworkSettings.DarknetCFG,
		WeightsFile:              settings.NeuralNetworkSettings.DarknetWeights,
		Threshold:                float32(settings.NeuralNetworkSettings.ConfThreshold),
	}
	err := neuralNet.Init()
	if err != nil {
		return nil, errors.Wrap(err, "Can't initialize neural network")
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
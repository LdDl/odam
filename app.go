package odam

import (
	"fmt"
	"log"
	"net/http"

	darknet "github.com/LdDl/go-darknet"
	blob "github.com/LdDl/gocv-blob/v2/blob"
	"github.com/hybridgroup/mjpeg"
	"github.com/pkg/errors"
)

type Application struct {
	neuralNetwork  *darknet.YOLONetwork
	blobiesStorage *blob.Blobies

	settings *AppSettings
}

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
	return &Application{
		neuralNetwork:  &neuralNet,
		blobiesStorage: blob.NewBlobiesDefaults(),
		settings:       settings,
	}, nil
}

func (app *Application) Close() {
	app.neuralNetwork.Close()
}

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

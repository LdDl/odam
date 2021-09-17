package odam

import (
	darknet "github.com/LdDl/go-darknet"
	"github.com/pkg/errors"
)

type Application struct {
	neuralNetwork *darknet.YOLONetwork
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
		neuralNetwork: &neuralNet,
	}, nil
}

func (app *Application) Close() {
	app.neuralNetwork.Close()
}

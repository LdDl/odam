package odam

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"os"
)

// NewSettings Create new AppSettings from content of configuration file
func NewSettings(fname string) (*AppSettings, error) {
	jsonFile, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	bytesValues, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	appsettings := AppSettings{}
	err = json.Unmarshal(bytesValues, &appsettings)
	if err != nil {
		return nil, err
	}

	if appsettings.VideoSettings.Width <= 0 {
		appsettings.VideoSettings.Width = 640
		fmt.Println("Field 'width' in 'video_settings' has not been provided (or <=0). Using default 640 width")
	}

	if appsettings.VideoSettings.Height <= 0 {
		appsettings.VideoSettings.Height = 360
		fmt.Println("Field 'height' in 'video_settings' has not been provided (or <=0). Using default 360 height")
	}

	if appsettings.VideoSettings.ReducedWidth <= 0 {
		appsettings.VideoSettings.ReducedWidth = appsettings.VideoSettings.Width
		fmt.Println("Field 'reduced_width' in 'video_settings' has not been provided (or <=0). Using default reduced_width = width")
	}
	if appsettings.VideoSettings.ReducedHeight <= 0 {
		appsettings.VideoSettings.ReducedHeight = appsettings.VideoSettings.Height
		fmt.Println("Field 'reduced_height' in 'video_settings' has not been provided (or <=0). Using default reduced_height = height")
	}

	if appsettings.VideoSettings.ReducedWidth > appsettings.VideoSettings.Width {
		appsettings.VideoSettings.ReducedWidth = appsettings.VideoSettings.Width
		fmt.Println("Field 'reduced_width' in 'video_settings' > 'width'. Using default reduced_width = width")
	}
	if appsettings.VideoSettings.ReducedHeight > appsettings.VideoSettings.Height {
		appsettings.VideoSettings.ReducedHeight = appsettings.VideoSettings.Height
		fmt.Println("Field 'reduced_height' in 'video_settings' > 'height'. Using default reduced_height = height")
	}

	appsettings.VideoSettings.ScaleX = float64(appsettings.VideoSettings.Width) / float64(appsettings.VideoSettings.ReducedWidth)
	appsettings.VideoSettings.ScaleY = float64(appsettings.VideoSettings.Height) / float64(appsettings.VideoSettings.ReducedHeight)

	if len(appsettings.TrackerSettings.LinesSettings) == 0 {
		fmt.Println("No 'lines_settings'? Please check it")
	}
	for i := range appsettings.TrackerSettings.LinesSettings {
		lsettings := &appsettings.TrackerSettings.LinesSettings[i]
		x1 := math.Round(float64(lsettings.Begin[0]) / appsettings.VideoSettings.ScaleX)
		y1 := math.Round(float64(lsettings.Begin[1]) / appsettings.VideoSettings.ScaleY)
		x2 := math.Round(float64(lsettings.End[0]) / appsettings.VideoSettings.ScaleX)
		y2 := math.Round(float64(lsettings.End[1]) / appsettings.VideoSettings.ScaleY)
		vline := VirtualLine{
			LeftPT:    image.Point{X: int(x1), Y: int(y1)},
			RightPT:   image.Point{X: int(x2), Y: int(y2)},
			Direction: true,
			Color:     color.RGBA{lsettings.RGBA[0], lsettings.RGBA[1], lsettings.RGBA[2], lsettings.RGBA[3]},
		}
		if lsettings.Direction == "from_detector" {
			vline.Direction = false
		}
		lsettings.VLine = &vline
	}
	return &appsettings, nil
}

// AppSettings Settings for application
type AppSettings struct {
	VideoSettings         VideoSettings         `json:"video_settings"`
	NeuralNetworkSettings NeuralNetworkSettings `json:"neural_network_settings"`
	CudaSettings          CudaSettings          `json:"cuda_settings"`
	MjpegSettings         MjpegSettings         `json:"mjpeg_settings"`
	GrpcSettings          GrpcSettings          `json:"grpc_settings"`
	TrackerSettings       TrackerSettings       `json:"tracker_settings"`
	PPROFSettings         PPROFSettings         `json:"matpprof_settings"`
}

// CudaSettings CUDA settings
type CudaSettings struct {
	Enable bool `json:"enable"`
}

// PPROFSettings pprof settings
type PPROFSettings struct {
	Enable bool `json:"enable"`
}

// GrpcSettings gRPC-server address
type GrpcSettings struct {
	Enable     bool   `json:"enable"`
	ServerIP   string `json:"server_ip"`
	ServerPort int    `json:"server_port"`
}

// MjpegSettings settings for output
type MjpegSettings struct {
	ImshowEnable bool `json:"imshow_enable"`
	Enable       bool `json:"enable"`
	Port         int  `json:"port"`
}

// NeuralNetworkSettings Neural network
type NeuralNetworkSettings struct {
	DarknetCFG     string `json:"darknet_cfg"`
	DarknetWeights string `json:"darknet_weights"`
	// DarknetClasses string   `json:"darknet_classes"`
	ConfThreshold float64  `json:"conf_threshold"`
	NmsThreshold  float64  `json:"nms_threshold"`
	TargetClasses []string `json:"target_classes"`
}

// TrackerSettings Object tracker settings
type TrackerSettings struct {
	LinesSettings []LinesSetting `json:"lines_settings"`
}

// LinesSetting Virtual lines
type LinesSetting struct {
	LineID        int      `json:"line_id"`
	Begin         [2]int   `json:"begin"`
	End           [2]int   `json:"end"`
	Direction     string   `json:"direction"`
	DetectClasses []string `json:"detect_classes"`
	RGBA          [4]uint8 `json:"rgba"`
	// Exported, but not from JSON
	VLine *VirtualLine `json:"-"`
}

// VideoSettings Settings for video
type VideoSettings struct {
	Source        string `json:"source"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	ReducedWidth  int    `json:"reduced_width"`
	ReducedHeight int    `json:"reduced_height"`
	CameraID      string `json:"camera_id"`

	// Exported, but not from JSON
	ScaleX float64 `json:"-"`
	ScaleY float64 `json:"-"`
}

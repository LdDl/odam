package odam

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"os"
	"strings"
	"sync"
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
		fmt.Println("[WARNING] Field 'width' in 'video_settings' has not been provided (or <=0). Using default 640 width")
	}

	if appsettings.VideoSettings.Height <= 0 {
		appsettings.VideoSettings.Height = 360
		fmt.Println("[WARNING] Field 'height' in 'video_settings' has not been provided (or <=0). Using default 360 height")
	}

	if appsettings.VideoSettings.ReducedWidth <= 0 {
		appsettings.VideoSettings.ReducedWidth = appsettings.VideoSettings.Width
		fmt.Println("[WARNING] Field 'reduced_width' in 'video_settings' has not been provided (or <=0). Using default reduced_width = width")
	}
	if appsettings.VideoSettings.ReducedHeight <= 0 {
		appsettings.VideoSettings.ReducedHeight = appsettings.VideoSettings.Height
		fmt.Println("[WARNING] Field 'reduced_height' in 'video_settings' has not been provided (or <=0). Using default reduced_height = height")
	}

	if appsettings.VideoSettings.ReducedWidth > appsettings.VideoSettings.Width {
		appsettings.VideoSettings.ReducedWidth = appsettings.VideoSettings.Width
		fmt.Println("[WARNING] Field 'reduced_width' in 'video_settings' > 'width'. Using default reduced_width = width")
	}
	if appsettings.VideoSettings.ReducedHeight > appsettings.VideoSettings.Height {
		appsettings.VideoSettings.ReducedHeight = appsettings.VideoSettings.Height
		fmt.Println("[WARNING] Field 'reduced_height' in 'video_settings' > 'height'. Using default reduced_height = height")
	}

	appsettings.VideoSettings.ScaleX = float64(appsettings.VideoSettings.Width) / float64(appsettings.VideoSettings.ReducedWidth)
	appsettings.VideoSettings.ScaleY = float64(appsettings.VideoSettings.Height) / float64(appsettings.VideoSettings.ReducedHeight)

	switch strings.ToLower(appsettings.TrackerSettings.TrackerType) {
	case "simple":
		appsettings.TrackerSettings.TrackerType = strings.ToLower(appsettings.TrackerSettings.TrackerType)
		appsettings.TrackerSettings.trackerType = TRACKER_SIMPLE
		break
	case "kalman":
		appsettings.TrackerSettings.TrackerType = strings.ToLower(appsettings.TrackerSettings.TrackerType)
		appsettings.TrackerSettings.trackerType = TRACKER_KALMAN
		break
	case "":
		fmt.Println("[WARNING]: Field 'tracker_type' is empty. Settings default value 'simple'")
		appsettings.TrackerSettings.TrackerType = "simple"
		appsettings.TrackerSettings.trackerType = TRACKER_SIMPLE
	default:
		fmt.Printf("[WARNING]: Value '%s' 'tracker_type' is not supported. Settings default value 'simple'\n", appsettings.TrackerSettings.TrackerType)
		appsettings.TrackerSettings.TrackerType = "simple"
		appsettings.TrackerSettings.trackerType = TRACKER_SIMPLE
		break
	}
	if len(appsettings.TrackerSettings.LinesSettings) == 0 {
		fmt.Println("[WARNING] No 'lines_settings'? Please check it")
	}
	for i := range appsettings.TrackerSettings.LinesSettings {
		lsettings := &appsettings.TrackerSettings.LinesSettings[i]
		vline := NewVirtualLine(lsettings.Begin[0], lsettings.Begin[1], lsettings.End[0], lsettings.End[1])
		vline.Scale(appsettings.VideoSettings.ScaleX, appsettings.VideoSettings.ScaleY)
		vline.Color = color.RGBA{lsettings.RGBA[0], lsettings.RGBA[1], lsettings.RGBA[2], lsettings.RGBA[3]}
		if lsettings.Direction == "from_detector" {
			vline.Direction = false
		}
		switch lsettings.CropMode {
		case "crop":
			vline.CropObject = true
			break
		case "no_crop":
			vline.CropObject = false
			break
		default:
			fmt.Printf("[WARNING] Field 'crop_mode' for line (id = '%d') can't be '%s'. Setting default value = 'crop'\n", lsettings.LineID, lsettings.CropMode)
			vline.CropObject = true
			break
		}
		lsettings.VLine = vline
	}

	if appsettings.TrackerSettings.MaxPointsInTrack < 1 {
		fmt.Printf("[WARNING] Field 'max_points_in_track' shoudle be >= 1, but got '%d'. Setting default value = 10\n", appsettings.TrackerSettings.MaxPointsInTrack)
		appsettings.TrackerSettings.MaxPointsInTrack = 10
	}

	// Prepare drawing options for each class defined in 'neural_network_settings'
	appsettings.ClassesDrawOptions = make(map[string]*DrawOptions)
	for _, class := range appsettings.NeuralNetworkSettings.TargetClasses {
		appsettings.ClassesDrawOptions[class] = nil
	}
	for _, classInfo := range appsettings.ClassesSettings {
		if _, ok := appsettings.ClassesDrawOptions[classInfo.ClassName]; !ok {
			// Class is not found in 'neural_network_settings'
			continue
		}
		appsettings.ClassesDrawOptions[classInfo.ClassName] = classInfo.PrepareDrawingOptions()
	}
	// Check if some of target classes haven't been described in classes settings
	// If there are some then prepare default settings for them
	for className, option := range appsettings.ClassesDrawOptions {
		if option == nil {
			appsettings.ClassesDrawOptions[className] = PrepareDrawingOptionsDefault()
		}
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
	ClassesSettings       []*ClassesSettings    `json:"classes_settings"`
	TrackerSettings       TrackerSettings       `json:"tracker_settings"`
	MatPPROFSettings      MatPPROFSettings      `json:"matpprof_settings"`

	sync.RWMutex
	// Exported, but not from JSON
	ClassesDrawOptions map[string]*DrawOptions `json:"-"`
}

func (settings *AppSettings) GetDrawOptions(className string) *DrawOptions {
	settings.Lock()
	found, ok := settings.ClassesDrawOptions[className]
	settings.Unlock()
	if ok {
		return found
	}
	return nil
}

// CudaSettings CUDA settings
type CudaSettings struct {
	Enable bool `json:"enable"`
}

// MatPPROFSettings pprof settings of gocv.Mat
type MatPPROFSettings struct {
	Enable bool `json:"enable"`
}

// GrpcSettings gRPC-server address
type GrpcSettings struct {
	Enable     bool   `json:"enable"`
	ServerIP   string `json:"server_ip"`
	ServerPort int    `json:"server_port"`
}

// ClassesSettings Settings for each possible class
type ClassesSettings struct {
	// Classname basically
	ClassName string `json:"class_name"`
	// Options for visual output (usefull when either imshow or mjpeg output is used)
	DrawingSettings *ObjectDrawingSettings `json:"drawing_settings"`
}

// ObjectDrawingSettings Drawing settings for MJPEG/imshow
type ObjectDrawingSettings struct {
	// Drawing options for detection rectangle
	BBoxSettings BBoxSettings `json:"bbox_settings"`
	// Drawing options for center of detection rectangle
	CentroidSettings CentroidSettings `json:"centroid_settings"`
	// Drawing options for text in top left corner of detection rectangle
	TextSettings TextSettings `json:"text_settings"`
	// Do you want to display ID of object (uuid)
	DisplayObjectID bool `json:"display_object_id"`
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

// TrackerSettings Object tracker settings
type TrackerSettings struct {
	TrackerType string `json:"tracker_type"`
	trackerType TRACKER_TYPE
	// Restriction for maximum points in single track
	MaxPointsInTrack        int                     `json:"max_points_in_track"`
	LinesSettings           []LinesSetting          `json:"lines_settings"`
	SpeedEstimationSettings SpeedEstimationSettings `json:"speed_estimation_settings"`
}

// GetTrackerType Returns enum for tracker type option
func (trs *TrackerSettings) GetTrackerType() TRACKER_TYPE {
	return trs.trackerType
}

// LinesSetting Virtual lines
type LinesSetting struct {
	LineID        int64    `json:"line_id"`
	Begin         [2]int   `json:"begin"`
	End           [2]int   `json:"end"`
	Direction     string   `json:"direction"`
	DetectClasses []string `json:"detect_classes"`
	RGBA          [4]uint8 `json:"rgba"`
	CropMode      string   `json:"crop_mode"`
	// Exported, but not from JSON
	VLine *VirtualLine `json:"-"`
}

// BBoxSettings Options for detection rectangle
type BBoxSettings struct {
	RGBA      [4]uint8 `json:"rgba"`
	Thickness int      `json:"thickness"`
}

// CentroidSettings Options for center of detection rectangle
type CentroidSettings struct {
	RGBA      [4]uint8 `json:"rgba"`
	Radius    int      `json:"radius"`
	Thickness int      `json:"thickness"`
}

// TextSettings Options for text in top left corner of detection rectangle
type TextSettings struct {
	RGBA      [4]uint8 `json:"rgba"`
	Scale     float64  `json:"scale"`
	Thickness int      `json:"thickness"`
	Font      string   `json:"font"` // Possible values are: hershey_simplex, hershey_plain, hershey_duplex, hershey_complex, hershey_triplex, hershey_complex_small, hershey_script_simplex, hershey_script_cddomplex, italic
}

// SpeedEstimationSettings Settings speed estimation
type SpeedEstimationSettings struct {
	// Is this feature enabled?
	Enabled bool `json:"enabled"`
	// Is gRPC sending needed? If yes make sure that 'grpc_settings.enable' is set to 'true' also
	SendGRPC bool `json:"send_grpc"`
	// Map image coordinates to GIS coordinates. EPSG 4326 is handled only currently
	Mapper []GISMapper `json:"mapper"`
}

// GISMapper Map image coordinates to GIS coordinates
type GISMapper struct {
	ImageCoordinates [2]float32 `json:"image_coordinates"`
	EPSG4326         [2]float32 `json:"epsg4326"`
}

type TRACKER_TYPE int

const (
	TRACKER_SIMPLE = TRACKER_TYPE(1)
	TRACKER_KALMAN = TRACKER_TYPE(2)
)

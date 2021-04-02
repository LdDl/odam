package odam

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"os"
	"strings"

	blob "github.com/LdDl/gocv-blob/v2/blob"
	"gocv.io/x/gocv"
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
			LeftPT:        image.Point{X: int(x1), Y: int(y1)},
			RightPT:       image.Point{X: int(x2), Y: int(y2)},
			SourceLeftPT:  image.Point{X: lsettings.Begin[0], Y: lsettings.Begin[1]},
			SourceRightPT: image.Point{X: lsettings.End[0], Y: lsettings.End[1]},
			Direction:     true,
			Color:         color.RGBA{lsettings.RGBA[0], lsettings.RGBA[1], lsettings.RGBA[2], lsettings.RGBA[3]},
		}
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
		lsettings.VLine = &vline
	}

	// Prepare drawing options
	if appsettings.TrackerSettings.DrawTrackSettings.MaxPointsInTrack < 1 {
		fmt.Printf("[WARNING] Field 'max_points_in_track' shoudle be >= 1, but got '%d'. Setting default value = 10\n", appsettings.TrackerSettings.DrawTrackSettings.MaxPointsInTrack)
		appsettings.TrackerSettings.DrawTrackSettings.MaxPointsInTrack = 10
	}
	bboxOpts := blob.DrawBBoxOptions{
		Color: color.RGBA{
			appsettings.TrackerSettings.DrawTrackSettings.BBoxSettings.RGBA[0],
			appsettings.TrackerSettings.DrawTrackSettings.BBoxSettings.RGBA[1],
			appsettings.TrackerSettings.DrawTrackSettings.BBoxSettings.RGBA[2],
			appsettings.TrackerSettings.DrawTrackSettings.BBoxSettings.RGBA[3],
		},
		Thickness: appsettings.TrackerSettings.DrawTrackSettings.BBoxSettings.Thickness,
	}
	cenOpts := blob.DrawCentroidOptions{
		Color: color.RGBA{
			appsettings.TrackerSettings.DrawTrackSettings.CentroidSettings.RGBA[0],
			appsettings.TrackerSettings.DrawTrackSettings.CentroidSettings.RGBA[1],
			appsettings.TrackerSettings.DrawTrackSettings.CentroidSettings.RGBA[2],
			appsettings.TrackerSettings.DrawTrackSettings.CentroidSettings.RGBA[3],
		},
		Radius:    appsettings.TrackerSettings.DrawTrackSettings.CentroidSettings.Radius,
		Thickness: appsettings.TrackerSettings.DrawTrackSettings.CentroidSettings.Thickness,
	}
	textOpts := blob.DrawTextOptions{
		Color: color.RGBA{
			appsettings.TrackerSettings.DrawTrackSettings.TextSettings.RGBA[0],
			appsettings.TrackerSettings.DrawTrackSettings.TextSettings.RGBA[1],
			appsettings.TrackerSettings.DrawTrackSettings.TextSettings.RGBA[2],
			appsettings.TrackerSettings.DrawTrackSettings.TextSettings.RGBA[3],
		},
		Scale:     appsettings.TrackerSettings.DrawTrackSettings.TextSettings.Scale,
		Thickness: appsettings.TrackerSettings.DrawTrackSettings.TextSettings.Thickness,
	}
	switch strings.ToLower(appsettings.TrackerSettings.DrawTrackSettings.TextSettings.Font) {
	case "hershey_simplex":
		textOpts.Font = gocv.FontHersheySimplex
		break
	case "hershey_plain":
		textOpts.Font = gocv.FontHersheyPlain
		break
	case "hershey_duplex":
		textOpts.Font = gocv.FontHersheyDuplex
		break
	case "hershey_complex":
		textOpts.Font = gocv.FontHersheyComplex
		break
	case "hershey_triplex":
		textOpts.Font = gocv.FontHersheyTriplex
		break
	case "hershey_complex_small":
		textOpts.Font = gocv.FontHersheyComplexSmall
		break
	case "hershey_script_simplex":
		textOpts.Font = gocv.FontHersheyScriptSimplex
		break
	case "hershey_script_complex":
		textOpts.Font = gocv.FontHersheyScriptComplex
		break
	case "italic":
		textOpts.Font = gocv.FontItalic
		break
	default:
		textOpts.Font = gocv.FontHersheyPlain
		break
	}

	drOpts := blob.NewDrawOptions(bboxOpts, cenOpts, textOpts)
	appsettings.TrackerSettings.DrawOptions = drOpts

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
	MatPPROFSettings      MatPPROFSettings      `json:"matpprof_settings"`
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
	LinesSettings     []LinesSetting    `json:"lines_settings"`
	DrawTrackSettings DrawTrackSettings `json:"draw_track_settings"`

	// Exported, but not from JSON
	DrawOptions *blob.DrawOptions `json:"-"`
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

// DrawTrackSettings Drawing settings for MJPEG/imshow
type DrawTrackSettings struct {
	// Restriction for maximum points in single track
	MaxPointsInTrack int `json:"max_points_in_track"`
	// Drawing options for detection rectangle
	BBoxSettings BBoxSettings `json:"bbox_settings"`
	// Drawing options for center of detection rectangle
	CentroidSettings CentroidSettings `json:"centroid_settings"`
	// Drawing options for text in top left corner of detection rectangle
	TextSettings TextSettings `json:"text_settings"`
	// Do you want to display ID of object (uuid)
	DisplayObjectID bool `json:"display_object_id"`
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

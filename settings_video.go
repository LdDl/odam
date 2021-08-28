package odam

import (
	"fmt"
)

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

// Prepare Prepares this structure for further usage
func (vs *VideoSettings) Prepare() {
	if vs.Width <= 0 {
		vs.Width = 640
		fmt.Println("[WARNING] Field 'width' in 'video_settings' has not been provided (or <=0). Using default 640 width")
	}
	if vs.Height <= 0 {
		vs.Height = 360
		fmt.Println("[WARNING] Field 'height' in 'video_settings' has not been provided (or <=0). Using default 360 height")
	}
	if vs.ReducedWidth <= 0 {
		vs.ReducedWidth = vs.Width
		fmt.Println("[WARNING] Field 'reduced_width' in 'video_settings' has not been provided (or <=0). Using default reduced_width = width")
	}
	if vs.ReducedHeight <= 0 {
		vs.ReducedHeight = vs.Height
		fmt.Println("[WARNING] Field 'reduced_height' in 'video_settings' has not been provided (or <=0). Using default reduced_height = height")
	}
	if vs.ReducedWidth > vs.Width {
		vs.ReducedWidth = vs.Width
		fmt.Println("[WARNING] Field 'reduced_width' in 'video_settings' > 'width'. Using default reduced_width = width")
	}
	if vs.ReducedHeight > vs.Height {
		vs.ReducedHeight = vs.Height
		fmt.Println("[WARNING] Field 'reduced_height' in 'video_settings' > 'height'. Using default reduced_height = height")
	}
	vs.ScaleX = float64(vs.Width) / float64(vs.ReducedWidth)
	vs.ScaleY = float64(vs.Height) / float64(vs.ReducedHeight)
}

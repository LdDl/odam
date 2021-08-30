package odam

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

type TRACKER_TYPE int

const (
	TRACKER_SIMPLE = TRACKER_TYPE(1)
	TRACKER_KALMAN = TRACKER_TYPE(2)
)

// TrackerSettings Object tracker settings
type TrackerSettings struct {
	TrackerType string `json:"tracker_type"`
	trackerType TRACKER_TYPE
	// Restriction for maximum points in single track
	MaxPointsInTrack        int                     `json:"max_points_in_track"`
	LinesSettings           []*LinesSetting         `json:"lines_settings"`
	PolygonsSettings        []*PolygonsSetting      `json:"polygons_settings"`
	SpeedEstimationSettings SpeedEstimationSettings `json:"speed_estimation_settings"`
}

// GetTrackerType Returns enum for tracker type option
func (trs *TrackerSettings) GetTrackerType() TRACKER_TYPE {
	return trs.trackerType
}

// Prepare Prepares this structure for further usage
func (trs *TrackerSettings) Prepare() {
	switch strings.ToLower(trs.TrackerType) {
	case "simple":
		trs.TrackerType = strings.ToLower(trs.TrackerType)
		trs.trackerType = TRACKER_SIMPLE
		break
	case "kalman":
		trs.TrackerType = strings.ToLower(trs.TrackerType)
		trs.trackerType = TRACKER_KALMAN
		break
	case "":
		fmt.Println("[WARNING]: Field 'tracker_type' is empty. Settings default value 'simple'")
		trs.TrackerType = "simple"
		trs.trackerType = TRACKER_SIMPLE
	default:
		fmt.Printf("[WARNING]: Value '%s' 'tracker_type' is not supported. Settings default value 'simple'\n", trs.TrackerType)
		trs.TrackerType = "simple"
		trs.trackerType = TRACKER_SIMPLE
		break
	}
	if len(trs.LinesSettings) == 0 {
		fmt.Println("[WARNING] No 'lines_settings'? Please check if it is true")
	}
	for _, lsettings := range trs.LinesSettings {
		vline := NewVirtualLine(lsettings.Begin[0], lsettings.Begin[1], lsettings.End[0], lsettings.End[1])
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
	if trs.MaxPointsInTrack < 1 {
		fmt.Printf("[WARNING] Field 'max_points_in_track' shoudle be >= 1, but got '%d'. Setting default value = 10\n", trs.MaxPointsInTrack)
		trs.MaxPointsInTrack = 10
	}
	if len(trs.PolygonsSettings) == 0 {
		fmt.Println("[WARNING] No 'polygons_settings'? Please check if it is true")
	}
	for _, psettings := range trs.PolygonsSettings {
		ptsCollected := make([]image.Point, len(psettings.Coordinates))
		for _, pair := range psettings.Coordinates {
			ptsCollected = append(ptsCollected, image.Point{X: pair[0], Y: pair[1]})
		}
		vpolygon := NewVirtualPolygon(ptsCollected...)
		vpolygon.Color = color.RGBA{psettings.RGBA[0], psettings.RGBA[1], psettings.RGBA[2], psettings.RGBA[3]}
		psettings.VPolygon = vpolygon
	}
}

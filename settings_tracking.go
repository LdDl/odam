package odam

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
	LinesSettings           []LinesSetting          `json:"lines_settings"`
	SpeedEstimationSettings SpeedEstimationSettings `json:"speed_estimation_settings"`
}

// GetTrackerType Returns enum for tracker type option
func (trs *TrackerSettings) GetTrackerType() TRACKER_TYPE {
	return trs.trackerType
}

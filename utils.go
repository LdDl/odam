package odam

import (
	"image"
	"math"

	blob "github.com/LdDl/gocv-blob/v2/blob"
	"gocv.io/x/gocv"
)

// Round Rounds float64 to int
func Round(v float64) int {
	if v >= 0 {
		return int(math.Floor(v + 0.5))
	}
	return int(math.Ceil(v - 0.5))
}

// FixRectForOpenCV Corrects rectangle's bounds for provided max-widtht and max-height
// Helps to avoid BBox error assertion
func FixRectForOpenCV(r *image.Rectangle, maxCols, maxRows int) {
	if r.Min.X <= 0 {
		r.Min.X = 0
	}
	if r.Min.Y < 0 {
		r.Min.Y = 0
	}
	if r.Max.X >= maxCols {
		r.Max.X = maxCols - 1
	}
	if r.Max.Y >= maxRows {
		r.Max.Y = maxRows - 1
	}
}

// ClassInfoGRPC Prepares gRPC message 'ClassInfo'
// Blob object should be provided
func ClassInfoGRPC(b blob.Blobie) *ClassInfo {
	return &ClassInfo{
		ClassId:   int32(b.GetClassID()),
		ClassName: b.GetClassName(),
	}
}

// DetectionInfoGRPC Prepares gRPC message 'Detection'
// BBox (x-leftop, y-leftop, width and height of bounding box) information should be provided
func DetectionInfoGRPC(xmin, ymin, width, height int32) *Detection {
	return &Detection{
		XLeft:  xmin,
		YTop:   ymin,
		Width:  width,
		Height: height,
	}
}

// VirtualLineInfoGRPC Prepares gRPC message 'VirtualLineInfo'
// Identifier of a line (int64) and its parameters (x0,y0 and x1,y1) should be provide
func VirtualLineInfoGRPC(lineID int64, virtualLine *VirtualLine) *VirtualLineInfo {
	return &VirtualLineInfo{
		Id:     lineID,
		LeftX:  int32(virtualLine.SourceLeftPT.X),
		LeftY:  int32(virtualLine.SourceLeftPT.Y),
		RightX: int32(virtualLine.SourceRightPT.X),
		RightY: int32(virtualLine.SourceRightPT.Y),
	}
}

// TrackInfoInfoGRPC Prepares gRPC message 'TrackInfo'
// Next data should be provided:
// Blob object for track extraction
// Key for extracting speed infromation
// Width/Height scale for EuclideanPoint correction to actual coordinates
// Coverter function (from pixel to WGS84)
func TrackInfoInfoGRPC(b blob.Blobie, speedKey string, scalex, scaley float32, gisConverter func(gocv.Point2f) gocv.Point2f) *TrackInfo {
	// Extract estimated speed information
	spd := float32(0.0)
	if spdInterface, ok := b.GetProperty(speedKey); ok {
		switch spdInterface.(type) { // Want to be sure that interface is float32
		case float32:
			spd = spdInterface.(float32)
			break
		default:
			break
		}
	}
	// spd := float32(0.0)
	// do, err := CastBlobToDetectedObject(b)
	// if err != nil {
	// 	fmt.Println("[WARNING] Can't cast blob.Blobie to *odam.DetectedObject:", err)
	// } else {
	// 	spd = do.GetSpeed()
	// }
	// Collect track information
	trackPixels := b.GetTrack()
	trackUnionInfo := make([]*Point, len(trackPixels))
	for i, stdPt := range trackPixels {
		// Convert point to spatial representation via provided converter function
		cvPt := STDPointToGoCVPoint2F(stdPt)
		gisPt := gisConverter(cvPt)
		// Collect point information
		trackUnionInfo[i] = &Point{
			EuclideanPoint: &EuclideanPoint{
				X: cvPt.X * scalex,
				Y: cvPt.Y * scaley,
			},
			Wgs84Point: &WGS84Point{
				Longitude: gisPt.X,
				Latitude:  gisPt.Y,
			},
		}
	}
	return &TrackInfo{
		EstimatedSpeed: spd,
		Points:         trackUnionInfo,
	}
}

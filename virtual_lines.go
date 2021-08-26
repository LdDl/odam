package odam

import (
	"image"
	"image/color"
	"math"

	blob "github.com/LdDl/gocv-blob/v2/blob"
	"gocv.io/x/gocv"
)

// VIRTUAL_LINE_TYPE Alias to int
type VIRTUAL_LINE_TYPE int

const (
	// HORIZONTAL_LINE Represents the line with Y{1} of (X{1}Y{1}) = Y{2} of (X{2}Y{2})
	HORIZONTAL_LINE = VIRTUAL_LINE_TYPE(iota + 1)
	// OBLIQUE_LINE Represents the line with Y{1} of (X{1}Y{1}) <> Y{2} of (X{2}Y{2}) (so it has some angle)
	OBLIQUE_LINE
)

// VirtualLine Detection line attributes
type VirtualLine struct {
	// Point on the left [scaled]
	LeftPT image.Point `json:"-"`
	// Point on the right [scaled]
	RightPT image.Point `json:"-"`
	// Color of line
	Color color.RGBA `json:"-"`
	// Direction of traffic flow
	Direction bool `json:"-"`
	// Is crossing object should be cropped for futher work with it?
	CropObject bool `json:"-"`
	// Point on the left [non-scaled]
	SourceLeftPT image.Point `json:"-"`
	// Point on the right [non-scaled]
	SourceRightPT image.Point `json:"-"`
	// Type of virtual line: could be horizontal or oblique
	LineType VIRTUAL_LINE_TYPE `json:"-"`
}

// Constructor for VirtualLine
// (x1, y1) - Left
// (x2, y2) - Right
func NewVirtualLine(x1, y1, x2, y2 int) *VirtualLine {
	vline := VirtualLine{
		LeftPT:        image.Point{X: x1, Y: y1},
		RightPT:       image.Point{X: x2, Y: y2},
		SourceLeftPT:  image.Point{X: x1, Y: y1},
		SourceRightPT: image.Point{X: x2, Y: y2},
		Direction:     true,
	}
	if y1 == y2 {
		vline.LineType = HORIZONTAL_LINE
	} else {
		vline.LineType = OBLIQUE_LINE
	}
	return &vline
}

// Scale Scales down (so scale factor can be > 1.0 ) virtual line
// (scaleX, scaleY) - How to scale source (x1,y1) and (x2,y2) coordinates
// Important notice:
// 1. Source coordinates won't be modified
// 2. Source coordinates would be used for scaling. So you can't scale line multiple times
func (vline *VirtualLine) Scale(scaleX, scaleY float64) {
	vline.LeftPT.X = int(math.Round(float64(vline.SourceLeftPT.X) / scaleX))
	vline.LeftPT.Y = int(math.Round(float64(vline.SourceLeftPT.Y) / scaleY))
	vline.RightPT.X = int(math.Round(float64(vline.SourceRightPT.X) / scaleX))
	vline.RightPT.Y = int(math.Round(float64(vline.SourceRightPT.Y) / scaleY))
}

// Draw Draw virtual line on image
func (vline *VirtualLine) Draw(img *gocv.Mat) {
	gocv.Line(img, vline.LeftPT, vline.RightPT, vline.Color, 3)
}

// IsBlobCrossedLine Wrapper around b.IsCrossedTheLine(y2,x1,y1,direction) and b.IsCrossedTheObliqueLine(x2,y2,x1,y1,direction).
// See ref. https://github.com/LdDl/gocv-blob/blob/master/v2/blob/line_cross.go
func (vline *VirtualLine) IsBlobCrossedLine(b blob.Blobie) bool {
	switch vline.LineType {
	case HORIZONTAL_LINE:
		return b.IsCrossedTheLine(vline.RightPT.Y, vline.LeftPT.X, vline.RightPT.X, vline.Direction)
	case OBLIQUE_LINE:
		return b.IsCrossedTheObliqueLine(vline.RightPT.X, vline.RightPT.Y, vline.LeftPT.X, vline.LeftPT.Y, vline.Direction)
	default:
		// This actually should not happen
		// Is this really needed to have error returning in this function?
		break
	}
	return false
}

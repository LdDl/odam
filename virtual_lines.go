package odam

import (
	"image"
	"image/color"
	"math"

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
// (scaleX, scaleY) - How to scale source (x1,y1) and (x2,y2) coordinates
// @todo scaling in black box - pretty bad idea. I guess we must have .Scale(x,y) method to call it if needed.
func NewVirtualLine(x1, y1, x2, y2 int, scaleX, scaleY float64) *VirtualLine {
	x1Scaled := int(math.Round(float64(x1) / scaleX))
	y1Scaled := int(math.Round(float64(y1) / scaleY))
	x2Scaled := int(math.Round(float64(x2) / scaleX))
	y2Scaled := int(math.Round(float64(y2) / scaleY))
	vline := VirtualLine{
		LeftPT:        image.Point{X: x1Scaled, Y: y1Scaled},
		RightPT:       image.Point{X: x2Scaled, Y: y2Scaled},
		SourceLeftPT:  image.Point{X: x1, Y: y1},
		SourceRightPT: image.Point{X: x2, Y: y2},
		Direction:     true,
	}
	if y1Scaled == y2Scaled {
		vline.LineType = HORIZONTAL_LINE
	} else {
		vline.LineType = OBLIQUE_LINE
	}
	return &vline
}

// Draw Draw virtual line on image
func (vline *VirtualLine) Draw(img *gocv.Mat) {
	gocv.Line(img, vline.LeftPT, vline.RightPT, vline.Color, 3)
}

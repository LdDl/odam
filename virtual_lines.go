package odam

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
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

// Draw Draw virtual line on image
func (vline *VirtualLine) Draw(img *gocv.Mat) {
	gocv.Line(img, vline.LeftPT, vline.RightPT, vline.Color, 3)
}

// VIRTUAL_LINE_TYPE Alias to int
type VIRTUAL_LINE_TYPE int

const (
	// HORIZONTAL_LINE Represents the line with Y{1} of (X{1}Y{1}) = Y{2} of (X{2}Y{2})
	HORIZONTAL_LINE = VIRTUAL_LINE_TYPE(iota + 1)
	// OBLIQUE_LINE Represents the line with Y{1} of (X{1}Y{1}) <> Y{2} of (X{2}Y{2}) (so it has some angle)
	OBLIQUE_LINE
)

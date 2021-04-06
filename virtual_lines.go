package odam

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

// VirtualLine Detection line attributes
type VirtualLine struct {
	LeftPT        image.Point `json:"-"`
	RightPT       image.Point `json:"-"`
	Color         color.RGBA  `json:"-"`
	Direction     bool        `json:"-"`
	CropObject    bool        `json:"-"`
	SourceLeftPT  image.Point `json:"-"`
	SourceRightPT image.Point `json:"-"`

	LineType VIRTUAL_LINE_TYPE `json:"-"`
}

// Draw Draw virtual line on image
func (vline *VirtualLine) Draw(img *gocv.Mat) {
	gocv.Line(img, vline.LeftPT, vline.RightPT, vline.Color, 3)
}

type VIRTUAL_LINE_TYPE int

const (
	HORIZONTAL_LINE = VIRTUAL_LINE_TYPE(1)
	OBLIQUE_LINE = VIRTUAL_LINE_TYPE(2)
)

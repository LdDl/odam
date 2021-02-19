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
}

// Draw Draw virtual line on image
func (vline *VirtualLine) Draw(img *gocv.Mat) {
	gocv.Line(img, vline.LeftPT, vline.RightPT, vline.Color, 3)
}

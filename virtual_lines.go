package odam

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

// VirtualLine Detection line attributes
type VirtualLine struct {
	LeftPT    image.Point
	RightPT   image.Point
	Color     color.RGBA
	Direction bool
}

// Draw Draw virtual line on image
func (vline *VirtualLine) Draw(img *gocv.Mat) {
	gocv.Line(img, vline.LeftPT, vline.RightPT, vline.Color, 3)
}

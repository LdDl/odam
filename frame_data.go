package odam

import (
	"gocv.io/x/gocv"
)

// FrameData Wrapper around gocv.Mat
type FrameData struct {
	ImgSource gocv.Mat //  Source image
	ImgScaled gocv.Mat // Scaled image
}

// NewFrameData Simplify creation of FrameData
func NewFrameData() *FrameData {
	fd := FrameData{
		ImgSource: gocv.NewMat(),
		ImgScaled: gocv.NewMat(),
	}
	return &fd
}

// Close Simplify memory management for each gocv.Mat of FrameData
func (fd *FrameData) Close() {
	fd.ImgSource.Close()
	fd.ImgScaled.Close()
}

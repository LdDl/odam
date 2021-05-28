package odam

import (
	"bytes"
	"image"
	"image/jpeg"

	"gocv.io/x/gocv"
)

// FrameData Wrapper around gocv.Mat
type FrameData struct {
	ImgSource gocv.Mat //  Source image
	ImgScaled gocv.Mat // Scaled image
	ImgSTD    image.Image
}

// NewFrameData Simplifies creation of FrameData
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

// Preprocess Scales image to given width and height
func (fd *FrameData) Preprocess(width, height int) error {
	gocv.Resize(fd.ImgSource, &fd.ImgScaled, image.Point{X: width, Y: height}, 0, 0, gocv.InterpolationDefault)
	stdImage, err := fd.ImgScaled.ToImage()
	if err != nil {
		return err
	}
	fd.ImgSTD = stdImage
	return nil
}

func matToBytes(im *gocv.Mat) (ans []byte, err error) {
	stdImage, err := im.ToImage()
	if err != nil {
		return ans, err
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, stdImage, nil)
	if err != nil {
		return ans, err
	}
	ans = buf.Bytes()
	return ans, nil
}

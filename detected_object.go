package odam

import (
	"image"
)

// DetectedObject Store detected object info
type DetectedObject struct {
	Rect       image.Rectangle
	ClassID    int
	ClassName  string
	Confidence float32
}

// DetectedObjects Just alias to slice of DetectedObject
type DetectedObjects []*DetectedObject

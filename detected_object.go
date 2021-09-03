package odam

import (
	"fmt"
	"image"
	reflect "reflect"

	blob "github.com/LdDl/gocv-blob/v2/blob"
)

// DetectedObject Store detected object info
type DetectedObject struct {
	// Bounding box
	Rect image.Rectangle
	// Class identifier
	ClassID int
	// Class description (in most of cases this is just another class unique identifier)
	ClassName string
	// The probability that an object belongs to the specified class
	Confidence float32

	// Unexported
	speed float32

	/*  Wrap blob.Blobie for duck typing */
	blob.Blobie
}

// DetectedObjects Just alias to slice of DetectedObject
type DetectedObjects []*DetectedObject

// CastBlobToDetectedObject Handy interface caster
// blob.Blobie -> *odam.DetectedObject
func CastBlobToDetectedObject(b blob.Blobie) (*DetectedObject, error) {
	switch b.(type) {
	case *DetectedObject:
		return b.(*DetectedObject), nil
	default:
		return nil, fmt.Errorf("Blob should be of type *odam.DetectedObject, but got %s", reflect.TypeOf(b))
	}
}

// SetSpeed Sets last measured speed value
func (do *DetectedObject) SetSpeed(v float32) {
	do.speed = v
}

// GetSpeed Returns last measured value of speed
func (do *DetectedObject) GetSpeed() float32 {
	return do.speed
}

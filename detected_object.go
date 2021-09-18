package odam

import (
	"fmt"
	"image"
	reflect "reflect"

	darknet "github.com/LdDl/go-darknet"
	blob "github.com/LdDl/gocv-blob/v2/blob"
	"github.com/pkg/errors"
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

// DetectObjects Detect objects for provided Go's image via neural network
//
// app - Application instance containing pointer to neural network for object detection
// imgSTD - image.Image from Go's standart library
// filters - List of classes for which you need to filter detected objects
//
func DetectObjects(app *Application, imgSTD image.Image, filters ...string) ([]*DetectedObject, error) {
	darknetImage, err := darknet.Image2Float32(imgSTD)
	if err != nil {
		return nil, errors.Wrap(err, "Can't convert image to Darknet's format")
	}
	dr, err := app.neuralNetwork.Detect(darknetImage)
	if err != nil {
		darknetImage.Close()
		return nil, errors.Wrap(err, "Can't make detection on Darknet image")
	}
	darknetImage.Close() // free the memory
	darknetImage = nil
	detectedRects := make([]*DetectedObject, 0, len(dr.Detections))
	for _, d := range dr.Detections {
		for i := range d.ClassIDs {
			if stringInSlice(&d.ClassNames[i], filters) {
				bBox := d.BoundingBox
				minX, minY := float64(bBox.StartPoint.X), float64(bBox.StartPoint.Y)
				maxX, maxY := float64(bBox.EndPoint.X), float64(bBox.EndPoint.Y)
				rect := DetectedObject{
					Rect:       image.Rect(Round(minX), Round(minY), Round(maxX), Round(maxY)),
					ClassName:  d.ClassNames[i],
					ClassID:    d.ClassIDs[i],
					Confidence: d.Probabilities[i],
				}
				detectedRects = append(detectedRects, &rect)
			}
		}
	}
	return detectedRects, nil
}

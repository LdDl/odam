package odam

import (
	"fmt"
	"image"
	reflect "reflect"

	blob "github.com/LdDl/gocv-blob/v2/blob"
	"gocv.io/x/gocv"
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

const (
	yoloScaleFactor = 1.0 / 255.0
	yoloHeight      = 608
	yoloWidth       = 608
)

var (
	yoloSize     = image.Point{yoloHeight, yoloWidth}
	yoloMean     = gocv.NewScalar(0.0, 0.0, 0.0, 0.0)
	yoloBlobName = ""
)

// DetectObjects Detect objects for provided Go's image via neural network
//
// app - Application instance containing pointer to neural network for object detection
// imgSTD - image.Image from Go's standart library
// filters - List of classes for which you need to filter detected objects
//
func DetectObjects(app *Application, img gocv.Mat, filters ...string) ([]*DetectedObject, error) {
	blobImg := gocv.BlobFromImage(img, yoloScaleFactor, yoloSize, yoloMean, true, false)
	app.neuralNetwork.SetInput(blobImg, yoloBlobName)
	detections := app.neuralNetwork.ForwardLayers(app.layersNames)
	return postprocess(detections, 0.5, 0.4, filters)
}

func postprocess(detections []gocv.Mat, confidenceThreshold, nmsThreshold float32, filters []string) ([]*DetectedObject, error) {
	// @todo
	return nil, nil
}

func getClassIDAndConfidence(x []float32) (int, float32) {
	res := 0
	max := float32(0.0)
	for i, y := range x {
		if y > max {
			max = y
			res = i
		}
	}
	return res, max
}

func calculateBoundingBox(netWidth, netHeight float32, row []float32) image.Rectangle {
	if len(row) < 4 {
		return image.Rect(0, 0, 0, 0)
	}
	centerX := int(row[0] * netWidth)
	centerY := int(row[1] * netHeight)
	width := int(row[2] * netWidth)
	height := int(row[3] * netHeight)
	left := (centerX - width/2)
	top := (centerY - height/2)
	return image.Rect(left, top, left+width, top+height)
}

package odam

import (
	"image"
	"sort"
)

// DetectedObject Store detected object info
type DetectedObject struct {
	Rect       image.Rectangle
	Classname  string
	ClassID    int
	ID         int64
	Confidence float64
}

type DetectedObjects []*DetectedObject

func (detections DetectedObjects) Len() int { return len(detections) }
func (detections DetectedObjects) Swap(i, j int) {
	detections[i], detections[j] = detections[j], detections[i]
}
func (detections DetectedObjects) Less(i, j int) bool {
	return detections[i].Confidence < detections[j].Confidence
}

func NMS(detections DetectedObjects, iouTreshold float64) DetectedObjects {
	sort.Sort(detections)
	nms := make(DetectedObjects, 0)
	if len(detections) == 0 {
		return nms
	}
	nms = append(nms, detections[0])
	for i := 1; i < len(detections); i++ {
		tocheck, del := len(nms), false
		for j := 0; j < tocheck; j++ {
			currIOU := IOUFloat64(detections[i].Rect, nms[j].Rect)
			if currIOU > iouTreshold && detections[i].ClassID == nms[j].ClassID {
				del = true
				break
			}
		}
		if !del {
			nms = append(nms, detections[i])
		}
	}
	return nms
}

// IOUFloat64 Intersection Over Union (x64)
func IOUFloat64(r1, r2 image.Rectangle) float64 {
	intersection := r1.Intersect(r2)
	interArea := intersection.Dx() * intersection.Dy()
	r1Area := r1.Dx() * r1.Dy()
	r2Area := r2.Dx() * r2.Dy()
	return float64(interArea) / float64(r1Area+r2Area-interArea)
}

// IOUFloat32 Intersection Over Union (x32)
func IOUFloat32(r1, r2 image.Rectangle) float32 {
	intersection := r1.Intersect(r2)
	interArea := intersection.Dx() * intersection.Dy()
	r1Area := r1.Dx() * r1.Dy()
	r2Area := r2.Dx() * r2.Dy()
	return float32(interArea) / float32(r1Area+r2Area-interArea)
}

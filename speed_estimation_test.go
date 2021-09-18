package odam

import (
	"testing"
	"time"

	"gocv.io/x/gocv"
)

func TestGetPerspectiveTransformer(t *testing.T) {
	src := []gocv.Point2f{
		{X: float32(1200), Y: float32(278)},
		{X: float32(87), Y: float32(328)},
		{X: float32(36), Y: float32(583)},
		{X: float32(1205), Y: float32(698)},
	}
	dst := []gocv.Point2f{
		{X: float32(6.602018), Y: float32(52.036769)},
		{X: float32(6.603227), Y: float32(52.036181)},
		{X: float32(6.603638), Y: float32(52.036558)},
		{X: float32(6.603560), Y: float32(52.036730)},
	}
	transformMat, converter := GetPerspectiveTransformer(src, dst)
	for i, p := range src {
		res := converter(p)
		if res != dst[i] {
			t.Errorf("Incorrect perspective transformation. Should be be %v, but got %v", dst[i], res)
		}
	}
	transformMat.Close()
}

func TestHaversine(t *testing.T) {
	src := gocv.Point2f{X: 6.602018, Y: 52.036769}
	dst := gocv.Point2f{X: 6.603560, Y: 52.036730}
	dist := Haversine(src, dst)
	correctGreatCircleDistance := float32(0.105567716)
	if dist != correctGreatCircleDistance {
		t.Errorf("Great circle distance should be %f, but got %f", correctGreatCircleDistance, dist)
	}
}

func TestEstimateSpeed(t *testing.T) {
	srcMat := []gocv.Point2f{
		{X: 1200, Y: 278},
		{X: 87, Y: 328},
		{X: 36, Y: 583},
		{X: 1205, Y: 698},
	}
	dstMat := []gocv.Point2f{
		{X: 6.602018, Y: 52.036769},
		{X: 6.603227, Y: 52.036181},
		{X: 6.603638, Y: 52.036558},
		{X: 6.603560, Y: 52.036730},
	}
	transformMat, converter := GetPerspectiveTransformer(srcMat, dstMat)
	src := gocv.Point2f{X: 1200, Y: 278}
	dst := gocv.Point2f{X: 1205, Y: 698}
	start := time.Now()
	finish := start.Add(6000 * time.Millisecond)
	res := EstimateSpeed(src, dst, start, finish, converter)
	correctSpeed := float32(63.30266)
	transformMat.Close()
	if res != correctSpeed {
		t.Errorf("Estimated speed should be %f, but got %f", res, correctSpeed)
	}
}

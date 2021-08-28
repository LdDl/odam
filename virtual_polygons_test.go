package odam

import (
	"image"
	"testing"

	blob "github.com/LdDl/gocv-blob/v2/blob"
)

func TestPolygonType(t *testing.T) {
	vpolygons := []*VirtualPolygon{
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 1, Y: 0},
			image.Point{X: 1, Y: 1},
			image.Point{X: 0, Y: 1},
		),
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 3, Y: 3},
			image.Point{X: 0, Y: 1},
		),
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 3, Y: 7},
			image.Point{X: 0, Y: 1},
		),
	}
	correctPolygonTypes := []VIRTUAL_POLYGON_TYPE{
		CONVEX_POLYGON,
		CONCAVE_POLYGON,
		CONVEX_POLYGON,
	}
	for i, vpolygon := range vpolygons {
		if vpolygon.PolygonType != correctPolygonTypes[i] {
			t.Errorf("Polygon with coordinates %v should be of type '%d' but got %d",
				vpolygon.Coordinates,
				correctPolygonTypes[i],
				vpolygon.PolygonType,
			)
		}
	}
}
func TestPolygonContainsPoint(t *testing.T) {
	vpolygons := []*VirtualPolygon{
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 0, Y: 5},
		),
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 0, Y: 5},
		),
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 5, Y: 0},
		),
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 5, Y: 0},
		),
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 5, Y: 0},
		),
		NewVirtualPolygon(
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 0, Y: 5},
		),
	}
	points := []image.Point{
		image.Point{X: 20, Y: 20},
		image.Point{X: 4, Y: 4},
		image.Point{X: 3, Y: 3},
		image.Point{X: 5, Y: 1},
		image.Point{X: 7, Y: 2},
		image.Point{X: -2, Y: 12},
	}
	correctAnswers := []bool{
		false,
		true,
		true,
		true,
		false,
		false,
	}
	for i, vpolygon := range vpolygons {
		if !vpolygon.isConvex() {
			t.Errorf("Polygon with coordinates %v should be of convex, but it is not. Got type: '%d'", vpolygon.Coordinates, vpolygon.PolygonType)
		}
		answer := vpolygon.ContainsPoint(points[i])
		if answer != correctAnswers[i] {
			if correctAnswers[i] {
				t.Errorf("Polygon with coordinates %v should contain point [%d, %d]. Actual answer is %t", vpolygon.Coordinates, points[i].X, points[i].Y, answer)
			} else {
				t.Errorf("Polygon with coordinates %v should not contain point [%d, %d]. Actual answer is %t", vpolygon.Coordinates, points[i].X, points[i].Y, answer)
			}
		}
	}
}

func TestPolygonContainsBlob(t *testing.T) {
	simpleA := blob.NewSimpleBlobie(image.Rect(26, 8, 44, 18), nil)
	simpleB := blob.NewSimpleBlobie(image.Rect(59, 8, 77, 23), nil)
	simpleC := blob.NewSimpleBlobie(image.Rect(40, 29, 61, 46), nil)
	blobies := []blob.Blobie{simpleA, simpleB, simpleC}
	vpolygon := NewVirtualPolygon(
		image.Point{X: 23, Y: 15},
		image.Point{X: 67, Y: 15},
		image.Point{X: 67, Y: 41},
		image.Point{X: 23, Y: 41},
	)
	correctAnswers := []bool{
		false,
		false,
		true,
	}
	for i, b := range blobies {
		center := b.GetCenter()
		answer := vpolygon.ContainsBlob(b)
		if answer != correctAnswers[i] {
			if correctAnswers[i] {
				t.Errorf("Polygon with coordinates %v should contain blob center at [%d, %d]. Actual answer is %t", vpolygon.Coordinates, center.X, center.Y, answer)
			} else {
				t.Errorf("Polygon with coordinates %v should not contain blob with center at [%d, %d]. Actual answer is %t", vpolygon.Coordinates, center.X, center.Y, answer)
			}
		}
	}
}

func TestPolygonBlobEnter(t *testing.T) {
	// @todo
}

func TestPolygonBlobLeft(t *testing.T) {
	// @todo
}

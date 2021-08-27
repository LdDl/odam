package odam

import (
	"image"
	"testing"
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

func TestPolygonContains(t *testing.T) {
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
	correctAnswer := []bool{
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
		if answer != correctAnswer[i] {
			if correctAnswer[i] {
				t.Errorf("Polygon with coordinates %v should contain point [%d, %d]. Actual answer is %t", vpolygon.Coordinates, points[i].X, points[i].Y, answer)
			} else {
				t.Errorf("Polygon with coordinates %v should not contain point [%d, %d]. Actual answer is %t", vpolygon.Coordinates, points[i].X, points[i].Y, answer)
			}
		}
	}
}

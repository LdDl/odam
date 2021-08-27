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
			image.Point{X: 7, Y: 7},
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

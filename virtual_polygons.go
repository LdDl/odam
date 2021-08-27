package odam

import (
	"image"
	"image/color"
	"math"
)

// POLYGON_TYPE Alias to int
type POLYGON_TYPE int

const (
	// CONVEX_POLYGON See ref. https://en.wikipedia.org/wiki/Convex_polygon
	CONVEX_POLYGON = POLYGON_TYPE(iota + 1)
	// CONCAVE_POLYGON See ref. https://en.wikipedia.org/wiki/Concave_polygon
	CONCAVE_POLYGON
)

// VirtualPolygon Detection polygon attributes
type VirtualPolygon struct {
	// Color of stroke line
	Color color.RGBA `json:"-"`
	// Information about coordinates [scaled]
	Coordinates []image.Point `json:"-"`
	// Information about coordinates [non-scaled]
	SourceCoordinates []image.Point `json:"-"`
	// Type of virtual polygon: could be convex or concave
	PolygonType POLYGON_TYPE `json:"-"`
}

// Constructor for VirtualPolygon
// (x1, y1) - Left
// (x2, y2) - Right
func NewVirtualPolygon(pairs ...[2]int) *VirtualPolygon {
	vpolygon := VirtualPolygon{
		Coordinates:       make([]image.Point, len(pairs)),
		SourceCoordinates: make([]image.Point, len(pairs)),
	}
	for i := range pairs {
		vpolygon.Coordinates[i] = image.Point{X: pairs[i][0], Y: pairs[i][1]}
		vpolygon.SourceCoordinates[i] = image.Point{X: pairs[i][0], Y: pairs[i][1]}
	}
	if vpolygon.isConvex() {
		vpolygon.PolygonType = CONVEX_POLYGON
	} else {
		vpolygon.PolygonType = CONCAVE_POLYGON
	}
	return &vpolygon
}

// isConvex check if polygon either convex or concave
func (vpolygon *VirtualPolygon) isConvex() bool {
	// time complexity: O(n)
	n := len(vpolygon.Coordinates)
	if n < 3 {
		// Well, this is not that strange if polygon have been prepared wrongly
		return false
	}
	previousCrossProduct := 0
	currentCrossProduct := 0
	for i := range vpolygon.Coordinates {
		currentCrossProduct = crossProduct(vpolygon.Coordinates[i], vpolygon.Coordinates[(i+1)%n], vpolygon.Coordinates[(i+2)%n])
		if currentCrossProduct != 0 {
			if currentCrossProduct*previousCrossProduct < 0 {
				return false
			} else {
				previousCrossProduct = currentCrossProduct
			}
		}
	}
	return true
}

// crossProduct Cross product of two vectors
func crossProduct(a image.Point, b image.Point, c image.Point) int {
	// direction of vector b.x -> a.x
	x1 := b.X - a.X
	// direction of vector b.y -> a.y
	y1 := b.Y - a.Y
	// direction of vector c.x -> a.x
	x2 := c.X - a.X
	// direction of vector c.y -> a.y
	y2 := c.Y - a.Y
	return x1*y2 + y1*x2
}

// Scale Scales down (so scale factor can be > 1.0 ) virtual polygon
// (scaleX, scaleY) - How to scale source (x1,y1) and (x2,y2) coordinates
// Important notice:
// 1. Source coordinates won't be modified
// 2. Source coordinates would be used for scaling. So you can't scale polygon multiple times
func (vpolygon *VirtualPolygon) Scale(scaleX, scaleY float64) {
	for i := range vpolygon.Coordinates {
		vpolygon.Coordinates[i].X = int(math.Round(float64(vpolygon.Coordinates[i].X) / scaleX))
		vpolygon.Coordinates[i].Y = int(math.Round(float64(vpolygon.Coordinates[i].Y) / scaleY))
	}
}

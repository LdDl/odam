package odam

import (
	"image"
	"image/color"
)

// POLYGON_TYPE Alias to int
type POLYGON_TYPE int

const (
	// CONVEX_POLYGON See ref. https://en.wikipedia.org/wiki/Convex_polygon
	CONVEX_POLYGON = POLYGON_TYPE(iota + 1)
	// NOT_CONVEX_POLYGON See ref. https://en.wikipedia.org/wiki/Concave_polygon
	NOT_CONVEX_POLYGON
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

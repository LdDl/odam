package odam

// POLYGON_TYPE Alias to int
type POLYGON_TYPE int

const (
	// CONVEX_POLYGON See ref. https://en.wikipedia.org/wiki/Convex_polygon
	CONVEX_POLYGON = POLYGON_TYPE(iota + 1)
	// NOT_CONVEX_POLYGON See ref. https://en.wikipedia.org/wiki/Concave_polygon
	NOT_CONVEX_POLYGON
)
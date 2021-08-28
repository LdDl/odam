package odam

import (
	"image"
	"image/color"
	"math"

	blob "github.com/LdDl/gocv-blob/v2/blob"
)

// VIRTUAL_POLYGON_TYPE Alias to int
type VIRTUAL_POLYGON_TYPE int

const (
	// CONVEX_POLYGON See ref. https://en.wikipedia.org/wiki/Convex_polygon
	CONVEX_POLYGON = VIRTUAL_POLYGON_TYPE(iota + 1)
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
	PolygonType VIRTUAL_POLYGON_TYPE `json:"-"`
}

// Constructor for VirtualPolygon
// (x1, y1) - Left
// (x2, y2) - Right
func NewVirtualPolygon(pairs ...image.Point) *VirtualPolygon {
	vpolygon := VirtualPolygon{
		Coordinates:       make([]image.Point, len(pairs)),
		SourceCoordinates: make([]image.Point, len(pairs)),
	}
	for i := range pairs {
		vpolygon.Coordinates[i] = image.Point{X: pairs[i].X, Y: pairs[i].Y}
		vpolygon.SourceCoordinates[i] = image.Point{X: pairs[i].X, Y: pairs[i].Y}
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
	return x1*y2 - y1*x2
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

// BlobEntered Checks if an object has entered the polygon
// Let's clarify for future questions: we are assuming the object is represented by a center, not a bounding box
// So object has entered polygon when its center had entered polygon too
func (vpolygon *VirtualPolygon) BlobEntered(b blob.Blobie) bool {
	track := b.GetTrack()
	n := len(track)
	if n < 2 {
		// Blob can't have one coordinates pair in track
		return false
	}
	lastPosition := track[len(track)-1]
	secondLastPosition := track[len(track)-2]
	// If P(xN-1,yN-1) is not inside of polygon and P(xN,yN) is inside of polygon then object has entered the polygon
	if !vpolygon.ContainsPoint(secondLastPosition) && vpolygon.ContainsPoint(lastPosition) {
		return true
	}
	return false
}

// BlobLeft Checks if an object has left the polygon
// Let's clarify for future questions: we are assuming the object is represented by a center, not a bounding box
// So object has left polygon when its center had left polygon too
func (vpolygon *VirtualPolygon) BlobLeft(b blob.Blobie) bool {
	track := b.GetTrack()
	n := len(track)
	if n < 2 {
		// Blob can't have one coordinates pair in track
		return false
	}
	lastPosition := track[len(track)-1]
	secondLastPosition := track[len(track)-2]
	// If P(xN-1,yN-1) is inside of polygon and P(xN,yN) is not inside of polygon then object has left the polygon
	if vpolygon.ContainsPoint(secondLastPosition) && !vpolygon.ContainsPoint(lastPosition) {
		return true
	}
	return false
}

// ContainsBlob Checks if polygon contains the given object
// Let's clarify for future questions: we are assuming the object is represented by a center, not a bounding box
// So object is inside of polygon when its center is inside of polygon too
func (vpolygon *VirtualPolygon) ContainsBlob(b blob.Blobie) bool {
	return vpolygon.ContainsPoint(b.GetCenter())
}

// ContainsPoint Checks if polygon contains the given point
func (vpolygon *VirtualPolygon) ContainsPoint(p image.Point) bool {
	if len(vpolygon.Coordinates) < 3 {
		// Well, this is not that strange if polygon have been prepared wrongly
		return false
	}
	switch vpolygon.PolygonType {
	case CONVEX_POLYGON:
		return vpolygon.convexContainsPoint(p)
	case CONCAVE_POLYGON:
		return vpolygon.concaveContainsPoint(p)
	default:
		// This actually should not happen
		// Is this really needed to have error returning in this function?
		break
	}
	return false
}

// convexContainsPoint Checks if CONVEX polygon contains the given point
// Heavily inspired by this: https://github.com/LdDl/gocv-blob/blob/master/v2/blob/line_cross.go#L5
func (vpolygon *VirtualPolygon) convexContainsPoint(p image.Point) bool {
	n := len(vpolygon.Coordinates)
	extremePoint := image.Point{
		X: 99999, // @todo: math.maxInt could lead to overflow obviously. Need good workaround. PRs are welcome
		Y: p.Y,
	}
	intersectionsCnt := 0
	previous := 0
	for {
		current := (previous + 1) % n
		// Check if the segment from given point P to extreme point intersects with the segment from polygon point on previous interation to  polygon point on current interation
		if isIntersects(
			vpolygon.Coordinates[previous].X, vpolygon.Coordinates[previous].Y,
			vpolygon.Coordinates[current].X, vpolygon.Coordinates[current].Y,
			p.X, p.Y,
			extremePoint.X, extremePoint.Y,
		) {
			orientation := getOrientation(
				vpolygon.Coordinates[previous].X, vpolygon.Coordinates[previous].Y,
				p.X, p.Y,
				vpolygon.Coordinates[current].X, vpolygon.Coordinates[current].Y,
			)
			// If given point P is collinear with segment from polygon point on previous interation to  polygon point on current interation
			if orientation == Collinear {
				// then check if it is on segment
				// 'True' will be returns if it lies on segment. Otherwise 'False' will be returned
				return isOnSegment(vpolygon.Coordinates[previous].X, vpolygon.Coordinates[previous].Y, p.X, p.Y, vpolygon.Coordinates[current].X, vpolygon.Coordinates[current].Y)
			}
			intersectionsCnt++
		}
		previous = current
		if previous == 0 {
			break
		}
	}
	// If ray intersects even number of times then return true
	// Otherwise return false
	if intersectionsCnt%2 == 1 {
		return true
	}
	return false
}

// concaveContainsPoint Checks if CONCAVE polygon contains the given point
func (vpolygon *VirtualPolygon) concaveContainsPoint(p image.Point) bool {
	// @todo
	return false
}

// isOnSegment Checks if point Q lies on segment PR
// Input: three colinear points Q, Q and R
func isOnSegment(Px, Py, Qx, Qy, Rx, Ry int) bool {
	if Qx <= maxInt(Px, Rx) && Qx >= minInt(Px, Rx) && Qy <= maxInt(Py, Ry) && Qy >= minInt(Py, Ry) {
		return true
	}
	return false
}

type PointsOrientation int

const (
	Collinear = iota
	Clockwise
	CounterClockwise
)

// getOrientation Gets orientations of points P -> Q -> R.
// Possible output values: Collinear / Clockwise or CounterClockwise
// Input: points P, Q and R in provided order
func getOrientation(Px, Py, Qx, Qy, Rx, Ry int) PointsOrientation {
	val := (Qy-Py)*(Rx-Qx) - (Qx-Px)*(Ry-Qy)
	if val == 0 {
		return Collinear
	}
	if val > 0 {
		return Clockwise
	}
	return CounterClockwise // if it's neither collinear nor clockwise
}

// isIntersects Checks if segments intersect each other
// Input:
// firstPx, firstPy, firstQx, firstQy === first segment
// secondPx, secondPy, secondQx, secondQy === second segment
/*
Notation
	P1 = (firstPx, firstPy)
	Q1 = (firstQx, firstQy)
	P2 = (secondPx, secondPy)
	Q2 = (secondQx, secondQy)
*/
func isIntersects(firstPx, firstPy, firstQx, firstQy, secondPx, secondPy, secondQx, secondQy int) bool {
	// Find the four orientations needed for general case and special ones
	o1 := getOrientation(firstPx, firstPy, firstQx, firstQy, secondPx, secondPy)
	o2 := getOrientation(firstPx, firstPy, firstQx, firstQy, secondQx, secondQy)
	o3 := getOrientation(secondPx, secondPy, secondQx, secondQy, firstPx, firstPy)
	o4 := getOrientation(secondPx, secondPy, secondQx, secondQy, firstQx, firstQy)

	// General case
	if o1 != o2 && o3 != o4 {
		return true
	}

	/* Special cases */
	// P1, Q1, P2 are colinear and P2 lies on segment P1-Q1
	if o1 == Collinear && isOnSegment(firstPx, firstPy, secondPx, secondPy, firstQx, firstQy) {
		return true
	}
	// P1, Q1 and Q2 are colinear and Q2 lies on segment P1-Q1
	if o2 == Collinear && isOnSegment(firstPx, firstPy, secondQx, secondQy, firstQx, firstQy) {
		return true
	}
	// P2, Q2 and P1 are colinear and P1 lies on segment P2-Q2
	if o3 == Collinear && isOnSegment(secondPx, secondPy, firstPx, firstPy, secondQx, secondQy) {
		return true
	}
	// P2, Q2 and Q1 are colinear and Q1 lies on segment P2-Q2
	if o4 == Collinear && isOnSegment(secondPx, secondPy, firstQx, firstQy, secondQx, secondQy) {
		return true
	}
	// Segments do not intersect
	return false
}

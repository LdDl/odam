package odam

import (
	"image"
	"testing"

	blob "github.com/LdDl/gocv-blob/v2/blob"
)

func TestPolygonType(t *testing.T) {
	vpolygons := []*VirtualPolygon{
		NewVirtualPolygon(
			1,
			image.Point{X: 0, Y: 0},
			image.Point{X: 1, Y: 0},
			image.Point{X: 1, Y: 1},
			image.Point{X: 0, Y: 1},
		),
		NewVirtualPolygon(
			2,
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 3, Y: 3},
			image.Point{X: 0, Y: 1},
		),
		NewVirtualPolygon(
			3,
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
			1,
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 0, Y: 5},
		),
		NewVirtualPolygon(
			2,
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 0, Y: 5},
		),
		NewVirtualPolygon(
			3,
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 5, Y: 0},
		),
		NewVirtualPolygon(
			4,
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 5, Y: 0},
		),
		NewVirtualPolygon(
			5,
			image.Point{X: 0, Y: 0},
			image.Point{X: 5, Y: 5},
			image.Point{X: 5, Y: 0},
		),
		NewVirtualPolygon(
			6,
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
		1,
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
				t.Errorf("Polygon with coordinates %v should contain blob with center at [%d, %d]. Actual answer is %t", vpolygon.Coordinates, center.X, center.Y, answer)
			} else {
				t.Errorf("Polygon with coordinates %v should not contain blob with center at [%d, %d]. Actual answer is %t", vpolygon.Coordinates, center.X, center.Y, answer)
			}
		}
	}
}

func TestPolygonBlobEnter(t *testing.T) {
	simpleATime0 := blob.NewSimpleBlobie(image.Rect(30, 2, 43, 13), nil)
	simpleATime1 := blob.NewSimpleBlobie(image.Rect(28, 8, 41, 18), nil)
	simpleATime2 := blob.NewSimpleBlobie(image.Rect(29, 17, 43, 26), nil)
	allblobiesEnter := blob.NewBlobiesDefaults()
	allblobiesEnter.MatchToExisting([]blob.Blobie{simpleATime0, simpleATime1, simpleATime2})
	vpolygon := NewVirtualPolygon(
		1,
		image.Point{X: 23, Y: 15},
		image.Point{X: 67, Y: 15},
		image.Point{X: 67, Y: 41},
		image.Point{X: 23, Y: 41},
	)
	for _, b := range allblobiesEnter.Objects {
		center := b.GetCenter()
		if !vpolygon.BlobEntered(b) {
			t.Errorf("Blob with center at [%d, %d] should has been entered to polygon with coordinates %v", center.X, center.Y, vpolygon.Coordinates)
		}
	}

	simpleBTime0 := blob.NewSimpleBlobie(image.Rect(38, 32, 52, 39), nil)
	simpleBTime1 := blob.NewSimpleBlobie(image.Rect(40, 35, 53, 42), nil)
	simpleBTime2 := blob.NewSimpleBlobie(image.Rect(42, 42, 56, 50), nil)
	allblobiesLeft := blob.NewBlobiesDefaults()
	allblobiesLeft.MatchToExisting([]blob.Blobie{simpleBTime0, simpleBTime1, simpleBTime2})
	for _, b := range allblobiesLeft.Objects {
		center := b.GetCenter()
		if vpolygon.BlobEntered(b) {
			t.Errorf("Blob with center at [%d, %d] should has NOT been entered to polygon with coordinates %v", center.X, center.Y, vpolygon.Coordinates)
		}
	}

	simpleCTime0 := blob.NewSimpleBlobie(image.Rect(49, 16, 64, 23), nil)
	simpleCTime1 := blob.NewSimpleBlobie(image.Rect(48, 20, 63, 27), nil)
	simpleCTime2 := blob.NewSimpleBlobie(image.Rect(48, 25, 63, 33), nil)
	allblobiesInside := blob.NewBlobiesDefaults()
	allblobiesInside.MatchToExisting([]blob.Blobie{simpleCTime0, simpleCTime1, simpleCTime2})
	for _, b := range allblobiesInside.Objects {
		center := b.GetCenter()
		if vpolygon.BlobEntered(b) {
			t.Errorf("Blob with center at [%d, %d] should has NOT been entered to polygon with coordinates %v", center.X, center.Y, vpolygon.Coordinates)
		}
	}

	simpleDTime0 := blob.NewSimpleBlobie(image.Rect(10, 9, 25, 18), nil)
	simpleDTime1 := blob.NewSimpleBlobie(image.Rect(12, 16, 27, 24), nil)
	simpleDTime2 := blob.NewSimpleBlobie(image.Rect(11, 21, 27, 30), nil)
	allblobiesOutside := blob.NewBlobiesDefaults()
	allblobiesOutside.MatchToExisting([]blob.Blobie{simpleDTime0, simpleDTime1, simpleDTime2})
	for _, b := range allblobiesOutside.Objects {
		center := b.GetCenter()
		if vpolygon.BlobEntered(b) {
			t.Errorf("Blob with center at [%d, %d] should has NOT been entered to polygon with coordinates %v", center.X, center.Y, vpolygon.Coordinates)
		}
	}
}

func TestPolygonBlobLeft(t *testing.T) {
	/* Using same data as in TestPolygonBlobEnter() */

	simpleATime0 := blob.NewSimpleBlobie(image.Rect(30, 2, 43, 13), nil)
	simpleATime1 := blob.NewSimpleBlobie(image.Rect(28, 8, 41, 18), nil)
	simpleATime2 := blob.NewSimpleBlobie(image.Rect(29, 17, 43, 26), nil)
	allblobiesEnter := blob.NewBlobiesDefaults()
	allblobiesEnter.MatchToExisting([]blob.Blobie{simpleATime0, simpleATime1, simpleATime2})
	vpolygon := NewVirtualPolygon(
		1,
		image.Point{X: 23, Y: 15},
		image.Point{X: 67, Y: 15},
		image.Point{X: 67, Y: 41},
		image.Point{X: 23, Y: 41},
	)
	for _, b := range allblobiesEnter.Objects {
		center := b.GetCenter()
		if vpolygon.BlobLeft(b) {
			t.Errorf("Blob with center at [%d, %d] should has NOT been left polygon with coordinates %v", center.X, center.Y, vpolygon.Coordinates)
		}
	}

	simpleBTime0 := blob.NewSimpleBlobie(image.Rect(38, 32, 52, 39), nil)
	simpleBTime1 := blob.NewSimpleBlobie(image.Rect(40, 35, 53, 42), nil)
	simpleBTime2 := blob.NewSimpleBlobie(image.Rect(42, 42, 56, 50), nil)
	allblobiesLeft := blob.NewBlobiesDefaults()
	allblobiesLeft.MatchToExisting([]blob.Blobie{simpleBTime0, simpleBTime1, simpleBTime2})
	for _, b := range allblobiesLeft.Objects {
		center := b.GetCenter()
		if !vpolygon.BlobLeft(b) {
			t.Errorf("Blob with center at [%d, %d] should has been left polygon with coordinates %v", center.X, center.Y, vpolygon.Coordinates)
		}
	}

	simpleCTime0 := blob.NewSimpleBlobie(image.Rect(49, 16, 64, 23), nil)
	simpleCTime1 := blob.NewSimpleBlobie(image.Rect(48, 20, 63, 27), nil)
	simpleCTime2 := blob.NewSimpleBlobie(image.Rect(48, 25, 63, 33), nil)
	allblobiesInside := blob.NewBlobiesDefaults()
	allblobiesInside.MatchToExisting([]blob.Blobie{simpleCTime0, simpleCTime1, simpleCTime2})
	for _, b := range allblobiesInside.Objects {
		center := b.GetCenter()
		if vpolygon.BlobLeft(b) {
			t.Errorf("Blob with center at [%d, %d] should has NOT been left polygon with coordinates %v", center.X, center.Y, vpolygon.Coordinates)
		}
	}

	simpleDTime0 := blob.NewSimpleBlobie(image.Rect(10, 9, 25, 18), nil)
	simpleDTime1 := blob.NewSimpleBlobie(image.Rect(12, 16, 27, 24), nil)
	simpleDTime2 := blob.NewSimpleBlobie(image.Rect(11, 21, 27, 30), nil)
	allblobiesOutside := blob.NewBlobiesDefaults()
	allblobiesOutside.MatchToExisting([]blob.Blobie{simpleDTime0, simpleDTime1, simpleDTime2})
	for _, b := range allblobiesOutside.Objects {
		center := b.GetCenter()
		if vpolygon.BlobLeft(b) {
			t.Errorf("Blob with center at [%d, %d] should has NOT been left polygon with coordinates %v", center.X, center.Y, vpolygon.Coordinates)
		}
	}
}

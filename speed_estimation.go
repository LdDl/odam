package odam

import (
	"image"
	"math"
	"time"

	"gocv.io/x/gocv"
)

const (
	earthRaidusKm = 6371 // radius of the earth in kilometers.
)

// GetPerspectiveTransformer Initializates gocv.Point2f for GIS conversion purposes
func GetPerspectiveTransformer(srcPoints, dstPoints []gocv.Point2f) (*gocv.Mat, func(gocv.Point2f) gocv.Point2f) {
	src := gocv.NewPoint2fVectorFromPoints(srcPoints)
	trgt := gocv.NewPoint2fVectorFromPoints(dstPoints)
	transformMat := gocv.GetPerspectiveTransform2f(src, trgt)
	return &transformMat, func(src gocv.Point2f) gocv.Point2f {
		pmat := gocv.NewMatWithSize(3, 1, gocv.MatTypeCV64F)
		pmat.SetDoubleAt(0, 0, float64(src.X))
		pmat.SetDoubleAt(1, 0, float64(src.Y))
		pmat.SetDoubleAt(2, 0, 1.0)
		answ := transformMat.MultiplyMatrix(pmat)
		pmat.Close() // Free memory
		scale := answ.GetDoubleAt(2, 0)
		xattr := answ.GetDoubleAt(0, 0)
		yattr := answ.GetDoubleAt(1, 0)
		answ.Close() // Free memory
		return gocv.Point2f{X: float32(xattr / scale), Y: float32(yattr / scale)}
	}
}

// EstimateSpeed Estimates speed approximately
func EstimateSpeed(firstPoint, lastPoint gocv.Point2f, start, end time.Time, perspectiveTransformer func(gocv.Point2f) gocv.Point2f) float32 {
	fpreal := perspectiveTransformer(firstPoint)
	lpreal := perspectiveTransformer(lastPoint)
	return Haversine(fpreal, lpreal) / float32(end.Sub(start).Hours())
}

// Haversine Calculates great circle distance between two points
// https://en.wikipedia.org/wiki/Great-circle_distance#:~:text=The%20great%2Dcircle%20distance%2C%20orthodromic,line%20through%20the%20sphere's%20interior).
func Haversine(src, dst gocv.Point2f) float32 {
	lat1 := degreesToRadians(src.Y)
	lon1 := degreesToRadians(src.X)
	lat2 := degreesToRadians(dst.Y)
	lon2 := degreesToRadians(dst.X)

	diffLat := lat2 - lat1
	diffLon := lon2 - lon1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	km := c * earthRaidusKm

	return float32(km)
}

// STDPointToGoCVPoint2F Convertes image.Point to gocv.Point2f
func STDPointToGoCVPoint2F(p image.Point) gocv.Point2f {
	return gocv.Point2f{X: float32(p.X), Y: float32(p.Y)}
}

func degreesToRadians(d float32) float64 {
	return float64(d) * math.Pi / 180
}

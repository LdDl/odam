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

func GetPerspectiveTransformer(srcPoints, dstPoints []gocv.Point2f) func(gocv.Point2f) gocv.Point2f {
	transformMat := gocv.GetPerspectiveTransform2f(srcPoints, dstPoints)
	return func(src gocv.Point2f) gocv.Point2f {
		pmat := gocv.NewMatWithSize(3, 1, gocv.MatTypeCV64F)
		pmat.SetDoubleAt(0, 0, float64(src.X))
		pmat.SetDoubleAt(1, 0, float64(src.Y))
		pmat.SetDoubleAt(2, 0, 1.0)
		answ := transformMat.MultiplyMatrix(pmat)
		scale := answ.GetDoubleAt(2, 0)
		return gocv.Point2f{X: float32(answ.GetDoubleAt(0, 0) / scale), Y: float32(answ.GetDoubleAt(1, 0) / scale)}
	}
}

func EstimateSpeed(firstPoint, lastPoint gocv.Point2f, start, end time.Time, perspectiveTransformer func(gocv.Point2f) gocv.Point2f) float32 {
	fpreal := perspectiveTransformer(firstPoint)
	lpreal := perspectiveTransformer(lastPoint)
	return Haversine(fpreal, lpreal) / float32(end.Sub(start).Hours())
}

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

func degreesToRadians(d float32) float64 {
	return float64(d) * math.Pi / 180
}

func STDPointToGoCVPoint2F(p image.Point) gocv.Point2f {
	return gocv.Point2f{X: float32(p.X), Y: float32(p.Y)}
}

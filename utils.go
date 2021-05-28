package odam

import (
	"image"
	"math"
)

// Round Rounds float64 to int
func Round(v float64) int {
	if v >= 0 {
		return int(math.Floor(v + 0.5))
	}
	return int(math.Ceil(v - 0.5))
}

// FixRectForOpenCV Corrects rectangle's bounds for provided max-widtht and max-height
// Helps to avoid BBox error assertion
func FixRectForOpenCV(r *image.Rectangle, maxCols, maxRows int) {
	if r.Min.X <= 0 {
		r.Min.X = 0
	}
	if r.Min.Y < 0 {
		r.Min.Y = 0
	}
	if r.Max.X >= maxCols {
		r.Max.X = maxCols - 1
	}
	if r.Max.Y >= maxRows {
		r.Max.Y = maxRows - 1
	}
}

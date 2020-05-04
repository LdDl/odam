package odam

import (
	"image"
	"math"
)

// Round - Round float64 to int
func Round(v float64) int {
	if v >= 0 {
		return int(math.Floor(v + 0.5))
	}
	return int(math.Ceil(v - 0.5))
}

func FixRectForOpenCV(r *image.Rectangle, cols, rows int) {
	if r.Min.X <= 0 {
		r.Min.X = 0
	}
	if r.Min.Y < 0 {
		r.Min.Y = 0
	}
	if r.Max.X >= cols {
		r.Max.X = cols - 1
	}
	if r.Max.Y >= rows {
		r.Max.Y = rows - 1
	}
}

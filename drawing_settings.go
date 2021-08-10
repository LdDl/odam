package odam

import (
	"image/color"
	"strings"

	blob "github.com/LdDl/gocv-blob/v2/blob"
	"gocv.io/x/gocv"
)

// Wrap blob.DrawOptions
type DrawOptions struct {
	*blob.DrawOptions
	DisplayObjectID bool
}

// PrepareDrawingOptions Prepares drawing options for blob library
func (classInfo *ClassesSettings) PrepareDrawingOptions() *DrawOptions {
	drOpts := &blob.DrawOptions{}
	if classInfo.DrawingSettings == nil {
		// If drawing settings are not defined for certain class in JSON, then apply default settings for it
		drOpts = blob.NewDrawOptionsDefault()
	} else {
		bboxOpts := blob.DrawBBoxOptions{
			Color: color.RGBA{
				classInfo.DrawingSettings.BBoxSettings.RGBA[0],
				classInfo.DrawingSettings.BBoxSettings.RGBA[1],
				classInfo.DrawingSettings.BBoxSettings.RGBA[2],
				classInfo.DrawingSettings.BBoxSettings.RGBA[3],
			},
			Thickness: classInfo.DrawingSettings.BBoxSettings.Thickness,
		}
		cenOpts := blob.DrawCentroidOptions{
			Color: color.RGBA{
				classInfo.DrawingSettings.CentroidSettings.RGBA[0],
				classInfo.DrawingSettings.CentroidSettings.RGBA[1],
				classInfo.DrawingSettings.CentroidSettings.RGBA[2],
				classInfo.DrawingSettings.CentroidSettings.RGBA[3],
			},
			Radius:    classInfo.DrawingSettings.CentroidSettings.Radius,
			Thickness: classInfo.DrawingSettings.CentroidSettings.Thickness,
		}
		textOpts := blob.DrawTextOptions{
			Color: color.RGBA{
				classInfo.DrawingSettings.TextSettings.RGBA[0],
				classInfo.DrawingSettings.TextSettings.RGBA[1],
				classInfo.DrawingSettings.TextSettings.RGBA[2],
				classInfo.DrawingSettings.TextSettings.RGBA[3],
			},
			Scale:     classInfo.DrawingSettings.TextSettings.Scale,
			Thickness: classInfo.DrawingSettings.TextSettings.Thickness,
		}
		switch strings.ToLower(classInfo.DrawingSettings.TextSettings.Font) {
		case "hershey_simplex":
			textOpts.Font = gocv.FontHersheySimplex
			break
		case "hershey_plain":
			textOpts.Font = gocv.FontHersheyPlain
			break
		case "hershey_duplex":
			textOpts.Font = gocv.FontHersheyDuplex
			break
		case "hershey_complex":
			textOpts.Font = gocv.FontHersheyComplex
			break
		case "hershey_triplex":
			textOpts.Font = gocv.FontHersheyTriplex
			break
		case "hershey_complex_small":
			textOpts.Font = gocv.FontHersheyComplexSmall
			break
		case "hershey_script_simplex":
			textOpts.Font = gocv.FontHersheyScriptSimplex
			break
		case "hershey_script_complex":
			textOpts.Font = gocv.FontHersheyScriptComplex
			break
		case "italic":
			textOpts.Font = gocv.FontItalic
			break
		default:
			textOpts.Font = gocv.FontHersheyPlain
			break
		}
		drOpts = blob.NewDrawOptions(bboxOpts, cenOpts, textOpts)
	}

	return &DrawOptions{
		drOpts,
		classInfo.DrawingSettings.DisplayObjectID,
	}
}

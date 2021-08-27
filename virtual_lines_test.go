package odam

import (
	"testing"
)

func TestLineType(t *testing.T) {
	vlines := []*VirtualLine{
		NewVirtualLine(10, 70, 120, 70),
		NewVirtualLine(10, 70, 120, 75),
	}
	correctLineTypes := []VIRTUAL_LINE_TYPE{
		HORIZONTAL_LINE,
		OBLIQUE_LINE,
	}
	for i, vline := range vlines {
		if vline.LineType != correctLineTypes[i] {
			t.Errorf("Line (x1,y1) = (%d, %d) and (x1,y1) = (%d, %d) should be of type '%d' but got %d",
				vline.LeftPT.X, vline.LeftPT.Y,
				vline.RightPT.X, vline.RightPT.Y,
				correctLineTypes[i],
				vline.LineType,
			)
		}
	}
}

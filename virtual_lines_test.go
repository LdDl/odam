package odam

import (
	"image"
	"testing"

	blob "github.com/LdDl/gocv-blob/v2/blob"
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
			t.Errorf("Line (x1,y1) = (%d, %d) and (x2,y2) = (%d, %d) should be of type '%d' but got %d",
				vline.LeftPT.X, vline.LeftPT.Y,
				vline.RightPT.X, vline.RightPT.Y,
				correctLineTypes[i],
				vline.LineType,
			)
		}
	}
}

// This test is insipred by https://github.com/LdDl/gocv-blob/blob/master/v2/blob/line_cross_test.go#L37
func TestLineCross(t *testing.T) {
	// true - "object is moving to us"
	// false - "object is moving from us"
	constDirections := []bool{
		true,
		true,
		true,
	}
	correctAnswers := []bool{
		true,
		true,
		false,
	}
	vlines := []*VirtualLine{
		NewVirtualLine(4, 35, 73, 35), // Horizontal
		NewVirtualLine(4, 35, 71, 31), // Oblique
		NewVirtualLine(4, 35, 71, 45), // Oblique
	}

	for i, vline := range vlines {
		vline.Direction = constDirections[i]
		/* Creating same set of blobies every time because b.IsCrossedTheLine(...) and b.(...) methods modifies IsCrossedTheObliqueLine private field of blob (blob.crossedLine) */
		allblobies := blob.NewBlobiesDefaults()
		simpleB_time0 := blob.NewSimpleBlobie(image.Rect(26, 8, 44, 18), nil)
		simpleB_time1 := blob.NewSimpleBlobie(image.Rect(26, 20, 44, 30), nil)
		simpleB_time2 := blob.NewSimpleBlobie(image.Rect(26, 32, 44, 42), nil)
		allblobies.MatchToExisting([]blob.Blobie{simpleB_time0, simpleB_time1, simpleB_time2})
		for _, b := range allblobies.Objects {
			if vline.IsBlobCrossedLine(b) != correctAnswers[i] {
				track := b.GetTrack()
				prevPos := track[len(track)-2]
				currentPos := track[len(track)-1]
				if vline.IsBlobCrossedLine(b) != true {
					t.Errorf("#%d Line (x1,y1) = (%d, %d) and (x2,y2) = (%d, %d) of type '%d' should be crossed by (xprev,yprev) = (%d %d) and (xcurrent,ycurrent) = (%d %d) with direction = %t",
						i+1,
						vline.LeftPT.X, vline.LeftPT.Y,
						vline.RightPT.X, vline.RightPT.Y,
						vline.LineType,
						prevPos.X, prevPos.Y,
						currentPos.X, currentPos.Y,
						vline.Direction,
					)
				} else {
					t.Errorf("#%d Line (x1,y1) = (%d, %d) and (x2,y2) = (%d, %d) of type '%d' should NOT be crossed by (xprev,yprev) = (%d %d) and (xcurrent,ycurrent) = (%d %d) with direction = %t",
						i+1,
						vline.LeftPT.X, vline.LeftPT.Y,
						vline.RightPT.X, vline.RightPT.Y,
						vline.LineType,
						prevPos.X, prevPos.Y,
						currentPos.X, currentPos.Y,
						vline.Direction,
					)
				}
			}

		}
	}

	correctAnswersOpposite := []bool{
		false,
		false,
		false,
	}
	// Opossite direction
	for i, vline := range vlines {
		vline.Direction = !constDirections[i]
		/* Creating same set of blobies every time because b.IsCrossedTheLine(...) and b.IsCrossedTheObliqueLine(...) methods modifies private field of blob (blob.crossedLine) */
		allblobies := blob.NewBlobiesDefaults()
		simpleB_time0 := blob.NewSimpleBlobie(image.Rect(26, 8, 44, 18), nil)
		simpleB_time1 := blob.NewSimpleBlobie(image.Rect(26, 20, 44, 30), nil)
		simpleB_time2 := blob.NewSimpleBlobie(image.Rect(26, 32, 44, 42), nil)
		allblobies.MatchToExisting([]blob.Blobie{simpleB_time0, simpleB_time1, simpleB_time2})
		for _, b := range allblobies.Objects {
			if vline.IsBlobCrossedLine(b) != correctAnswersOpposite[i] {
				track := b.GetTrack()
				prevPos := track[len(track)-2]
				currentPos := track[len(track)-1]
				if vline.IsBlobCrossedLine(b) != true {
					t.Errorf("#%d Line (x1,y1) = (%d, %d) and (x2,y2) = (%d, %d) of type '%d' should be crossed by (xprev,yprev) = (%d %d) and (xcurrent,ycurrent) = (%d %d) with direction = %t",
						i+1,
						vline.LeftPT.X, vline.LeftPT.Y,
						vline.RightPT.X, vline.RightPT.Y,
						vline.LineType,
						prevPos.X, prevPos.Y,
						currentPos.X, currentPos.Y,
						vline.Direction,
					)
				} else {
					t.Errorf("#%d Line (x1,y1) = (%d, %d) and (x2,y2) = (%d, %d) of type '%d' should NOT be crossed by (xprev,yprev) = (%d %d) and (xcurrent,ycurrent) = (%d %d) with direction = %t",
						i+1,
						vline.LeftPT.X, vline.LeftPT.Y,
						vline.RightPT.X, vline.RightPT.Y,
						vline.LineType,
						prevPos.X, prevPos.Y,
						currentPos.X, currentPos.Y,
						vline.Direction,
					)
				}
			}

		}
	}
}

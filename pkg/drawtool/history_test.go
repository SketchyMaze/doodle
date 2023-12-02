package drawtool

import (
	"testing"

	"git.kirsle.net/go/render"
)

func TestHistory(t *testing.T) {
	// Test assertion helpers.
	shouldBool := func(note string, expect, actual bool) {
		if actual != expect {
			t.Errorf(
				"Unexpected boolean result (%s)\n"+
					"Expected: %+v\n"+
					"     Got: %+v",
				note,
				expect,
				actual,
			)
		}
	}
	shouldInt := func(note string, expect, actual int) {
		if actual != expect {
			t.Errorf(
				"Unexpected integer result (%s)\n"+
					"Expected: %+v\n"+
					"     Got: %+v",
				note,
				expect,
				actual,
			)
		}
	}
	shouldPoint := func(note string, expect render.Point, actual *Stroke) {
		if actual == nil {
			t.Errorf("Missing history stroke for shouldPoint(%s)", note)
			return
		}
		if actual.PointA != expect {
			t.Errorf(
				"Unexpected point result (%s)\n"+
					"Expected: %+v\n"+
					"     Got: %+v",
				note,
				expect,
				actual.PointA,
			)
		}
	}

	var H = NewHistory(10)

	// Add and remove and re-add the first element.
	H.AddStroke(&Stroke{
		PointA: render.NewPoint(999, 999),
	})
	shouldInt("first element", 1, H.Size())
	shouldBool("can undo first element", true, H.Undo())
	shouldBool("latest should be null", true, H.Latest() == nil)

	H = NewHistory(10)

	shouldBool("can't Undo with fresh history", false, H.Undo())
	shouldInt("size should be zero", 0, H.Size())

	H.AddStroke(&Stroke{
		PointA: render.NewPoint(1, 1),
	})

	shouldInt("after first stroke", 1, H.Size())
	shouldPoint("head is the newest point", render.NewPoint(1, 1), H.Latest())

	H.AddStroke(&Stroke{
		PointA: render.NewPoint(2, 2),
	})

	shouldInt("after second stroke", 2, H.Size())
	shouldPoint("head is the newest point", render.NewPoint(2, 2), H.Latest())

	// Undo.
	shouldBool("undo second stroke", true, H.Undo())
	shouldInt("after undo the future stroke is still part of the size", 2, H.Size())
	shouldPoint("after undo, the newest point", render.NewPoint(1, 1), H.Latest())

	// Redo.
	shouldBool("redo second stroke", true, H.Redo())
	shouldInt("after redo second stroke, size is still the same", 2, H.Size())
	shouldPoint("after redo, the newest point", render.NewPoint(2, 2), H.Latest())

	// Another redo must fail.
	shouldBool("redo when there is nothing to redo", false, H.Redo())

	// Add a few more points.
	for i := 3; i <= 6; i++ {
		H.AddStroke(&Stroke{
			PointA: render.NewPoint(i, i),
		})
	}
	shouldInt("after adding more strokes", 6, H.Size())
	shouldPoint("last point added", render.NewPoint(6, 6), H.Latest())

	// Undo a few times.
	shouldBool("undo^1", true, H.Undo())
	shouldBool("undo^2", true, H.Undo())
	shouldBool("undo^3", true, H.Undo())
	shouldInt("after a few undos, the size still contains future history", 6, H.Size())

	// A new stroke invalidates the future history.
	H.AddStroke(&Stroke{
		PointA: render.NewPoint(7, 7),
	})
	shouldInt("after new history, size is recapped to tail", 4, H.Size())
	shouldBool("can't Redo after new point added", false, H.Redo())

	// Overflow past our history size to test rollover.
	for i := 8; i <= 16; i++ {
		H.AddStroke(&Stroke{
			PointA: render.NewPoint(i, i),
		})
	}
	shouldInt("after tons of new history, size is capped out", 10, H.Size())
	shouldPoint("after overflow, latest point", render.NewPoint(16, 16), H.Latest())
	shouldPoint("after overflow, first point", render.NewPoint(7, 7), H.Oldest())

	// Undo back to beginning.
	for i := 0; i < H.Size(); i++ {
		shouldBool("bulk undo to beginning", true, H.Undo())
	}
	shouldBool("after bulk undo, tail", true, H.Latest() == nil)
	shouldBool("can't undo further", false, H.Undo())
}

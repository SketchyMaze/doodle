package uix

import (
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/shmem"
)

// canvas_strokes.go: functions related to drawtool.Stroke and the Canvas.

// AddStroke installs a new Stroke to be superimposed over drawing data
// in the canvas.
//
// The stroke is added to the canvas's map by its ID so it can be removed later.
// The stroke must have a non-zero ID value set or this function will panic.
// drawtool.NewStroke() creates an initialized Stroke object to use here.
func (w *Canvas) AddStroke(stroke *drawtool.Stroke) {
	if stroke.ID == 0 {
		panic("Canvas.AddStroke: the Stroke is missing an ID; was it initialized properly?")
	}

	w.strokes[stroke.ID] = stroke
}

// RemoveStroke uninstalls a Stroke from the canvas using its ID.
//
// Returns true if the stroke existed to begin with, false if not.
func (w *Canvas) RemoveStroke(stroke *drawtool.Stroke) bool {
	if _, ok := w.strokes[stroke.ID]; ok {
		delete(w.strokes, stroke.ID)
		return true
	}
	return false
}

// UndoStroke rolls back the level's UndoHistory and deletes the pixels last
// added to the level. Returns false and emits a warning to the log if the
// canvas has no level loaded properly.
func (w *Canvas) UndoStroke() bool {
	if w.level == nil {
		log.Error("Canvas.UndoStroke: no Level currently available to the canvas")
		return false
	}

	latest := w.level.UndoHistory.Latest()
	if latest != nil {
		for point := range latest.IterPoints() {
			w.chunks.Delete(point)
		}
	}
	return w.level.UndoHistory.Undo()
}

// RedoStroke rolls the level's UndoHistory forwards again and replays the
// recently undone changes.
func (w *Canvas) RedoStroke() bool {
	if w.level == nil {
		log.Error("Canvas.UndoStroke: no Level currently available to the canvas")
		return false
	}

	ok := w.level.UndoHistory.Redo()
	if !ok {
		return false
	}

	latest := w.level.UndoHistory.Latest()

	// We stored the ActiveSwatch on this stroke as we drew it. Recover it
	// and place the pixels back down.
	if swatch, ok := latest.ExtraData.(*level.Swatch); ok {
		for point := range latest.IterPoints() {
			w.chunks.Set(point, swatch)
		}
		return true
	}

	log.Error("Canvas.UndoStroke: undo was successful but no Swatch was stored on the Stroke.ExtraData!")

	return ok
}

// presentStrokes is called as part of Present() and draws the strokes whose
// pixels are currently visible within the viewport.
func (w *Canvas) presentStrokes(e render.Engine) {
	// Turn stroke map into a list.
	var strokes []*drawtool.Stroke
	for _, stroke := range w.strokes {
		strokes = append(strokes, stroke)
	}
	w.drawStrokes(e, strokes)

	// Dynamic actor links visible in the ActorTool and LinkTool.
	if w.Tool == drawtool.ActorTool || w.Tool == drawtool.LinkTool {
		w.presentActorLinks(e)
	}
}

// presentActorLinks draws strokes connecting actors together by their links.
// TODO: the strokes are computed dynamically every tick in here, might be a
// way to better optimize later.
func (w *Canvas) presentActorLinks(e render.Engine) {
	var (
		strokes  = []*drawtool.Stroke{}
		actorMap = map[string]*Actor{}
	)

	// Loop over actors and collect linked ones into the map.
	for _, actor := range w.actors {
		if len(actor.Actor.Links) > 0 {
			actorMap[actor.ID()] = actor
		}
	}

	// If no links, stop.
	if len(actorMap) == 0 {
		return
	}

	// The glow colored line. Huge hacky block of code but makes for some
	// basic visualization for now.
	var color = balance.LinkLineColor
	var lightenStep = float64(balance.LinkLighten) / 16
	var step = shmem.Tick % balance.LinkAnimSpeed
	if step < 32 {
		for i := uint64(0); i < step; i++ {
			color = color.Lighten(int(lightenStep))
		}
		if step > 16 {
			for i := uint64(0); i < step-16; i++ {
				color = color.Darken(int(lightenStep))
			}
		}
	}

	// Loop over the linked actors and draw stroke lines.
	for _, actor := range actorMap {
		for _, linkID := range actor.Actor.Links {
			if _, ok := actorMap[linkID]; !ok {
				continue
			}

			var (
				aP = actor.Position()
				aS = actor.Size()
				bP = actorMap[linkID].Position()
				bS = actorMap[linkID].Size()
			)

			// Draw a line connecting the centers of each actor together.
			stroke := drawtool.NewStroke(drawtool.Line, color)
			stroke.PointA = render.Point{
				X: aP.X + (aS.W / 2),
				Y: aP.Y + (aS.H / 2),
			}
			stroke.PointB = render.Point{
				X: bP.X + (bS.W / 2),
				Y: bP.Y + (bS.H / 2),
			}

			strokes = append(strokes, stroke)

			// Make it double thick.
			double := stroke.Copy()
			double.PointA = render.NewPoint(stroke.PointA.X, stroke.PointA.Y+1)
			double.PointB = render.NewPoint(stroke.PointB.X, stroke.PointB.Y+1)
			strokes = append(strokes, double)
			double = stroke.Copy()
			double.PointA = render.NewPoint(stroke.PointA.X+1, stroke.PointA.Y)
			double.PointB = render.NewPoint(stroke.PointB.X+1, stroke.PointB.Y)
			strokes = append(strokes, double)
		}
	}

	w.drawStrokes(e, strokes)
}

// drawStrokes is the common base function behind presentStrokes and
// presentActorLinks to actually draw the lines to the canvas.
func (w *Canvas) drawStrokes(e render.Engine, strokes []*drawtool.Stroke) {
	var (
		P  = ui.AbsolutePosition(w) // w.Point()    // Canvas point in UI
		VP = w.ViewportRelative()   // Canvas scroll viewport
	)

	for _, stroke := range strokes {
		// If none of this stroke is in our viewport, don't waste time
		// looping through it.
		if stroke.Shape == drawtool.Freehand {
			if len(stroke.Points) >= 2 {
				if !stroke.Points[0].Inside(VP) && !stroke.Points[len(stroke.Points)-1].Inside(VP) {
					continue
				}
			}
		} else {
			// TODO: a very long line that starts and ends outside the viewport
			// but passes thru it would disappear when both ends are out of
			// view.
			if !stroke.PointA.Inside(VP) && !stroke.PointB.Inside(VP) {
				continue
			}
		}

		// Iter the points and draw what's visible.
		for point := range stroke.IterPoints() {
			if !point.Inside(VP) {
				continue
			}

			dest := render.Point{
				X: P.X + w.Scroll.X + w.BoxThickness(1) + point.X,
				Y: P.Y + w.Scroll.Y + w.BoxThickness(1) + point.Y,
			}

			if balance.DebugCanvasStrokeColor != render.Invisible {
				e.DrawPoint(balance.DebugCanvasStrokeColor, dest)
			} else {
				e.DrawPoint(stroke.Color, dest)
			}
		}
	}
}
package doodle

// XXX REFACTOR XXX
// This function only uses EditorUI and not Doodle and is a candidate for
// refactor into a subpackage if EditorUI itself can ever be decoupled.

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// DraggableActor is a Doodad being dragged from the Doodad palette.
type DraggableActor struct {
	canvas *uix.Canvas
	doodad *doodads.Doodad // if a new one from the palette
	actor  *level.Actor    // if a level actor
}

// startDragActor begins the drag event for a Doodad onto a level.
// actor may be nil (if you drag a new doodad from the palette) or otherwise
// is an existing actor from the level.
func (u *EditorUI) startDragActor(doodad *doodads.Doodad, actor *level.Actor) {
	u.Supervisor.DragStart()

	if doodad == nil {
		if actor != nil {
			obj, err := doodads.LoadFile(actor.Filename)
			if err != nil {
				log.Error("startDragExistingActor: actor doodad name %s not found: %s", actor.Filename, err)
				return
			}
			doodad = obj
		} else {
			panic("EditorUI.startDragActor: doodad AND/OR actor is required, but neither were given")
		}
	}

	// Create the canvas to render on the mouse cursor.
	drawing := uix.NewCanvas(doodad.Layers[0].Chunker.Size, false)
	drawing.LoadDoodad(doodad)
	drawing.Resize(doodad.Rect())
	drawing.SetBackground(render.RGBA(0, 0, 1, 0)) // TODO: invisible becomes white
	drawing.MaskColor = balance.DragColor          // blueprint effect
	u.DraggableActor = &DraggableActor{
		canvas: drawing,
		doodad: doodad,
		actor:  actor,
	}
}

// setupDoodadFrame configures the Doodad Palette tab for Edit Mode.
// This is a subroutine of editor_ui.go#SetupPalette()
//
// Can return an error if userdir.ListDoodads() returns an error (like directory
// not found), but it will *ALWAYS* return a valid ui.Frame -- it will just be
// empty and uninitialized.
func (u *EditorUI) setupDoodadFrame(e render.Engine, window *ui.Window) (*ui.Frame, error) {
	var (
		frame  = ui.NewFrame("Doodad Tab")
		perRow = balance.UIDoodadsPerRow
	)

	// Pager buttons on top of the doodad list.
	pager := ui.NewFrame("Doodad Pager")
	pager.SetBackground(render.RGBA(255, 0, 0, 20)) // TODO: if I don't set a background color,
	// this frame will light up the same color as the Link button on mouse
	// over. somewhere some memory might be shared between the recent widgets
	{
		leftBtn := ui.NewButton("Prev Page", ui.NewLabel(ui.Label{
			Text: "<",
			Font: balance.MenuFont,
		}))
		leftBtn.Handle(ui.Click, func(ed ui.EventData) error {
			u.scrollDoodadFrame(-1)
			return nil
		})
		u.Supervisor.Add(leftBtn)
		pager.Pack(leftBtn, ui.Pack{
			Side: ui.W,
		})

		scroller := ui.NewFrame("Doodad Scroll Progressbar")
		scroller.Configure(ui.Config{
			Width:      20,
			Height:     20,
			Background: render.RGBA(128, 128, 128, 128),
		})
		pager.Pack(scroller, ui.Pack{
			Side: ui.W,
		})
		u.doodadScroller = scroller

		rightBtn := ui.NewButton("Next Page", ui.NewLabel(ui.Label{
			Text: ">",
			Font: balance.MenuFont,
		}))
		rightBtn.Handle(ui.Click, func(ed ui.EventData) error {
			u.scrollDoodadFrame(1)
			return nil
		})
		u.Supervisor.Add(rightBtn)
		pager.Pack(rightBtn, ui.Pack{
			Side: ui.E,
		})
	}
	u.doodadPager = pager
	frame.Pack(pager, ui.Pack{
		Side: ui.N,
		Fill: true,
		PadY: 5,
	})

	doodadsAvailable, err := doodads.ListDoodads()
	if err != nil {
		return frame, fmt.Errorf(
			"setupDoodadFrame: doodads.ListDoodads: %s",
			err,
		)
	}

	var buttonSize = (paletteWidth - window.BoxThickness(2)) / perRow
	u.doodadButtonSize = buttonSize

	// Load all the doodads, skip hidden ones.
	var items []*doodads.Doodad
	for _, filename := range doodadsAvailable {
		doodad, err := doodads.LoadFile(filename)
		if err != nil {
			log.Error(err.Error())
			doodad = doodads.New(balance.DoodadSize)
		}

		// Skip hidden doodads.
		if doodad.Hidden && !balance.ShowHiddenDoodads {
			log.Info("skip %s: hidden doodad", filename)
			continue
		}

		doodad.Filename = filename
		items = append(items, doodad)
	}

	// Draw the doodad buttons in a grid `perRow` buttons wide.
	var (
		row      *ui.Frame
		rowCount int             // for labeling the ui.Frame for each row
		btnRows  = []*ui.Frame{} // Collect the row frames for the buttons.
	)
	for i, doodad := range items {
		doodad := doodad

		if row == nil || i%perRow == 0 {
			rowCount++
			row = ui.NewFrame(fmt.Sprintf("Doodad Row %d", rowCount))
			row.SetBackground(balance.WindowBackground)
			btnRows = append(btnRows, row)
			frame.Pack(row, ui.Pack{
				Side: ui.N,
				Fill: true,
			})
		}

		can := uix.NewCanvas(int(buttonSize), true)
		can.Name = doodad.Title
		can.SetBackground(render.White)
		can.LoadDoodad(doodad)

		btn := ui.NewButton(doodad.Title, can)
		btn.Resize(render.NewRect(
			buttonSize-2, // TODO: without the -2 the button border
			buttonSize-2, // rests on top of the window border.
		))
		row.Pack(btn, ui.Pack{
			Side: ui.W,
		})

		// Tooltip hover to show the doodad's name.
		ui.NewTooltip(btn, ui.Tooltip{
			Text: doodad.Title,
			Edge: ui.Top,
		})

		// Begin the drag event to grab this Doodad.
		// NOTE: The drag target is the EditorUI.Canvas in
		// editor_ui.go#SetupCanvas()
		btn.Handle(ui.MouseDown, func(ed ui.EventData) error {
			log.Warn("MouseDown on doodad %s (%s)", doodad.Filename, doodad.Title)
			u.startDragActor(doodad, nil)
			return nil
		})
		u.Supervisor.Add(btn)

		// Resize the canvas to fill the button interior.
		btnSize := btn.Size()
		can.Resize(render.NewRect(
			btnSize.W-btn.BoxThickness(2),
			btnSize.H-btn.BoxThickness(2),
		),
		)

		btn.Compute(e)
	}

	u.doodadRows = btnRows
	u.scrollDoodadFrame(0)

	return frame, nil
}

// scrollDoodadFrame handles the Page Up/Down buttons to adjust the number of
// Doodads visible on screen.
//
// rows is the number of rows to scroll. Positive values mean scroll *down*
// the list.
func (u *EditorUI) scrollDoodadFrame(rows int) {
	u.doodadSkip += rows
	if u.doodadSkip < 0 {
		u.doodadSkip = 0
	}

	// Calculate about how many rows we can see given our current window size.
	var (
		maxVisibleHeight = u.d.height - 86
		calculatedHeight int
		rowsBefore       int // count of rows hidden before
		rowsVisible      int
		rowsAfter        int                                     // count of rows hidden after
		rowsEstimated    = maxVisibleHeight / u.doodadButtonSize // estimated number rows shown
		maxSkip          = ((len(u.doodadRows) * int(u.doodadButtonSize)) - int(u.doodadButtonSize*rowsEstimated)) / int(u.doodadButtonSize)
	)

	if maxSkip < 0 {
		maxSkip = 0
	}

	if u.doodadSkip > maxSkip {
		u.doodadSkip = maxSkip
	}

	// If the window is big enough to encompass all the doodads, don't show
	// the pager toolbar, its just confusing.
	if maxSkip == 0 {
		u.doodadPager.Hide()
	} else {
		u.doodadPager.Show()
	}

	// Comb through the doodads and show/hide the relevant buttons.
	for i, row := range u.doodadRows {
		if i < u.doodadSkip {
			row.Hide()
			rowsBefore++
			continue
		}

		calculatedHeight += u.doodadButtonSize
		if calculatedHeight > maxVisibleHeight {
			row.Hide()
			rowsAfter++
		} else {
			row.Show()
			rowsVisible++
		}
	}

	var viewPercent = float64(rowsBefore+rowsVisible) / float64(len(u.doodadRows))
	u.doodadScroller.Configure(ui.Config{
		Width: int(float64(paletteWidth-50) * viewPercent), // TODO: hacky magic number
	})

}

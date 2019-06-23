package doodle

// XXX REFACTOR XXX
// This function only uses EditorUI and not Doodle and is a candidate for
// refactor into a subpackage if EditorUI itself can ever be decoupled.

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
)

// DraggableActor is a Doodad being dragged from the Doodad palette.
type DraggableActor struct {
	canvas *uix.Canvas
	doodad *doodads.Doodad
}

// startDragActor begins the drag event for a Doodad onto a level.
func (u *EditorUI) startDragActor(doodad *doodads.Doodad) {
	u.Supervisor.DragStart()

	// Create the canvas to render on the mouse cursor.
	drawing := uix.NewCanvas(doodad.Layers[0].Chunker.Size, false)
	drawing.LoadDoodad(doodad)
	drawing.Resize(doodad.Rect())
	drawing.SetBackground(render.RGBA(0, 0, 1, 0)) // TODO: invisible becomes white
	drawing.MaskColor = balance.DragColor          // blueprint effect
	u.DraggableActor = &DraggableActor{
		canvas: drawing,
		doodad: doodad,
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

	frame.SetBackground(render.RGBA(0, 153, 255, 153))

	// Toolbar on top of the Doodad panel.
	toolbar := ui.NewFrame("Doodad Palette Toolbar")
	toolbar.Configure(ui.Config{
		Background:  render.Grey,
		BorderSize:  2,
		BorderStyle: ui.BorderRaised,
		Height:      24,
	})
	{
		// Link button.
		linkButton := ui.NewButton("Link", ui.NewLabel(ui.Label{
			Text: "Link Doodads",
		}))
		linkButton.Handle(ui.Click, func(p render.Point) {
			u.Canvas.LinkStart()
			u.d.Flash("Click on the first Doodad to link to another one.")
		})
		u.Supervisor.Add(linkButton)

		toolbar.Pack(linkButton, ui.Pack{
			Anchor: ui.N,
			FillX:  true,
		})
	}
	frame.Pack(toolbar, ui.Pack{
		Anchor: ui.N,
		Fill:   true,
	})

	// Pager buttons on top of the doodad list.
	pager := ui.NewFrame("Doodad Pager")
	pager.SetBackground(render.RGBA(255, 0, 0, 20)) // TODO: if I don't set a background color,
	// this frame will light up the same color as the Link button on mouse
	// over. somewhere some memory might be shared between the recent widgets
	{
		leftBtn := ui.NewButton("Prev Page", ui.NewLabel(ui.Label{
			Text: "<",
		}))
		u.Supervisor.Add(leftBtn)
		pager.Pack(leftBtn, ui.Pack{
			Anchor: ui.W,
		})

		pageLabel := ui.NewLabel(ui.Label{
			Text: "    Page 1 of 20",
		})
		pager.Pack(pageLabel, ui.Pack{
			Anchor: ui.W,
			Expand: true,
		})

		rightBtn := ui.NewButton("Next Page", ui.NewLabel(ui.Label{
			Text: ">",
		}))
		u.Supervisor.Add(rightBtn)
		pager.Pack(rightBtn, ui.Pack{
			Anchor: ui.W,
		})
	}
	frame.Pack(pager, ui.Pack{
		Anchor: ui.N,
		Fill:   true,
		PadY:   5,
	})

	doodadsAvailable, err := doodads.ListDoodads()
	if err != nil {
		return frame, fmt.Errorf(
			"setupDoodadFrame: doodads.ListDoodads: %s",
			err,
		)
	}

	var buttonSize = (paletteWidth - window.BoxThickness(2)) / int32(perRow)

	// Draw the doodad buttons in a grid `perRow` buttons wide.
	var (
		row      *ui.Frame
		rowCount int // for labeling the ui.Frame for each row
	)
	for i, filename := range doodadsAvailable {
		if row == nil || i%perRow == 0 {
			rowCount++
			row = ui.NewFrame(fmt.Sprintf("Doodad Row %d", rowCount))
			row.SetBackground(balance.WindowBackground)
			frame.Pack(row, ui.Pack{
				Anchor: ui.N,
				Fill:   true,
			})
		}

		func(filename string) {
			doodad, err := doodads.LoadFile(filename)
			if err != nil {
				log.Error(err.Error())
				doodad = doodads.New(balance.DoodadSize)
			}

			can := uix.NewCanvas(int(buttonSize), true)
			can.Name = filename
			can.SetBackground(render.White)
			can.LoadDoodad(doodad)

			btn := ui.NewButton(filename, can)
			btn.Resize(render.NewRect(
				buttonSize-2, // TODO: without the -2 the button border
				buttonSize-2, // rests on top of the window border.
			))
			row.Pack(btn, ui.Pack{
				Anchor: ui.W,
			})

			// Begin the drag event to grab this Doodad.
			// NOTE: The drag target is the EditorUI.Canvas in
			// editor_ui.go#SetupCanvas()
			btn.Handle(ui.MouseDown, func(e render.Point) {
				u.startDragActor(doodad)
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
		}(filename)
	}

	return frame, nil
}

package doodle

import (
	"fmt"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/uix"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// setupInventoryHud configures the Inventory HUD.
func (s *PlayScene) setupInventoryHud() {
	s.invenFrame = ui.NewFrame("Inventory")
	s.invenDoodads = map[string]*uix.Canvas{}
	s.invenFrame.Configure(ui.Config{
		BorderStyle: ui.BorderRaised,
		BorderSize:  2,
		Background:  render.RGBA(128, 128, 128, 60),
	})

	// Items label.
	label := ui.NewLabel(ui.Label{
		Text: "Items:",
		Font: balance.LabelFont,
	})
	label.Compute(s.d.Engine)

	// Configure the label tall enough to cover typical 32x32 item doodads.
	// This pushes the Frame height tall enough.
	label.Configure(ui.Config{
		Height: 36,
		Width:  label.Size().W + 2,
	})
	s.invenFrame.Pack(label, ui.Pack{
		Side: ui.W,
		PadX: 2,
		PadY: 4,
	})

	// Add the inventory frame to the screen frame.
	s.screen.Place(s.invenFrame, ui.Place{
		Top:   40,
		Right: 40,
	})

	// Hide inventory if empty.
	if len(s.invenItems) == 0 {
		s.invenFrame.Hide()
	}
}

// computeInventory adjusts the inventory HUD when the player's inventory changes.
func (s *PlayScene) computeInventory() {
	items := s.Player.ListItems()

	// Inventory has changed! See which doodads we have
	// and which we need to load.
	var seen = map[string]interface{}{}
	for _, filename := range items {
		seen[filename] = nil

		if _, ok := s.invenDoodads[filename]; !ok {
			// Cache miss. Load the doodad here.
			doodad, err := doodads.LoadFile(filename)
			if err != nil {
				s.d.FlashError("Inventory item '%s' error: %s", filename, err)
				continue
			}

			canvas := uix.NewCanvas(doodad.ChunkSize8(), false)
			canvas.SetBackground(render.RGBA(1, 0, 0, 0))
			canvas.LoadDoodad(doodad)
			canvas.Resize(render.NewRect(
				doodad.ChunkSize(), doodad.ChunkSize(),
			))
			s.invenFrame.Pack(canvas, ui.Pack{
				Side: ui.W,

				// TODO: work around a weird padding bug. item had too
				// tall a top margin when added to the inventory frame!
				PadX: 8,
			})
			s.invenDoodads[filename] = canvas
		}

		// For items with >1 quantity, show the quantity in the corner.
		if qty := s.Player.HasItem(filename); qty > 0 {
			s.invenDoodads[filename].CornerLabel = fmt.Sprintf("%d", qty)
		} else {
			s.invenDoodads[filename].CornerLabel = ""
		}

		s.invenDoodads[filename].Show()
	}

	// Hide any doodad that used to be in the inventory but now is not.
	for filename, canvas := range s.invenDoodads {
		if _, ok := seen[filename]; !ok {
			canvas.Hide()
		}
	}

	// Recompute the size of the inventory frame.
	// TODO: this works around a bug in ui.Frame, at the bottom of
	// compute_packed, a frame Resize's itself to fit the children but this
	// trips the "manually set size" boolean... packing more items after a
	// computer doesn't resize the frame. So here, we resize-auto it to
	// reset that boolean so the next compute, picks the right size.
	s.invenFrame.Configure(ui.Config{
		AutoResize: true,
		Width:      1,
		Height:     1,
	})
	s.invenFrame.Compute(s.d.Engine)

	// If we removed all items, hide the frame.
	if len(items) == 0 {
		s.invenFrame.Hide()
	} else {
		s.invenFrame.Show()
	}

	// Cache the item list so we don't run the above logic every single tick.
	s.invenItems = items

	// Compute the inventory frame so it positions and wraps the items.
	s.screen.Compute(s.d.Engine)
}

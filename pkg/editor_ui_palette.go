package doodle

import (
	"fmt"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/uix"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Width of the panel frame.
var paletteWidth = 50

// SetupPalette sets up the palette panel.
func (u *EditorUI) SetupPalette(d *Doodle) *ui.Window {
	window := ui.NewWindow("Palette")
	window.ConfigureTitle(balance.TitleConfig)
	_, label := window.TitleBar()
	label.Font = balance.TitleFont
	window.Configure(ui.Config{
		Background:  balance.WindowBackground,
		BorderColor: balance.WindowBorder,
	})

	// Color Palette Frame.
	u.PaletteTab = u.setupPaletteFrame(window)
	window.Pack(u.PaletteTab, ui.Pack{
		Side: ui.N,
		Fill: true,
	})

	return window
}

// setupPaletteFrame configures the Color Palette tab for Edit Mode.
// This is a subroutine of editor_ui.go#SetupPalette()
func (u *EditorUI) setupPaletteFrame(window *ui.Window) *ui.Frame {
	frame := ui.NewFrame("Palette Tab")
	frame.SetBackground(balance.WindowBackground)

	var (
		packAlign  = ui.N
		packConfig = ui.Pack{
			Side: packAlign,
			Fill: true,
			PadY: 1,
		}
		tooltipEdge = ui.Left
		buttonSize  = 32
		twoColumn   = true // To place in two columns, halves buttonSize to /2
	)
	if usercfg.Current.HorizontalToolbars {
		packAlign = ui.W
		packConfig = ui.Pack{
			Side: packAlign,
			Fill: true,
			PadX: 2,
		}
		tooltipEdge = ui.Top
		buttonSize = 24
		twoColumn = false
	}

	// Handler function for the radio buttons being clicked.
	onClick := func(ed ui.EventData) error {
		name := u.selectedSwatch
		swatch, ok := u.Canvas.Palette.Get(name)
		if !ok {
			log.Error("Palette onClick: couldn't get swatch named '%s' from palette", name)
			return nil
		}
		log.Info("Set swatch: %s", swatch)
		u.Canvas.SetSwatch(swatch)
		return nil
	}

	// Draw the radio buttons for the palette.
	var row *ui.Frame
	if u.Canvas != nil && u.Canvas.Palette != nil {
		for i, swatch := range u.Canvas.Palette.Swatches {
			swatch := swatch
			var width = buttonSize

			// Drawing buttons in two-column mode? (default right-side palette layout)
			if twoColumn {
				width = buttonSize / 2
				if row == nil || i%2 == 0 {
					row = ui.NewFrame(fmt.Sprintf("Swatch(%s) Button Frame", swatch.Name))
					frame.Pack(row, packConfig)
				}
			} else {
				row = ui.NewFrame(fmt.Sprintf("Swatch(%s) Button Frame", swatch.Name))
				frame.Pack(row, packConfig)
			}

			// Fancy colorbox: show the color AND the texture of each swatch.
			var (
				colorbox = uix.NewCanvas(width, false)
				chunker  = level.NewChunker(width)
				size     = render.NewRect(width, width)
			)
			chunker.SetRect(size, swatch)
			colorbox.Resize(size)
			colorbox.Load(u.Canvas.Palette, chunker)

			btn := ui.NewRadioButton("palette", &u.selectedSwatch, swatch.Name, colorbox)
			btn.Configure(ui.Config{
				BorderColor: swatch.Color.Darken(20),
				BorderSize:  2,
				OutlineSize: 0,
			})
			btn.Handle(ui.Click, onClick)
			u.Supervisor.Add(btn)

			// Add a tooltip showing the swatch attributes.
			ui.NewTooltip(btn, ui.Tooltip{
				Text: fmt.Sprintf("Name: %s\nAttributes: %s", swatch.Name, swatch.Attributes()),
				Edge: tooltipEdge,
			})

			btn.Compute(u.d.Engine)

			row.Pack(btn, ui.Pack{
				Side: ui.W,
				PadX: 1,
			})
		}
	}

	// Draw the Edit Palette button.
	btn := ui.NewButton("Edit Palette", ui.NewLabel(ui.Label{
		Text: "Edit",
		Font: balance.MenuFont,
	}))
	btn.Handle(ui.Click, func(ed ui.EventData) error {
		u.OpenPaletteWindow()
		return nil
	})
	u.Supervisor.Add(btn)

	btn.Compute(u.d.Engine)
	frame.Pack(btn, packConfig)

	return frame
}

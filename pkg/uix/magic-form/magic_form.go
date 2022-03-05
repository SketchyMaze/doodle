// Package magicform helps create simple form layouts with go/ui.
package magicform

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
)

type Type int

const (
	Auto   Type = iota
	Text        // free, wide Label row
	Frame       // custom frame from the caller
	Button      // Single button with a label
	Textbox
	Checkbox
	Radiobox
	Selectbox
)

// Form configuration.
type Form struct {
	Supervisor *ui.Supervisor // Required for most useful forms
	Engine     render.Engine

	// For vertical forms.
	Vertical   bool
	LabelWidth int // size of left frame for labels.
	PadY       int // spacer between (vertical) forms
	PadX       int
}

/*
Field for your form (or form-aligned label sections, etc.)

The type of Form control to render is inferred based on bound
variables and other configuration.
*/
type Field struct {
	// Type may be inferred by presence of other params.
	Type Type

	// Set a text string and font for simple labels or paragraphs.
	Label string
	Font  render.Text

	// Easy button row: make Buttons an array of Button fields
	Buttons     []Field
	ButtonStyle *style.Button

	// Easy Paginator. DO NOT SUPERVISE, let the Create do so!
	Pager *ui.Pager

	// If you send a *ui.Frame to insert, the Type is inferred
	// to be Frame.
	Frame *ui.Frame

	// Variable bindings, the type may infer to be:
	BoolVariable *bool       // Checkbox
	TextVariable *string     // Textbox
	IntVariable  *int        // Textbox
	Options      []Option    // Selectbox
	SelectValue  interface{} // Selectbox default choice

	// Tooltip to add to a form control.
	// Checkbox only for now.
	Tooltip ui.Tooltip // config for the tooltip only

	// Handlers you can configure
	OnSelect func(value interface{}) // Selectbox
	OnClick  func()                  // Button
}

// Option used in Selectbox or Radiobox fields.
type Option struct {
	Value interface{}
	Label string
}

/*
Create the form field and populate it into the given Frame.

Renders the form vertically.
*/
func (form Form) Create(into *ui.Frame, fields []Field) {
	for n, row := range fields {
		row := row

		if row.Frame != nil {
			into.Pack(row.Frame, ui.Pack{
				Side:  ui.N,
				FillX: true,
			})
			continue
		}

		frame := ui.NewFrame(fmt.Sprintf("Line %d", n))
		into.Pack(frame, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  form.PadY,
		})

		// Pager row?
		if row.Pager != nil {
			row.Pager.Compute(form.Engine)
			form.Supervisor.Add(row.Pager)
			frame.Pack(row.Pager, ui.Pack{
				Side:   ui.W,
				Expand: true,
			})

		}

		// Buttons row?
		if row.Buttons != nil && len(row.Buttons) > 0 {
			for _, row := range row.Buttons {
				row := row

				btn := ui.NewButton(row.Label, ui.NewLabel(ui.Label{
					Text: row.Label,
					Font: row.Font,
				}))
				if row.ButtonStyle != nil {
					btn.SetStyle(row.ButtonStyle)
				}

				btn.Handle(ui.Click, func(ed ui.EventData) error {
					if row.OnClick != nil {
						row.OnClick()
					} else {
						log.Error("No OnClick handler for button %s", row.Label)
					}
					return nil
				})

				btn.Compute(form.Engine)
				form.Supervisor.Add(btn)

				frame.Pack(btn, ui.Pack{
					Side: ui.W,
					PadX: 4,
					PadY: 2,
				})
			}

			continue
		}

		// Infer the type of the form field.
		if row.Type == Auto {
			row.Type = row.Infer()
			if row.Type == Auto {
				continue
			}
		}

		// Is there a label frame to the left?
		// - Checkbox gets a full row.
		if row.Label != "" && row.Type != Checkbox {
			labFrame := ui.NewFrame("Label Frame")
			labFrame.Configure(ui.Config{
				Width: form.LabelWidth,
			})
			frame.Pack(labFrame, ui.Pack{
				Side: ui.W,
			})

			// Draw the label text into it.
			label := ui.NewLabel(ui.Label{
				Text: row.Label,
				Font: row.Font,
			})
			labFrame.Pack(label, ui.Pack{
				Side: ui.W,
			})
		}

		// Buttons and Text fields (for now).
		if row.Type == Button || row.Type == Textbox {
			btn := ui.NewButton("Button", ui.NewLabel(ui.Label{
				Text:         row.Label,
				Font:         row.Font,
				TextVariable: row.TextVariable,
				IntVariable:  row.IntVariable,
			}))
			form.Supervisor.Add(btn)
			frame.Pack(btn, ui.Pack{
				Side:   ui.W,
				FillX:  true,
				Expand: true,
			})

			// Tooltip? TODO - make nicer.
			if row.Tooltip.Text != "" || row.Tooltip.TextVariable != nil {
				ui.NewTooltip(btn, row.Tooltip)
			}

			// Handlers
			btn.Handle(ui.Click, func(ed ui.EventData) error {
				if row.OnClick != nil {
					row.OnClick()
				}
				return nil
			})
		}

		// Checkbox?
		if row.Type == Checkbox {
			cb := ui.NewCheckbox("Checkbox", row.BoolVariable, ui.NewLabel(ui.Label{
				Text: row.Label,
				Font: row.Font,
			}))
			cb.Supervise(form.Supervisor)
			frame.Pack(cb, ui.Pack{
				Side:  ui.W,
				FillX: true,
			})

			// Tooltip? TODO - make nicer.
			if row.Tooltip.Text != "" || row.Tooltip.TextVariable != nil {
				ui.NewTooltip(cb, row.Tooltip)
			}

			// Handlers
			cb.Handle(ui.Click, func(ed ui.EventData) error {
				if row.OnClick != nil {
					row.OnClick()
				}
				return nil
			})
		}

		// Selectbox? also Radiobox for now.
		if row.Type == Selectbox || row.Type == Radiobox {
			btn := ui.NewSelectBox("Select", ui.Label{
				Font: row.Font,
			})
			frame.Pack(btn, ui.Pack{
				Side:   ui.W,
				FillX:  true,
				Expand: true,
			})

			if row.Options != nil {
				for _, option := range row.Options {
					btn.AddItem(option.Label, option.Value, func() {})
				}
			}

			if row.SelectValue != nil {
				btn.SetValue(row.SelectValue)
			}

			btn.Handle(ui.Change, func(ed ui.EventData) error {
				if selection, ok := btn.GetValue(); ok {
					if row.OnSelect != nil {
						row.OnSelect(selection.Value)
					}
				}
				return nil
			})

			btn.Supervise(form.Supervisor)
			form.Supervisor.Add(btn)
		}
	}
}

/*
Infer the type if the field was of type Auto.

Returns the first Type inferred from the field by checking in
this order:

- Frame if the field has a *Frame
- Checkbox if there is a *BoolVariable
- Selectbox if there are Options
- Textbox if there is a *TextVariable
- Text if there is a Label

May return Auto if none of the above and be ignored.
*/
func (field Field) Infer() Type {
	if field.Frame != nil {
		return Frame
	}

	if field.BoolVariable != nil {
		return Checkbox
	}

	if field.Options != nil && len(field.Options) > 0 {
		return Selectbox
	}

	if field.TextVariable != nil || field.IntVariable != nil {
		return Textbox
	}

	if field.Label != "" {
		return Text
	}

	return Auto
}

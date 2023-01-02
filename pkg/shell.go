package doodle

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/keybind"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal/loadscreen"
	"git.kirsle.net/SketchyMaze/doodle/pkg/physics"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
	"github.com/dop251/goja"
)

// Flash a message to the user.
func (d *Doodle) Flash(template string, v ...interface{}) {
	log.Warn(template, v...)
	d.shell.Write(fmt.Sprintf(template, v...))
}

// FlashError flashes an error-colored message to the user.
func (d *Doodle) FlashError(template string, v ...interface{}) {
	log.Error(template, v...)
	d.shell.WriteColorful(fmt.Sprintf(template, v...), balance.FlashErrorColor)
}

// Prompt the user for a question in the dev console.
func (d *Doodle) Prompt(question string, callback func(string)) {
	d.shell.Prompt = question
	d.shell.callback = callback
	d.shell.Open = true
}

// PromptPre prompts with a pre-filled value.
func (d *Doodle) PromptPre(question string, prefilled string, callback func(string)) {
	d.shell.Text = prefilled
	d.shell.Prompt = question
	d.shell.callback = callback
	d.shell.Open = true
}

// FindLikelySupervisor will locate a most likely ui.Supervisor depending on the current Scene,
// if it understands the Scene and knows where it keeps its Supervisor.
func (d *Doodle) FindLikelySupervisor() (*ui.Supervisor, error) {
	switch scene := d.Scene.(type) {
	case *EditorScene:
		return scene.UI.Supervisor, nil
	case *PlayScene:
		return scene.Supervisor, nil
	case *MainScene:
		return scene.Supervisor, nil
	}
	return nil, errors.New("couldn't find a Supervisor")
}

// Shell implements the developer console in-game.
type Shell struct {
	parent *Doodle

	Open     bool
	Prompt   string
	Repl     bool
	callback func(string) // for prompt answers only
	Text     string
	History  []string
	Output   []string
	Flashes  []Flash

	// Blinky cursor variables.
	cursor     byte   // cursor symbol
	cursorFlip uint64 // ticks until cursor flip
	cursorRate uint64

	// Paging through history variables.
	historyPaging bool
	historyIndex  int

	// JavaScript shell interpreter.
	js *goja.Runtime
}

// Flash holds a message to flash on screen.
type Flash struct {
	Text    string
	Expires uint64 // tick that it expires
	Color   render.Color
}

// NewShell initializes the shell helper (the "Shellper").
func NewShell(d *Doodle) Shell {
	s := Shell{
		parent:     d,
		History:    []string{},
		Output:     []string{},
		Flashes:    []Flash{},
		Prompt:     ">",
		cursor:     '_',
		cursorRate: balance.ShellCursorBlinkRate,
		js:         goja.New(),
	}

	// Make the Doodle instance available to the shell.
	bindings := map[string]interface{}{
		"d":       d,
		"Execute": s.Execute,
		"RGBA":    render.RGBA,
		"Point":   render.NewPoint,
		"Vector":  physics.NewVector,
		"Rect":    render.NewRect,
		"Tree": func(w ui.Widget) string {
			for _, row := range ui.WidgetTree(w) {
				d.Flash(row)
			}
			return ""
		},
		"loadscreen": map[string]interface{}{
			"Show":             loadscreen.Show,
			"ShowWithProgress": loadscreen.ShowWithProgress,
			"Hide":             loadscreen.Hide,
			"IsActive":         loadscreen.IsActive,
			"SetProgress":      loadscreen.SetProgress,
		},
	}
	for name, v := range bindings {
		err := s.js.Set(name, v)
		if err != nil {
			log.Error("Failed to make `%s` available to JS shell: %s", name, err)
		}
	}

	return s
}

// Close the shell, resetting its internal state.
func (s *Shell) Close() {
	log.Debug("Shell: closing shell")
	s.Open = false
	s.Repl = false
	s.Prompt = ">"
	s.callback = nil
	s.Text = ""
	s.historyPaging = false
	s.historyIndex = 0
}

// Execute a command in the shell.
func (s *Shell) Execute(input string) {
	command := s.Parse(input)

	if command.Raw != "" {
		s.Output = append(s.Output, s.Prompt+command.Raw)
		s.History = append(s.History, command.Raw)
	}

	// Are we answering a Prompt?
	if s.callback != nil {
		log.Info("Invoking prompt callback:")
		s.callback(command.Raw)
		s.Close()
		return
	}

	if command.Command == "clear" {
		s.Output = []string{}
	} else {
		err := command.Run(s.parent)
		if err != nil {
			s.Write(err.Error())
		}
	}

	// Reset the text buffer in the shell.
	if s.Repl {
		s.Text = "$ "
	} else {
		s.Text = ""
	}
}

// Write a line of output text to the console.
func (s *Shell) Write(line string) {
	s.Output = append(s.Output, line)
	s.Flashes = append(s.Flashes, Flash{
		Text:    line,
		Expires: shmem.Tick + balance.FlashTTL,
	})
}

// WriteError writes a line of error (red) text to the console.
func (s *Shell) WriteColorful(line string, color render.Color) {
	s.Output = append(s.Output, line)
	s.Flashes = append(s.Flashes, Flash{
		Text:    line,
		Color:   color,
		Expires: shmem.Tick + balance.FlashTTL,
	})
}

// Parse the command line.
func (s *Shell) Parse(input string) Command {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return Command{}
	}

	var (
		inQuote bool
		buffer  = bytes.NewBuffer([]byte{})
		words   = []string{}
	)
	for i := 0; i < len(input); i++ {
		char := input[i]
		switch char {
		case ' ':
			if inQuote {
				buffer.WriteByte(char)
				continue
			}

			if word := buffer.String(); word != "" {
				words = append(words, word)
				buffer.Reset()
			}
		case '"':
			if !inQuote {
				// An opening quote character.
				inQuote = true
			} else {
				// The closing quote.
				inQuote = false

				if word := buffer.String(); word != "" {
					words = append(words, word)
					buffer.Reset()
				}
			}
		default:
			buffer.WriteByte(char)
		}
	}

	if remainder := buffer.String(); remainder != "" {
		words = append(words, remainder)
	}

	return Command{
		Raw:         input,
		Command:     words[0],
		Args:        words[1:],
		ArgsLiteral: strings.TrimSpace(input[len(words[0]):]),
	}
}

// Draw the shell.
func (s *Shell) Draw(d *Doodle, ev *event.State) error {
	// Compute the line height we can draw.
	lineHeight := balance.ShellFontSize + int(balance.ShellPadding)

	// If the console is open, draw the console.
	if s.Open {
		if ev.Escape {
			s.Close()
			return nil
		} else if keybind.Enter(ev) {
			s.Execute(s.Text)

			// Auto-close the console unless in REPL mode.
			if !s.Repl {
				s.Close()
			}

			return nil
		} else if (ev.Up || ev.Down) && len(s.History) > 0 {
			// Paging through history.
			if !s.historyPaging {
				s.historyPaging = true
				s.historyIndex = len(s.History)
			}

			// Consume the inputs and make convenient variables.
			isUp := ev.Up
			ev.Down = false
			ev.Up = false

			// Scroll through the input history.
			if isUp {
				s.historyIndex--
				if s.historyIndex < 0 {
					s.historyIndex = 0
				}
			} else {
				s.historyIndex++
				if s.historyIndex >= len(s.History) {
					s.historyIndex = len(s.History) - 1
				}
			}

			s.Text = s.History[s.historyIndex]

		}

		// Cursor flip?
		if shmem.Tick > s.cursorFlip {
			s.cursorFlip = shmem.Tick + s.cursorRate
			if s.cursor == ' ' {
				s.cursor = '_'
			} else {
				s.cursor = ' '
			}
		}

		// Read a character from the keyboard.
		for _, key := range ev.KeysDown(true) {
			// Backspace?
			if key == `\b` {
				if len(s.Text) > 0 {
					s.Text = s.Text[:len(s.Text)-1]
				}
			} else {
				s.Text += key
			}
			// HACK: I wanted to do:
			// ev.SetKeyDown(key, false)
			// But, ev.KeysDown(shifted=true) returns letter keys
			// like 'M' when the key we wanted to unset was 'm',
			// or we got '$' when we want to unset '5'... so all
			// shifted chars got duplicated 3+ times on key press!
			// So, just reset ALL key press states to work around it:
			ev.ResetKeyDown()
		}

		// How tall is the box?
		boxHeight := (lineHeight * (balance.ShellHistoryLineCount + 1)) + balance.ShellPadding

		// Draw the background color.
		d.Engine.DrawBox(
			balance.ShellBackgroundColor,
			render.Rect{
				X: 0,
				Y: d.height - boxHeight,
				W: d.width,
				H: boxHeight,
			},
		)

		// Draw the recent commands.
		outputY := d.height - (lineHeight * 2)
		for i := 0; i < balance.ShellHistoryLineCount; i++ {
			if len(s.Output) > i {
				line := s.Output[len(s.Output)-1-i]
				d.Engine.DrawText(
					render.Text{
						FontFilename: balance.ShellFontFilename,
						Text:         line,
						Size:         balance.ShellFontSize,
						Color:        balance.ShellForegroundColor,
					},
					render.Point{
						X: balance.ShellPadding,
						Y: outputY,
					},
				)
			}
			outputY -= lineHeight
		}

		// Draw the command prompt.
		d.Engine.DrawText(
			render.Text{
				FontFilename: balance.ShellFontFilename,
				Text:         s.Prompt + s.Text + string(s.cursor),
				Size:         balance.ShellFontSize,
				Color:        balance.ShellPromptColor,
			},
			render.Point{
				X: balance.ShellPadding,
				Y: d.height - balance.ShellFontSize - balance.ShellPadding,
			},
		)
	} else if len(s.Flashes) > 0 {
		// Otherwise, just draw flashed messages.
		valid := false // Did we actually draw any?

		outputY := d.height - (lineHeight * 2) - 16
		for i := len(s.Flashes); i > 0; i-- {
			flash := s.Flashes[i-1]
			if shmem.Tick >= flash.Expires {
				continue
			}

			var text = balance.FlashFont(flash.Text)
			if !flash.Color.IsZero() {
				text.Color = flash.Color
				text.Stroke = text.Color.Darken(balance.FlashStrokeDarken)
				text.Shadow = text.Color.Darken(balance.FlashShadowDarken)
			}

			d.Engine.DrawText(
				text,
				render.Point{
					X: balance.ShellPadding + toolbarWidth,
					Y: outputY,
				},
			)
			outputY -= lineHeight
			valid = true
		}

		// If we've exhausted all flashes, free up the memory.
		if !valid {
			s.Flashes = []Flash{}
		}
	}

	return nil
}

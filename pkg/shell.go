package doodle

import (
	"bytes"
	"fmt"
	"strings"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/robertkrimen/otto"
)

// Flash a message to the user.
func (d *Doodle) Flash(template string, v ...interface{}) {
	log.Warn(template, v...)
	d.shell.Write(fmt.Sprintf(template, v...))
}

// Prompt the user for a question in the dev console.
func (d *Doodle) Prompt(question string, callback func(string)) {
	d.shell.Prompt = question
	d.shell.callback = callback
	d.shell.Open = true
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
	js *otto.Otto
}

// Flash holds a message to flash on screen.
type Flash struct {
	Text    string
	Expires uint64 // tick that it expires
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
		js:         otto.New(),
	}

	// Make the Doodle instance available to the shell.
	bindings := map[string]interface{}{
		"d":     d,
		"log":   log.Logger,
		"RGBA":  render.RGBA,
		"Point": render.NewPoint,
		"Rect":  render.NewRect,
		"Tree": func(w ui.Widget) string {
			for _, row := range ui.WidgetTree(w) {
				d.Flash(row)
			}
			return ""
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
		Expires: s.parent.ticks + balance.FlashTTL,
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
func (s *Shell) Draw(d *Doodle, ev *events.State) error {
	// Compute the line height we can draw.
	lineHeight := balance.ShellFontSize + int(balance.ShellPadding)

	// If the console is open, draw the console.
	if s.Open {
		if ev.EscapeKey.Read() {
			s.Close()
			return nil
		} else if ev.EnterKey.Read() || ev.EscapeKey.Read() {
			s.Execute(s.Text)

			// Auto-close the console unless in REPL mode.
			if !s.Repl {
				s.Close()
			}

			return nil
		} else if (ev.Up.Now || ev.Down.Now) && len(s.History) > 0 {
			// Paging through history.
			if !s.historyPaging {
				s.historyPaging = true
				s.historyIndex = len(s.History)
			}

			// Consume the inputs and make convenient variables.
			ev.Down.Read()
			isUp := ev.Up.Read()

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
		if d.ticks > s.cursorFlip {
			s.cursorFlip = d.ticks + s.cursorRate
			if s.cursor == ' ' {
				s.cursor = '_'
			} else {
				s.cursor = ' '
			}
		}

		// Read a character from the keyboard.
		if key := ev.ReadKey(); key != "" {
			// Backspace?
			if key == `\b` {
				if len(s.Text) > 0 {
					s.Text = s.Text[:len(s.Text)-1]
				}
			} else {
				s.Text += key
			}
		}

		// How tall is the box?
		boxHeight := int32(lineHeight*(balance.ShellHistoryLineCount+1)) + balance.ShellPadding

		// Draw the background color.
		d.Engine.DrawBox(
			balance.ShellBackgroundColor,
			render.Rect{
				X: 0,
				Y: int32(d.height) - boxHeight,
				W: int32(d.width),
				H: boxHeight,
			},
		)

		// Draw the recent commands.
		outputY := int32(d.height - (lineHeight * 2))
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
			outputY -= int32(lineHeight)
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
				Y: int32(d.height-balance.ShellFontSize) - balance.ShellPadding,
			},
		)
	} else if len(s.Flashes) > 0 {
		// Otherwise, just draw flashed messages.
		valid := false // Did we actually draw any?

		outputY := int32(d.height - (lineHeight * 2))
		for i := len(s.Flashes); i > 0; i-- {
			flash := s.Flashes[i-1]
			if d.ticks >= flash.Expires {
				continue
			}

			d.Engine.DrawText(
				render.Text{
					Text:   flash.Text,
					Size:   balance.ShellFontSize,
					Color:  render.SkyBlue,
					Stroke: render.Grey,
					Shadow: render.Black,
				},
				render.Point{
					X: balance.ShellPadding,
					Y: outputY,
				},
			)
			outputY -= int32(lineHeight)
			valid = true
		}

		// If we've exhausted all flashes, free up the memory.
		if !valid {
			s.Flashes = []Flash{}
		}
	}

	return nil
}

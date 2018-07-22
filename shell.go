package doodle

import (
	"bytes"
	"strings"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
)

// Shell implements the developer console in-game.
type Shell struct {
	parent     *Doodle
	Open       bool
	Prompt     string
	Text       string
	History    []string
	Output     []string
	Flashes    []Flash
	Cursor     string
	cursorFlip uint64 // ticks until cursor flip
	cursorRate uint64
}

// Flash holds a message to flash on screen.
type Flash struct {
	Text    string
	Expires uint64 // tick that it expires
}

// NewShell initializes the shell helper (the "Shellper").
func NewShell(d *Doodle) Shell {
	return Shell{
		parent:     d,
		History:    []string{},
		Output:     []string{},
		Flashes:    []Flash{},
		Prompt:     ">",
		Cursor:     "_",
		cursorRate: balance.ShellCursorBlinkRate,
	}
}

// Close the shell, resetting its internal state.
func (s *Shell) Close() {
	log.Debug("Shell: closing shell")
	s.Open = false
	s.Prompt = ">"
	s.Text = ""
}

// Execute a command in the shell.
func (s *Shell) Execute(input string) {
	command := s.Parse(input)
	err := command.Run(s.parent)
	if err != nil {
		s.Write(err.Error())
	}

	if command.Raw != "" {
		s.History = append(s.History, command.Raw)
	}

	// Reset the text buffer in the shell.
	s.Text = ""
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
	if ev.EscapeKey.Read() {
		s.Close()
		return nil
	} else if ev.EnterKey.Read() || ev.EscapeKey.Read() {
		s.Execute(s.Text)
		s.Close()
		return nil
	}

	// Compute the line height we can draw.
	lineHeight := balance.ShellFontSize + int(balance.ShellPadding)

	// If the console is open, draw the console.
	if s.Open {
		// Cursor flip?
		if d.ticks > s.cursorFlip {
			s.cursorFlip = d.ticks + s.cursorRate
			if s.Cursor == "" {
				s.Cursor = "_"
			} else {
				s.Cursor = ""
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
				Y: d.height - boxHeight,
				W: d.width,
				H: boxHeight,
			},
		)

		// Draw the recent commands.
		outputY := d.height - int32(lineHeight*2)
		for i := 0; i < balance.ShellHistoryLineCount; i++ {
			if len(s.Output) > i {
				line := s.Output[len(s.Output)-1-i]
				d.Engine.DrawText(
					render.Text{
						Text:  line,
						Size:  balance.ShellFontSize,
						Color: render.Grey,
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
				Text:  s.Prompt + s.Text + s.Cursor,
				Size:  balance.ShellFontSize,
				Color: balance.ShellForegroundColor,
			},
			render.Point{
				X: balance.ShellPadding,
				Y: d.height - int32(balance.ShellFontSize) - balance.ShellPadding,
			},
		)
	} else if len(s.Flashes) > 0 {
		// Otherwise, just draw flashed messages.
		valid := false // Did we actually draw any?

		outputY := d.height - int32(lineHeight*2)
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

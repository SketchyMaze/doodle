package doodle

import (
	"errors"
	"fmt"
)

// Command is a parsed shell command.
type Command struct {
	Raw         string   // The complete raw command the user typed.
	Command     string   // The first word of their command.
	Args        []string // The shell-args array of parameters.
	ArgsLiteral string   // The args portion of the command literally.
}

// Run the command.
func (c Command) Run(d *Doodle) error {
	if len(c.Raw) == 0 {
		return nil
	}

	switch c.Command {
	case "new":
		return c.New(d)
	case "save":
		return c.Save(d)
	case "edit":
		return c.Edit(d)
	case "play":
		return c.Play(d)
	case "exit":
	case "quit":
		return c.Quit()
	default:
		return c.Default()
	}
	return nil
}

// New opens a new map in the editor mode.
func (c Command) New(d *Doodle) error {
	d.shell.Write("Starting a new map")
	d.NewMap()
	return nil
}

// Save the current map to disk.
func (c Command) Save(d *Doodle) error {
	if scene, ok := d.scene.(*EditorScene); ok {
		filename := ""
		if len(c.Args) > 0 {
			filename = c.Args[0]
		} else if scene.filename != "" {
			filename = scene.filename
		} else {
			return errors.New("usage: save <filename.json>")
		}

		d.shell.Write("Saving to file: " + filename)
		scene.SaveLevel(filename)
	} else {
		return errors.New("save: only available in Edit Mode")
	}

	return nil
}

// Edit a map from disk.
func (c Command) Edit(d *Doodle) error {
	if len(c.Args) == 0 {
		return errors.New("Usage: edit <file name>")
	}

	filename := c.Args[0]
	d.shell.Write("Editing level: " + filename)
	d.EditLevel(filename)
	return nil
}

// Play a map.
func (c Command) Play(d *Doodle) error {
	if len(c.Args) == 0 {
		return errors.New("Usage: play <file name>")
	}

	filename := c.Args[0]
	d.shell.Write("Playing level: " + filename)
	d.PlayLevel(filename)
	return nil
}

// Quit the command line shell.
func (c Command) Quit() error {
	return nil
}

// Default command.
func (c Command) Default() error {
	return fmt.Errorf("%s: command not found. Try `help` for help",
		c.Command,
	)
}

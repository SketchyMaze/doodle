package doodle

import (
	"errors"
	"fmt"
	"strconv"

	"git.kirsle.net/apps/doodle/enum"
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
	case "echo":
		d.Flash(c.ArgsLiteral)
		return nil
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
	case "help":
		return c.Help(d)
	case "reload":
		d.Goto(d.Scene)
		return nil
	case "guitest":
		d.Goto(&GUITestScene{})
		return nil
	case "eval":
	case "$":
		out, err := d.shell.js.Run(c.ArgsLiteral)
		d.Flash("%+v", out)
		return err
	case "boolProp":
		return c.BoolProp(d)
	default:
		return c.Default()
	}
	return nil
}

// New opens a new map in the editor mode.
func (c Command) New(d *Doodle) error {
	d.Flash("Starting a new map")
	d.NewMap()
	return nil
}

// Help prints the help info.
func (c Command) Help(d *Doodle) error {
	if len(c.Args) == 0 {
		d.Flash("Available commands: new save edit play quit echo clear help")
		d.Flash("Type `help` and then the command, like: `help edit`")
		return nil
	}

	switch c.Args[0] {
	case "new":
		d.Flash("Usage: new")
		d.Flash("Create a new drawing in Edit Mode")
	case "save":
		d.Flash("Usage: save [filename.json]")
		d.Flash("Save the map to disk (in Edit Mode only)")
	case "edit":
		d.Flash("Usage: edit <filename.json>")
		d.Flash("Open a file on disk in Edit Mode")
	case "play":
		d.Flash("Usage: play <filename.json>")
		d.Flash("Open a map from disk in Play Mode")
	case "echo":
		d.Flash("Usage: echo <message>")
		d.Flash("Flash a message back to the console")
	case "quit":
	case "exit":
		d.Flash("Usage: quit")
		d.Flash("Closes the dev console")
	case "clear":
		d.Flash("Usage: clear")
		d.Flash("Clears the terminal output history")
	case "help":
		d.Flash("Usage: help <command>")
	default:
		d.Flash("Unknown help topic.")
	}

	return nil
}

// Save the current map to disk.
func (c Command) Save(d *Doodle) error {
	if scene, ok := d.Scene.(*EditorScene); ok {
		filename := ""
		if len(c.Args) > 0 {
			filename = c.Args[0]
		} else if scene.filename != "" {
			filename = scene.filename
		} else {
			return errors.New("usage: save <filename>")
		}

		switch scene.DrawingType {
		case enum.LevelDrawing:
			d.shell.Write("Saving Level: " + filename)
			scene.SaveLevel(filename)
		case enum.DoodadDrawing:
			d.shell.Write("Saving Doodad: " + filename)
			scene.SaveDoodad(filename)
		}
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
	d.shell.Write("Editing file: " + filename)
	d.EditFile(filename)
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

// BoolProp command sets available boolean variables.
func (c Command) BoolProp(d *Doodle) error {
	if len(c.Args) != 2 {
		return errors.New("Usage: boolProp <name> <true or false>")
	}

	var (
		name   = c.Args[0]
		value  = c.Args[1]
		truthy = value[0] == 't' || value[0] == 'T' || value[0] == '1'
		ok     = true
	)

	switch name {
	case "Debug":
	case "D":
		d.Debug = truthy
	case "DebugOverlay":
	case "DO":
		DebugOverlay = truthy
	case "DebugCollision":
	case "DC":
		DebugCollision = truthy
	default:
		ok = false
	}

	if ok {
		d.Flash("Set boolProp %s=%s", name, strconv.FormatBool(truthy))
	} else {
		d.Flash("Unknown boolProp name %s", name)
	}

	return nil
}

// Default command.
func (c Command) Default() error {
	return fmt.Errorf("%s: command not found. Try `help` for help",
		c.Command,
	)
}

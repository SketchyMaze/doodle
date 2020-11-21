package doodle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/bindata"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/modal"
	"github.com/robertkrimen/otto"
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

	// Cheat codes
	if cheat := c.cheatCommand(d); cheat {
		return nil
	}

	switch c.Command {
	case "echo":
		d.Flash(c.ArgsLiteral)
		return nil
	case "alert":
		modal.Alert(c.ArgsLiteral)
		return nil
	case "new":
		return c.New(d)
	case "save":
		return c.Save(d)
	case "edit":
		return c.Edit(d)
	case "play":
		return c.Play(d)
	case "close":
		return c.Close(d)
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
		fallthrough
	case "$":
		out, err := c.RunScript(d, c.ArgsLiteral)
		d.Flash("%+v", out)
		return err
	case "repl":
		d.shell.Repl = true
		d.shell.Text = "$ "
	case "boolProp":
		return c.BoolProp(d)
	case "extract-bindata":
		// Undocumented command to extract the binary of its assets.
		return c.ExtractBindata(d, c.ArgsLiteral)
	default:
		return c.Default()
	}
	return nil
}

// New opens a new map in the editor mode.
func (c Command) New(d *Doodle) error {
	d.GotoNewMenu()
	return nil
}

// Close returns to the Main Scene.
func (c Command) Close(d *Doodle) error {
	main := &MainScene{}
	d.Goto(main)
	return nil
}

// ExtractBindata dumps the app's embedded bindata to the filesystem.
func (c Command) ExtractBindata(d *Doodle, path string) error {
	if len(path) == 0 || path[0] != '/' {
		d.Flash("Required: an absolute path to a directory to extract to.")
		return nil
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		d.Flash("MkdirAll: %s", err)
		return err
	}

	for _, filename := range bindata.AssetNames() {
		outfile := filepath.Join(path, filename)
		log.Info("Extracting bindata: %s    to: %s", filename, outfile)

		data, err := bindata.Asset(filename)
		if err != nil {
			d.Flash("error on file %s: %s", filename, err)
			continue
		}

		// Fill out the directory path.
		if _, err := os.Stat(filepath.Dir(outfile)); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(outfile), 0755)
		}

		fh, err := os.Create(outfile)
		if err != nil {
			d.Flash("error writing file %s: %s", outfile, err)
			continue
		}
		fh.Write(data)
		fh.Close()
	}

	d.Flash("Bindata extracted to %s", path)
	return nil
}

// Help prints the help info.
func (c Command) Help(d *Doodle) error {
	if len(c.Args) == 0 {
		d.Flash("Available commands: new save edit play quit echo")
		d.Flash("     alert clear help boolProp eval repl")
		d.Flash("Type `help` and then the command, like: `help edit`")
		return nil
	}

	switch c.Args[0] {
	case "echo":
		d.Flash("Usage: echo <message>")
		d.Flash("Flash a message back to the console")
	case "alert":
		d.Flash("Usage: alert <message>")
		d.Flash("Pop up an Alert box with a custom message")
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
	case "quit":
		fallthrough
	case "exit":
		d.Flash("Usage: quit")
		d.Flash("Closes the dev console (alias: exit)")
	case "clear":
		d.Flash("Usage: clear")
		d.Flash("Clears the console output history")
	case "eval":
		fallthrough
	case "$":
		d.Flash("Evaluate a line of JavaScript on the in-game interpreter")
	case "repl":
		d.Flash("Enter a JavaScript shell on the in-game interpreter")
	case "boolProp":
		d.Flash("Toggle boolean values. `boolProp list` lists available")
	case "help":
		d.Flash("Usage: help <command>")
		d.Flash("Gets further help on a command")
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
	return d.EditFile(filename)
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
	if len(c.Args) == 1 {
		// Showing the value of a boolProp. Only supported for those defined
		// in balance/boolprops.go
		value, err := balance.GetBoolProp(c.Args[0])
		if err != nil {
			return err
		}
		d.Flash("%s: %+v", c.Args[0], value)
		return nil
	}

	if len(c.Args) != 2 {
		return errors.New("Usage: boolProp <name> [true or false]")
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
		// Try the global boolProps in balance package.
		if err := balance.BoolProp(name, truthy); err != nil {
			d.Flash("%s", err)
		} else {
			d.Flash("%s: %+v", name, truthy)
		}
	}

	return nil
}

// RunScript evaluates some JavaScript code safely.
func (c Command) RunScript(d *Doodle, code interface{}) (otto.Value, error) {
	defer func() {
		if err := recover(); err != nil {
			d.Flash("Panic: %s", err)
		}
	}()
	out, err := d.shell.js.Run(code)
	return out, err
}

// Default command.
func (c Command) Default() error {
	return fmt.Errorf("%s: command not found. Try `help` for help",
		c.Command,
	)
}

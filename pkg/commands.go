package doodle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/assets"
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/chatbot"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/scripting/exceptions"
	"github.com/dop251/goja"
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

	switch strings.ToLower(c.Command) {
	case "echo":
		d.Flash(c.ArgsLiteral)
		return nil
	case "error":
		d.FlashError(c.ArgsLiteral)
		return nil
	case "alert":
		modal.Alert(c.ArgsLiteral)
		return nil
	case "confirm":
		modal.Confirm(c.ArgsLiteral).Then(func() {
			d.Flash("Confirmed.")
		})
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
	case "titlescreen":
		return c.TitleScreen(d)
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
	case "boolprop":
		return c.BoolProp(d)
	case "extract-bindata":
		// Undocumented command to extract the binary of its assets.
		return c.ExtractBindata(d, c.ArgsLiteral)
	case "throw":
		// Test exception catcher with custom message.
		exceptions.Catch(strings.ReplaceAll(c.ArgsLiteral, "\\n", "\n"))
		return nil
	case "throw2":
		// Stress test exception catcher.
		exceptions.Catch(
			"This is a test of the Exception Catcher modal.\n\nIt should be able to display a decent amount " +
				"of text with character wrapping. Multiple lines, too.\n\nIt might not show the full message, so " +
				"click the  'Copy' button to copy to clipboard and read the whole message. There is more text " +
				"than is shown on screen.\n\nThis text for example was not on screen, but copied to your " +
				"clipboard anyway. :)",
		)
		return nil
	case "throw3":
		// Realistic exception.
		exceptions.Catch(
			"Error in main() for actor trapdoor-down.doodad:\n\n" +
				"TypeError: Cannot read property 'zz' of undefined at main (<eval>:25:14(77))\n\n" +
				"Actor ID: c3aa346b-be51-4bc4-94bb-f3adf5643830\n" +
				"Filename: trapdoor-down.doodad\n" +
				"Position: 643,266",
		)
	case "flush-textures":
		// Flush all textures.
		native.FreeTextures(d.Engine)
		d.Flash("All textures freed.")
	default:
		return c.Default(d)
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
		d.FlashError("Required: an absolute path to a directory to extract to.")
		return nil
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		d.FlashError("MkdirAll: %s", err)
		return err
	}

	for _, filename := range assets.AssetNames() {
		outfile := filepath.Join(path, filename)
		log.Info("Extracting bindata: %s    to: %s", filename, outfile)

		data, err := assets.Asset(filename)
		if err != nil {
			d.FlashError("error on file %s: %s", filename, err)
			continue
		}

		// Fill out the directory path.
		if _, err := os.Stat(filepath.Dir(outfile)); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(outfile), 0755)
		}

		fh, err := os.Create(outfile)
		if err != nil {
			d.FlashError("error writing file %s: %s", outfile, err)
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
		d.Flash("Available commands: new save edit play quit echo error")
		d.Flash("     alert clear help boolProp eval repl")
		d.Flash("Type `help` and then the command, like: `help edit`")
		return nil
	}

	switch strings.ToLower(c.Args[0]) {
	case "echo":
		d.Flash("Usage: echo <message>")
		d.Flash("Flash a message back to the console")
	case "error":
		d.Flash("Usage: error <message>")
		d.Flash("Flash an error message back to the console")
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
	case "boolprop":
		d.Flash("Toggle boolean values. `boolProp list` lists available")
	case "titlescreen":
		d.Flash("Usage: titlescreen <filename.level>")
		d.Flash("Open the title screen with a level")
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

// TitleScreen loads the title with a custom user level.
func (c Command) TitleScreen(d *Doodle) error {
	if len(c.Args) == 0 {
		return errors.New("Usage: titlescreen <level name.level>")
	}

	filename := c.Args[0]
	d.shell.Write("Playing level: " + filename)
	d.Goto(&MainScene{
		LevelFilename: filename,
	})
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
		return errors.New("Usage: boolProp <name> [true, false, flip]")
	}

	var (
		name   = c.Args[0]
		value  = c.Args[1]
		truthy = value[0] == 't' || value[0] == 'T' || value[0] == '1'
		flip   = value == "flip"
		ok     = true
	)

	switch name {
	case "Debug":
	case "D":
		d.Debug = truthy
	case "DebugOverlay":
	case "DO":
		if flip {
			DebugOverlay = !DebugOverlay
		} else {
			DebugOverlay = truthy
		}
	case "DebugCollision":
	case "DC":
		if flip {
			DebugCollision = !DebugCollision
		} else {
			DebugCollision = truthy
		}
	default:
		ok = false
	}

	if ok {
		if flip {
			d.Flash("Toggled boolProp %s", name)
		} else {
			d.Flash("Set boolProp %s=%s", name, strconv.FormatBool(truthy))
		}
	} else {
		// Try the global boolProps in balance package.
		if err := balance.BoolProp(name, truthy); err != nil {
			d.FlashError("%s", err)
		} else {
			d.Flash("%s: %+v", name, truthy)
		}
	}

	return nil
}

// RunScript evaluates some JavaScript code safely.
func (c Command) RunScript(d *Doodle, code string) (goja.Value, error) {
	defer func() {
		if err := recover(); err != nil {
			d.FlashError("Command.RunScript: Panic: %s", err)
		}
	}()

	out, err := d.shell.js.RunString(code)

	// If we're in Play Mode, consider it cheating if the player is
	// messing with any in-game structures.
	if scene, ok := d.Scene.(*PlayScene); ok {
		scene.SetCheated()
	}

	return out, err
}

// Default command.
func (c Command) Default(d *Doodle) error {
	// Give the easter egg RiveScript bot a chance.
	if reply, err := chatbot.Handle(c.Raw); err == nil {
		for _, reply := range strings.Split(reply, "\n") {
			d.Flash(reply)
		}
		return nil
	} else {
		log.Error("RiveScript error: %s", err)
	}

	return fmt.Errorf("%s: command not found. Try `help` for help",
		c.Command,
	)
}

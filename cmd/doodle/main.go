package main

import (
	"flag"
	"runtime"

	_ "image/png"

	"git.kirsle.net/apps/doodle/lib/render/sdl"
	doodle "git.kirsle.net/apps/doodle/pkg"
	"git.kirsle.net/apps/doodle/pkg/balance"
)

// Build number is the git commit hash.
var Build string

// Command line args
var (
	debug   bool
	edit    bool
	guitest bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.BoolVar(&edit, "edit", false, "Edit the map given on the command line. Default is to play the map.")
	flag.BoolVar(&guitest, "guitest", false, "Enter the GUI Test scene.")
}

func main() {
	runtime.LockOSThread()
	flag.Parse()

	args := flag.Args()
	var filename string
	if len(args) > 0 {
		filename = args[0]
	}

	// SDL engine.
	engine := sdl.New(
		"Doodle v"+doodle.Version,
		balance.Width,
		balance.Height,
	)

	app := doodle.New(debug, engine)
	app.SetupEngine()
	if guitest {
		app.Goto(&doodle.GUITestScene{})
	} else if filename != "" {
		if edit {
			app.EditFile(filename)
		} else {
			app.PlayLevel(filename)
		}
	}
	app.Run()
}

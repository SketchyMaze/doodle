package main

import (
	"flag"
	"runtime"

	"git.kirsle.net/apps/doodle"
	"git.kirsle.net/apps/doodle/render/sdl"
)

// Build number is the git commit hash.
var Build string

// Command line args
var (
	debug bool
	edit  bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.BoolVar(&edit, "edit", false, "Edit the map given on the command line. Default is to play the map.")
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
		800,
		600,
	)

	app := doodle.New(debug, engine)
	app.SetupEngine()
	if filename != "" {
		if edit {
			app.EditFile(filename)
		} else {
			app.PlayLevel(filename)
		}
	}
	app.Run()
}

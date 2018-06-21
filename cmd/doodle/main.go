package main

import (
	"flag"
	"runtime"

	"git.kirsle.net/apps/doodle"
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

	app := doodle.New(debug)
	if filename != "" {
		if edit {
			app.EditLevel(filename)
		} else {
			app.PlayLevel(filename)
		}
	}
	app.Run()
}

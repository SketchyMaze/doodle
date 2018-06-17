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
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Debug mode")
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
		app.LoadLevel(filename)
	}
	app.Run()
}

package main

import (
	"flag"
	"runtime"

	"github.com/kirsle/doodle"
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

	app := doodle.New(debug)
	app.Run()
}

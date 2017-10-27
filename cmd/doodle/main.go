package main

import (
	"flag"

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
	flag.Parse()

	app := doodle.New(debug)
	_ = app
}

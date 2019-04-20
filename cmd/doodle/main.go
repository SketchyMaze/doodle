package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"time"

	"git.kirsle.net/apps/doodle/lib/render/sdl"
	doodle "git.kirsle.net/apps/doodle/pkg"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"github.com/urfave/cli"

	_ "image/png"
)

// Build number is the git commit hash.
var (
	Build     = "<dynamic>"
	BuildDate string
)

func init() {
	if BuildDate == "" {
		BuildDate = time.Now().Format(time.RFC3339)
	}
}

func main() {
	runtime.LockOSThread()

	app := cli.NewApp()
	app.Name = "doodle"
	app.Usage = "command line interface for Doodle"

	var freeLabel string
	if balance.FreeVersion {
		freeLabel = " (shareware)"
	}

	app.Version = fmt.Sprintf("%s build %s%s. Built on %s",
		doodle.Version,
		Build,
		freeLabel,
		BuildDate,
	)

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug level logging",
		},
		cli.BoolFlag{
			Name:  "edit, e",
			Usage: "edit the map given on the command line (instead of play it)",
		},
		cli.BoolFlag{
			Name:  "guitest",
			Usage: "enter the GUI Test scene on startup",
		},
	}

	app.Action = func(c *cli.Context) error {
		var filename string
		if c.NArg() > 0 {
			filename = c.Args().Get(0)
		}

		// SDL engine.
		engine := sdl.New(
			"Doodle v"+doodle.Version,
			balance.Width,
			balance.Height,
		)

		game := doodle.New(c.Bool("debug"), engine)
		game.SetupEngine()
		if c.Bool("guitest") {
			game.Goto(&doodle.GUITestScene{})
		} else if filename != "" {
			if c.Bool("edit") {
				game.EditFile(filename)
			} else {
				game.PlayLevel(filename)
			}
		}
		game.Run()
		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// doodad is the command line developer tool for Doodle.
package main

import (
	"log"
	"os"
	"sort"

	"git.kirsle.net/apps/doodle"
	"git.kirsle.net/apps/doodle/cmd/doodad/commands"
	"github.com/urfave/cli"
)

var Build = "N/A"

func main() {
	app := cli.NewApp()
	app.Name = "doodad"
	app.Usage = "command line interface for Doodle"
	app.Version = doodle.Version + " build " + Build

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug level logging",
		},
	}

	app.Commands = []cli.Command{
		commands.Convert,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

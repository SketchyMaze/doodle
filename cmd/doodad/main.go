// doodad is the command line developer tool for Doodle.
package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"git.kirsle.net/apps/doodle/cmd/doodad/commands"
	doodle "git.kirsle.net/apps/doodle/pkg"
	"github.com/urfave/cli"
)

// Build variables.
var (
	Build     = "N/A"
	BuildDate string
)

func init() {
	if BuildDate == "" {
		BuildDate = time.Now().Format(time.RFC3339)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "doodad"
	app.Usage = "command line interface for Doodle"
	app.Version = fmt.Sprintf("%s build %s. Built on %s",
		doodle.Version,
		Build,
		BuildDate,
	)

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

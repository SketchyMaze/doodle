// doodle-admin performs secret admin tasks like generating license keys.
package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/cmd/doodle-admin/command"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"github.com/urfave/cli/v2"
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
	app.Name = "doodle-admin"
	app.Usage = "Admin tasks for Sketchy Maze."

	app.Version = fmt.Sprintf("%s build %s. Built on %s",
		branding.Version,
		Build,
		BuildDate,
	)

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug level logging",
		},
	}

	app.Commands = []*cli.Command{
		command.Key,
		command.Sign,
		command.Verify,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// doodad is the command line developer tool for Doodle.
package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/cmd/doodad/commands"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding/builds"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/bootstrap"
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
	bootstrap.InitPlugins()

	app := cli.NewApp()
	app.Name = "doodad"
	app.Usage = "command line interface for Doodle"

	app.Version = fmt.Sprintf("%s build %s. Built on %s",
		builds.Version,
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
		commands.Convert,
		commands.Show,
		commands.Resave,
		commands.EditLevel,
		commands.EditDoodad,
		commands.InstallScript,
		commands.LevelPack,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Error("Fatal: %s", err)
		os.Exit(1)
	}
}

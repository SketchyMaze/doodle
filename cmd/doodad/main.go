// doodad is the command line developer tool for Doodle.
package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"git.kirsle.net/apps/doodle/cmd/doodad/commands"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/license"
	"git.kirsle.net/apps/doodle/pkg/log"
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
	app.Name = "doodad"
	app.Usage = "command line interface for Doodle"

	var freeLabel string
	if !license.IsRegistered() {
		freeLabel = " (shareware)"
	}

	app.Version = fmt.Sprintf("%s build %s%s. Built on %s",
		branding.Version,
		Build,
		freeLabel,
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

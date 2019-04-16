package commands

import (
	"fmt"
	"io/ioutil"

	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/urfave/cli"
)

// InstallScript to add the script to a doodad file.
var InstallScript cli.Command

func init() {
	InstallScript = cli.Command{
		Name:      "install-script",
		Usage:     "install the JavaScript source to a doodad",
		ArgsUsage: "<index.js> <filename.doodad>",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "key",
				Usage: "chroma key color for transparency on input image files",
				Value: "#ffffff",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() != 2 {
				return cli.NewExitError(
					"Usage: doodad install-script <script.js> <filename.doodad>",
					1,
				)
			}

			var (
				args       = c.Args()
				scriptFile = args[0]
				doodadFile = args[1]
			)

			// Read the JavaScript source.
			javascript, err := ioutil.ReadFile(scriptFile)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			doodad, err := doodads.LoadJSON(doodadFile)
			if err != nil {
				return cli.NewExitError(
					fmt.Sprintf("Failed to read doodad file: %s", err),
					1,
				)
			}
			doodad.Script = string(javascript)
			doodad.WriteJSON(doodadFile)
			log.Info("Installed script successfully")

			return nil
		},
	}
}
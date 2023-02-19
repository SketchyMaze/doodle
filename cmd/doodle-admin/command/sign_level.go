package command

import (
	"fmt"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/license"
	"git.kirsle.net/SketchyMaze/doodle/pkg/license/levelsigning"
	"github.com/urfave/cli/v2"
)

// SignLevel a license key for Sketchy Maze.
var SignLevel *cli.Command

func init() {
	SignLevel = &cli.Command{
		Name:  "sign-level",
		Usage: "sign a level file so that it may use embedded assets in free versions of the game.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Aliases:  []string{"k"},
				Usage:    "Private key .pem file for signing",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "Input file name (.level or .levelpack)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file, default outputs to console",
			},
		},
		Action: func(c *cli.Context) error {
			key, err := license.AdminLoadPrivateKey(c.String("key"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			var (
				filename = c.String("input")
				output   = c.String("output")
			)
			if output == "" {
				output = filename
			}

			// Sign a level?
			if strings.HasSuffix(filename, ".level") {
				lvl, err := level.LoadJSON(filename)
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				// Sign it.
				if sig, err := levelsigning.SignLevel(key, lvl); err != nil {
					return cli.Exit(fmt.Errorf("couldn't sign level: %s", err), 1)
				} else {
					lvl.Signature = sig
					err := lvl.WriteFile(output)
					if err != nil {
						return cli.Exit(err.Error(), 1)
					}
				}
			} else if strings.HasSuffix(filename, ".levelpack") {
				lp, err := levelpack.LoadFile(filename)
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				// Sign it.
				if sig, err := levelsigning.SignLevelPack(key, lp); err != nil {
					return cli.Exit(fmt.Errorf("couldn't sign levelpack: %s", err), 1)
				} else {
					lp.Signature = sig
					err := lp.WriteZipfile(output)
					if err != nil {
						return cli.Exit(err.Error(), 1)
					}
				}
			}

			return nil
		},
	}
}

package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"github.com/urfave/cli/v2"
)

// Resave a Level or Doodad to adapt to file format upgrades.
var Resave *cli.Command

func init() {
	Resave = &cli.Command{
		Name:      "resave",
		Usage:     "load and re-save a level or doodad file to migrate to newer file format versions",
		ArgsUsage: "<.level or .doodad>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "actors",
				Usage: "print verbose actor data in Level files",
			},
			&cli.BoolFlag{
				Name:  "chunks",
				Usage: "print verbose data about all the pixel chunks in a file",
			},
			&cli.BoolFlag{
				Name:  "script",
				Usage: "print the script from a doodad file and exit",
			},
			&cli.StringFlag{
				Name:    "attachment",
				Aliases: []string{"a"},
				Usage:   "print the contents of the attached filename to terminal",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "print verbose output (all verbose flags enabled)",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit(
					"Usage: doodad resave <.level .doodad ...>",
					1,
				)
			}

			filenames := c.Args().Slice()
			for _, filename := range filenames {
				switch strings.ToLower(filepath.Ext(filename)) {
				case enum.LevelExt:
					if err := resaveLevel(c, filename); err != nil {
						log.Error(err.Error())
						return cli.Exit("Error", 1)
					}
				case enum.DoodadExt:
					if err := resaveDoodad(c, filename); err != nil {
						log.Error(err.Error())
						return cli.Exit("Error", 1)
					}
				default:
					log.Error("File %s: not a level or doodad", filename)
				}
			}
			return nil
		},
	}
}

// resaveLevel shows data about a level file.
func resaveLevel(c *cli.Context, filename string) error {
	lvl, err := level.LoadJSON(filename)
	if err != nil {
		return err
	}

	log.Info("Loaded level from file: %s", filename)
	log.Info("Last saved game version: %s", lvl.GameVersion)

	log.Info("Saving back to disk")
	if err := lvl.WriteJSON(filename); err != nil {
		return fmt.Errorf("couldn't write %s: %s", filename, err)
	}
	return showLevel(c, filename)
}

func resaveDoodad(c *cli.Context, filename string) error {
	dd, err := doodads.LoadJSON(filename)
	if err != nil {
		return err
	}

	log.Info("Loaded doodad from file: %s", filename)
	log.Info("Last saved game version: %s", dd.GameVersion)

	log.Info("Saving back to disk")
	if err := dd.WriteJSON(filename); err != nil {
		return fmt.Errorf("couldn't write %s: %s", filename, err)
	}
	return showDoodad(c, filename)
}

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
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "write to a different file than the input",
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

	// Different output filename?
	if output := c.String("output"); output != "" {
		log.Info("Output will be saved to: %s", output)
		filename = output
	}

	if err := lvl.Vacuum(); err != nil {
		log.Error("Vacuum error: %s", err)
	} else {
		log.Info("Run vacuum on level file.")
	}

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

	// Different output filename?
	if output := c.String("output"); output != "" {
		log.Info("Output will be saved to: %s", output)
		filename = output
	}

	if err := dd.Vacuum(); err != nil {
		log.Error("Vacuum error: %s", err)
	} else {
		log.Info("Run vacuum on doodad file.")
	}

	log.Info("Saving back to disk")
	if err := dd.WriteJSON(filename); err != nil {
		return fmt.Errorf("couldn't write %s: %s", filename, err)
	}
	return showDoodad(c, filename)
}

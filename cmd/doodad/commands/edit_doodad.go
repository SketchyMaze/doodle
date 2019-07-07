package commands

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/urfave/cli"
)

// EditDoodad allows writing doodad metadata.
var EditDoodad cli.Command

func init() {
	EditDoodad = cli.Command{
		Name:      "edit-doodad",
		Usage:     "update metadata for a Doodad file",
		ArgsUsage: "<filename.doodad>",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "title",
				Usage: "set the doodad title",
			},
			cli.StringFlag{
				Name:  "author",
				Usage: "set the doodad author",
			},
			cli.BoolFlag{
				Name:  "hide",
				Usage: "Hide the doodad from the palette",
			},
			cli.BoolFlag{
				Name:  "unhide",
				Usage: "Unhide the doodad from the palette",
			},
			cli.BoolFlag{
				Name:  "lock",
				Usage: "write-lock the level file",
			},
			cli.BoolFlag{
				Name:  "unlock",
				Usage: "remove the write-lock on the level file",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.NewExitError(
					"Usage: doodad edit-doodad <filename.doodad>",
					1,
				)
			}

			var filenames = c.Args()
			for _, filename := range filenames {
				if err := editDoodad(c, filename); err != nil {
					log.Error("%s: %s", filename, err)
				}
			}

			return nil
		},
	}
}

func editDoodad(c *cli.Context, filename string) error {
	var modified bool

	dd, err := doodads.LoadJSON(filename)
	if err != nil {
		return fmt.Errorf("Failed to load %s: %s", filename, err)
	}

	log.Info("File: %s", filename)

	/***************************
	* Update level properties *
	***************************/

	if c.String("title") != "" {
		dd.Title = c.String("title")
		log.Info("Set title: %s", dd.Title)
		modified = true
	}

	if c.String("author") != "" {
		dd.Author = c.String("author")
		log.Info("Set author: %s", dd.Author)
		modified = true
	}

	if c.Bool("hide") {
		dd.Hidden = true
		log.Info("Marked doodad Hidden")
		modified = true
	} else if c.Bool("unhide") {
		dd.Hidden = false
		log.Info("Doodad is no longer Hidden")
		modified = true
	}

	if c.Bool("lock") {
		dd.Locked = true
		log.Info("Write lock enabled.")
		modified = true
	} else if c.Bool("unlock") {
		dd.Locked = false
		log.Info("Write lock disabled.")
		modified = true
	}

	/******************************
	* Save level changes to disk *
	******************************/

	if modified {
		if err := dd.WriteFile(filename); err != nil {
			return cli.NewExitError(fmt.Sprintf("Write error: %s", err), 1)
		}
	} else {
		log.Warn("Note: No changes made to level")
	}

	return showDoodad(c, filename)
}

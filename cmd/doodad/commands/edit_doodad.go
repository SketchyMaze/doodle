package commands

import (
	"fmt"
	"os"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/urfave/cli"
)

// EditDoodad allows writing doodad metadata.
var EditDoodad *cli.Command

func init() {
	EditDoodad = &cli.Command{
		Name:      "edit-doodad",
		Usage:     "update metadata for a Doodad file",
		ArgsUsage: "<filename.doodad>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "limit output (don't show doodad data at the end)",
			},
			&cli.StringFlag{
				Name:  "title",
				Usage: "set the doodad title",
			},
			&cli.StringFlag{
				Name:  "author",
				Usage: "set the doodad author",
			},
			&cli.StringSliceFlag{
				Name:    "tag",
				Aliases: []string{"t"},
				Usage:   "set a key/value tag on the doodad, in key=value format. Empty value deletes the tag.",
			},
			&cli.BoolFlag{
				Name:  "hide",
				Usage: "Hide the doodad from the palette",
			},
			&cli.BoolFlag{
				Name:  "unhide",
				Usage: "Unhide the doodad from the palette",
			},
			&cli.BoolFlag{
				Name:  "lock",
				Usage: "write-lock the level file",
			},
			&cli.BoolFlag{
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

			var filenames = c.Args().Slice()
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

	log.Info("Edit Doodad: %s", filename)

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

	// Tags.
	tags := c.StringSlice("tag")
	if len(tags) > 0 {
		for _, tag := range tags {
			parts := strings.SplitN(tag, "=", 2)
			if len(parts) != 2 {
				log.Error("--tag: must be in format `key=value`. Value may be blank to delete a tag. len=%d", len(parts))
				os.Exit(1)
			}

			var (
				key   = parts[0]
				value = parts[1]
			)
			if value == "" {
				log.Debug("Delete tag '%s'", key)
				delete(dd.Tags, key)
			} else {
				log.Debug("Set tag '%s' to '%s'", key, value)
				dd.Tags[key] = value
			}

			modified = true
		}
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
		if err := dd.WriteJSON(filename); err != nil {
			return cli.NewExitError(fmt.Sprintf("Write error: %s", err), 1)
		}
	} else {
		log.Warn("Note: No changes made to level")
	}

	if c.Bool("quiet") {
		return nil
	}

	return showDoodad(c, filename)
}

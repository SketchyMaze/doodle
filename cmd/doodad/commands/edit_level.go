package commands

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"github.com/urfave/cli/v2"
)

// EditLevel allows writing level metadata.
var EditLevel *cli.Command

func init() {
	EditLevel = &cli.Command{
		Name:      "edit-level",
		Usage:     "update metadata for a Level file",
		ArgsUsage: "<filename.level>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "limit output (don't show doodad data at the end)",
			},
			&cli.StringFlag{
				Name:  "title",
				Usage: "set the level title",
			},
			&cli.StringFlag{
				Name:  "author",
				Usage: "set the level author",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "set the level password",
			},
			&cli.StringFlag{
				Name:  "type",
				Usage: "set the page type. One of: Unbounded, Bounded, NoNegativeSpace, Bordered",
			},
			&cli.StringFlag{
				Name:  "max-size",
				Usage: "set the page max size (WxH format, like 2550x3300)",
			},
			&cli.StringFlag{
				Name:  "wallpaper",
				Usage: "set the wallpaper filename",
			},
			&cli.BoolFlag{
				Name:  "lock",
				Usage: "write-lock the level file",
			},
			&cli.BoolFlag{
				Name:  "unlock",
				Usage: "remove the write-lock on the level file",
			},
			&cli.StringFlag{
				Name:  "remove-actor",
				Usage: "Remove all instances of the actor from the level. Value is their filename or UUID.",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit(
					"Usage: doodad edit-level <filename.level>",
					1,
				)
			}

			var filenames = c.Args().Slice()
			for _, filename := range filenames {
				if err := editLevel(c, filename); err != nil {
					log.Error("%s: %s", filename, err)
				}
			}

			return nil
		},
	}
}

func editLevel(c *cli.Context, filename string) error {
	var modified bool

	lvl, err := level.LoadJSON(filename)
	if err != nil {
		return fmt.Errorf("Failed to load %s: %s", filename, err)
	}

	log.Info("File: %s", filename)

	/***************************
	* Update level properties *
	***************************/

	if c.String("title") != "" {
		lvl.Title = c.String("title")
		log.Info("Set title: %s", lvl.Title)
		modified = true
	}

	if c.String("author") != "" {
		lvl.Author = c.String("author")
		log.Info("Set author: %s", lvl.Author)
		modified = true
	}

	if c.String("password") != "" {
		lvl.Password = c.String("password")
		log.Info("Updated level password")
		modified = true
	}

	if c.String("max-size") != "" {
		w, h, err := render.ParseResolution(c.String("max-size"))
		if err != nil {
			log.Error("-max-size: %s", err)
		} else {
			lvl.MaxWidth = int64(w)
			lvl.MaxHeight = int64(h)
			modified = true
		}
	}

	if c.Bool("lock") {
		lvl.Locked = true
		log.Info("Write lock enabled.")
		modified = true
	} else if c.Bool("unlock") {
		lvl.Locked = false
		log.Info("Write lock disabled.")
		modified = true
	}

	if c.String("type") != "" {
		if pageType, ok := level.PageTypeFromString(c.String("type")); ok {
			lvl.PageType = pageType
			log.Info("Page Type set to %s", pageType)
			modified = true
		} else {
			log.Error("Invalid -type value. Should be like Unbounded, Bounded, NoNegativeSpace, Bordered")
		}
	}

	if c.String("wallpaper") != "" {
		lvl.Wallpaper = c.String("wallpaper")
		log.Info("Set wallpaper: %s", c.String("wallpaper"))
		modified = true
	}

	if c.String("remove-actor") != "" {
		var (
			match     = c.String("remove-actor")
			removeIDs = []string{}
		)

		for id, actor := range lvl.Actors {
			if id == match || actor.Filename == match {
				removeIDs = append(removeIDs, id)
			}
		}

		if len(removeIDs) > 0 {
			for _, id := range removeIDs {
				delete(lvl.Actors, id)
			}
			log.Info("Removed %d instances of actor %s from the level.", len(removeIDs), match)
			modified = true
		} else {
			log.Error("Did not find any actors like %s in the level.", match)
		}
	}

	/******************************
	* Save level changes to disk *
	******************************/

	if modified {
		if err := lvl.WriteFile(filename); err != nil {
			return cli.Exit(fmt.Sprintf("Write error: %s", err), 1)
		}
	} else {
		log.Warn("Note: No changes made to level")
	}

	if c.Bool("quiet") {
		return nil
	}

	return showLevel(c, filename)
}

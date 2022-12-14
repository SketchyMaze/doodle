package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"github.com/urfave/cli/v2"
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
			&cli.StringFlag{
				Name:  "hitbox",
				Usage: "set the doodad hitbox (X,Y,W,H or W,H format)",
			},
			&cli.StringFlag{
				Name:    "tag",
				Aliases: []string{"t"},
				Usage:   "set a key/value tag on the doodad, in key=value format. Empty value deletes the tag.",
			},
			&cli.StringFlag{
				Name:    "option",
				Aliases: []string{"o"},
				Usage:   "set an option on the doodad, in key=type=default format, e.g. active=bool=true, speed=int=10, name=str. Value types are bool, str, int.",
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
			&cli.BoolFlag{
				Name:  "touch",
				Usage: "simply load and re-save the doodad, to migrate it to a zipfile",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit(
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

	if c.Bool("touch") {
		log.Info("Just touching and resaving the file")
		modified = true
	}

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

	if c.String("hitbox") != "" {
		// Setting a hitbox, parse it out.
		parts := strings.Split(c.String("hitbox"), ",")
		var ints []int
		for _, part := range parts {
			a, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return err
			}
			ints = append(ints, a)
		}

		if len(ints) == 2 {
			dd.Hitbox = render.NewRect(ints[0], ints[1])
			modified = true
		} else if len(ints) == 4 {
			dd.Hitbox = render.Rect{
				X: ints[0],
				Y: ints[1],
				W: ints[2],
				H: ints[3],
			}
			modified = true
		} else {
			return cli.Exit("Hitbox should be in X,Y,W,H or just W,H format, 2 or 4 numbers.", 1)
		}
	}

	// Tags.
	tag := c.String("tag")
	if len(tag) > 0 {
		parts := strings.SplitN(tag, "=", 3)
		if len(parts) != 2 {
			log.Error("--tag: must be in format `key=value`. Value may be blank to delete a tag. len=%d tag=%s got=%+v", len(parts), tag, parts)
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

	// Options.
	opt := c.String("option")
	if len(opt) > 0 {
		parts := strings.SplitN(opt, "=", 3)
		if len(parts) < 2 {
			log.Error("--option: must be in format `name=type` or `name=type=value`")
			os.Exit(1)
		}

		var (
			name     = parts[0]
			dataType = parts[1]
			value    string
		)
		if len(parts) == 3 {
			value = parts[2]
		}

		// Validate the data types.
		if dataType != "bool" && dataType != "str" && dataType != "int" {
			log.Error("--option: invalid type, should be a bool, str or int")
			os.Exit(1)
		}

		value = dd.SetOption(name, dataType, value)
		log.Info("Set option %s (%s) = %s", name, dataType, value)

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
		if err := dd.WriteJSON(filename); err != nil {
			return cli.Exit(fmt.Sprintf("Write error: %s", err), 1)
		}
	} else {
		log.Warn("Note: No changes made to level")
	}

	if c.Bool("quiet") {
		return nil
	}

	return showDoodad(c, filename)
}

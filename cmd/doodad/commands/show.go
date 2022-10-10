package commands

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"github.com/urfave/cli/v2"
)

// Show information about a Level or Doodad file.
var Show *cli.Command

func init() {
	Show = &cli.Command{
		Name:      "show",
		Usage:     "show information about a level or doodad file",
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
					"Usage: doodad show <.level .doodad ...>",
					1,
				)
			}

			filenames := c.Args().Slice()
			for _, filename := range filenames {
				switch strings.ToLower(filepath.Ext(filename)) {
				case enum.LevelExt:
					if err := showLevel(c, filename); err != nil {
						log.Error(err.Error())
						return cli.Exit("Error", 1)
					}
				case enum.DoodadExt:
					if err := showDoodad(c, filename); err != nil {
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

// showLevel shows data about a level file.
func showLevel(c *cli.Context, filename string) error {
	lvl, err := level.LoadJSON(filename)
	if err != nil {
		return err
	}

	// Are we printing an attached file?
	if filename := c.String("attachment"); filename != "" {
		if data, err := lvl.GetFile(filename); err == nil {
			fmt.Print(string(data))
			return nil
		} else {
			fmt.Printf("Couldn't get attached file '%s': %s\n", filename, err)
			return err
		}
	}

	// Is it a new zipfile format?
	var fileType = "json or gzip"
	if lvl.Zipfile != nil {
		fileType = "zipfile"
	}

	fmt.Printf("===== Level: %s =====\n", filename)

	fmt.Println("Headers:")
	fmt.Printf("   File format: %s\n", fileType)
	fmt.Printf("  File version: %d\n", lvl.Version)
	fmt.Printf("  Game version: %s\n", lvl.GameVersion)
	fmt.Printf("   Level title: %s\n", lvl.Title)
	fmt.Printf("        Author: %s\n", lvl.Author)
	fmt.Printf("      Password: %s\n", lvl.Password)
	fmt.Printf("        Locked: %+v\n", lvl.Locked)
	fmt.Println("")

	fmt.Println("Game Rules:")
	fmt.Printf("  Difficulty: %s (%d)\n", lvl.GameRule.Difficulty, lvl.GameRule.Difficulty)
	fmt.Printf("    Survival: %+v\n", lvl.GameRule.Survival)
	fmt.Println("")

	showPalette(lvl.Palette)

	fmt.Println("Level Settings:")
	fmt.Printf("  Page type: %s\n", lvl.PageType.String())
	fmt.Printf("   Max size: %dx%d\n", lvl.MaxWidth, lvl.MaxHeight)
	fmt.Printf("  Wallpaper: %s\n", lvl.Wallpaper)
	fmt.Println("")

	fmt.Println("Attached Files:")
	if files := lvl.ListFiles(); len(files) > 0 {
		for _, v := range files {
			data, _ := lvl.GetFile(v)
			fmt.Printf("  %s: %d bytes\n", v, len(data))
		}
		fmt.Println("")
	} else {
		fmt.Printf("  None\n\n")
	}

	// Print the actor information.
	fmt.Println("Actors:")
	fmt.Printf("  Level contains %d actors\n", len(lvl.Actors))
	if c.Bool("actors") || c.Bool("verbose") {
		fmt.Println("  List of Actors:")
		for id, actor := range lvl.Actors {
			fmt.Printf("  -  Name: %s\n", actor.Filename)
			fmt.Printf("     UUID: %s\n", id)
			fmt.Printf("       At: %s\n", actor.Point)
			if len(actor.Options) > 0 {
				var ordered = []string{}
				for name := range actor.Options {
					ordered = append(ordered, name)
				}
				sort.Strings(ordered)

				fmt.Println("     Options:")
				for _, name := range ordered {
					val := actor.Options[name]
					fmt.Printf("         %s %s = %v\n", val.Type, val.Name, val.Value)
				}
			}
			if c.Bool("links") {
				for _, link := range actor.Links {
					if other, ok := lvl.Actors[link]; ok {
						fmt.Printf("     Link: %s (%s)\n", link, other.Filename)
					} else {
						fmt.Printf("     Link: %s (**UNRESOLVED**)", link)
					}
				}
			}
		}
		fmt.Println("")
	} else {
		fmt.Print("  Use -actors or -verbose to serialize Actors\n\n")
	}

	// Serialize chunk information.
	showChunker(c, lvl.Chunker)

	fmt.Println("")
	return nil
}

func showDoodad(c *cli.Context, filename string) error {
	dd, err := doodads.LoadJSON(filename)
	if err != nil {
		return err
	}

	if c.Bool("script") {
		fmt.Printf("// %s.js\n", filename)
		fmt.Println(strings.TrimSpace(dd.Script))
		return nil
	}

	// Is it a new zipfile format?
	var fileType = "json or gzip"
	if dd.Zipfile != nil {
		fileType = "zipfile"
	}

	fmt.Printf("===== Doodad: %s =====\n", filename)

	fmt.Println("Headers:")
	fmt.Printf("   File format: %s\n", fileType)
	fmt.Printf("  File version: %d\n", dd.Version)
	fmt.Printf("  Game version: %s\n", dd.GameVersion)
	fmt.Printf("  Doodad title: %s\n", dd.Title)
	fmt.Printf("        Author: %s\n", dd.Author)
	fmt.Printf("        Hitbox: %s\n", dd.Hitbox)
	fmt.Printf("        Locked: %+v\n", dd.Locked)
	fmt.Printf("        Hidden: %+v\n", dd.Hidden)
	fmt.Printf("   Script size: %d bytes\n", len(dd.Script))
	fmt.Println("")

	if len(dd.Tags) > 0 {
		fmt.Println("Tags:")
		for k, v := range dd.Tags {
			fmt.Printf("  %s: %s\n", k, v)
		}
		fmt.Println("")
	}

	if len(dd.Options) > 0 {
		var ordered = []string{}
		for name := range dd.Options {
			ordered = append(ordered, name)
		}
		sort.Strings(ordered)

		fmt.Println("Options:")
		for _, name := range ordered {
			opt := dd.Options[name]
			fmt.Printf("   %s %s = %v\n", opt.Type, opt.Name, opt.Default)
		}
		fmt.Println("")
	}

	showPalette(dd.Palette)

	for i, layer := range dd.Layers {
		fmt.Printf("Layer %d: %s\n", i, layer.Name)
		showChunker(c, layer.Chunker)
	}

	fmt.Println("")
	return nil
}

func showPalette(pal *level.Palette) {
	fmt.Println("Palette:")
	for _, sw := range pal.Swatches {
		fmt.Printf("  - Swatch name: %s\n", sw.Name)
		fmt.Printf("    Attributes:  %s\n", sw.Attributes())
		fmt.Printf("    Color:       %s\n", sw.Color.ToHex())
	}
	fmt.Println("")
}

func showChunker(c *cli.Context, ch *level.Chunker) {
	var worldSize = ch.WorldSize()
	var width = worldSize.W - worldSize.X
	var height = worldSize.H - worldSize.Y
	fmt.Println("Chunks:")
	fmt.Printf("  Pixels Per Chunk: %d^2\n", ch.Size)
	fmt.Printf("  Number Generated: %d\n", len(ch.Chunks))
	fmt.Printf("  Coordinate Range: (%d,%d) ... (%d,%d)\n",
		worldSize.X,
		worldSize.Y,
		worldSize.W,
		worldSize.H,
	)
	fmt.Printf("  World Dimensions: %dx%d\n", width, height)

	// Verbose chunk information.
	if c.Bool("chunks") || c.Bool("verbose") {
		fmt.Println("  Chunk Details:")
		for point, chunk := range ch.Chunks {
			fmt.Printf("  - Coord: %s\n", point)
			fmt.Printf("     Type: %s\n", chunkTypeToName(chunk.Type))
			fmt.Printf("    Range: (%d,%d) ... (%d,%d)\n",
				int(point.X)*ch.Size,
				int(point.Y)*ch.Size,
				(int(point.X)*ch.Size)+ch.Size,
				(int(point.Y)*ch.Size)+ch.Size,
			)
		}
	} else {
		fmt.Println("  Use -chunks or -verbose to serialize Chunks")
	}
	fmt.Println("")
}

func chunkTypeToName(v int) string {
	switch v {
	case level.MapType:
		return "map"
	case level.GridType:
		return "grid"
	default:
		return fmt.Sprintf("type %d", v)
	}
}

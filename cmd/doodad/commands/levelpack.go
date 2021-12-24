package commands

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/levelpack"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"github.com/urfave/cli/v2"
)

// LevelPack creation and management.
var LevelPack *cli.Command

func init() {
	LevelPack = &cli.Command{
		Name:      "levelpack",
		Usage:     "create and manage .levelpack archives",
		ArgsUsage: "-o output.levelpack <list of .level files>",
		Subcommands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "create a new .levelpack file from source files",
				ArgsUsage: "<output.levelpack> <input.level> [input.level...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "title",
						Aliases: []string{"t"},
						Usage:   "set a title for your levelpack, default will use the first level's title",
					},
					&cli.StringFlag{
						Name:    "author",
						Aliases: []string{"a"},
						Usage:   "set an author for your levelpack, default will use the first level's author",
					},
					&cli.StringFlag{
						Name:    "description",
						Aliases: []string{"d"},
						Usage:   "set a description for your levelpack",
					},
					&cli.IntFlag{
						Name:    "free",
						Aliases: []string{"f"},
						Usage:   "set number of free levels (levels unlocked by default), 0 means all unlocked",
					},
					&cli.StringFlag{
						Name:    "doodads",
						Aliases: []string{"D"},
						Usage:   "which doodads to embed: none, custom, all",
						Value:   "all",
					},
				},
				Action: levelpackCreate,
			},
			{
				Name:      "show",
				Usage:     "print details about a levelpack file",
				ArgsUsage: "<input.levelpack>",
				Action:    levelpackShow,
			},
		},
		Flags: []cli.Flag{},
	}
}

// Subcommand `levelpack show`
func levelpackShow(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit(
			"Usage: doodad levelpack show <file.levelpack>",
			1,
		)
	}

	var filename = c.Args().Slice()[0]
	if !strings.HasSuffix(filename, ".levelpack") {
		return cli.Exit("file must name a .levelpack", 1)
	}

	lp, err := levelpack.LoadFile(filename)
	if err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Printf("===== Levelpack: %s =====\n", filename)

	fmt.Println("Headers:")
	fmt.Printf("        Title: %s\n", lp.Title)
	fmt.Printf("       Author: %s\n", lp.Author)
	fmt.Printf("  Description: %s\n", lp.Description)
	fmt.Printf("  Free levels: %d\n", lp.FreeLevels)

	// List the levels.
	fmt.Println("\nLevels:")
	for i, lvl := range lp.Levels {
		fmt.Printf("%d. %s: %s\n", i+1, lvl.Filename, lvl.Title)
	}

	// List the doodads.
	dl := lp.ListFiles("doodads/")
	if len(dl) > 0 {
		fmt.Println("\nDoodads:")
		for i, doodad := range dl {
			fmt.Printf("%d. %s\n", i, doodad)
		}
	}

	return nil
}

// Subcommand `levelpack create`
func levelpackCreate(c *cli.Context) error {
	if c.NArg() < 2 {
		return cli.Exit(
			"Usage: doodad levelpack create <out.levelpack> <in.level ...>",
			1,
		)
	}

	var (
		args         = c.Args().Slice()
		outfile      = args[0]
		infiles      = args[1:]
		title        = c.String("title")
		author       = c.String("author")
		description  = c.String("description")
		free         = c.Int("free")
		embedDoodads = c.String("doodads")
	)

	// Validate params.
	if !strings.HasSuffix(outfile, ".levelpack") {
		return cli.Exit("Output file must have a .levelpack extension", 1)
	}
	if embedDoodads != "none" && embedDoodads != "custom" && embedDoodads != "all" {
		return cli.Exit(
			"--doodads: must be one of all, custom, none",
			1,
		)
	}

	var lp = levelpack.LevelPack{
		Title:       title,
		Author:      author,
		Description: description,
		FreeLevels:  free,
		Created:     time.Now().UTC(),
	}

	// Create a temp directory to work with.
	workdir, err := os.MkdirTemp(userdir.CacheDirectory, "levelpack-*")
	if err != nil {
		return cli.Exit(
			fmt.Sprintf("Couldn't make temp folder: %s", err),
			1,
		)
	}
	log.Info("Working directory: %s", workdir)
	defer os.RemoveAll(workdir)

	// Useful folders inside the working directory.
	var (
		levelDir  = filepath.Join(workdir, "levels")
		doodadDir = filepath.Join(workdir, "doodads")
		assets    = []string{
			"index.json",
		}
	)
	os.MkdirAll(levelDir, 0755)
	os.MkdirAll(doodadDir, 0755)

	// Get the list of the game's builtin doodads.
	builtins, err := doodads.ListBuiltin()
	if err != nil {
		return cli.Exit(err, 1)
	}

	// Read the input levels.
	for i, filename := range infiles {
		if !strings.HasSuffix(filename, ".level") {
			return cli.Exit(
				fmt.Sprintf("input file at position %d (%s) was not a .level file", i, filename),
				1,
			)
		}

		lvl, err := level.LoadJSON(filename)
		if err != nil {
			return cli.Exit(
				fmt.Sprintf("%s: %s", filename, err),
				1,
			)
		}

		// Fill in defaults for --title, --author
		if lp.Title == "" {
			lp.Title = lvl.Title
		}
		if lp.Author == "" {
			lp.Author = lvl.Author
		}

		// Log the level in the index.json list.
		lp.Levels = append(lp.Levels, levelpack.Level{
			Title:    lvl.Title,
			Author:   lvl.Author,
			Filename: filepath.Base(filename),
		})

		// Grab all the level's doodads to embed in the zip folder.
		for _, actor := range lvl.Actors {
			// What was the user's embeds request? (--doodads)
			if embedDoodads == "none" {
				break
			} else if embedDoodads == "custom" {
				// Custom doodads only.
				if isBuiltinDoodad(builtins, actor.Filename) {
					log.Warn("Doodad %s is a built-in, skipping embed", actor.Filename)
					continue
				}
			}

			if _, err := os.Stat(filepath.Join(doodadDir, actor.Filename)); !os.IsNotExist(err) {
				continue
			}

			log.Info("New doodad: %s", actor.Filename)

			// Get this doodad from the game's built-ins or the user's
			// profile directory only. Pulling embedded doodads out of
			// the level is NOT supported.
			asset, err := doodads.LoadFile(actor.Filename)
			if err != nil {
				return cli.Exit(
					fmt.Sprintf("%s: Doodad file '%s': %s", filename, asset.Filename, err),
					1,
				)
			}

			var targetFile = filepath.Join(doodadDir, actor.Filename)
			assets = append(assets, targetFile)
			log.Debug("Write doodad: %s", targetFile)
			err = asset.WriteFile(filepath.Join(doodadDir, actor.Filename))
			if err != nil {
				return cli.Exit(
					fmt.Sprintf("Writing doodad %s: %s", actor.Filename, err),
					1,
				)
			}
		}

		// Copy the level in.
		var targetFile = filepath.Join(levelDir, filepath.Base(filename))
		assets = append(assets, targetFile)
		log.Info("Write level: %s", filename)
		err = copyFile(filename, filepath.Join(levelDir, filepath.Base(filename)))
		if err != nil {
			return cli.Exit(
				fmt.Sprintf("couldn't copy %s to %s: %s", filename, targetFile, err),
				1,
			)
		}
	}

	log.Info("Writing index.json")
	if err := lp.WriteFile(filepath.Join(workdir, "index.json")); err != nil {
		return cli.Exit(err, 1)
	}

	// Zip the levelpack directory.
	log.Info("Creating levelpack file: %s", outfile)
	zipf, err := os.Create(outfile)
	if err != nil {
		return cli.Exit(
			fmt.Sprintf("failed to create %s: %s", outfile, err),
			1,
		)
	}

	zipper := zip.NewWriter(zipf)
	defer zipper.Close()

	// Embed all the assets.
	sort.Strings(assets)
	for _, asset := range assets {
		asset = strings.TrimPrefix(asset, workdir+"/")
		log.Info("Zip: %s", asset)
		err := zipFile(zipper, asset, filepath.Join(workdir, asset))
		if err != nil {
			return cli.Exit(err, 1)
		}
	}

	log.Info("Written: %s", outfile)
	return cli.Exit("", 0)
}

// copyFile copies a file on disk to another location.
func copyFile(source, target string) error {
	input, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(target, input, 0644)
}

// zipFile reads a file on disk to add to a zip file.
// The `key` is the filepath inside the ZIP file, filename is the actual source file on disk.
func zipFile(zf *zip.Writer, key, filename string) error {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	writer, err := zf.Create(key)
	if err != nil {
		return err
	}

	_, err = writer.Write(input)
	return err
}

// Helper function to test whether a filename is part of the builtin doodads.
func isBuiltinDoodad(doodads []string, filename string) bool {
	for _, cmp := range doodads {
		if cmp == filename {
			return true
		}
	}
	return false
}

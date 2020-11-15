package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"time"

	doodle "git.kirsle.net/apps/doodle/pkg"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/bindata"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/sound"
	"git.kirsle.net/go/render/sdl"
	"github.com/urfave/cli/v2"

	_ "image/png"
)

// Build number is the git commit hash.
var (
	Build     = "<dynamic>"
	BuildDate string
)

func init() {
	if BuildDate == "" {
		BuildDate = time.Now().Format(time.RFC3339)
	}

	// Use all the CPU cores for collision detection and other load balanced
	// goroutine work in the app.
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	runtime.LockOSThread()

	app := cli.NewApp()
	app.Name = "doodle"
	app.Usage = fmt.Sprintf("%s - %s", branding.AppName, branding.Summary)

	var freeLabel string
	if balance.FreeVersion {
		freeLabel = " (shareware)"
	}

	app.Version = fmt.Sprintf("%s build %s%s. Built on %s",
		branding.Version,
		Build,
		freeLabel,
		BuildDate,
	)

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Usage:   "enable debug level logging",
		},
		&cli.BoolFlag{
			Name:    "edit",
			Aliases: []string{"e"},
			Usage:   "edit the map given on the command line (instead of play it)",
		},
		&cli.StringFlag{
			Name:    "window",
			Aliases: []string{"w"},
			Usage:   "set the window size (e.g. -w 1024x768) or special value: desktop, mobile, landscape, maximized",
		},
		&cli.BoolFlag{
			Name:  "guitest",
			Usage: "enter the GUI Test scene on startup",
		},
	}

	app.Action = func(c *cli.Context) error {
		var filename string
		if c.NArg() > 0 {
			filename = c.Args().Get(0)
		}

		// Setting a custom resolution?
		if c.String("window") != "" {
			if err := setResolution(c.String("window")); err != nil {
				panic(err)
			}
		}

		// SDL engine.
		engine := sdl.New(
			fmt.Sprintf("%s v%s", branding.AppName, branding.Version),
			balance.Width,
			balance.Height,
		)

		// Load the SDL fonts in from bindata storage.
		if fonts, err := bindata.AssetDir("assets/fonts"); err == nil {
			for _, file := range fonts {
				data, err := bindata.Asset("assets/fonts/" + file)
				if err != nil {
					panic(err)
				}

				sdl.InstallFont(file, data)
			}
		} else {
			panic(err)
		}

		// Preload all sound effects.
		sound.PreloadAll()

		game := doodle.New(c.Bool("debug"), engine)
		game.SetupEngine()
		if c.Bool("guitest") {
			game.Goto(&doodle.GUITestScene{})
		} else if filename != "" {
			if c.Bool("edit") {
				game.EditFile(filename)
			} else {
				game.PlayLevel(filename)
			}
		}

		// Maximizing the window? with `-w maximized`
		if c.String("window") == "maximized" {
			log.Info("Maximize main window")
			engine.Maximize()
		}

		game.Run()
		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
	}
}

func setResolution(value string) error {
	switch value {
	case "desktop", "maximized":
		return nil
	case "mobile":
		balance.Width = 375
		balance.Height = 812
	case "landscape":
		balance.Width = 812
		balance.Height = 375
	default:
		var re = regexp.MustCompile(`^(\d+?)x(\d+?)$`)
		m := re.FindStringSubmatch(value)
		if len(m) == 0 {
			return errors.New("--window: must be of the form WIDTHxHEIGHT, i.e. " +
				"1024x768, or special keywords desktop, mobile, or landscape.")
		}

		w, _ := strconv.Atoi(m[1])
		h, _ := strconv.Atoi(m[2])
		balance.Width = w
		balance.Height = h
	}
	return nil
}

package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"time"

	"git.kirsle.net/apps/doodle/assets"
	doodle "git.kirsle.net/apps/doodle/pkg"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/license"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/apps/doodle/pkg/sound"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
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

	// Seed the random number generator.
	rand.Seed(time.Now().UnixNano())
}

func main() {
	runtime.LockOSThread()

	app := cli.NewApp()
	app.Name = "doodle"
	app.Usage = fmt.Sprintf("%s - %s", branding.AppName, branding.Summary)

	var freeLabel string
	if !license.IsRegistered() {
		freeLabel = " (shareware)"
	}

	// Load user settings from disk ASAP.
	if err := usercfg.Load(); err != nil {
		log.Error("Error loading user settings (defaults will be used): %s", err)
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
		&cli.StringFlag{
			Name:  "chdir",
			Usage: "working directory for the game's runtime package",
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
		&cli.BoolFlag{
			Name:  "experimental",
			Usage: "enable experimental Feature Flags",
		},
		&cli.BoolFlag{
			Name:  "offline",
			Usage: "offline mode, disables check for new updates",
		},
	}

	app.Action = func(c *cli.Context) error {
		// --chdir into a different working directory? e.g. for Flatpak especially.
		if doodlePath := c.String("chdir"); doodlePath != "" {
			if err := os.Chdir(doodlePath); err != nil {
				log.Error("--chdir: couldn't enter '%s': %s", doodlePath, err)
				return err
			}
		}

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

		// Enable feature flags?
		if c.Bool("experimental") {
			balance.FeaturesOn()
		}

		// Offline mode?
		if c.Bool("offline") {
			shmem.OfflineMode = true
		}

		// SDL engine.
		engine := sdl.New(
			fmt.Sprintf("%s v%s", branding.AppName, branding.Version),
			balance.Width,
			balance.Height,
		)

		// Load the SDL fonts in from bindata storage.
		if fonts, err := assets.AssetDir("assets/fonts"); err == nil {
			for _, file := range fonts {
				data, err := assets.Asset("assets/fonts/" + file)
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

		// Log what Doodle thinks its working directory is, for debugging.
		pwd, _ := os.Getwd()
		log.Debug("PWD: %s", pwd)

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
		if !usercfg.Current.Initialized {
			usercfg.Current.HorizontalToolbars = true
		}
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

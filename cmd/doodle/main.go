package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/assets"
	doodle "git.kirsle.net/SketchyMaze/doodle/pkg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding/builds"
	"git.kirsle.net/SketchyMaze/doodle/pkg/chatbot"
	"git.kirsle.net/SketchyMaze/doodle/pkg/gamepad"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/bootstrap"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sound"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sprites"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	golog "git.kirsle.net/go/log"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/sdl"
	"github.com/urfave/cli/v2"
	sdl2 "github.com/veandco/go-sdl2/sdl"

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

	bootstrap.InitPlugins()

	app := cli.NewApp()
	app.Name = "doodle"
	app.Usage = fmt.Sprintf("%s - %s", branding.AppName, branding.Summary)

	// Load user settings from disk ASAP.
	if err := usercfg.Load(); err != nil {
		log.Error("Error loading user settings (defaults will be used): %s", err)
	}

	// Set default user settings.
	if usercfg.Current.CrosshairColor == render.Invisible {
		usercfg.Current.CrosshairColor = balance.DefaultCrosshairColor
		usercfg.Save()
	}

	// Set GameController style.
	gamepad.SetStyle(gamepad.Style(usercfg.Current.ControllerStyle))

	app.Version = fmt.Sprintf("%s build %s. Built on %s",
		builds.Version,
		Build,
		BuildDate,
	)

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Usage:   "enable debug level logging",
		},
		&cli.StringFlag{
			Name:    "log",
			Aliases: []string{"o"},
			Usage:   "path on disk to copy the game's standard output logs (default goes to your game profile directory)",
		},
		&cli.StringFlag{
			Name:  "pprof",
			Usage: "record pprof metrics to a filename",
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
			Name:    "touch",
			Aliases: []string{"t"},
			Usage:   "force TouchScreenMode to be on at all times, which hides the mouse cursor",
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
		// Set the log level now if debugging is enabled.
		if c.Bool("debug") {
			log.Logger.Config.Level = golog.DebugLevel
		}

		// Write the game's log to disk.
		if err := initLogFile(c.String("log")); err != nil {
			log.Error("Couldn't write logs to disk: %s", err)
		}

		log.Info("Starting %s %s", app.Name, app.Version)

		// Print registration information, + also this sets the DefaultAuthor field.
		if reg, err := dpp.Driver.GetRegistration(); err == nil {
			log.Info("Registered to %s", reg.Name)
		}

		// --chdir into a different working directory? e.g. for Flatpak especially.
		if err := setWorkingDirectory(c); err != nil {
			log.Error("Couldn't set working directory: %s", err)
		}

		// Recording pprof stats?
		if cpufile := c.String("pprof"); cpufile != "" {
			log.Info("Saving CPU profiling data to %s", cpufile)
			fh, err := os.Create(cpufile)
			if err != nil {
				log.Error("--pprof: can't create file: %s", err)
				return err
			}
			defer fh.Close()

			if err := pprof.StartCPUProfile(fh); err != nil {
				log.Error("pprof: %s", err)
				return err
			}
			defer pprof.StopCPUProfile()
		}

		var filename string
		if c.NArg() > 0 {
			filename = c.Args().Get(0)
		}

		// Setting a custom resolution?
		var maximize = true
		if c.String("window") != "" {
			if err := setResolution(c.String("window")); err != nil {
				panic(err)
			}
			maximize = false
		}

		// Enable feature flags?
		if c.Bool("experimental") || usercfg.Current.EnableFeatures {
			balance.FeaturesOn()
		}

		// Set other program flags.
		shmem.OfflineMode = c.Bool("offline")
		native.ForceTouchScreenModeAlwaysOn = c.Bool("touch")

		// SDL engine.
		engine := sdl.New(
			fmt.Sprintf("%s v%s", branding.AppName, branding.Version),
			balance.Width,
			balance.Height,
		)

		// Activate game controller event support.
		sdl2.GameControllerEventState(1)

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

		// Start with maximized window unless -w was given.
		if maximize {
			log.Info("Maximize window")
			engine.Maximize()
		}

		// Reload usercfg - if their settings.json doesn't exist, we try and pick a
		// default "hide touch hints" based on touch device presence - which is only
		// known after SetupEngine.
		usercfg.Load()

		// Hide the mouse cursor over the window, we draw our own sprite image for it.
		engine.ShowCursor(false)

		// Set the app window icon.
		if engine, ok := game.Engine.(*sdl.Renderer); ok {
			if icon, err := sprites.LoadImage(game.Engine, balance.WindowIcon); err == nil {
				engine.SetWindowIcon(icon.Image)
			} else {
				log.Error("Couldn't load WindowIcon (%s): %s", balance.WindowIcon, err)
			}
		}

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
		log.Info("Program's working directory is: %s", pwd)

		// Initialize the developer shell chatbot easter egg.
		chatbot.Setup()

		// Log some basic environment details.
		w, h := engine.WindowSize()
		log.Info("Window size: %dx%d", w, h)

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

// Set the app's working directory to find the runtime rtp assets.
func setWorkingDirectory(c *cli.Context) error {
	// If they used the --chdir CLI option, go there.
	if doodlePath := c.String("chdir"); doodlePath != "" {
		return os.Chdir(doodlePath)
	}

	var test = func(paths ...string) bool {
		paths = append(paths, filepath.Join("rtp", "Credits.txt"))
		_, err := os.Stat(filepath.Join(paths...))
		return err == nil
	}

	// If the rtp/ folder is already here, nothing is needed.
	if test() {
		return nil
	}

	// Get the path to the executable and search around from there.
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("couldn't find the path to current executable: %s", err)
	}
	exPath := filepath.Dir(ex)

	log.Debug("Trying to locate rtp/ folder relative to game's executable path: %s", exPath)

	// Test a few relative paths around the executable's folder.
	paths := []string{
		exPath,                                   // same directory, e.g. Linux /opt/sketchymaze root or Windows zipfile
		filepath.Join(exPath, ".."),              // parent directory, e.g. from the git clone root
		filepath.Join(exPath, "..", "Resources"), // e.g. in a macOS .app bundle.

		// Some well-known installed paths to check.
		"/opt/sketchymaze",       // Linux deb/rpm package
		"/app/share/sketchymaze", // Linux flatpak package
	}
	for _, testPath := range paths {
		if test(testPath) {
			log.Info("Found rtp folder in: %s", testPath)
			return os.Chdir(testPath)
		}
	}

	return nil
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

func initLogFile(filename string) error {
	// Default log file to disk goes to your profile directory.
	if filename == "" {
		filename = userdir.LogFile
	}

	fh, err := golog.NewFileTee(filename)
	if err != nil {
		return err
	}

	log.Logger.Config.Writer = fh
	return nil
}

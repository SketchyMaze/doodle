package command

import (
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/license"
	"git.kirsle.net/SketchyMaze/doodle/pkg/license/levelsigning"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"github.com/urfave/cli/v2"
)

// VerifyLevel a license key for Sketchy Maze.
var VerifyLevel *cli.Command

func init() {
	VerifyLevel = &cli.Command{
		Name:  "verify-level",
		Usage: "check the signature on a level or levelpack file.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Aliases:  []string{"k"},
				Usage:    "Public key .pem file that signed the level",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "filename",
				Aliases:  []string{"f"},
				Usage:    "File name of the .level or .levelpack",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			key, err := license.AdminLoadPublicKey(c.String("key"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			filename := c.String("filename")
			if strings.HasSuffix(filename, ".level") {
				lvl, err := level.LoadJSON(filename)
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				// Verify it.
				if ok := levelsigning.VerifyLevel(key, lvl); !ok {
					log.Error("Signature is not valid!")
					return cli.Exit("", 1)
				} else {
					log.Info("Level signature is OK!")
				}
			} else if strings.HasSuffix(filename, ".levelpack") {
				lp, err := levelpack.LoadFile(filename)
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}

				// Verify it.
				if ok := levelsigning.VerifyLevelPack(key, lp); !ok {
					log.Error("Signature is not valid!")
					return cli.Exit("", 1)
				} else {
					log.Info("Levelpack signature is OK!")
				}
			}

			return nil
		},
	}
}

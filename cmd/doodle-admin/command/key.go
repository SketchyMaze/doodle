package command

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/license"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"github.com/urfave/cli/v2"
)

// Key a license key for Sketchy Maze.
var Key *cli.Command

func init() {
	Key = &cli.Command{
		Name:  "key",
		Usage: "generate an admin ECDSA signing key",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "public",
				Usage:    "Filename to write the public key to (.pem)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "private",
				Usage:    "Filename to write the private key to (.pem)",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			key, err := license.AdminGenerateKeys()
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			err = license.AdminWriteKeys(key, c.String("private"), c.String("public"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			log.Info("Written private key: %s", c.String("private"))
			log.Info("Written public key: %s", c.String("public"))
			return nil
		},
	}
}

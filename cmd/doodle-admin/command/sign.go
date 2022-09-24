package command

import (
	"fmt"
	"io/ioutil"

	"git.kirsle.net/SketchyMaze/doodle/pkg/license"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"github.com/urfave/cli/v2"
)

// Sign a license key for Sketchy Maze.
var Sign *cli.Command

func init() {
	Sign = &cli.Command{
		Name:  "sign",
		Usage: "sign a license key for the paid version of Sketchy Maze.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Aliases:  []string{"k"},
				Usage:    "Private key .pem file for signing",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Usage:    "User name for certificate",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "email",
				Aliases:  []string{"e"},
				Usage:    "User email address",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file, default outputs to console",
			},
		},
		Action: func(c *cli.Context) error {
			key, err := license.AdminLoadPrivateKey(c.String("key"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			reg := license.Registration{
				Name:  c.String("name"),
				Email: c.String("email"),
			}

			result, err := license.AdminSignRegistration(key, reg)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			// Writing to an output file?
			if output := c.String("output"); output != "" {
				log.Info("Write to: %s", output)
				if err := ioutil.WriteFile(output, []byte(result), 0644); err != nil {
					return cli.Exit(err, 1)
				}
			} else {
				fmt.Println(result)
			}

			return nil
		},
	}
}

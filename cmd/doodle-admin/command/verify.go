package command

import (
	"io/ioutil"
	"time"

	"git.kirsle.net/apps/doodle/pkg/license"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/urfave/cli/v2"
)

// Verify a license key for Sketchy Maze.
var Verify *cli.Command

func init() {
	Verify = &cli.Command{
		Name:  "verify",
		Usage: "check the signature on a license key",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Aliases:  []string{"k"},
				Usage:    "Public key .pem file that signed the JWT",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "filename",
				Aliases:  []string{"f"},
				Usage:    "File name of the license file to validate",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			key, err := license.AdminLoadPublicKey(c.String("key"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			jwt, err := ioutil.ReadFile(c.String("filename"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			reg, err := license.Validate(key, string(jwt))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			log.Info("Registration valid")
			log.Info("      Name: %s", reg.Name)
			log.Info("     Email: %s", reg.Email)
			log.Info("    Issued: %s", time.Unix(reg.IssuedAt, 0))
			log.Info("       NBF: %s", time.Unix(reg.NotBefore, 0))
			log.Info("Raw:\n%+v", reg)

			return nil
		},
	}
}

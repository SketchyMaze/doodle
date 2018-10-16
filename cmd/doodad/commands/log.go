package commands

import "github.com/kirsle/golog"

var log *golog.Logger

func init() {
	log = golog.GetLogger("doodad")
	log.Configure(&golog.Config{
		Level:  golog.InfoLevel,
		Theme:  golog.DarkTheme,
		Colors: golog.ExtendedColor,
	})
}

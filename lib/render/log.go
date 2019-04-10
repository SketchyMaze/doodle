package render

import "github.com/kirsle/golog"

var log *golog.Logger

func init() {
	log = golog.GetLogger("doodle")
	log.Configure(&golog.Config{
		Level:  golog.DebugLevel,
		Theme:  golog.DarkTheme,
		Colors: golog.ExtendedColor,
	})
}

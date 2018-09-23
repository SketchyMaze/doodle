package doodle

import "github.com/kirsle/golog"

var log *golog.Logger

func init() {
	log = golog.GetLogger("doodle")
	log.Configure(&golog.Config{
		Level:      golog.DebugLevel,
		Theme:      golog.DarkTheme,
		Colors:     golog.ExtendedColor,
		TimeFormat: "2006-01-02 15:04:05.000000",
	})
}

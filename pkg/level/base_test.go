package level_test

import "github.com/kirsle/golog"

func init() {
	log := golog.GetLogger("doodle")
	log.Config.Level = golog.ErrorLevel
}

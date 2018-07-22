package scene

import "github.com/kirsle/golog"

var log *golog.Logger

func init() {
	log = golog.GetLogger("doodle")
}

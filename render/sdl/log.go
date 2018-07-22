package sdl

import "github.com/kirsle/golog"

var log *golog.Logger

// Verbose debug logging.
var (
	DebugMouseEvents = false
	DebugClickEvents = false
	DebugKeyEvents   = false
)

func init() {
	log = golog.GetLogger("doodle")
}

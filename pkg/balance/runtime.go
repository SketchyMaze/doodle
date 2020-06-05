package balance

import "runtime"

// Runtime environment settings.
var (
	Runtime rtc

	GuidebookPath = "./guidebook/index.html"
)

type rtc struct {
	Platform platform
	Arch     string
}

// Platform type the app was built for.
type platform string

// Runtime.Platform constants.
const (
	Linux   platform = "linux"
	Windows platform = "windows"
	Darwin  platform = "darwin"
	Web     platform = "web"
)

func init() {
	Runtime = rtc{
		Platform: platform(runtime.GOOS),
		Arch:     runtime.GOARCH,
	}
}

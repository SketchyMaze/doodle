package branding

import (
	"fmt"
	"runtime"
)

// Constants for branding and version information.
const (
	AppName   = "Sketchy Maze"
	Summary   = "A drawing-based maze game"
	Version   = "0.13.0"
	Website   = "https://www.sketchymaze.com"
	Copyright = "2022 Noah Petherbridge"
	Byline    = "a game by Noah Petherbridge."

	// Update check URL
	UpdateCheckJSON = "https://download.sketchymaze.com/version.json"
	GuidebookURL    = "https://www.sketchymaze.com/guidebook/"
)

// UserAgent controls the HTTP User-Agent header when the game checks
// for updates on startup, to collect basic statistics of which game
// versions are out in the wild. Only static data (the --version string)
// about version, architecture, build number is included but no user
// specific data.
func UserAgent() string {
	return fmt.Sprintf("%s v%s on %s/%s",
		AppName,
		Version,
		runtime.GOOS,
		runtime.GOARCH,
	)
}

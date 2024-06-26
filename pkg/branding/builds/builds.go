// Package builds handles build-specific branding strings.
package builds

import (
	"fmt"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
)

var (
	/*
		Version string for user display.

		It may look like the following:

		- "v1.2.3 (open source)" for FOSS builds of the game.
		- "v1.2.3 (shareware)" for unregistered Doodle++ builds.
		- "v1.2.3" for registered Doodle++ builds.
	*/
	Version       = branding.Version
	VersionSuffix = " (unknown)"
)

func init() {
	if !balance.DPP {
		VersionSuffix = " (open source)"
	} else if !dpp.Driver.IsRegistered() {
		VersionSuffix = " (shareware)"
	} else {
		VersionSuffix = " (registered)"
	}

	Version = fmt.Sprintf("v%s%s", branding.Version, VersionSuffix)
}

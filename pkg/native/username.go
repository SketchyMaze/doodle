package native

import (
	"os"

	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
)

var USER string = os.Getenv("USER")

/*
DefaultAuthor will return the local user's name to be the default Author
for levels and doodads they create.

If they have registered the game, use the name from their license JWT token.

Otherwise fall back to their native operating system user.
*/
func DefaultAuthor() string {
	// Are we registered? TODO: get from registration
	if dpp.Driver.IsRegistered() {
		if reg, err := dpp.Driver.GetRegistration(); err == nil {
			return reg.Name
		}
	}

	// Return OS username
	return os.Getenv("USER")
}

package native

import (
	"os"

	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
)

var USER string = os.Getenv("USER")

/*
DefaultAuthor will return the local user's name to be the default Author
for levels and doodads they create.

If they have registered the game, use the name from their license JWT token.

Otherwise fall back to their native operating system user.
*/
func DefaultAuthor() string {
	// Are we registered?
	if plus.IsRegistered() {
		if reg, err := plus.GetRegistration(); err == nil {
			return reg.Name
		}
	}

	// Return OS username
	return os.Getenv("USER")
}

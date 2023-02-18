package uix

import "strings"

// Some debugging functions for the Canvas reachable via dev console in-game.

// GetCanvasesByActorName searches a (level) canvas's installed actors and returns any of
// them having this Title or Filename, with filename being more precise.
func (c *Canvas) GetCanvasesByActorName(filename string) []*Canvas {
	var (
		byFilename = []*Canvas{}
		byTitle    = []*Canvas{}
		lower      = strings.ToLower(filename)
	)

	for _, a := range c.actors {
		var doodad = a.Doodad()
		if doodad.Filename == filename {
			byFilename = append(byFilename, a.Canvas)
		} else if strings.ToLower(doodad.Title) == lower {
			byTitle = append(byTitle, a.Canvas)
		}
	}

	return append(byFilename, byTitle...)
}

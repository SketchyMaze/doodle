package level

import (
	"bytes"
	"image"
	"image/png"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/go/ui"
)

// Helper functions to get a level's embedded screenshot PNGs as textures.

// HasScreenshot returns whether screenshots exist for the level.
func (lvl *Level) HasScreenshot() bool {
	var filenames = []string{
		balance.LevelScreenshotLargeFilename,
		balance.LevelScreenshotMediumFilename,
		balance.LevelScreenshotSmallFilename,
	}
	for _, filename := range filenames {
		if lvl.Files.Exists("assets/screenshots/" + filename) {
			return true
		}
	}
	return false
}

// GetScreenshotImage returns a screenshot from the level's data as a Go Image.
// The filename is like "large.png" or "medium.png" and is appended to "assets/screenshots"
func (lvl *Level) GetScreenshotImage(filename string) (image.Image, error) {
	data, err := lvl.Files.Get("assets/screenshots/" + filename)
	if err != nil {
		return nil, err
	}

	return png.Decode(bytes.NewBuffer(data))
}

// GetScreenshotImageAsUIImage returns a ui.Image texture of a screenshot.
func (lvl *Level) GetScreenshotImageAsUIImage(filename string) (*ui.Image, error) {
	// Have it cached recently?
	if lvl.cacheImages == nil {
		lvl.cacheImages = map[string]*ui.Image{}
	} else if img, ok := lvl.cacheImages[filename]; ok {
		return img, nil
	}

	img, err := lvl.GetScreenshotImage(filename)
	if err != nil {
		return nil, err
	}

	result, err := ui.ImageFromImage(img)
	if err != nil {
		return nil, err
	}

	lvl.cacheImages[filename] = result
	return result, nil
}

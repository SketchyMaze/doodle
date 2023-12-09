package giant_screenshot

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
	"golang.org/x/image/draw"
)

// CroppedScreenshot returns a rendered RGBA image of the level.
func CroppedScreenshot(lvl *level.Level, viewport render.Rect) (image.Image, error) {
	// Lock this to one user at a time.
	if locked {
		return nil, errors.New("a screenshot is still being processed; try later...")
	}
	locked = true
	defer func() {
		locked = false
	}()

	// How big will our image be?
	var (
		size      = render.NewRect(viewport.W-viewport.X, viewport.H-viewport.Y)
		chunkSize = int(lvl.Chunker.Size)
		// worldSize = viewport
	)

	// Create the image.
	img := image.NewRGBA(image.Rect(0, 0, size.W, size.H))

	// Render the wallpaper onto it.
	log.Debug("CroppedScreenshot: Render wallpaper to image (%s)...", size)
	img = WallpaperToImage(lvl, img, size.W, size.H, viewport.Point())

	// Render the chunks.
	log.Debug("CroppedScreenshot: Render level chunks...")
	for coord := range lvl.Chunker.IterViewportChunks(viewport) {
		if chunk, ok := lvl.Chunker.GetChunk(coord); ok {

			// Get this chunk's rendered bitmap.
			rgba, ok := chunk.CachedBitmap(render.Invisible).(*image.RGBA)
			if !ok {
				log.Error("CroppedScreenshot: couldn't turn chunk to RGBA")
			}
			log.Debug("Blot chunk %s onto image", coord)

			// Compute where on the output image to copy this bitmap to.
			dst := image.Pt(
				// (X * W) multiplies the chunk coord by its size,
				// Then subtract the Viewport (level scroll position)
				(coord.X*chunkSize)-viewport.X,
				(coord.Y*chunkSize)-viewport.Y,
			)

			log.Debug("Copy chunk: %s to %s", coord, dst)
			img = blotImage(img, rgba, dst)
		}
	}

	// Render the doodads.
	log.Debug("CroppedScreenshot: Render actors...")
	for _, actor := range lvl.Actors {
		doodad, err := doodads.LoadFromEmbeddable(actor.Filename, lvl, false)
		if err != nil {
			log.Error("CroppedScreenshot: Load doodad: %s", err)
			continue
		}

		// Offset the doodad position by the viewport (scroll position).
		drawAt := render.NewPoint(actor.Point.X, actor.Point.Y)
		drawAt.X -= viewport.X
		drawAt.Y -= viewport.Y

		// TODO: usually doodad sprites start at 0,0 and the chunkSize
		// is the same as their sprite size.
		if len(doodad.Layers) > 0 && doodad.Layers[0].Chunker != nil {
			var chunker = doodad.Layers[0].Chunker
			chunk, ok := chunker.GetChunk(render.Origin)
			if !ok {
				continue
			}

			// TODO: we always use RGBA but is risky:
			rgba, ok := chunk.CachedBitmap(render.Invisible).(*image.RGBA)
			if !ok {
				log.Error("CroppedScreenshot: couldn't turn chunk to RGBA")
			}
			img = blotImage(img, rgba, image.Pt(drawAt.X, drawAt.Y))
		}

	}

	return img, nil
}

// SaveCroppedScreenshot will take a screenshot and write it to a file on disk,
// returning the filename relative to ~/.config/doodle/screenshots
func SaveCroppedScreenshot(level *level.Level, viewport render.Rect) (string, error) {
	var filename = time.Now().Format("2006-01-02_15-04-05.png")

	img, err := CroppedScreenshot(level, viewport)
	if err != nil {
		return "", err
	}

	fh, err := os.Create(filepath.Join(userdir.ScreenshotDirectory, filename))
	if err != nil {
		return "", err
	}

	png.Encode(fh, img)
	return filename, nil
}

// UpdateLevelScreenshots will generate and embed the screenshot PNGs into the level data.
func UpdateLevelScreenshots(lvl *level.Level, scroll render.Point) error {
	// Take screenshots.
	large, medium, small, err := CreateLevelScreenshots(lvl, scroll)
	if err != nil {
		return err
	}

	// Save the images into the level's filesystem.
	for filename, img := range map[string]image.Image{
		balance.LevelScreenshotLargeFilename:  large,
		balance.LevelScreenshotMediumFilename: medium,
		balance.LevelScreenshotSmallFilename:  small,
	} {
		var fh = bytes.NewBuffer([]byte{})
		if err := png.Encode(fh, img); err != nil {
			return fmt.Errorf("encode %s: %s", filename, err)
		}

		log.Debug("UpdateLevelScreenshots: add %s", filename)
		lvl.Files.Set(
			fmt.Sprintf("assets/screenshots/%s", filename),
			fh.Bytes(),
		)
	}

	return nil
}

// CreateLevelScreenshots generates a screenshot to save with the level data.
//
// This is called by the editor upon level save, and outputs the screenshots that
// will be embedded within the level data itself.
//
// Returns the large, medium and small images.
func CreateLevelScreenshots(lvl *level.Level, scroll render.Point) (large, medium, small image.Image, err error) {
	// Viewport to screenshot.
	viewport := render.Rect{
		X: scroll.X,
		W: scroll.X + balance.LevelScreenshotLargeSize.W,
		Y: scroll.Y,
		H: scroll.Y + balance.LevelScreenshotLargeSize.H,
	}

	// Get the full size screenshot as an image.
	large, err = CroppedScreenshot(lvl, viewport)
	if err != nil {
		return
	}

	// Scale the medium and small versions.
	medium = Scale(large, image.Rect(0, 0, balance.LevelScreenshotMediumSize.W, balance.LevelScreenshotMediumSize.H), draw.ApproxBiLinear)
	small = Scale(large, image.Rect(0, 0, balance.LevelScreenshotSmallSize.W, balance.LevelScreenshotSmallSize.H), draw.ApproxBiLinear)
	return large, medium, small, nil
}

// Scale down an image. Example:
//
// scaled := Scale(src, image.Rect(0, 0, 200, 200), draw.ApproxBiLinear)
func Scale(src image.Image, rect image.Rectangle, scale draw.Scaler) image.Image {
	dst := image.NewRGBA(rect)
	copyRect := image.Rect(
		rect.Min.X,
		rect.Min.Y,
		rect.Min.X+rect.Max.X,
		rect.Min.Y+rect.Max.Y,
	)
	scale.Scale(dst, copyRect, src, src.Bounds(), draw.Over, nil)
	return dst
}

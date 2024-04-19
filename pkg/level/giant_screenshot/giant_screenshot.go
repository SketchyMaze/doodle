package giant_screenshot

import (
	"errors"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/SketchyMaze/doodle/pkg/wallpaper"
	"git.kirsle.net/go/render"
)

/*
Giant Screenshot functionality for the Level Editor.
*/

var locked bool

// GiantScreenshot returns a rendered RGBA image of the entire level.
//
// Only one thread should be doing this at a time. A sync.Mutex will cause
// an error to return if another goroutine is already in the process of
// generating a screenshot, and you'll have to wait and try later.
func GiantScreenshot(lvl *level.Level) (image.Image, error) {
	// Lock this to one user at a time.
	if locked {
		return nil, errors.New("a giant screenshot is still being processed; try later...")
	}
	locked = true
	defer func() {
		locked = false
	}()

	shmem.Flash("Saving a giant screenshot (this takes a moment)...")

	// How big will our image be?
	var (
		size                = lvl.Chunker.WorldSizePositive()
		chunkSize           = int(lvl.Chunker.Size)
		chunkLow, chunkHigh = lvl.Chunker.Bounds()
		worldSize           = render.Rect{
			X: chunkLow.X,
			Y: chunkLow.Y,
			W: chunkHigh.X,
			H: chunkHigh.Y,
		}
		x int
		y int
	)

	// Bounded levels: set the image output size precisely.
	if lvl.PageType == level.Bounded || lvl.PageType == level.Bordered {
		size = render.NewRect(int(lvl.MaxWidth), int(lvl.MaxHeight))
	}

	// Levels without negative space: set the lower chunk coord to 0,0
	if lvl.PageType > level.Unbounded {
		worldSize.X = 0
		worldSize.Y = 0
	}

	// Create the image.
	img := image.NewRGBA(image.Rect(0, 0, size.W, size.H))

	// Render the wallpaper onto it.
	log.Debug("GiantScreenshot: Render wallpaper to image (%s)...", size)
	img = WallpaperToImage(lvl, img, size.W, size.H, render.Origin)

	// Render the chunks.
	log.Debug("GiantScreenshot: Render level chunks...")
	for chunkX := worldSize.X; chunkX <= worldSize.W; chunkX++ {
		y = 0
		for chunkY := worldSize.Y; chunkY <= worldSize.H; chunkY++ {
			if chunk, ok := lvl.Chunker.GetChunk(render.NewPoint(chunkX, chunkY)); ok {
				// TODO: we always use RGBA but is risky:
				rgba, ok := chunk.CachedBitmap(render.Invisible).(*image.RGBA)
				if !ok {
					log.Error("GiantScreenshot: couldn't turn chunk to RGBA")
				}
				img = blotImage(img, rgba, image.Pt(x, y))
			}
			y += chunkSize
		}
		x += chunkSize
	}

	// Render the doodads.
	log.Debug("GiantScreenshot: Render actors...")
	for _, actor := range lvl.Actors {
		doodad, err := plus.DoodadFromEmbeddable(actor.Filename, lvl, false)
		if err != nil {
			log.Error("GiantScreenshot: Load doodad: %s", err)
			continue
		}

		// Offset the doodad position if the image is displaying
		// negative coordinates.
		drawAt := render.NewPoint(actor.Point.X, actor.Point.Y)
		if worldSize.X < 0 {
			var offset = render.AbsInt(worldSize.X) * chunkSize
			drawAt.X += offset
		}
		if worldSize.Y < 0 {
			var offset = render.AbsInt(worldSize.Y) * chunkSize
			drawAt.Y += offset
		}

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
				log.Error("GiantScreenshot: couldn't turn chunk to RGBA")
			}
			img = blotImage(img, rgba, image.Pt(drawAt.X, drawAt.Y))
		}

	}

	return img, nil
}

// SaveGiantScreenshot will take a screenshot and write it to a file on disk,
// returning the filename relative to ~/.config/doodle/screenshots
func SaveGiantScreenshot(level *level.Level) (string, error) {
	var filename = time.Now().Format("2006-01-02_15-04-05.png")

	img, err := GiantScreenshot(level)
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

// WallpaperToImage accurately draws the wallpaper into an Image.
//
// The image is assumed to have a rect of (0,0,width,height) and that
// width and height are positive. Used for the Giant Screenshot feature.
//
// Pass an offset point to 'scroll' the wallpaper (for the Cropped Screenshot).
func WallpaperToImage(lvl *level.Level, target *image.RGBA, width, height int, offset render.Point) *image.RGBA {
	wp, err := wallpaper.FromFile("assets/wallpapers/"+lvl.Wallpaper, lvl)
	if err != nil {
		log.Error("GiantScreenshot: wallpaper load: %s", err)
	}

	var size = wp.QuarterRect()
	var resultImage = target

	// Handling the scroll offsets: wallpapers are made up of small-ish
	// repeating squares so the scroll offset needs only be a modulus within
	// the size of one square.
	absOffset := offset
	if offset != render.Origin {
		offset.X = render.AbsInt(offset.X % size.W)
		offset.Y = render.AbsInt(offset.Y % size.H)
	}

	// Tile the repeat texture. Go one tile extra in case of offset.
	for x := 0; x < width+size.W; x += size.W {
		for y := 0; y < height+size.H; y += size.H {
			dst := image.Pt(x-offset.X, y-offset.Y)
			resultImage = blotImage(resultImage, wp.Repeat(), dst)
		}
	}

	// Tile the left edge for bounded lvls.
	if lvl.PageType > level.Unbounded {
		// The left edge (unless off screen)
		if absOffset.X < size.W {
			for y := 0; y < height; y += size.H {
				dst := image.Pt(0-offset.X, y-offset.Y)
				resultImage = blotImage(resultImage, wp.Left(), dst)
			}
		}

		// The top edge.
		if absOffset.Y < size.H {
			for x := 0; x < width; x += size.W {
				dst := image.Pt(x-offset.X, 0-offset.Y)
				resultImage = blotImage(resultImage, wp.Top(), dst)
			}
		}

		// The top left corner.
		if absOffset.X < size.W && absOffset.Y < size.H {
			resultImage = blotImage(resultImage, wp.Corner(), image.Pt(
				-offset.X,
				-offset.Y,
			))
		}
	}

	return resultImage
}

func blotImage(target, source *image.RGBA, offset image.Point) *image.RGBA {
	b := target.Bounds()
	newImg := image.NewRGBA(b)
	draw.Draw(newImg, b, target, image.Point{}, draw.Src)
	draw.Draw(newImg, source.Bounds().Add(offset), source, image.Point{}, draw.Over)
	return newImg
}

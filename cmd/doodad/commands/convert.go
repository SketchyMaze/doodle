package commands

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"image/png"

	"git.kirsle.net/apps/doodle/lib/render"
	doodle "git.kirsle.net/apps/doodle/pkg"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/urfave/cli"
	"golang.org/x/image/bmp"
)

// Convert between image files (png or bitmap) and Doodle drawing files (levels
// and doodads)
var Convert cli.Command

func init() {
	Convert = cli.Command{
		Name:      "convert",
		Usage:     "convert between images and Doodle drawing files",
		ArgsUsage: "<input> <output>",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "key",
				Usage: "chroma key color for transparency on input image files",
				Value: "#ffffff",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() != 2 {
				return cli.NewExitError(
					"Usage: doodad convert <input.png> <output.level>\n"+
						"   Image file types: png, bmp\n"+
						"   Drawing file types: level, doodad",
					1,
				)
			}

			// Parse the chroma key.
			chroma, err := render.HexColor(c.String("key"))
			if err != nil {
				return cli.NewExitError(
					"Chrome key not a valid color: "+err.Error(),
					1,
				)
			}

			args := c.Args()
			var (
				inputFile  = args[0]
				inputType  = strings.ToLower(filepath.Ext(inputFile))
				outputFile = args[1]
				outputType = strings.ToLower(filepath.Ext(outputFile))
			)

			if inputType == extPNG || inputType == extBMP {
				if outputType == extLevel || outputType == extDoodad {
					if err := imageToDrawing(c, chroma, inputFile, outputFile); err != nil {
						return cli.NewExitError(err.Error(), 1)
					}
					return nil
				}
				return cli.NewExitError("Image inputs can only output to Doodle drawings", 1)
			} else if inputType == extLevel || inputType == extDoodad {
				if outputType == extPNG || outputType == extBMP {
					if err := drawingToImage(c, chroma, inputFile, outputFile); err != nil {
						return cli.NewExitError(err.Error(), 1)
					}
					return nil
				}
				return cli.NewExitError("Doodle drawing inputs can only output to image files", 1)
			}

			return cli.NewExitError("File types must be: png, bmp, level, doodad", 1)
		},
	}
}

func imageToDrawing(c *cli.Context, chroma render.Color, inputFile, outputFile string) error {
	reader, err := os.Open(inputFile)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	img, format, err := image.Decode(reader)
	log.Info("format: %s", format)
	_ = img
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	// Get the bounding box information of the source image.
	var (
		bounds    = img.Bounds()
		imageSize = bounds.Size()
		chunkSize int // the square shape for Doodad chunk size
	)
	if imageSize.X > imageSize.Y {
		chunkSize = imageSize.X
	} else {
		chunkSize = imageSize.Y
	}

	// Generate the output drawing file.
	switch strings.ToLower(filepath.Ext(outputFile)) {
	case extDoodad:
		log.Info("Output is a Doodad file (chunk size %d): %s", chunkSize, outputFile)
		doodad := doodads.New(chunkSize)
		doodad.GameVersion = doodle.Version
		doodad.Title = "Converted Doodad"
		doodad.Author = os.Getenv("USER")
		doodad.Palette = imageToChunker(img, chroma, doodad.Layers[0].Chunker)

		err := doodad.WriteJSON(outputFile)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
	case extLevel:
		log.Info("Output is a Level file: %s", outputFile)

		lvl := level.New()
		lvl.GameVersion = doodle.Version
		lvl.Title = "Converted Level"
		lvl.Author = os.Getenv("USER")
		lvl.Palette = imageToChunker(img, chroma, lvl.Chunker)

		err := lvl.WriteJSON(outputFile)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
	default:
		return cli.NewExitError("invalid output file: not a Doodle drawing", 1)
	}

	return nil
}

func drawingToImage(c *cli.Context, chroma render.Color, inputFile, outputFile string) error {
	var palette *level.Palette
	var chunker *level.Chunker

	switch strings.ToLower(filepath.Ext(inputFile)) {
	case extLevel:
		log.Info("Load Level: %s", inputFile)
		m, err := level.LoadJSON(inputFile)
		if err != nil {
			return fmt.Errorf("load level: %s", err.Error())
		}
		chunker = m.Chunker
		palette = m.Palette
	case extDoodad:
		log.Info("Load Doodad: %s", inputFile)
		d, err := doodads.LoadJSON(inputFile)
		if err != nil {
			return fmt.Errorf("load doodad: %s", err.Error())
		}
		chunker = d.Layers[0].Chunker // TODO: layers
		palette = d.Palette
	default:
		return fmt.Errorf("%s: not a level or doodad file", inputFile)
	}

	_ = chunker
	_ = palette

	// Create an image for the full world size.
	canvas := chunker.WorldSizePositive()
	img := image.NewRGBA(image.Rectangle{
		Min: image.Point{
			X: int(canvas.X),
			Y: int(canvas.Y),
		},
		Max: image.Point{
			X: int(canvas.W),
			Y: int(canvas.H),
		},
	})

	// Blank out the pixels.
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			img.Set(x, y, render.White.ToColor())
		}
	}

	// Transcode all pixels onto it.
	for px := range chunker.IterPixels() {
		img.Set(int(px.X), int(px.Y), px.Swatch.Color.ToColor())
	}

	// Write the output file.
	switch strings.ToLower(filepath.Ext(outputFile)) {
	case ".png":
		fh, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer fh.Close()
		return png.Encode(fh, img)
	case ".bmp":
		fh, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer fh.Close()
		return bmp.Encode(fh, img)
	}

	return errors.New("not valid output image type")
}

// imageToChunker implements a generic transcoding of an image.Image to a Chunker
// and returns the Palette, ready to plug into a Doodad or Level drawing.
//
// img: input image like a PNG
// chroma: transparent color
func imageToChunker(img image.Image, chroma render.Color, chunker *level.Chunker) *level.Palette {
	var (
		palette = level.NewPalette()
		bounds  = img.Bounds()
	)

	// Cache a palette of unique colors as we go.
	var uniqueColor = map[string]*level.Swatch{}

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			px := img.At(x, y)
			color := render.FromColor(px)
			if color == chroma || color.Transparent() { // invisible pixel
				continue
			}

			// New color for the palette?
			swatch, ok := uniqueColor[color.String()]
			if !ok {
				log.Info("New color: %s", color)
				swatch = &level.Swatch{
					Name:  color.String(),
					Color: color,
				}
				uniqueColor[color.String()] = swatch
			}

			chunker.Set(render.NewPoint(int32(x), int32(y)), swatch)
		}
	}

	// Order the palette.
	var sortedColors []string
	for k := range uniqueColor {
		sortedColors = append(sortedColors, k)
	}
	sort.Strings(sortedColors)
	for _, hex := range sortedColors {
		palette.Swatches = append(palette.Swatches, uniqueColor[hex])
	}

	return palette
}
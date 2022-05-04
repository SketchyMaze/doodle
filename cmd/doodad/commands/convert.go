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

	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"github.com/urfave/cli/v2"
	"golang.org/x/image/bmp"
)

// Convert between image files (png or bitmap) and Doodle drawing files (levels
// and doodads)
var Convert *cli.Command

func init() {
	Convert = &cli.Command{
		Name:      "convert",
		Usage:     "convert between images and Doodle drawing files",
		ArgsUsage: "<input> <output>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "key",
				Usage: "chroma key color for transparency on input image files, e.g. #ffffff",
				Value: "",
			},
			&cli.StringFlag{
				Name:    "title",
				Aliases: []string{"t"},
				Usage:   "set the title of the level or doodad being created",
			},
			&cli.StringFlag{
				Name:    "palette",
				Aliases: []string{"p"},
				Usage:   "use a palette JSON to define color swatch properties",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				return cli.Exit(
					"Usage: doodad convert <input.png...> <output.doodad>\n"+
						"   Image file types: png, bmp\n"+
						"   Drawing file types: level, doodad",
					1,
				)
			}

			// Parse the chroma key.
			var chroma = render.Invisible
			if key := c.String("key"); key != "" {
				color, err := render.HexColor(c.String("key"))
				if err != nil {
					return cli.Exit(
						"Chrome key not a valid color: "+err.Error(),
						1,
					)
				}
				chroma = color
			}

			args := c.Args().Slice()
			var (
				inputFiles = args[:len(args)-1]
				inputType  = strings.ToLower(filepath.Ext(inputFiles[0]))
				outputFile = args[len(args)-1]
				outputType = strings.ToLower(filepath.Ext(outputFile))
			)

			if inputType == extPNG || inputType == extBMP {
				if outputType == extLevel || outputType == extDoodad {
					if err := imageToDrawing(c, chroma, inputFiles, outputFile); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				}
				return cli.Exit("Image inputs can only output to Doodle drawings", 1)
			} else if inputType == extLevel || inputType == extDoodad {
				if outputType == extPNG || outputType == extBMP {
					if err := drawingToImage(c, chroma, inputFiles, outputFile); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				}
				return cli.Exit("Doodle drawing inputs can only output to image files", 1)
			}

			return cli.Exit("File types must be: png, bmp, level, doodad", 1)
		},
	}
}

func imageToDrawing(c *cli.Context, chroma render.Color, inputFiles []string, outputFile string) error {
	// Read the source images. Ensure they all have the same boundaries.
	var (
		imageBounds image.Point
		chunkSize   int // the square shape for the Doodad chunk size
		images      []image.Image
	)

	for i, filename := range inputFiles {
		reader, err := os.Open(filename)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		img, format, err := image.Decode(reader)
		log.Info("Parsed image %d of %d. Format: %s", i+1, len(inputFiles), format)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		// Get the bounding box information of the source image.
		var (
			bounds    = img.Bounds()
			imageSize = bounds.Size()
		)

		// Validate all images are the same size.
		if i == 0 {
			imageBounds = imageSize
			if imageSize.X > imageSize.Y {
				chunkSize = imageSize.X
			} else {
				chunkSize = imageSize.Y
			}
		} else if imageSize != imageBounds {
			return cli.Exit("your source images are not all the same dimensions", 1)
		}

		images = append(images, img)
	}

	// Helper function to translate image filenames into layer names.
	toLayerName := func(filename string) string {
		ext := filepath.Ext(filename)
		return strings.TrimSuffix(filepath.Base(filename), ext)
	}

	// Generate the output drawing file.
	switch strings.ToLower(filepath.Ext(outputFile)) {
	case extDoodad:
		log.Info("Output is a Doodad file (chunk size %d): %s", chunkSize, outputFile)
		doodad := doodads.New(chunkSize)
		doodad.GameVersion = branding.Version
		doodad.Title = c.String("title")
		if doodad.Title == "" {
			doodad.Title = "Converted Doodad"
		}
		doodad.Author = os.Getenv("USER")

		// Write the first layer and gather its palette.
		log.Info("Converting first layer to drawing and getting the palette")
		palette, layer0 := imageToChunker(images[0], chroma, nil, chunkSize)
		doodad.Palette = palette
		doodad.Layers[0].Chunker = layer0
		doodad.Layers[0].Name = toLayerName(inputFiles[0])

		// Write any additional layers.
		if len(images) > 1 {
			for i := 1; i < len(images); i++ {
				img := images[i]
				log.Info("Converting extra layer %d", i)
				_, chunker := imageToChunker(img, chroma, palette, chunkSize)
				doodad.AddLayer(toLayerName(inputFiles[i]), chunker)
			}
		}

		err := doodad.WriteJSON(outputFile)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}
	case extLevel:
		log.Info("Output is a Level file: %s", outputFile)
		if len(images) > 1 {
			log.Warn("Notice: levels only support one layer so only your first image will be used")
		}

		lvl := level.New()
		lvl.GameVersion = branding.Version
		lvl.Title = c.String("title")
		if lvl.Title == "" {
			lvl.Title = "Converted Level"
		}
		lvl.Author = os.Getenv("USER")
		palette, chunker := imageToChunker(images[0], chroma, nil, lvl.Chunker.Size)
		lvl.Palette = palette
		lvl.Chunker = chunker

		err := lvl.WriteJSON(outputFile)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}
	default:
		return cli.Exit("invalid output file: not a Doodle drawing", 1)
	}

	return nil
}

func drawingToImage(c *cli.Context, chroma render.Color, inputFiles []string, outputFile string) error {
	var palette *level.Palette
	var chunker *level.Chunker
	inputFile := inputFiles[0]

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
func imageToChunker(img image.Image, chroma render.Color, palette *level.Palette, chunkSize int) (*level.Palette, *level.Chunker) {
	var (
		chunker = level.NewChunker(chunkSize)
		bounds  = img.Bounds()
	)

	if palette == nil {
		palette = level.NewPalette()
	}

	// Cache a palette of unique colors as we go.
	var uniqueColor = map[string]*level.Swatch{}
	var newColors = map[string]*level.Swatch{} // new ones discovered this time
	for _, swatch := range palette.Swatches {
		uniqueColor[swatch.Color.String()] = swatch
	}

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
				newColors[color.String()] = swatch
			}

			chunker.Set(render.NewPoint(x, y), swatch)
		}
	}

	// Order the palette.
	var sortedColors []string
	for k := range uniqueColor {
		sortedColors = append(sortedColors, k)
	}
	sort.Strings(sortedColors)
	for _, hex := range sortedColors {
		if _, ok := newColors[hex]; ok {
			palette.Swatches = append(palette.Swatches, uniqueColor[hex])
		}
	}
	palette.Inflate()

	return palette, chunker
}

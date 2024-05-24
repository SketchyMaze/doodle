// Package rle contains support for Run-Length Encoding of level chunks.
package rle

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/render"
)

const NullColor = 0xFFFF

// Grid is a 2D array of nullable integers to store a flat bitmap of a chunk.
type Grid [][]*uint64

// NewGrid will return an initialized 2D grid of equal dimensions of the given size.
//
// The grid is indexed in [Y][X] notation, or: by row first and then column.
func NewGrid(size int) (Grid, error) {
	if size == 0 {
		return nil, errors.New("no size given for RLE Grid: the chunker was probably not initialized")
	}

	var grid = make([][]*uint64, size+1)
	for i := 0; i < size+1; i++ {
		grid[i] = make([]*uint64, size+1)
	}

	return grid, nil
}

func MustGrid(size int) Grid {
	grid, err := NewGrid(size)
	if err != nil {
		panic(err)
	}
	return grid
}

type Pixel struct {
	Point   render.Point
	Palette int
}

// Size of the grid.
func (g Grid) Size() int {
	return len(g[0])
}

// Compress the grid into a byte stream of RLE compressed data.
//
// The compressed format is a stream of:
//
// - A Uvarint for the palette index (0-255) or 0xffff (65535) for null.
// - A Uvarint for how many pixels to repeat that color.
func (g Grid) Compress() ([]byte, error) {
	log.Error("BEGIN Compress()")
	// log.Warn("Visualized:\n%s", g.Visualize())

	// Run-length encode the grid.
	var (
		compressed []byte // final result
		lastColor  uint64 // last color seen (current streak)
		runLength  uint64 // current streak for the last color
		buffering  bool   // detect end of grid

		// Flush the buffer
		flush = func() {
			// log.Info("flush: %d for %d length", lastColor, runLength)
			compressed = binary.AppendUvarint(compressed, lastColor)
			compressed = binary.AppendUvarint(compressed, runLength)
		}
	)

	for y, row := range g {
		for x, nullableIndex := range row {
			var index uint64
			if nullableIndex == nil {
				index = NullColor
			} else {
				index = *nullableIndex
			}

			// First color of the grid
			if y == 0 && x == 0 {
				// log.Info("First color @ %dx%d is %d", x, y, index)
				lastColor = index
				runLength = 1
				continue
			}

			// Buffer it until we get a change of color or EOF.
			if index != lastColor {
				// log.Info("Color %d streaks for %d until %dx%d", lastColor, runLength, x, y)
				flush()
				lastColor = index
				runLength = 1
				buffering = false
				continue
			}

			buffering = true
			runLength++
		}
	}

	// Flush the final buffer when we got to EOF on the grid.
	if buffering {
		flush()
	}

	// log.Error("RLE compressed: %v", compressed)

	return compressed, nil
}

// Decompress the RLE byte stream back into a populated 2D grid.
func (g Grid) Decompress(compressed []byte) error {
	log.Error("BEGIN Decompress()")
	// log.Warn("Visualized:\n%s", g.Visualize())

	// Prepare the 2D grid to decompress the RLE stream into.
	var (
		size         = g.Size()
		x, y, cursor int
	)

	var reader = bytes.NewBuffer(compressed)

	for {
		var (
			paletteIndexRaw, err1 = binary.ReadUvarint(reader)
			repeatCount, err2     = binary.ReadUvarint(reader)
		)

		if err1 != nil || err2 != nil {
			break
		}

		// Handle the null color.
		var paletteIndex *uint64
		if paletteIndexRaw != NullColor {
			paletteIndex = &paletteIndexRaw
		}

		// log.Warn("RLE index %v for %dpx", paletteIndexRaw, repeatCount)

		for i := uint64(0); i < repeatCount; i++ {
			cursor++
			if cursor%size == 0 {
				y++
				x = 0
			}

			point := render.NewPoint(int(x), int(y))
			if point.Y >= size || point.X >= size {
				continue
			}
			g[point.Y][point.X] = paletteIndex

			x++
		}
	}

	// log.Warn("Visualized:\n%s", g.Visualize())

	return nil
}

// Visualize the state of the 2D grid.
func (g Grid) Visualize() string {
	var lines []string
	for _, row := range g {
		var line = "["
		for _, col := range row {
			if col == nil {
				line += " "
			} else {
				line += fmt.Sprintf("%x", *col)
			}
		}
		lines = append(lines, line+"]")
	}
	return strings.Join(lines, "\n")
}

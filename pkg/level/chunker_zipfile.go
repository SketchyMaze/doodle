package level

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/render"
)

// Zipfile interactions for the Chunker to cache and manage which
// chunks of large levels need be in active memory.

var (
	zipChunkfileRegexp = regexp.MustCompile(`^chunks/(\d+)/(.+?)\.(bin|json)$`)
)

// MigrateZipfile is called on save to migrate old-style ChunkMap
// chunks into external zipfile members and free up space in the
// master Level or Doodad struct.
func (c *Chunker) MigrateZipfile(zf *zip.Writer) error {
	// Identify if any chunks in active memory had been completely erased.
	var (
		// Chunks that have become empty and are to be REMOVED from zip.
		erasedChunks = map[render.Point]interface{}{}

		// Unique chunks we added to the zip file so we don't add duplicates.
		chunksZipped = map[render.Point]interface{}{}
	)
	for coord, chunk := range c.Chunks {
		if chunk.Len() == 0 {
			log.Debug("Chunker.MigrateZipfile: %s has become empty, remove from zip", coord)
			erasedChunks[coord] = nil
		}
	}

	// Copy all COLD STORED chunks from our original zipfile into the new one.
	// These are chunks that are NOT actively loaded (those are written next),
	// and erasedChunks are not written to the zipfile at all.
	if c.Zipfile != nil {
		log.Debug("MigrateZipfile: Copying chunk files from old zip to new zip")
		for _, file := range c.Zipfile.File {
			m := zipChunkfileRegexp.FindStringSubmatch(file.Name)
			if len(m) > 0 {
				var (
					mLayer, _ = strconv.Atoi(m[1])
					coord     = m[2]
					ext       = m[3]
				)

				// Will we need to do a format conversion now?
				var reencode bool
				if ext == "json" && balance.BinaryChunkerEnabled {
					reencode = true
				} else if ext == "bin" && !balance.BinaryChunkerEnabled {
					reencode = true
				}

				// Not our layer, not our problem.
				if mLayer != c.Layer {
					continue
				}

				point, err := render.ParsePoint(coord)
				if err != nil {
					return err
				}

				// Don't create zip files for empty (0 pixel) chunks.
				if _, ok := erasedChunks[point]; ok {
					log.Debug("Skip copying %s: chunk is empty", coord)
					continue
				}

				// Don't ever write duplicate files.
				if _, ok := chunksZipped[point]; ok {
					log.Debug("Skip copying duplicate chunk %s", coord)
					continue
				}
				chunksZipped[point] = nil

				// Don't copy the chunks we have currently in memory: those
				// are written next. Apparently zip files are allowed to
				// have duplicate named members!
				if _, ok := c.Chunks[point]; ok {
					log.Debug("Skip chunk %s (in memory)", coord)
					continue
				}

				// Verify that this chunk file in the old ZIP was not empty.
				chunk, err := c.ChunkFromZipfile(point)
				if err == nil && chunk.Len() == 0 {
					log.Debug("Skip chunk %s (old zipfile chunk was empty)", coord)
					continue
				}

				// Are we simply copying the existing chunk, or re-encoding it too?
				if reencode {
					log.Debug("Re-encoding existing chunk %s into target format", file.Name)
					if err := chunk.Inflate(c.pal); err != nil {
						return fmt.Errorf("couldn't inflate cold storage chunk for reencode: %s", err)
					}

					if err := chunk.ToZipfile(zf, mLayer, point); err != nil {
						return err
					}
				} else {
					log.Debug("Copy existing chunk %s", file.Name)
					if err := zf.Copy(file); err != nil {
						return err
					}
				}
			}
		}
	} else {
		log.Debug("Chunker.MigrateZipfile: the drawing did not give me a zipfile!")
	}

	if len(c.Chunks) == 0 {
		return nil
	}

	log.Debug("MigrateZipfile: chunker has %d in memory, exporting to zipfile", len(c.Chunks))

	// Flush in-memory chunks out to zipfile.
	for coord, chunk := range c.Chunks {
		if _, ok := erasedChunks[coord]; ok {
			continue
		}

		// Are we encoding chunks as JSON?
		log.Debug("Flush in-memory chunks %s to zip", coord)
		chunk.ToZipfile(zf, c.Layer, coord)
	}

	// Flush the chunkmap out.
	// TODO: do similar to move old attached files (wallpapers) too
	c.Chunks = ChunkMap{}

	return nil
}

// ClearChunkCache completely flushes the ChunkMap from memory. BE CAREFUL.
// If the level is a Zipfile the chunks will reload as needed, but old style
// levels this will nuke the whole drawing!
func (c *Chunker) ClearChunkCache() {
	c.chunkMu.Lock()
	c.Chunks = ChunkMap{}
	c.chunkMu.Unlock()
}

// CacheSize returns the number of chunks in memory.
func (c *Chunker) CacheSize() int {
	return len(c.Chunks)
}

// GCSize returns the number of chunks pending free (not accessed in 2+ ticks)
func (c *Chunker) GCSize() int {
	return len(c.chunksToFree)
}

// ToZipfile writes just a chunk's data into a zipfile.
//
// It will write a file like "chunks/{layer}/{coord}.json" if using JSON
// format or a .bin file for binary format based on the BinaryChunkerEnabled
// game config constant.
func (c *Chunk) ToZipfile(zf *zip.Writer, layer int, coord render.Point) error {
	// File name?
	ext := ".json"
	if balance.BinaryChunkerEnabled {
		ext = ".bin"
	}
	filename := fmt.Sprintf("chunks/%d/%s%s", layer, coord, ext)

	writer, err := zf.Create(filename)
	if err != nil {
		return err
	}

	// Are we writing it as binary format?
	var data []byte
	if balance.BinaryChunkerEnabled {
		if bytes, err := c.MarshalBinary(); err != nil {
			return err
		} else {
			data = bytes
		}
	} else {
		return errors.New("Chunk.ToZipfile: JSON chunk format no longer supported for writing")
	}

	// Write the file contents to zip whether binary or json.
	n, err := writer.Write(data)
	if err != nil {
		return err
	}

	log.Debug("Written chunk to zipfile: %s (%d bytes)", filename, n)
	return nil
}

// ChunkFromZipfile loads a chunk from a zipfile.
func (c *Chunker) ChunkFromZipfile(coord render.Point) (*Chunk, error) {
	// Grab the chunk (bin or json) from the Zipfile.
	ext, bin, err := c.RawChunkFromZipfile(coord)
	if err != nil {
		return nil, err
	}

	var chunk = NewChunk()
	chunk.Point = coord
	chunk.Size = c.Size

	switch ext {
	case ".bin":
		// New style .bin compressed format:
		// Either a MapAccessor compressed bin, or RLE compressed.
		err = chunk.UnmarshalBinary(bin)
		if err != nil {
			log.Error("ChunkFromZipfile(%s): %s", coord, err)
			return nil, err
		}
	case ".json":
		// Legacy style plain .json file (MapAccessor only).
		err = chunk.UnmarshalJSON(bin)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected filetype found for this chunk: %s", ext)
	}

	return chunk, nil
}

// RawChunkFromZipfile loads a chunk from a zipfile and returns its raw binary content.
//
// Returns the file extension (".bin" or ".json"), raw bytes, and an error.
func (c *Chunker) RawChunkFromZipfile(coord render.Point) (string, []byte, error) {
	// File names?
	var (
		zf    = c.Zipfile
		layer = c.Layer

		binfile  = fmt.Sprintf("chunks/%d/%s.bin", layer, coord)
		jsonfile = fmt.Sprintf("chunks/%d/%s.json", layer, coord)
	)

	// Read from the new binary format.
	if file, err := zf.Open(binfile); err == nil {
		data, err := io.ReadAll(file)
		return ".bin", data, err
	} else if file, err := zf.Open(jsonfile); err == nil {
		data, err := io.ReadAll(file)
		return ".json", data, err
	}

	return "", nil, errors.New("not found in zipfile")
}

// ChunksInZipfile returns the list of chunk coordinates in a zipfile.
func ChunksInZipfile(zf *zip.Reader, layer int) []render.Point {
	var (
		result = []render.Point{}
		sLayer = fmt.Sprintf("%d", layer)
	)

	for _, file := range zf.File {
		m := zipChunkfileRegexp.FindStringSubmatch(file.Name)
		if len(m) > 0 {
			var (
				mLayer = m[1]
				mPoint = m[2]
			)

			// Not our layer?
			if mLayer != sLayer {
				continue
			}

			if point, err := render.ParsePoint(mPoint); err == nil {
				result = append(result, point)
			} else {
				log.Error("ChunksInZipfile: file '%s' didn't parse as a point: %s", file.Name, err)
			}
		}
	}

	return result
}

// ChunkInZipfile tests whether the chunk exists in the zipfile.
func ChunkInZipfile(zf *zip.Reader, layer int, coord render.Point) bool {
	for _, chunk := range ChunksInZipfile(zf, layer) {
		if chunk == coord {
			return true
		}
	}
	return false
}

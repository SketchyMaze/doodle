package level

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
)

// Zipfile interactions for the Chunker to cache and manage which
// chunks of large levels need be in active memory.

var (
	zipChunkfileRegexp = regexp.MustCompile(`^chunks/(\d+)/(.+?)\.json$`)
)

// MigrateZipfile is called on save to migrate old-style ChunkMap
// chunks into external zipfile members and free up space in the
// master Level or Doodad struct.
func (c *Chunker) MigrateZipfile(zf *zip.Writer) error {
	// Identify if any chunks in active memory had been completely erased.
	var (
		erasedChunks = map[render.Point]interface{}{}
		chunksZipped = map[render.Point]interface{}{}
	)
	for coord, chunk := range c.Chunks {
		if chunk.Len() == 0 {
			log.Info("Chunker.MigrateZipfile: %s has become empty, remove from zip", coord)
			erasedChunks[coord] = nil
		}
	}

	// Copy all COLD STORED chunks from our original zipfile into the new one.
	// These are chunks that are NOT actively loaded (those are written next),
	// and erasedChunks are not written to the zipfile at all.
	if c.Zipfile != nil {
		log.Info("MigrateZipfile: Copying chunk files from old zip to new zip")
		for _, file := range c.Zipfile.File {
			m := zipChunkfileRegexp.FindStringSubmatch(file.Name)
			if len(m) > 0 {
				mLayer, _ := strconv.Atoi(m[1])
				coord := m[2]

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

				log.Info("Copy existing chunk %s", file.Name)
				if err := zf.Copy(file); err != nil {
					return err
				}
			}
		}
	} else {
		log.Warn("Chunker.MigrateZipfile: the drawing did not give me a zipfile!")
	}

	if len(c.Chunks) == 0 {
		return nil
	}

	log.Info("MigrateZipfile: chunker has %d in memory, exporting to zipfile", len(c.Chunks))

	// Flush in-memory chunks out to zipfile.
	for coord, chunk := range c.Chunks {
		filename := fmt.Sprintf("chunks/%d/%s.json", c.Layer, coord.String())
		log.Info("Flush in-memory chunks to %s", filename)
		chunk.ToZipfile(zf, filename)
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

// ToZipfile writes just a chunk's data into a zipfile.
func (c *Chunk) ToZipfile(zf *zip.Writer, filename string) error {
	writer, err := zf.Create(filename)
	if err != nil {
		return err
	}

	json, err := c.MarshalJSON()
	if err != nil {
		return err
	}

	n, err := writer.Write(json)
	if err != nil {
		return err
	}

	log.Debug("Written chunk to zipfile: %s (%d bytes)", filename, n)
	return nil
}

// ChunkFromZipfile loads a chunk from a zipfile.
func ChunkFromZipfile(zf *zip.Reader, layer int, coord render.Point) (*Chunk, error) {
	filename := fmt.Sprintf("chunks/%d/%s.json", layer, coord)

	file, err := zf.Open(filename)
	if err != nil {
		return nil, err
	}

	bin, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var chunk = NewChunk()
	err = chunk.UnmarshalJSON(bin)
	if err != nil {
		return nil, err
	}

	return chunk, nil
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

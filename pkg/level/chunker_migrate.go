package level

import (
	"runtime"
	"sync"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
)

/* Functions to migrate Chunkers between different implementations. */

// OptimizeChunkerAccessors will evaluate all of the chunks of your drawing
// and possibly migrate them to a different Accessor implementation when
// saving on disk.
func (c *Chunker) OptimizeChunkerAccessors() {
	log.Info("Optimizing Chunker Accessors")

	// TODO: parallelize this with goroutines
	var (
		chunks = make(chan *Chunk, len(c.Chunks))
		wg     sync.WaitGroup
	)

	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunks {
				var point = chunk.Point

				// Upgrade all MapTypes into RLE compressed MapTypes?
				if balance.RLEBinaryChunkerEnabled {
					if chunk.Type == MapType {
						log.Info("Optimizing chunk %s accessor from Map to RLE", point)
						ma, _ := chunk.Accessor.(*MapAccessor)
						rle := NewRLEAccessor(chunk).FromMapAccessor(ma)

						c.Chunks[point].Type = RLEType
						c.Chunks[point].Accessor = rle
					}
				}
			}
		}()
	}

	// Feed it the chunks.
	for point := range c.IterChunks() {
		chunk, ok := c.GetChunk(point)
		if !ok {
			continue
		}
		chunks <- chunk
	}

	close(chunks)
	wg.Wait()

}

// FromMapAccessor migrates from a MapAccessor to RLE.
func (a *RLEAccessor) FromMapAccessor(ma *MapAccessor) *RLEAccessor {
	return &RLEAccessor{
		chunk: a.chunk,
		acc:   ma,
	}
}

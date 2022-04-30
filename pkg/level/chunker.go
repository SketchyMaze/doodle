package level

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
)

// Chunker is the data structure that manages the chunks of a level, and
// provides the API to interact with the pixels using their absolute coordinates
// while abstracting away the underlying details.
type Chunker struct {
	// Layer is optional for the caller, levels use only 0 and
	// doodads use them for frames. When chunks are exported to
	// zipfile the Layer keeps them from overlapping.
	Layer int
	Size  int `json:"size"`

	// A Zipfile reference for new-style levels and doodads which
	// keep their chunks in external parts of a zip file.
	Zipfile *zip.Reader `json:"-"`

	// Chunks, oh boy.
	// The v1 drawing format had all the chunks in the JSON file.
	// New drawings write them to zips. Legacy drawings can be converted
	// simply by loading and resaving: their Chunks loads from JSON and
	// is committed to zipfile on save. This makes Chunks also a good
	// cache even when we have a zipfile to fall back on.
	Chunks  ChunkMap `json:"chunks"`
	chunkMu sync.RWMutex

	// If we have a zipfile, only keep chunks warm in memory if they
	// are actively wanted by the game.
	lastTick              uint64 // NOTE: tracks from shmem.Tick
	chunkRequestsThisTick map[render.Point]interface{}
	requestsN1            map[render.Point]interface{}
	requestsN2            map[render.Point]interface{}
	requestMu             sync.Mutex

	// The palette reference from first call to Inflate()
	pal *Palette
}

// NewChunker creates a new chunk manager with a given chunk size.
func NewChunker(size int) *Chunker {
	return &Chunker{
		Size:   size,
		Chunks: ChunkMap{},

		chunkRequestsThisTick: map[render.Point]interface{}{},
		requestsN1:            map[render.Point]interface{}{},
		requestsN2:            map[render.Point]interface{}{},
	}
}

// Inflate iterates over the pixels in the (loaded) chunks and expands any
// Sparse Swatches (which have only their palette index, from the file format
// on disk) to connect references to the swatches in the palette.
func (c *Chunker) Inflate(pal *Palette) error {
	c.pal = pal

	c.chunkMu.RLock()
	defer c.chunkMu.RUnlock()
	for coord, chunk := range c.Chunks {
		chunk.Point = coord
		chunk.Size = c.Size
		chunk.Inflate(pal)
	}
	return nil
}

// IterViewport returns a channel to iterate every point that exists within
// the viewport rect.
func (c *Chunker) IterViewport(viewport render.Rect) <-chan Pixel {
	pipe := make(chan Pixel)
	go func() {
		// Get the chunk box coordinates.
		var (
			topLeft     = c.ChunkCoordinate(render.NewPoint(viewport.X, viewport.Y))
			bottomRight = c.ChunkCoordinate(render.Point{
				X: viewport.X + viewport.W,
				Y: viewport.Y + viewport.H,
			})
		)
		for cx := topLeft.X; cx <= bottomRight.X; cx++ {
			for cy := topLeft.Y; cy <= bottomRight.Y; cy++ {
				if chunk, ok := c.GetChunk(render.NewPoint(cx, cy)); ok {
					for px := range chunk.Iter() {

						// Verify this pixel is also in range.
						if px.Point().Inside(viewport) {
							pipe <- px
						}
					}
				}
			}
		}
		close(pipe)
	}()
	return pipe
}

// IterChunks returns a channel to iterate over all chunks in the drawing.
func (c *Chunker) IterChunks() <-chan render.Point {
	var (
		pipe = make(chan render.Point)
		sent = map[render.Point]interface{}{}
	)

	go func() {
		c.chunkMu.RLock()

		// Send the chunk coords we have in working memory.
		// v1 levels: had all their chunks there in their JSON data
		// v2 levels: chunks are in zipfile, cached ones are here
		for point := range c.Chunks {
			sent[point] = nil
			pipe <- point
		}

		c.chunkMu.RUnlock()

		// If we have a zipfile, send any remaining chunks that are
		// in colder storage.
		if c.Zipfile != nil {
			for _, point := range ChunksInZipfile(c.Zipfile, c.Layer) {
				if _, ok := sent[point]; ok {
					continue // Already sent from active memory
				}
				pipe <- point
			}
		}

		close(pipe)
	}()
	return pipe
}

/*
IterChunksThemselves iterates all chunks in the drawing rather than coords.

Note: this will mark every chunk as "touched" this frame, so in a zipfile
level will load ALL chunks into memory.
*/
func (c *Chunker) IterChunksThemselves() <-chan *Chunk {
	pipe := make(chan *Chunk)
	go func() {
		for coord := range c.IterChunks() {
			if chunk, ok := c.GetChunk(coord); ok {
				pipe <- chunk
			}
		}
		close(pipe)
	}()
	return pipe
}

// IterCachedChunks iterates ONLY over the chunks currently cached in memory,
// e.g. so they can be torn down without loading extra chunks by looping normally.
func (c *Chunker) IterCachedChunks() <-chan *Chunk {
	pipe := make(chan *Chunk)
	go func() {
		c.chunkMu.RLock()
		defer c.chunkMu.RUnlock()

		for _, chunk := range c.Chunks {
			pipe <- chunk
		}
		close(pipe)
	}()
	return pipe
}

// IterViewportChunks returns a channel to iterate over the Chunk objects that
// appear within the viewport rect, instead of the pixels in each chunk.
func (c *Chunker) IterViewportChunks(viewport render.Rect) <-chan render.Point {
	pipe := make(chan render.Point)
	go func() {
		sent := make(map[render.Point]interface{})

		for x := viewport.X; x < viewport.W; x += (c.Size / 4) {
			for y := viewport.Y; y < viewport.H; y += (c.Size / 4) {

				// Constrain this chunksize step to a point within the bounds
				// of the viewport. This can yield partial chunks on the edges
				// of the viewport.
				point := render.NewPoint(x, y)
				if point.X < viewport.X {
					point.X = viewport.X
				} else if point.X > viewport.X+viewport.W {
					point.X = viewport.X + viewport.W
				}
				if point.Y < viewport.Y {
					point.Y = viewport.Y
				} else if point.Y > viewport.Y+viewport.H {
					point.Y = viewport.Y + viewport.H
				}

				// Translate to a chunk coordinate, dedupe and send it.
				coord := c.ChunkCoordinate(render.NewPoint(x, y))
				if _, ok := sent[coord]; ok {
					continue
				}
				sent[coord] = nil

				if _, ok := c.GetChunk(coord); ok {
					pipe <- coord
				}
			}
		}

		close(pipe)
	}()
	return pipe
}

// IterPixels returns a channel to iterate over every pixel in the entire
// chunker.
func (c *Chunker) IterPixels() <-chan Pixel {
	pipe := make(chan Pixel)
	go func() {
		for chunk := range c.IterChunksThemselves() {
			for px := range chunk.Iter() {
				pipe <- px
			}
		}
		close(pipe)
	}()
	return pipe
}

// WorldSize returns the bounding coordinates that the Chunker has chunks to
// manage: the lowest pixels from the lowest chunks to the highest pixels of
// the highest chunks.
func (c *Chunker) WorldSize() render.Rect {
	chunkLowest, chunkHighest := c.Bounds()
	return render.Rect{
		X: chunkLowest.X * c.Size,
		Y: chunkLowest.Y * c.Size,
		W: (chunkHighest.X * c.Size) + (c.Size - 1),
		H: (chunkHighest.Y * c.Size) + (c.Size - 1),
	}
}

// WorldSizePositive returns the WorldSize anchored to 0,0 with only positive
// coordinates.
func (c *Chunker) WorldSizePositive() render.Rect {
	S := c.WorldSize()
	return render.Rect{
		X: 0,
		Y: 0,
		W: int(math.Abs(float64(S.X))) + S.W,
		H: int(math.Abs(float64(S.Y))) + S.H,
	}
}

// Bounds returns the boundary points of the lowest and highest chunk which
// have any data in them.
func (c *Chunker) Bounds() (low, high render.Point) {
	for coord := range c.IterChunks() {
		if coord.X < low.X {
			low.X = coord.X
		}
		if coord.Y < low.Y {
			low.Y = coord.Y
		}

		if coord.X > high.X {
			high.X = coord.X
		}
		if coord.Y > high.Y {
			high.Y = coord.Y
		}
	}

	return low, high
}

/*
GetChunk gets a chunk at a certain position. Returns false if not found.

This should be the centralized function to request a Chunk from the Chunker
(or IterChunksThemselves). On old-style levels all of the chunks were just
in memory as part of the JSON struct, in Zip files we can load/unload them
at will from external files.
*/
func (c *Chunker) GetChunk(p render.Point) (*Chunk, bool) {
	// It's currently cached in memory?
	c.chunkMu.RLock()
	chunk, ok := c.Chunks[p]
	c.chunkMu.RUnlock()

	if ok {
		// An empty chunk? We hang onto these until save time to commit
		// the empty chunk to ZIP.
		if chunk.Len() == 0 {
			return nil, false
		}

		c.logChunkAccess(p, chunk) // for the LRU cache
		return chunk, ok
	}

	// Hit the zipfile for it.
	if c.Zipfile != nil {
		if chunk, err := ChunkFromZipfile(c.Zipfile, c.Layer, p); err == nil {
			log.Debug("GetChunk(%s) cache miss, read from zip", p)
			c.SetChunk(p, chunk)       // cache it
			c.logChunkAccess(p, chunk) // for the LRU cache
			if c.pal != nil {
				chunk.Point = p
				chunk.Size = c.Size
				chunk.Inflate(c.pal)
			}
			return chunk, true
		}
	}

	// Is our chunk cache getting too full? e.g. on full level
	// sweeps where a whole zip file's worth of chunks are scanned.
	if balance.ChunkerLRUCacheMax > 0 && len(c.Chunks) > balance.ChunkerLRUCacheMax {
		log.Error("Chunks in memory (%d) exceeds LRU cache cap of %d, freeing random chunks")
		c.chunkMu.Lock()
		defer c.chunkMu.Unlock()

		var (
			i     = 0
			limit = len(c.Chunks) - balance.ChunkerLRUCacheMax
		)
		for coord := range c.Chunks {
			if i < limit {
				delete(c.Chunks, coord)
			}
			i++
		}
	}

	return nil, false
}

// LRU cache for chunks from zipfiles: log which chunks were accessed
// this tick, so they can be compared to the tick prior, and then freed
// up after that.
func (c *Chunker) logChunkAccess(p render.Point, chunk *Chunk) {
	// Record this point.
	c.requestMu.Lock()
	if c.chunkRequestsThisTick == nil {
		c.chunkRequestsThisTick = map[render.Point]interface{}{}
	}
	c.chunkRequestsThisTick[p] = nil
	c.requestMu.Unlock()
}

// FreeCaches unloads chunks that have not been requested in 2 frames.
//
// Only on chunkers that have zipfiles, old-style levels without zips
// wouldn't be able to restore their chunks otherwise! Returns -1 if
// no Zipfile, otherwise number of chunks freed.
func (c *Chunker) FreeCaches() int {
	if c.Zipfile == nil {
		return -1
	}

	var thisTick = shmem.Tick

	// Very first tick this chunker has seen?
	if c.lastTick == 0 {
		c.lastTick = thisTick
	}

	// A new tick?
	if (thisTick-c.lastTick)%4 == 0 {
		c.requestMu.Lock()
		c.chunkMu.Lock()
		defer c.requestMu.Unlock()
		defer c.chunkMu.Unlock()

		var (
			requestsThisTick = c.chunkRequestsThisTick
			requestsN2       = c.requestsN2
			delete_coords    = []render.Point{}
		)

		// Chunks not requested this last tick, unload from the cache.
		for coord := range requestsN2 {
			// Old point not requested recently?
			if _, ok := requestsThisTick[coord]; !ok {
				delete_coords = append(delete_coords, coord)
			}
		}

		for _, coord := range delete_coords {
			c.FreeChunk(coord)
		}

		// Rotate the cached ticks and clean the slate.
		c.requestsN2 = c.requestsN1
		c.requestsN1 = requestsThisTick
		c.chunkRequestsThisTick = map[render.Point]interface{}{}

		c.lastTick = thisTick

		return len(delete_coords)
	}

	return 0
}

// SetChunk writes the chunk into the cache dict and nothing more.
//
// This function should be the singular writer to the chunk cache.
func (c *Chunker) SetChunk(p render.Point, chunk *Chunk) {
	c.chunkMu.Lock()
	c.Chunks[p] = chunk
	c.chunkMu.Unlock()

	c.logChunkAccess(p, chunk)
}

// FreeChunk unloads a chunk from active memory for zipfile-backed levels.
//
// Not thread safe: it is assumed the caller has the lock on c.Chunks.
func (c *Chunker) FreeChunk(p render.Point) bool {
	if c.Zipfile == nil {
		return false
	}

	// Don't delete empty chunks, hang on until next zipfile save.
	if chunk, ok := c.Chunks[p]; ok && chunk.Len() == 0 {
		return false
	}

	delete(c.Chunks, p)
	return true
}

// Redraw marks every chunk as dirty and invalidates all their texture caches,
// forcing the drawing to re-generate from scratch.
func (c *Chunker) Redraw() {
	for chunk := range c.IterChunksThemselves() {
		chunk.SetDirty()
	}
}

// Prerender visits every chunk and fetches its texture, in order to pre-load
// the whole drawing for smooth gameplay rather than chunks lazy rendering as
// they enter the screen.
func (c *Chunker) Prerender() {
	for chunk := range c.IterChunksThemselves() {
		_ = chunk.CachedBitmap(render.Invisible)
	}
}

// PrerenderN will pre-render the texture for N number of chunks and then
// yield back to the caller. Returns the number of chunks that still need
// textures rendered; zero when the last chunk has been prerendered.
func (c *Chunker) PrerenderN(n int) (remaining int) {
	var (
		total         int // total no. of chunks available
		totalRendered int // no. of chunks with textures
		modified      int // number modified this call
	)

	for chunk := range c.IterChunksThemselves() {
		total++
		if chunk.bitmap != nil {
			totalRendered++
			continue
		}

		if modified < n {
			_ = chunk.CachedBitmap(render.Invisible)
			totalRendered++
			modified++
		}
	}

	remaining = total - totalRendered
	return
}

// Get a pixel at the given coordinate. Returns the Palette entry for that
// pixel or else returns an error if not found.
func (c *Chunker) Get(p render.Point) (*Swatch, error) {
	// Compute the chunk coordinate.
	coord := c.ChunkCoordinate(p)
	if chunk, ok := c.GetChunk(coord); ok {
		return chunk.Get(p)
	}
	return nil, fmt.Errorf("no chunk %s exists for point %s", coord, p)
}

// Set a pixel at the given coordinate.
func (c *Chunker) Set(p render.Point, sw *Swatch) error {
	coord := c.ChunkCoordinate(p)
	chunk, ok := c.GetChunk(coord)
	if !ok {
		chunk = NewChunk()
		chunk.Point = coord
		chunk.Size = c.Size
		c.SetChunk(coord, chunk)
	}

	return chunk.Set(p, sw)
}

// SetRect sets a rectangle of pixels to a color all at once.
func (c *Chunker) SetRect(r render.Rect, sw *Swatch) error {
	var (
		xMin = r.X
		yMin = r.Y
		xMax = r.X + r.W
		yMax = r.Y + r.H
	)
	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			c.Set(render.NewPoint(x, y), sw)
		}
	}

	return nil
}

// Delete a pixel at the given coordinate.
func (c *Chunker) Delete(p render.Point) error {
	coord := c.ChunkCoordinate(p)

	if chunk, ok := c.GetChunk(coord); ok {
		return chunk.Delete(p)
	}
	return fmt.Errorf("no chunk %s exists for point %s", coord, p)
}

// DeleteRect deletes a rectangle of pixels between two points.
// The rect is a relative one with a width and height, and the X,Y values are
// an absolute world coordinate.
func (c *Chunker) DeleteRect(r render.Rect) error {
	var (
		xMin = r.X
		yMin = r.Y
		xMax = r.X + r.W
		yMax = r.Y + r.H
	)
	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			c.Delete(render.NewPoint(x, y))
		}
	}

	return nil
}

// ChunkCoordinate computes a chunk coordinate from an absolute coordinate.
func (c *Chunker) ChunkCoordinate(abs render.Point) render.Point {
	if c.Size == 0 {
		return render.Point{}
	}

	size := float64(c.Size)
	return render.NewPoint(
		int(math.Floor(float64(abs.X)/size)),
		int(math.Floor(float64(abs.Y)/size)),
	)
}

// ChunkMap maps a chunk coordinate to its chunk data.
type ChunkMap map[render.Point]*Chunk

// MarshalJSON to convert the chunk map to JSON. This is needed for writing so
// the JSON encoder knows how to serializes a `map[Point]*Chunk` but the inverse
// is not necessary to implement.
func (c ChunkMap) MarshalJSON() ([]byte, error) {
	dict := map[string]*Chunk{}
	for point, chunk := range c {
		dict[point.String()] = chunk
	}

	out, err := json.Marshal(dict)
	return out, err
}

package uix

import (
	"runtime"
	"sync"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
)

// Memory optimization features of the Canvas.

/*
LoadUnloadChunks optimizes memory for (level) canvases by warming up chunk images
that fall within the LoadingViewport and freeing chunks that are outside of it.
*/
func (w *Canvas) LoadUnloadChunks() {
	if w.level == nil || shmem.Tick%balance.CanvasLoadUnloadModuloTicks != 0 || !balance.Feature.LoadUnloadChunk {
		return
	}

	var (
		vp             = w.LoadingViewport()
		chunks         = make(chan render.Point)
		chunksInside   = map[render.Point]interface{}{}
		chunksTeardown = []*level.Chunk{}
		cores          = runtime.NumCPU()
		wg             sync.WaitGroup

		// Collect metrics for the debug overlay.
		resultInside  int
		resultOutside int
	)

	// Collect the chunks that are inside the viewport so we know which ones are not.
	for chunk := range w.level.Chunker.IterViewportChunks(vp) {
		chunksInside[chunk] = nil
	}

	// Spawn background goroutines to process the chunks quickly.
	for i := 0; i < cores; i++ {
		wg.Add(1)
		go func(i int) {
			for coord := range chunks {
				if chunk, ok := w.level.Chunker.GetChunk(coord); ok {
					chunk := chunk

					if _, ok := chunksInside[coord]; ok {
						// Preload its bitmap image.
						_ = chunk.CachedBitmap(render.Invisible)
						resultInside++
					} else {
						// Unload its bitmap and texture.
						chunksTeardown = append(chunksTeardown, chunk)
						resultOutside++
					}
				}
			}
			wg.Done()
		}(i)
	}

	for chunk := range w.level.Chunker.IterChunks() {
		chunks <- chunk
	}
	close(chunks)
	wg.Wait()

	// Tear down the SDL2 textures of chunks to free.
	for i, chunk := range chunksTeardown {
		if chunk == nil {
			log.Error("LoadUnloadChunks: chunksTeardown#%d was nil??", i)
			continue
		}

		chunk.Teardown()
	}

	// Export the metrics for the debug overlay.
	w.loadUnloadInside = resultInside
	w.loadUnloadOutside = resultOutside
}

// LoadUnloadMetrics returns the canvas's stored metrics from the LoadUnloadChunks
// function, for the debug overlay.
func (w *Canvas) LoadUnloadMetrics() (inside, outside int) {
	return w.loadUnloadInside, w.loadUnloadOutside
}

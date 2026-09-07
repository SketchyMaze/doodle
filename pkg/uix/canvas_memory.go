package uix

import (
	"runtime"
	"sync"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
)

// Memory optimization features of the Canvas.

/*
LoadUnloadChunks optimizes memory for (level) canvases by warming up chunk images
that fall within the LoadingViewport and freeing chunks that are outside of it.
*/
func (w *Canvas) LoadUnloadChunks(force ...bool) {
	if !(len(force) > 0 && force[0]) {
		if w.level == nil || shmem.Tick%balance.CanvasLoadUnloadModuloTicks != 0 || !balance.Feature.LoadUnloadChunk || (len(force) > 0 && force[0]) {
			return
		}
	}

	var (
		vp           = w.LoadingViewport()
		chunks       = make(chan render.Point)
		chunksInside = map[render.Point]interface{}{}
		cores        = runtime.NumCPU()
		wg           sync.WaitGroup

		// Collect metrics for the debug overlay.
		resultInside  int
		resultOutside int
	)

	// Collect the chunks that are inside the viewport so we know which ones are not.
	for _, chunk := range w.level.Chunker.IterViewportChunks(vp) {
		chunksInside[chunk] = nil
	}

	// Spawn background goroutines to process the chunks quickly. Each worker
	// tallies into its own counters (summed below) since resultInside and
	// resultOutside would otherwise be incremented concurrently from
	// multiple goroutines with no synchronization.
	var counts = make([][2]int, cores) // [i] = {inside, outside}
	for i := 0; i < cores; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for coord := range chunks {
				if _, ok := chunksInside[coord]; ok {
					// This chunk is INSIDE our viewport, preload its bitmap.
					if chunk, ok := w.level.Chunker.GetChunk(coord); ok {
						_ = chunk.CachedBitmap(render.Invisible)
						counts[i][0]++
						continue
					}
				}

				// Chunks outside the viewport, we won't load them and
				// the Chunker will flush them out to (zip) file.
				counts[i][1]++
			}
		}(i)
	}

	for chunk := range w.level.Chunker.IterChunks() {
		chunks <- chunk
	}
	close(chunks)
	wg.Wait()

	for _, c := range counts {
		resultInside += c[0]
		resultOutside += c[1]
	}

	// Note: chunks outside the viewport are torn down separately, on their own
	// schedule, by Chunker.FreeCaches() (see canvas.go's Loop()) which frees
	// chunks that haven't been requested by GetChunk() in the last couple of
	// ticks. This function only warms the bitmap cache for chunks coming into
	// view.

	// Export the metrics for the debug overlay.
	w.loadUnloadInside = resultInside
	w.loadUnloadOutside = resultOutside
}

// LoadUnloadMetrics returns the canvas's stored metrics from the LoadUnloadChunks
// function, for the debug overlay.
func (w *Canvas) LoadUnloadMetrics() (inside, outside int) {
	return w.loadUnloadInside, w.loadUnloadOutside
}

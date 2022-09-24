package wasm

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
)

// HTTPGet fetches a path via ajax request.
func HTTPGet(filename string) ([]byte, error) {
	// Already cached?
	jsonData, ok := shmem.AjaxCache[filename]
	if ok {
		return jsonData, nil
	}

	// Fetch the URI.
	resp, err := http.Get(filename)
	if err != nil {
		log.Error("http error: %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Error?
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, fmt.Errorf("failed to load URI %s: HTTP %d response",
			filename,
			resp.StatusCode,
		)
	}

	// Parse and store the response in cache.
	jsonData, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	shmem.AjaxCache[filename] = jsonData

	return jsonData, nil
}

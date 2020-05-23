package sound

import (
	"io/ioutil"
	"path/filepath"
)

// PreloadAll looks in the SoundRoot and MusicRoot folders and preloads all
// supported files into the caches.
func PreloadAll() {
	if engine == nil || !Enabled {
		return
	}

	// Preload sound effects.
	if files, err := ioutil.ReadDir(SoundRoot); err == nil {
		for _, file := range files {
			if filepath.Ext(file.Name()) != ".wav" {
				continue
			}

			LoadSound(file.Name())
		}
	}
}

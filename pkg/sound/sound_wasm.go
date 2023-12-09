//go:build js && wasm
// +build js,wasm

package sound

import (
	"git.kirsle.net/go/audio"
	"git.kirsle.net/go/audio/null"
)

// Globals for WASM sound engine.
var (
	engine *null.Engine
)

func init() {
	engine = null.New()
}

// LoadMusic loads filename from the MusicRoot into the global music cache.
// If the music is already loaded, does nothing.
func LoadMusic(filename string) audio.Playable {
	return null.Playable{}
}

// LoadSound loads filename from the SoundRoot into the global SFX cache.
// If the sound is already loaded, does nothing.
func LoadSound(filename string) audio.Playable {
	return null.Playable{}
}

// PlaySound plays the named sound.
func PlaySound(filename string) {}

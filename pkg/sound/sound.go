// Package sound provides audio functions for Doodle.
package sound

import (
	"path/filepath"
	"sync"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/audio/sdl"
	"github.com/veandco/go-sdl2/mix"
)

// Globals.
var (
	// If enabled is false, all sound functions are no-ops.
	Enabled bool

	// Root folder on disk where sound and music files should live.
	SoundRoot = filepath.Join("rtp", "sfx")
	MusicRoot = filepath.Join("rtp", "music")

	// Cache of loaded music and sound effects.
	music  = map[string]*sdl.Track{}
	sounds = map[string]*sdl.Track{}
	mu     sync.RWMutex

	engine *sdl.Engine
)

// Initialize SDL2 Audio at startup.
func init() {
	eng, err := sdl.New(mix.INIT_MP3 | mix.INIT_OGG)
	if err != nil {
		log.Error("sound.init(): error initializing SDL2 audio: %s", err)
		return
	}

	err = eng.Setup()
	if err != nil {
		log.Error("sound.init(): error setting up SDL2 audio: %s", err)
		return
	}

	engine = eng
	Enabled = true
}

// LoadMusic loads filename from the MusicRoot into the global music cache.
// If the music is already loaded, does nothing.
func LoadMusic(filename string) *sdl.Track {
	if engine == nil || !Enabled {
		return nil
	}

	// Check if the music is already loaded.
	mu.RLock()
	mus, ok := music[filename]
	mu.RUnlock()
	if ok {
		return mus
	}

	// Load the music in.
	track, err := engine.LoadMusic(filepath.Join(MusicRoot, filename))
	if err != nil {
		log.Error("sound.LoadMusic: failed to load file %s: %s", filename, err)
		return nil
	}

	mu.Lock()
	music[filename] = &track
	mu.Unlock()

	return &track
}

// LoadSound loads filename from the SoundRoot into the global SFX cache.
// If the sound is already loaded, does nothing.
func LoadSound(filename string) *sdl.Track {
	if engine == nil || !Enabled {
		return nil
	}

	// Check if the music is already loaded.
	mu.RLock()
	sfx, ok := sounds[filename]
	mu.RUnlock()
	if ok {
		return sfx
	}

	// Load the sound in.
	log.Info("Loading sound: %s", filename)
	track, err := engine.LoadSound(filepath.Join(SoundRoot, filename))
	if err != nil {
		log.Error("sound.LoadSound: failed to load file %s: %s", filename, err)
		return nil
	}

	mu.Lock()
	sounds[filename] = &track
	mu.Unlock()

	return &track
}

// PlaySound plays the named sound.
func PlaySound(filename string) {
	log.Debug("Play sound: %s", filename)
	sound := LoadSound(filename)
	if sound != nil {
		sound.Play(1)
	}
}

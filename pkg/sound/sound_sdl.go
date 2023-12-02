//go:build !js && !wasm
// +build !js,!wasm

package sound

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/audio"
	"git.kirsle.net/go/audio/sdl"
	"github.com/veandco/go-sdl2/mix"
)

// SDL engine globals.
var (
	engine *sdl.Engine

	// Cache of loaded music and sound effects.
	music  = map[string]*sdl.Track{}
	sounds = map[string]*sdl.Track{}
	mu     sync.RWMutex

	// Supported file extensions, in preference order.
	SupportedSoundExtensions = []string{
		".wav",
		".ogg",
		".mp3",
	}
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
func LoadMusic(filename string) audio.Playable {
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
func LoadSound(filename string) audio.Playable {
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

	// Resolve the filepath on disk to this sound.
	fullpath, err := ResolveFilename(filename)
	if err != nil {
		log.Error("Loading sound: %s: %s", filename, err)
		return nil
	}

	// Load the sound in.
	log.Info("Loading sound: %s", filename)
	track, err := engine.LoadSound(fullpath)
	if err != nil {
		log.Error("sound.LoadSound: failed to load file %s: %s", filename, err)
		return nil
	}

	mu.Lock()
	sounds[filename] = &track
	mu.Unlock()

	return &track
}

// PlaySound plays the named sound. It will de-duplicate if the same sound is already playing.
func PlaySound(filename string) {
	log.Debug("Play sound: %s", filename)
	sound := LoadSound(filename)
	if sound != nil && !sound.Playing() {
		sound.Play(1)
	}
}

// ResolveFilename resolves the filename to a sound file on disk.
//
// Ogg has been found more reliable than MP3 for cross-platform distribution. A doodad might
// request a "sound.mp3" but the filename on disk is actually "sound.ogg" and this function
// will return the latter, if "sound.mp3" does not exist.
//
// If the exact filename does exist, it is returned; otherwise a preference order of
// WAV > OGG > MP3 will be tried and returned if those versions exist.
//
// Returns the full path on disk (e.g. "rtp/sfx/sound.ogg")
func ResolveFilename(filename string) (string, error) {
	// Does the exact file exist?
	if _, err := os.Stat(filepath.Join(SoundRoot, filename)); !os.IsNotExist(err) {
		return filepath.Join(SoundRoot, filename), nil
	}

	// If the filename ends with a supported extension, trim it to the basename.
	var basename = filename
	for _, ext := range SupportedSoundExtensions {
		if filepath.Ext(filename) == ext {
			basename = strings.TrimSuffix(filename, ext)
		}
	}

	// Try the basename + suffixes.
	for _, ext := range SupportedSoundExtensions {
		check := filepath.Join(SoundRoot, basename+ext)
		if _, err := os.Stat(check); !os.IsNotExist(err) {
			log.Info("Sound(%s): resolved to nearest match %s", filename, check)
			return check, nil
		}
	}

	// No luck.
	return "", errors.New("no suitable sound file found")
}

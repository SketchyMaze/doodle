package sdl

import (
	"git.kirsle.net/apps/doodle/level"
)

// Frames to cache for FPS calculation.
const (
	maxSamples = 100
	TargetFPS  = 1000 / 60
)

var (
	fpsCurrentTicks uint32 // current time we get sdl.GetTicks()
	fpsLastTime     uint32 // last time we printed the fpsCurrentTicks
	fpsCurrent      int
	fpsFrames       int
	fpsSkipped      uint32
	fpsInterval     uint32 = 1000
)

var pixelHistory []level.Pixel

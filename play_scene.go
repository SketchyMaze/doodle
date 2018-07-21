package doodle

import (
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"github.com/veandco/go-sdl2/sdl"
)

// PlayScene manages the "Edit Level" game mode.
type PlayScene struct {
	canvas Grid

	// Canvas size
	width  int32
	height int32

	// Player position and velocity.
	x  int32
	y  int32
	vx int32
	vy int32
}

// Name of the scene.
func (s *PlayScene) Name() string {
	return "Play"
}

// Setup the play scene.
func (s *PlayScene) Setup(d *Doodle) error {
	s.x = 10
	s.y = 10

	if s.canvas == nil {
		s.canvas = Grid{}
	}
	s.width = d.width // TODO: canvas width = copy the window size
	s.height = d.height
	return nil
}

// Loop the editor scene.
func (s *PlayScene) Loop(d *Doodle) error {
	s.PollEvents(d.events)

	// Apply gravity.

	return s.Draw(d)
}

// Draw the pixels on this frame.
func (s *PlayScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.renderer.SetDrawColor(255, 255, 255, 255)
	d.renderer.Clear()

	d.renderer.SetDrawColor(0, 0, 0, 255)
	for pixel := range s.canvas {
		d.renderer.DrawPoint(pixel.x, pixel.y)
	}

	// Draw our hero.
	d.renderer.SetDrawColor(0, 0, 255, 255)
	d.renderer.DrawRect(&sdl.Rect{
		X: s.x,
		Y: s.y,
		W: 16,
		H: 16,
	})

	// Draw the FPS.
	d.DrawDebugOverlay()
	d.renderer.Present()

	return nil
}

// PollEvents checks the event state and updates variables.
func (s *PlayScene) PollEvents(ev *events.State) {
	if ev.Down.Now {
		s.y += 4
	}
	if ev.Left.Now {
		s.x -= 4
	}
	if ev.Right.Now {
		s.x += 4
	}
	if ev.Up.Now {
		s.y -= 4
	}
}

// LoadLevel loads a level from disk.
func (s *PlayScene) LoadLevel(filename string) error {
	s.canvas = Grid{}

	m, err := level.LoadJSON(filename)
	if err != nil {
		return err
	}

	for _, point := range m.Pixels {
		pixel := Pixel{
			x: point.X,
			y: point.Y,
		}
		s.canvas[pixel] = nil
	}

	return nil
}

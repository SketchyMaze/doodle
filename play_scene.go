package doodle

import (
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
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
func (s *PlayScene) Loop(d *Doodle, ev *events.State) error {
	s.movePlayer(ev)
	return nil
}

// Draw the pixels on this frame.
func (s *PlayScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	for pixel := range s.canvas {
		d.Engine.DrawPoint(render.Black, render.Point{pixel.x, pixel.y})
	}

	// Draw our hero.
	d.Engine.DrawRect(render.Magenta, render.Rect{s.x, s.y, 16, 16})

	return nil
}

// movePlayer updates the player's X,Y coordinate based on key pressed.
func (s *PlayScene) movePlayer(ev *events.State) {
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

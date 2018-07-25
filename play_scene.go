package doodle

import (
	"git.kirsle.net/apps/doodle/doodads"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

// PlayScene manages the "Edit Level" game mode.
type PlayScene struct {
	// Configuration attributes.
	Filename string
	Canvas   render.Grid

	// Private variables.
	canvas render.Grid

	// Canvas size
	width  int32
	height int32

	// Player character
	player doodads.Doodad
}

// Name of the scene.
func (s *PlayScene) Name() string {
	return "Play"
}

// Setup the play scene.
func (s *PlayScene) Setup(d *Doodle) error {
	// Given a filename or map data to play?
	if s.Canvas != nil {
		log.Debug("PlayScene.Setup: loading map from given canvas")
		s.canvas = s.Canvas

	} else if s.Filename != "" {
		log.Debug("PlayScene.Setup: loading map from file %s", s.Filename)
		s.LoadLevel(s.Filename)
		s.Filename = ""
	}

	s.player = doodads.NewPlayer()

	if s.canvas == nil {
		log.Debug("PlayScene.Setup: no grid given, initializing empty grid")
		s.canvas = render.Grid{}
	}

	s.width = d.width // TODO: canvas width = copy the window size
	s.height = d.height

	d.Flash("Entered Play Mode. Press 'E' to edit this map.")

	return nil
}

// Loop the editor scene.
func (s *PlayScene) Loop(d *Doodle, ev *events.State) error {
	// Switching to Edit Mode?
	if ev.KeyName.Read() == "e" {
		log.Info("Edit Mode, Go!")
		d.Goto(&EditorScene{
			Canvas: s.canvas,
		})
		return nil
	}

	s.movePlayer(ev)
	return nil
}

// Draw the pixels on this frame.
func (s *PlayScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	s.canvas.Draw(d.Engine)

	// Draw our hero.
	s.player.Draw(d.Engine)

	return nil
}

// movePlayer updates the player's X,Y coordinate based on key pressed.
func (s *PlayScene) movePlayer(ev *events.State) {
	delta := s.player.Position()
	var playerSpeed int32 = 8
	var gravity int32 = 2

	if ev.Down.Now {
		delta.Y += playerSpeed
	}
	if ev.Left.Now {
		delta.X -= playerSpeed
	}
	if ev.Right.Now {
		delta.X += playerSpeed
	}
	if ev.Up.Now {
		delta.Y -= playerSpeed
	}

	// Apply gravity.
	delta.Y += gravity

	// Draw a ray and check for collision.
	var lastOk = s.player.Position()
	for point := range render.IterLine2(s.player.Position(), delta) {
		s.player.MoveTo(point)
		if _, ok := doodads.CollidesWithGrid(s.player, &s.canvas); ok {
			s.player.MoveTo(lastOk)
		} else {
			lastOk = s.player.Position()
		}
	}

	s.player.MoveTo(lastOk)
}

// LoadLevel loads a level from disk.
func (s *PlayScene) LoadLevel(filename string) error {
	s.canvas = render.Grid{}

	m, err := level.LoadJSON(filename)
	if err != nil {
		return err
	}

	for _, point := range m.Pixels {
		pixel := level.Pixel{
			X: point.X,
			Y: point.Y,
		}
		s.canvas[pixel] = nil
	}

	return nil
}

// Destroy the scene.
func (s *PlayScene) Destroy() error {
	return nil
}

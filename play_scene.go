package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/doodads"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

// PlayScene manages the "Edit Level" game mode.
type PlayScene struct {
	// Configuration attributes.
	Filename string
	Level    *level.Level

	// Private variables.
	drawing *level.Canvas

	// Player character
	Player doodads.Doodad
}

// Name of the scene.
func (s *PlayScene) Name() string {
	return "Play"
}

// Setup the play scene.
func (s *PlayScene) Setup(d *Doodle) error {
	s.drawing = level.NewCanvas(balance.ChunkSize, false)
	s.drawing.MoveTo(render.Origin)
	s.drawing.Resize(render.NewRect(d.width, d.height))
	s.drawing.Compute(d.Engine)

	// Given a filename or map data to play?
	if s.Level != nil {
		log.Debug("PlayScene.Setup: received level from scene caller")
		s.drawing.LoadLevel(s.Level)
	} else if s.Filename != "" {
		log.Debug("PlayScene.Setup: loading map from file %s", s.Filename)
		s.LoadLevel(s.Filename)
	}

	s.Player = doodads.NewPlayer()

	if s.Level == nil {
		log.Debug("PlayScene.Setup: no grid given, initializing empty grid")
		s.Level = level.New()
		s.drawing.LoadLevel(s.Level)
	}

	d.Flash("Entered Play Mode. Press 'E' to edit this map.")

	return nil
}

// Loop the editor scene.
func (s *PlayScene) Loop(d *Doodle, ev *events.State) error {
	// Switching to Edit Mode?
	if ev.KeyName.Read() == "e" {
		log.Info("Edit Mode, Go!")
		d.Goto(&EditorScene{
			Filename: s.Filename,
			Level:    s.Level,
		})
		return nil
	}

	// s.drawing.Loop(ev)
	s.movePlayer(ev)
	return nil
}

// Draw the pixels on this frame.
func (s *PlayScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	// Draw the level.
	s.drawing.Present(d.Engine, s.drawing.Point())

	// Draw our hero.
	s.Player.Draw(d.Engine)

	// Draw out bounding boxes.
	d.DrawCollisionBox(s.Player)

	return nil
}

// movePlayer updates the player's X,Y coordinate based on key pressed.
func (s *PlayScene) movePlayer(ev *events.State) {
	delta := s.Player.Position()
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
	// var onFloor bool

	info, ok := doodads.CollidesWithGrid(s.Player, s.Level.Chunker, delta)
	if ok {
		// Collision happened with world.
	}
	delta = info.MoveTo

	// Apply gravity if not grounded.
	if !s.Player.Grounded() {
		// Gravity has to pipe through the collision checker, too, so it
		// can't give us a cheated downward boost.
		delta.Y += gravity
	}

	s.Player.MoveTo(delta)
}

// LoadLevel loads a level from disk.
func (s *PlayScene) LoadLevel(filename string) error {
	s.Filename = filename

	level, err := level.LoadJSON(filename)
	if err != nil {
		return fmt.Errorf("PlayScene.LoadLevel(%s): %s", filename, err)
	}

	s.Level = level
	s.drawing.LoadLevel(s.Level)

	return nil
}

// Destroy the scene.
func (s *PlayScene) Destroy() error {
	return nil
}

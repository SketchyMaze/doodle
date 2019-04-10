package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
)

// PlayScene manages the "Edit Level" game mode.
type PlayScene struct {
	// Configuration attributes.
	Filename string
	Level    *level.Level

	// Private variables.
	d       *Doodle
	drawing *uix.Canvas

	// Custom debug labels.
	debPosition   *string
	debViewport   *string
	debScroll     *string
	debWorldIndex *string

	// Player character
	Player doodads.Actor
}

// Name of the scene.
func (s *PlayScene) Name() string {
	return "Play"
}

// Setup the play scene.
func (s *PlayScene) Setup(d *Doodle) error {
	s.d = d

	// Initialize debug overlay values.
	s.debPosition = new(string)
	s.debViewport = new(string)
	s.debScroll = new(string)
	s.debWorldIndex = new(string)
	customDebugLabels = []debugLabel{
		{"Pixel:", s.debWorldIndex},
		{"Player:", s.debPosition},
		{"Viewport:", s.debViewport},
		{"Scroll:", s.debScroll},
	}

	// Initialize the drawing canvas.
	s.drawing = uix.NewCanvas(balance.ChunkSize, false)
	s.drawing.MoveTo(render.Origin)
	s.drawing.Resize(render.NewRect(int32(d.width), int32(d.height)))
	s.drawing.Compute(d.Engine)

	// Given a filename or map data to play?
	if s.Level != nil {
		log.Debug("PlayScene.Setup: received level from scene caller")
		s.drawing.LoadLevel(d.Engine, s.Level)
	} else if s.Filename != "" {
		log.Debug("PlayScene.Setup: loading map from file %s", s.Filename)
		s.LoadLevel(s.Filename)
	}

	s.Player = doodads.NewPlayer()

	if s.Level == nil {
		log.Debug("PlayScene.Setup: no grid given, initializing empty grid")
		s.Level = level.New()
		s.drawing.LoadLevel(d.Engine, s.Level)
	}

	d.Flash("Entered Play Mode. Press 'E' to edit this map.")

	return nil
}

// Loop the editor scene.
func (s *PlayScene) Loop(d *Doodle, ev *events.State) error {
	// Update debug overlay values.
	*s.debWorldIndex = s.drawing.WorldIndexAt(render.NewPoint(ev.CursorX.Now, ev.CursorY.Now)).String()
	*s.debPosition = s.Player.Position().String()
	*s.debViewport = s.drawing.Viewport().String()
	*s.debScroll = s.drawing.Scroll.String()

	// Has the window been resized?
	if resized := ev.Resized.Read(); resized {
		w, h := d.Engine.WindowSize()
		if w != d.width || h != d.height {
			d.width = w
			d.height = h
			s.drawing.Resize(render.NewRect(int32(d.width), int32(d.height)))
			return nil
		}
	}

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
	s.drawing.LoadLevel(s.d.Engine, s.Level)

	return nil
}

// Destroy the scene.
func (s *PlayScene) Destroy() error {
	return nil
}

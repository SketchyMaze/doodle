package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/apps/doodle/pkg/uix"
)

// PlayScene manages the "Edit Level" game mode.
type PlayScene struct {
	// Configuration attributes.
	Filename string
	Level    *level.Level
	CanEdit  bool // i.e. you came from the Editor Mode
	HasNext  bool // has a next level to load next

	// Private variables.
	d         *Doodle
	drawing   *uix.Canvas
	scripting *scripting.Supervisor
	running   bool

	// UI widgets.
	supervisor *ui.Supervisor
	editButton *ui.Button

	// The alert box shows up when the level goal is reached and includes
	// buttons what to do next.
	alertBox          *ui.Window
	alertReplayButton *ui.Button // Replay level
	alertEditButton   *ui.Button // Edit Level
	alertNextButton   *ui.Button // Next Level
	alertExitButton   *ui.Button // Exit to menu

	// Custom debug labels.
	debPosition   *string
	debViewport   *string
	debScroll     *string
	debWorldIndex *string

	// Player character
	Player *uix.Actor
}

// Name of the scene.
func (s *PlayScene) Name() string {
	return "Play"
}

// Setup the play scene.
func (s *PlayScene) Setup(d *Doodle) error {
	s.d = d
	s.scripting = scripting.NewSupervisor()
	s.supervisor = ui.NewSupervisor()

	// Level Exit handler.
	s.SetupAlertbox()
	s.scripting.OnLevelExit(func() {
		d.Flash("Hurray!")

		// Pause the simulation.
		s.running = false

		// Toggle the relevant buttons on.
		if s.CanEdit {
			s.alertEditButton.Show()
		}
		if s.HasNext {
			s.alertNextButton.Show()
		}

		// Always-visible buttons.
		s.alertReplayButton.Show()
		s.alertExitButton.Show()

		// Show the alert box.
		s.alertBox.Show()
	})

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

	// Initialize the "Edit Map" button.
	s.editButton = ui.NewButton("Edit", ui.NewLabel(ui.Label{
		Text: "Edit (E)",
		Font: balance.PlayButtonFont,
	}))
	s.editButton.Handle(ui.Click, func(p render.Point) {
		s.EditLevel()
	})
	s.supervisor.Add(s.editButton)

	// Initialize the drawing canvas.
	s.drawing = uix.NewCanvas(balance.ChunkSize, false)
	s.drawing.Name = "play-canvas"
	s.drawing.MoveTo(render.Origin)
	s.drawing.Resize(render.NewRect(int32(d.width), int32(d.height)))
	s.drawing.Compute(d.Engine)

	// Given a filename or map data to play?
	if s.Level != nil {
		log.Debug("PlayScene.Setup: received level from scene caller")
		s.drawing.LoadLevel(d.Engine, s.Level)
		s.drawing.InstallActors(s.Level.Actors)
	} else if s.Filename != "" {
		log.Debug("PlayScene.Setup: loading map from file %s", s.Filename)
		// NOTE: s.LoadLevel also calls s.drawing.InstallActors
		s.LoadLevel(s.Filename)
	}

	if s.Level == nil {
		log.Debug("PlayScene.Setup: no grid given, initializing empty grid")
		s.Level = level.New()
		s.drawing.LoadLevel(d.Engine, s.Level)
		s.drawing.InstallActors(s.Level.Actors)
	}

	// Load all actor scripts.
	s.drawing.SetScriptSupervisor(s.scripting)
	if err := s.scripting.InstallScripts(s.Level); err != nil {
		log.Error("PlayScene.Setup: failed to InstallScripts: %s", err)
	}

	// Load in the player character.
	player, err := doodads.LoadFile("azu-blu.doodad")
	if err != nil {
		log.Error("PlayScene.Setup: failed to load player doodad: %s", err)
		player = doodads.NewDummy(32)
	}

	s.Player = uix.NewActor("PLAYER", &level.Actor{}, player)
	s.Player.MoveTo(render.NewPoint(128, 128))
	s.drawing.AddActor(s.Player)
	s.drawing.FollowActor = s.Player.ID()

	// Set up the player character's script in the VM.
	if err := s.scripting.AddLevelScript(s.Player.ID()); err != nil {
		log.Error("PlayScene.Setup: scripting.InstallActor(player) failed: %s", err)
	}

	// Run all the actor scripts' main() functions.
	if err := s.drawing.InstallScripts(); err != nil {
		log.Error("PlayScene.Setup: failed to drawing.InstallScripts: %s", err)
	}

	d.Flash("Entered Play Mode. Press 'E' to edit this map.")
	s.running = true

	return nil
}

// SetupAlertbox configures the alert box UI.
func (s *PlayScene) SetupAlertbox() {
	window := ui.NewWindow("Level Completed")
	window.Configure(ui.Config{
		Width:      320,
		Height:     160,
		Background: render.Grey,
	})
	window.Compute(s.d.Engine)

	{
		frame := ui.NewFrame("Open Drawing Frame")
		window.Pack(frame, ui.Pack{
			Anchor: ui.N,
			Fill:   true,
			Expand: true,
		})

		/******************
		 * Frame for selecting User Levels
		 ******************/

		label1 := ui.NewLabel(ui.Label{
			Text: "Congratulations on clearing the level!",
			Font: balance.LabelFont,
		})
		frame.Pack(label1, ui.Pack{
			Anchor: ui.N,
			FillX:  true,
			PadY:   16,
		})

		/******************
		 * Confirm/cancel buttons.
		 ******************/

		bottomFrame := ui.NewFrame("Button Frame")
		frame.Pack(bottomFrame, ui.Pack{
			Anchor: ui.N,
			FillX:  true,
			PadY:   8,
		})

		// Button factory for the various options.
		makeButton := func(text string, handler func()) *ui.Button {
			btn := ui.NewButton(text, ui.NewLabel(ui.Label{
				Font: balance.LabelFont,
				Text: text,
			}))
			btn.Handle(ui.Click, func(p render.Point) {
				handler()
			})
			bottomFrame.Pack(btn, ui.Pack{
				Anchor: ui.W,
				PadX:   2,
			})
			s.supervisor.Add(btn)
			btn.Hide() // all buttons hidden by default
			return btn
		}

		s.alertReplayButton = makeButton("Play Again", func() {
			s.RestartLevel()
		})
		s.alertEditButton = makeButton("Edit Level", func() {
			s.EditLevel()
		})
		s.alertNextButton = makeButton("Next Level", func() {
			s.d.Flash("Not Implemented")
		})
		s.alertExitButton = makeButton("Exit to Menu", func() {
			s.d.Goto(&MainScene{})
		})
	}

	s.alertBox = window
	s.alertBox.Hide()
}

// EditLevel toggles out of Play Mode to edit the level.
func (s *PlayScene) EditLevel() {
	log.Info("Edit Mode, Go!")
	s.d.Goto(&EditorScene{
		Filename: s.Filename,
		Level:    s.Level,
	})
}

// RestartLevel starts the level over again.
func (s *PlayScene) RestartLevel() {
	log.Info("Restart Level")
	s.d.Goto(&PlayScene{
		Filename: s.Filename,
		Level:    s.Level,
		CanEdit:  s.CanEdit,
	})
}

// Loop the editor scene.
func (s *PlayScene) Loop(d *Doodle, ev *events.State) error {
	// Update debug overlay values.
	*s.debWorldIndex = s.drawing.WorldIndexAt(render.NewPoint(ev.CursorX.Now, ev.CursorY.Now)).String()
	*s.debPosition = s.Player.Position().String() + " vel " + s.Player.Velocity().String()
	*s.debViewport = s.drawing.Viewport().String()
	*s.debScroll = s.drawing.Scroll.String()

	s.supervisor.Loop(ev)

	// Has the window been resized?
	if resized := ev.Resized.Now; resized {
		w, h := d.Engine.WindowSize()
		if w != d.width || h != d.height {
			d.width = w
			d.height = h
			s.drawing.Resize(render.NewRect(int32(d.width), int32(d.height)))
			return nil
		}
	}

	// Switching to Edit Mode?
	if s.CanEdit && ev.KeyName.Read() == "e" {
		s.EditLevel()
		return nil
	}

	// Is the simulation still running?
	if s.running {
		// Loop the script supervisor so timeouts/intervals can fire in scripts.
		if err := s.scripting.Loop(); err != nil {
			log.Error("PlayScene.Loop: scripting.Loop: %s", err)
		}

		s.movePlayer(ev)
		if err := s.drawing.Loop(ev); err != nil {
			log.Error("Drawing loop error: %s", err.Error())
		}
	}

	return nil
}

// Draw the pixels on this frame.
func (s *PlayScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	// Draw the level.
	s.drawing.Present(d.Engine, s.drawing.Point())

	// Draw out bounding boxes.
	d.DrawCollisionBox(s.Player)

	// Draw the Edit button.
	var (
		canSize       = s.drawing.Size()
		size          = s.editButton.Size()
		padding int32 = 8
	)
	s.editButton.Present(d.Engine, render.Point{
		X: canSize.W - size.W - padding,
		Y: canSize.H - size.H - padding,
	})

	// Draw the alert box window.
	if !s.alertBox.Hidden() {
		s.alertBox.Compute(d.Engine)
		s.alertBox.MoveTo(render.Point{
			X: int32(d.width/2) - (s.alertBox.Size().W / 2),
			Y: int32(d.height/2) - (s.alertBox.Size().H / 2),
		})
		s.alertBox.Present(d.Engine, s.alertBox.Point())
	}

	return nil
}

// movePlayer updates the player's X,Y coordinate based on key pressed.
func (s *PlayScene) movePlayer(ev *events.State) {
	var playerSpeed = int32(balance.PlayerMaxVelocity)
	// var gravity = int32(balance.Gravity)

	var velocity render.Point

	if ev.Down.Now {
		velocity.Y = playerSpeed
	}
	if ev.Left.Now {
		velocity.X = -playerSpeed
	}
	if ev.Right.Now {
		velocity.X = playerSpeed
	}
	if ev.Up.Now {
		velocity.Y = -playerSpeed
	}

	// // Apply gravity if not grounded.
	// if !s.Player.Grounded() {
	// 	// Gravity has to pipe through the collision checker, too, so it
	// 	// can't give us a cheated downward boost.
	// 	velocity.Y += gravity
	// }

	s.Player.SetVelocity(velocity)

	// TODO: invoke the player OnKeypress for animation testing
	// if velocity != render.Origin {
	s.scripting.To(s.Player.ID()).Events.RunKeypress(ev)
	// }
}

// Drawing returns the private world drawing, for debugging with the console.
func (s *PlayScene) Drawing() *uix.Canvas {
	return s.drawing
}

// LoadLevel loads a level from disk.
func (s *PlayScene) LoadLevel(filename string) error {
	s.Filename = filename

	level, err := level.LoadFile(filename)
	if err != nil {
		return fmt.Errorf("PlayScene.LoadLevel(%s): %s", filename, err)
	}

	s.Level = level
	s.drawing.LoadLevel(s.d.Engine, s.Level)
	s.drawing.InstallActors(s.Level.Actors)

	return nil
}

// Destroy the scene.
func (s *PlayScene) Destroy() error {
	return nil
}

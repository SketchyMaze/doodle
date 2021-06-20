package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/collision"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/keybind"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/physics"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
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
	screen     *ui.Frame // A window sized invisible frame to position UI elements.
	editButton *ui.Button

	// The alert box shows up when the level goal is reached and includes
	// buttons what to do next.
	alertBox          *ui.Window
	alertBoxLabel     *ui.Label
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
	Player            *uix.Actor
	playerPhysics     *physics.Mover
	antigravity       bool // Cheat: disable player gravity
	noclip            bool // Cheat: disable player clipping
	playerJumpCounter int  // limit jump length

	// Inventory HUD. Impl. in play_inventory.go
	invenFrame   *ui.Frame
	invenItems   []string // item list
	invenDoodads map[string]*uix.Canvas
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

	// Create an invisible 'screen' frame for UI elements to use for positioning.
	s.screen = ui.NewFrame("Screen")
	s.screen.Resize(render.NewRect(d.width, d.height))

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
	s.editButton.Handle(ui.Click, func(ed ui.EventData) error {
		s.EditLevel()
		return nil
	})
	s.supervisor.Add(s.editButton)

	// Set up the inventory HUD.
	s.setupInventoryHud()

	// Initialize the drawing canvas.
	s.drawing = uix.NewCanvas(balance.ChunkSize, false)
	s.drawing.Name = "play-canvas"
	s.drawing.MoveTo(render.Origin)
	s.drawing.Resize(render.NewRect(d.width, d.height))
	s.drawing.Compute(d.Engine)

	// Handler when an actor touches water or fire.
	s.drawing.OnLevelCollision = func(a *uix.Actor, col *collision.Collide) {
		if col.InFire != "" {
			a.Canvas.MaskColor = render.Black
			if a.ID() == "PLAYER" { // only the player dies in fire.
				s.DieByFire(col.InFire)
			}
		} else if col.InWater {
			a.Canvas.MaskColor = render.DarkBlue
		} else {
			a.Canvas.MaskColor = render.Invisible
		}
	}

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
	s.setupPlayer()

	// Run all the actor scripts' main() functions.
	if err := s.drawing.InstallScripts(); err != nil {
		log.Error("PlayScene.Setup: failed to drawing.InstallScripts: %s", err)
	}

	if s.CanEdit {
		d.Flash("Entered Play Mode. Press 'E' to edit this map.")
	} else {
		d.Flash("%s", s.Level.Title)
	}

	s.running = true

	return nil
}

// setupPlayer creates and configures the Player Character in the level.
func (s *PlayScene) setupPlayer() {
	// Load in the player character.
	player, err := doodads.LoadFile(balance.PlayerCharacterDoodad)
	if err != nil {
		log.Error("PlayScene.Setup: failed to load player doodad: %s", err)
		player = doodads.NewDummy(32)
	}

	// Find the spawn point of the player. Search the level for the
	// "start-flag.doodad"
	var (
		spawn     render.Point
		flagCount int
	)
	for actorID, actor := range s.Level.Actors {
		if actor.Filename == "start-flag.doodad" {
			if flagCount > 1 {
				break
			}

			// TODO: start-flag.doodad is 86x86 pixels but we can't tell that
			// from right here.
			size := render.NewRect(86, 86)
			log.Info("Found start-flag.doodad at %s (ID %s)", actor.Point, actorID)
			spawn = render.NewPoint(
				// X: centered inside the flag.
				actor.Point.X+(size.W/2)-(player.Layers[0].Chunker.Size/2),

				// Y: the bottom of the flag, 4 pixels from the floor.
				actor.Point.Y+size.H-4-(player.Layers[0].Chunker.Size),
			)
			flagCount++
		}
	}

	// Surface warnings around the spawn flag.
	if flagCount == 0 {
		s.d.Flash("Warning: this level contained no Start Flag.")
	} else if flagCount > 1 {
		s.d.Flash("Warning: this level contains multiple Start Flags. Player spawn point is ambiguous.")
	}

	s.Player = uix.NewActor("PLAYER", &level.Actor{}, player)
	s.Player.MoveTo(spawn)
	s.drawing.AddActor(s.Player)
	s.drawing.FollowActor = s.Player.ID()

	// Set up the movement physics for the player.
	s.playerPhysics = &physics.Mover{
		MaxSpeed: physics.NewVector(balance.PlayerMaxVelocity, balance.PlayerMaxVelocity),
		// Gravity:      physics.NewVector(balance.Gravity, balance.Gravity),
		Acceleration: 0.025,
		Friction:     0.1,
	}

	// Set up the player character's script in the VM.
	if err := s.scripting.AddLevelScript(s.Player.ID()); err != nil {
		log.Error("PlayScene.Setup: scripting.InstallActor(player) failed: %s", err)
	}
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
			Side:   ui.N,
			Fill:   true,
			Expand: true,
		})

		/******************
		 * Frame for selecting User Levels
		 ******************/

		s.alertBoxLabel = ui.NewLabel(ui.Label{
			Text: "Congratulations on clearing the level!",
			Font: balance.LabelFont,
		})
		frame.Pack(s.alertBoxLabel, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  16,
		})

		/******************
		 * Confirm/cancel buttons.
		 ******************/

		bottomFrame := ui.NewFrame("Button Frame")
		frame.Pack(bottomFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  8,
		})

		// Button factory for the various options.
		makeButton := func(text string, handler func()) *ui.Button {
			btn := ui.NewButton(text, ui.NewLabel(ui.Label{
				Font: balance.LabelFont,
				Text: text,
			}))
			btn.Handle(ui.Click, func(ed ui.EventData) error {
				handler()
				return nil
			})
			bottomFrame.Pack(btn, ui.Pack{
				Side: ui.W,
				PadX: 2,
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

// DieByFire ends the level by "fire", or w/e the swatch is named.
func (s *PlayScene) DieByFire(name string) {
	log.Info("Watch out for %s!", name)
	s.alertBox.Title = "You've died!"
	s.alertBoxLabel.Text = fmt.Sprintf("Watch out for %s!", name)

	s.alertReplayButton.Show()
	if s.CanEdit {
		s.alertEditButton.Show()
	}
	s.alertExitButton.Show()

	s.alertBox.Show()

	// Stop the simulation.
	s.running = false
}

// Loop the editor scene.
func (s *PlayScene) Loop(d *Doodle, ev *event.State) error {
	// Update debug overlay values.
	*s.debWorldIndex = s.drawing.WorldIndexAt(render.NewPoint(ev.CursorX, ev.CursorY)).String()
	*s.debPosition = s.Player.Position().String() + " vel " + s.Player.Velocity().String()
	*s.debViewport = s.drawing.Viewport().String()
	*s.debScroll = s.drawing.Scroll.String()

	s.supervisor.Loop(ev)

	// Has the window been resized?
	if ev.WindowResized {
		w, h := d.Engine.WindowSize()
		if w != d.width || h != d.height {
			d.width = w
			d.height = h
			s.drawing.Resize(render.NewRect(d.width, d.height))
			return nil
		}
	}

	// Switching to Edit Mode?
	if s.CanEdit && keybind.GotoEdit(ev) {
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

		// Update the inventory HUD.
		s.computeInventory()
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
	if DebugCollision {
		for _, actor := range s.drawing.Actors() {
			d.DrawCollisionBox(s.drawing, actor)
		}
	}

	// Draw the UI screen and any widgets that attached to it.
	s.screen.Compute(d.Engine)
	s.screen.Present(d.Engine, render.Origin)

	// Draw the Edit button.
	var (
		canSize = s.drawing.Size()
		size    = s.editButton.Size()
		padding = 8
	)
	s.editButton.MoveTo(render.Point{
		X: canSize.W - size.W - padding,
		Y: canSize.H - size.H - padding,
	})
	s.editButton.Present(d.Engine, s.editButton.Point())

	// Draw the alert box window.
	if !s.alertBox.Hidden() {
		s.alertBox.Compute(d.Engine)
		s.alertBox.MoveTo(render.Point{
			X: (d.width / 2) - (s.alertBox.Size().W / 2),
			Y: (d.height / 2) - (s.alertBox.Size().H / 2),
		})
		s.alertBox.Present(d.Engine, s.alertBox.Point())
	}

	return nil
}

// movePlayer updates the player's X,Y coordinate based on key pressed.
func (s *PlayScene) movePlayer(ev *event.State) {
	var (
		playerSpeed = float64(balance.PlayerMaxVelocity)
		velocity    = s.Player.Velocity()
		direction   float64
		jumping     bool
	)

	// Antigravity: player can move anywhere with arrow keys.
	if s.antigravity {
		velocity.X = 0
		velocity.Y = 0

		// Shift to slow your roll to 1 pixel per tick.
		if keybind.Shift(ev) {
			playerSpeed = 1
		}

		if keybind.Left(ev) {
			velocity.X = -playerSpeed
		} else if keybind.Right(ev) {
			velocity.X = playerSpeed
		}
		if keybind.Up(ev) {
			velocity.Y = -playerSpeed
		} else if keybind.Down(ev) {
			velocity.Y = playerSpeed
		}
	} else {
		// Moving left or right.
		if keybind.Left(ev) {
			direction = -1
		} else if keybind.Right(ev) {
			direction = 1
		}

		// Up button to signal they want to jump.
		if keybind.Up(ev) && (s.Player.Grounded() || s.playerJumpCounter >= 0) {
			jumping = true

			if s.Player.Grounded() {
				// Allow them to sustain the jump this many ticks.
				s.playerJumpCounter = 32
			}
		}

		// Moving left or right? Interpolate their velocity by acceleration.
		if direction != 0 {
			// TODO: fast turn-around if they change directions so they don't
			// slip and slide while their velocity updates.
			velocity.X = physics.Lerp(
				velocity.X,
				direction*s.playerPhysics.MaxSpeed.X,
				s.playerPhysics.Acceleration,
			)
		} else {
			// Slow them back to zero using friction.
			velocity.X = physics.Lerp(
				velocity.X,
				0,
				s.playerPhysics.Friction,
			)
		}

		// Moving upwards (jumping): give them full acceleration upwards.
		if jumping {
			velocity.Y = -playerSpeed
		}

		// While in the air, count down their jump counter; when zero they
		// cannot jump again until they touch ground.
		if !s.Player.Grounded() {
			s.playerJumpCounter--
		}
	}

	// Move the player unless frozen.
	// TODO: if Y=0 then gravity fails, but not doing this allows the
	// player to jump while frozen. Not a HUGE deal right now as only Warp Doors
	// freeze the player currently but do address this later.
	if s.Player.IsFrozen() {
		velocity.X = 0
	}
	s.Player.SetVelocity(velocity)

	// If the "Use" key is pressed, set an actor flag on the player.
	s.Player.SetUsing(keybind.Use(ev))

	s.scripting.To(s.Player.ID()).Events.RunKeypress(keybind.FromEvent(ev))
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

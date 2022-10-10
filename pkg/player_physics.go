package doodle

// Subset of the PlayScene that is responsible for movement of the player character.

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/keybind"
	"git.kirsle.net/SketchyMaze/doodle/pkg/physics"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/go/render/event"
)

// movePlayer updates the player's X,Y coordinate based on key pressed.
func (s *PlayScene) movePlayer(ev *event.State) {
	var (
		playerSpeed = float64(balance.PlayerMaxVelocity)
		velocity    = s.Player.Velocity()
		direction   float64
		jumping     bool
		phys        = s.playerPhysics
		// holdingJump bool // holding down the jump button vs. tapping it
	)

	if s.slippery {
		phys = s.slipperyPhysics
	}

	// Antigravity: player can move anywhere with arrow keys.
	if s.antigravity || !s.Player.HasGravity() {
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
		if keybind.Up(ev) {
			if s.Player.IsWet() {
				// If they are holding Up put a cooldown in how fast they can swim
				// to the surface. Tapping the Jump button allows a faster ascent.
				if shmem.Tick > s.jumpCooldownUntil {
					s.jumpCooldownUntil = shmem.Tick + balance.SwimJumpCooldown
					velocity.Y = balance.SwimJumpVelocity
				}
			} else if s.Player.Grounded() {
				velocity.Y = balance.PlayerJumpVelocity
			}
		} else {
			s.jumpCooldownUntil = 0
			if velocity.Y < 0 {
				velocity.Y = 0
			}
		}

		// Moving left or right? Interpolate their velocity by acceleration.
		if direction != 0 {
			if s.playerLastDirection != direction {
				velocity.X = 0
			}

			// TODO: fast turn-around if they change directions so they don't
			// slip and slide while their velocity updates.
			velocity.X = physics.Lerp(
				velocity.X,
				direction*phys.MaxSpeed.X,
				phys.Acceleration,
			)
		} else {
			// Slow them back to zero using friction.
			velocity.X = physics.Lerp(
				velocity.X,
				0,
				phys.Friction,
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

	s.playerLastDirection = direction

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

	// Camera behaviors: Anvils can take the camera's focus while they're falling
	// but player inputs will take control back to the player. Most anvils will fall
	// a couple pixels upon level load - prevent them taking the camera's focus for
	// the first few frames of gameplay.
	if s.mustFollowPlayerUntil == 0 {
		s.mustFollowPlayerUntil = shmem.Tick + balance.FollowPlayerFirstTicks
	}

	// If we insist that the canvas follow the player doodad.
	if shmem.Tick < s.mustFollowPlayerUntil || keybind.Up(ev) || keybind.Left(ev) || keybind.Right(ev) || keybind.Use(ev) {
		s.drawing.FollowActor = s.Player.ID()
	}

	s.scripting.To(s.Player.ID()).Events.RunKeypress(keybind.FromEvent(ev))
}

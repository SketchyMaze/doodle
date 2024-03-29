package doodle

import (
	"strings"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal/loadscreen"
	"git.kirsle.net/SketchyMaze/doodle/pkg/scripting"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/windows"
	"git.kirsle.net/go/ui"
	"github.com/dop251/goja"
)

// IsDefaultPlayerCharacter checks whether the DefaultPlayerCharacter doodad has
// been modified

// MakeCheatsWindow initializes the windows/cheats_menu.go window from anywhere you need it,
// binding all the variables in. If you pass a nil Supervisor, this function will attempt to
// find one based on your Scene and
func (d *Doodle) MakeCheatsWindow(supervisor *ui.Supervisor) *ui.Window {
	// If not given a supervisor, try and find one.
	if supervisor == nil {
		if v, err := d.FindLikelySupervisor(); err != nil {
			d.FlashError("Couldn't make cheats window: %s", err)
			return nil
		} else {
			supervisor = v
		}
	}

	cfg := windows.CheatsMenu{
		Supervisor: supervisor,
		Engine:     d.Engine,
		SceneName: func() string {
			return d.Scene.Name()
		},
		RunCommand: func(command string) {
			// If we are in Play Mode, every command out of here is cheating.
			if playScene, ok := d.Scene.(*PlayScene); ok {
				playScene.SetCheated()
			}
			d.shell.Execute(command)
		},
		OnSetPlayerCharacter: func(doodad string) {
			if scene, ok := d.Scene.(*PlayScene); ok {
				scene.SetCheated()
				scene.SetPlayerCharacter(doodad)
			} else {
				shmem.FlashError("This only works during Play Mode.")
			}
		},
	}
	return windows.MakeCheatsMenu(cfg)
}

// SetPlayerCharacter -- this is designed to be called in-game with the developer
// console. Sets your player character to whatever doodad you want, not just the
// few that have cheat codes. If you set an invalid filename, you become the
// dummy default doodad sprite (a red "X").
func (d *Doodle) SetPlayerCharacter(filename string) {
	balance.PlayerCharacterDoodad = filename
	if playScene, isPlay := d.Scene.(*PlayScene); isPlay {
		playScene.SetPlayerCharacter(balance.PlayerCharacterDoodad)
	}
}

// cheatCommand is a subroutine of the Command.Run() method of the Doodle
// developer shell (commands.go). It looks for special cheat codes entered
// into the command shell and executes them.
//
// Returns true if a cheat was intercepted, false if the command is not a cheat.
func (c Command) cheatCommand(d *Doodle) bool {
	// Some cheats only work in Play Mode.
	playScene, isPlay := d.Scene.(*PlayScene)

	// If a character cheat is used during Play Mode, replace the player NOW.
	var setPlayerCharacter bool

	// Cheat codes
	switch c.Raw {
	case balance.CheatUncapFPS:
		if fpsDoNotCap {
			d.Flash("Reset frame rate throttle to factory default FPS")
		} else {
			d.Flash("Unleashing as many frames as we can render!")
		}
		fpsDoNotCap = !fpsDoNotCap

	case balance.CheatEditDuringPlay:
		if isPlay {
			playScene.drawing.Editable = true
			playScene.SetCheated()
			d.Flash("Level canvas is now editable. Don't edit and drive!")
		} else {
			d.FlashError("Use this cheat in Play Mode to make the level canvas editable.")
		}

	case balance.CheatScrollDuringPlay:
		if isPlay {
			playScene.drawing.Scrollable = true
			playScene.SetCheated()
			d.Flash("Level canvas is now scrollable with the arrow keys.")
		} else {
			d.FlashError("Use this cheat in Play Mode to make the level scrollable.")
		}

	case balance.CheatAntigravity:
		if isPlay {
			playScene.SetCheated()

			playScene.antigravity = !playScene.antigravity
			playScene.Player.SetGravity(!playScene.antigravity)

			if playScene.antigravity {
				d.Flash("Gravity disabled for player character.")
			} else {
				d.Flash("Gravity restored for player character.")
			}
		} else {
			d.FlashError("Use this cheat in Play Mode to disable gravity for the player character.")
		}

	case balance.CheatNoclip:
		if isPlay {
			playScene.SetCheated()

			playScene.noclip = !playScene.noclip
			playScene.Player.SetNoclip(playScene.noclip)

			playScene.antigravity = playScene.noclip
			playScene.Player.SetGravity(!playScene.antigravity)

			if playScene.noclip {
				d.Flash("Clipping disabled for player character.")
			} else {
				d.Flash("Clipping and gravity restored for player character.")
			}
		} else {
			d.FlashError("Use this cheat in Play Mode to disable clipping for the player character.")
		}

	case balance.CheatShowAllActors:
		if isPlay {
			playScene.SetCheated()
			for _, actor := range playScene.drawing.Actors() {
				actor.Show()
			}
			d.Flash("All invisible actors made visible.")
		} else {
			d.FlashError("Use this cheat in Play Mode to show hidden actors, such as technical doodads.")
		}

	case balance.CheatGiveKeys:
		if isPlay {
			playScene.SetCheated()
			playScene.Player.AddItem("key-red.doodad", 0)
			playScene.Player.AddItem("key-blue.doodad", 0)
			playScene.Player.AddItem("key-green.doodad", 0)
			playScene.Player.AddItem("key-yellow.doodad", 0)
			playScene.Player.AddItem("small-key.doodad", 99)
			d.Flash("Given all keys to the player character.")
		} else {
			d.FlashError("Use this cheat in Play Mode to get all colored keys.")
		}

	case balance.CheatGiveGems:
		if isPlay {
			playScene.SetCheated()
			playScene.Player.AddItem("gem-red.doodad", 1)
			playScene.Player.AddItem("gem-green.doodad", 1)
			playScene.Player.AddItem("gem-blue.doodad", 1)
			playScene.Player.AddItem("gem-yellow.doodad", 1)
			d.Flash("Given all gemstones to the player character.")
		} else {
			d.FlashError("Use this cheat in Play Mode to get all gemstones.")
		}

	case balance.CheatDropItems:
		if isPlay {
			playScene.SetCheated()
			playScene.Player.ClearInventory()
			d.Flash("Cleared inventory of player character.")
		} else {
			d.FlashError("Use this cheat in Play Mode to clear your inventory.")
		}

	case balance.CheatGodMode:
		if isPlay {
			d.Flash("God mode toggled")
			playScene.SetCheated()
			playScene.godMode = !playScene.godMode
			if playScene.godMode {
				d.FlashError("God mode enabled.")
			} else {
				d.Flash("God mode disabled.")
			}
		} else {
			d.FlashError("Use this cheat in Play Mode to toggle invincibility.")
		}

	case balance.CheatFreeEnergy:
		if isPlay {
			playScene.SetCheated()
			d.Flash("Power toggle sent to all actors in the level.")
			for _, a := range playScene.Canvas().Actors() {
				// Hacky stuff here - just a fun cheat code anyway.
				vm := playScene.ScriptSupervisor().To(a.ID())
				value := vm.Get("__tesla")
				if value == nil || !value.ToBoolean() {
					vm.Set("__tesla", true)
					value = vm.Get("__tesla")
				} else if value.ToBoolean() {
					vm.Set("__tesla", false)
					value = vm.Get("__tesla")
				}
				vm.Inbound <- scripting.Message{
					Name:     "power",
					SenderID: a.ID(),
					Args:     []goja.Value{value},
				}
			}
		} else {
			d.FlashError("Use this cheat in Play Mode to send power to all actors (chaotic!).")
		}

	case balance.CheatDebugLoadScreen:
		loadscreen.ShowWithProgress()
		loadscreen.SetSubtitle("Loading: /dev/null", "Loadscreen testing.")
		go func() {
			var i float64
			for i = 0; i < 100; i++ {
				time.Sleep(100 * time.Millisecond)
				loadscreen.SetProgress(i / 100)
			}
			loadscreen.Hide()
		}()

	case balance.CheatDebugWaitScreen:
		m := modal.Wait("Crunching some numbers...").WithTitle("Please hold").Then(func() {
			d.Flash("Wait modal dismissed.")
		})
		go func() {
			time.Sleep(10 * time.Second)
			m.Dismiss(true)
		}()

	case balance.CheatUnlockLevels:
		balance.CheatEnabledUnlockLevels = !balance.CheatEnabledUnlockLevels
		if balance.CheatEnabledUnlockLevels {
			d.Flash("All locked Story Mode levels can now be played.")
		} else {
			d.Flash("All locked Story Mode levels are again locked.")
		}

	case balance.CheatSkipLevel:
		if isPlay {
			playScene.SetCheated()
			playScene.ShowEndLevelModal(
				true,
				"Level Completed",
				"Great job, you cheated and 'won' the level!",
			)
		} else {
			d.Flash("Use this cheat in Play Mode to instantly win the level.")
		}

	default:
		// See if it was an endorsed actor cheat.
		if filename, ok := balance.CheatActors[strings.ToLower(c.Raw)]; ok {
			d.Flash("Set default player character to %s.", filename)
			balance.PlayerCharacterDoodad = filename
			setPlayerCharacter = true
		} else {
			// Not a cheat code.
			return false
		}
	}

	// If we're setting the player character and in Play Mode, do it.
	if setPlayerCharacter && isPlay {
		playScene.SetCheated()
		playScene.SetPlayerCharacter(balance.PlayerCharacterDoodad)
	}

	return true
}

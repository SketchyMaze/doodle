package doodle

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
)

// IsDefaultPlayerCharacter checks whether the DefaultPlayerCharacter doodad has
// been modified

// cheatCommand is a subroutine of the Command.Run() method of the Doodle
// developer shell (commands.go). It looks for special cheat codes entered
// into the command shell and executes them.
//
// Returns true if a cheat was intercepted, false if the command is not a cheat.
func (c Command) cheatCommand(d *Doodle) bool {
	// Some cheats only work in Play Mode.
	playScene, isPlay := d.Scene.(*PlayScene)

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

	case balance.CheatDropItems:
		if isPlay {
			playScene.SetCheated()
			playScene.Player.ClearInventory()
			d.Flash("Cleared inventory of player character.")
		} else {
			d.FlashError("Use this cheat in Play Mode to clear your inventory.")
		}

	case balance.CheatPlayAsBird:
		balance.PlayerCharacterDoodad = "bird-red.doodad"
		d.Flash("Set default player character to Bird (red)")

	case balance.CheatPlayAsBoy:
		balance.PlayerCharacterDoodad = "boy.doodad"
		d.Flash("Set default player character to Boy")

	case balance.CheatPlayAsAzuBlue:
		balance.PlayerCharacterDoodad = "azu-blu.doodad"
		d.Flash("Set default player character to Blue Azulian")

	case balance.CheatPlayAsThief:
		balance.PlayerCharacterDoodad = "thief.doodad"
		d.Flash("Set default player character to Thief")

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

	default:
		return false
	}

	return true
}

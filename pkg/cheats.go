package doodle

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
)

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
	case "unleash the beast":
		if fpsDoNotCap {
			d.Flash("Reset frame rate throttle to factory default FPS")
		} else {
			d.Flash("Unleashing as many frames as we can render!")
		}
		fpsDoNotCap = !fpsDoNotCap

	case "don't edit and drive":
		if isPlay {
			playScene.drawing.Editable = true
			d.Flash("Level canvas is now editable. Don't edit and drive!")
		} else {
			d.Flash("Use this cheat in Play Mode to make the level canvas editable.")
		}

	case "scroll scroll scroll your boat":
		if isPlay {
			playScene.drawing.Scrollable = true
			d.Flash("Level canvas is now scrollable with the arrow keys.")
		} else {
			d.Flash("Use this cheat in Play Mode to make the level scrollable.")
		}

	case "import antigravity":
		if isPlay {
			playScene.antigravity = !playScene.antigravity
			playScene.Player.SetGravity(!playScene.antigravity)

			if playScene.antigravity {
				d.Flash("Gravity disabled for player character.")
			} else {
				d.Flash("Gravity restored for player character.")
			}
		} else {
			d.Flash("Use this cheat in Play Mode to disable gravity for the player character.")
		}

	case "ghost mode":
		if isPlay {
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
			d.Flash("Use this cheat in Play Mode to disable clipping for the player character.")
		}

	case "show all actors":
		if isPlay {
			for _, actor := range playScene.drawing.Actors() {
				actor.Show()
			}
			d.Flash("All invisible actors made visible.")
		} else {
			d.Flash("Use this cheat in Play Mode to show hidden actors, such as technical doodads.")
		}

	case "give all keys":
		if isPlay {
			playScene.Player.AddItem("key-red.doodad", 0)
			playScene.Player.AddItem("key-blue.doodad", 0)
			playScene.Player.AddItem("key-green.doodad", 0)
			playScene.Player.AddItem("key-yellow.doodad", 0)
			playScene.Player.AddItem("small-key.doodad", 99)
			d.Flash("Given all keys to the player character.")
		} else {
			d.Flash("Use this cheat in Play Mode to get all colored keys.")
		}

	case "drop all items":
		if isPlay {
			playScene.Player.ClearInventory()
			d.Flash("Cleared inventory of player character.")
		} else {
			d.Flash("Use this cheat in Play Mode to clear your inventory.")
		}

	case "fly like a bird":
		balance.PlayerCharacterDoodad = "bird-red.doodad"
		d.Flash("Set default player character to Bird (red)")

	case "pinocchio":
		balance.PlayerCharacterDoodad = "boy.doodad"
		d.Flash("Set default player character to Boy")

	case "the cell":
		balance.PlayerCharacterDoodad = "azu-blu.doodad"
		d.Flash("Set default player character to Blue Azulian")

	case "play as thief":
		balance.PlayerCharacterDoodad = "thief.doodad"
		d.Flash("Set default player character to Thief")

	default:
		return false
	}

	return true
}

package balance

import magicform "git.kirsle.net/SketchyMaze/doodle/pkg/uix/magic-form"

// Store a copy of the PlayerCharacterDoodad original value.
var playerCharacterDefault string

func init() {
	playerCharacterDefault = PlayerCharacterDoodad
}

// IsPlayerCharacterDefault returns whether the balance.PlayerCharacterDoodad
// has been modified at runtime away from its built-in default. This is a cheat
// detection method: high scores could be tainted if you `fly like a bird` right
// to the exit in a couple of seconds.
func IsPlayerCharacterDefault() bool {
	return PlayerCharacterDoodad == playerCharacterDefault
}

// The game's cheat codes
var (
	CheatUncapFPS         = "unleash the beast"
	CheatEditDuringPlay   = "don't edit and drive"
	CheatScrollDuringPlay = "scroll scroll scroll your boat"
	CheatAntigravity      = "import antigravity"
	CheatNoclip           = "ghost mode"
	CheatShowAllActors    = "show all actors"
	CheatGiveKeys         = "give all keys"
	CheatGiveGems         = "give all gems"
	CheatDropItems        = "drop all items"
	CheatPlayAsBird       = "fly like a bird"
	CheatGodMode          = "god mode"
	CheatDebugLoadScreen  = "test load screen"
	CheatDebugWaitScreen  = "test wait screen"
	CheatUnlockLevels     = "master key"
	CheatSkipLevel        = "warp whistle"
	CheatFreeEnergy       = "tesla"
)

// Global cheat boolean states.
var (
	CheatEnabledUnlockLevels bool
)

// Actor replacement cheats
var CheatActors = map[string]string{
	"pinocchio":       PlayerCharacterDoodad,
	"the cell":        "azu-blu",
	"super azulian":   "azu-red",
	"hyper azulian":   "azu-white",
	"fly like a bird": "bird-red",
	"bluebird":        "bird-blue",
	"megaton weight":  "anvil",
	"play as thief":   "thief",
}

// Options for the "Play as:" drop-down in the Cheat Menu window.
var CheatMenuActors = []magicform.Option{
	{
		Value: "",
		Label: "Play as . . .",
	},
	{
		Value: "boy.doodad",
		Label: "Boy",
	},
	{
		Value: "thief.doodad",
		Label: "Thief",
	},
	{
		Value: "azu-blu.doodad",
		Label: "Azulian",
	},
	{
		Value: "bird-red.doodad",
		Label: "Bird",
	},
	{
		Value: "crusher.doodad",
		Label: "Crusher",
	},
	{
		Value: "snake.doodad",
		Label: "Snake",
	},
	{
		Value: "anvil.doodad",
		Label: "Anvil",
	},
}

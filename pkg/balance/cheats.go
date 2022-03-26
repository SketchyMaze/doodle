package balance

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
	CheatDropItems        = "drop all items"
	CheatPlayAsBird       = "fly like a bird"
	CheatPlayAsBoy        = "pinocchio"
	CheatPlayAsAzuBlue    = "the cell"
	CheatPlayAsThief      = "play as thief"
	CheatPlayAsAnvil      = "megaton weight"
	CheatGodMode          = "god mode"
	CheatDebugLoadScreen  = "test load screen"
	CheatUnlockLevels     = "master key"
)

// Global cheat boolean states.
var (
	CheatEnabledUnlockLevels bool
)

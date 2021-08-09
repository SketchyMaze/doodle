package scripting

/*
RegisterEventHooks attaches the supervisor level event hooks into a JS VM.

Names registered:

- EndLevel(): for a doodad to exit the level. Panics if the OnLevelExit
  handler isn't defined.
*/
func RegisterEventHooks(s *Supervisor, vm *VM) {
	vm.Set("EndLevel", func() {
		if s.onLevelFail == nil {
			panic("JS FailLevel(): No OnLevelFail handler attached to script supervisor")
		}
		s.onLevelExit()
	})
	vm.Set("FailLevel", func(message string) {
		if s.onLevelFail == nil {
			panic("JS FailLevel(): No OnLevelFail handler attached to script supervisor")
		}
		s.onLevelFail(message)
	})
}

// OnLevelExit registers an event hook for when a Level Exit doodad is reached.
func (s *Supervisor) OnLevelExit(handler func()) {
	s.onLevelExit = handler
}

// OnLevelFail registers an event hook for level failures (doodads killing the player).
func (s *Supervisor) OnLevelFail(handler func(string)) {
	s.onLevelFail = handler
}

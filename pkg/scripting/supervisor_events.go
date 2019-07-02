package scripting

/*
RegisterEventHooks attaches the supervisor level event hooks into a JS VM.

Names registered:

- EndLevel(): for a doodad to exit the level. Panics if the OnLevelExit
  handler isn't defined.
*/
func RegisterEventHooks(s *Supervisor, vm *VM) {
	vm.Set("EndLevel", func() {
		if s.onLevelExit == nil {
			panic("JS EndLevel(): no OnLevelExit handler attached to script supervisor")
		}
		s.onLevelExit()
	})
}

// OnLevelExit registers an event hook for when a Level Exit doodad is reached.
func (s *Supervisor) OnLevelExit(handler func()) {
	s.onLevelExit = handler
}

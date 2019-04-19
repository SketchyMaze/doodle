// Package scripting manages the JavaScript VMs for Doodad
// scripts.
package scripting

import (
	"errors"
	"fmt"
	"time"

	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
)

// Supervisor manages the JavaScript VMs for each doodad by its
// unique ID.
type Supervisor struct {
	scripts map[string]*VM
}

// NewSupervisor creates a new JavaScript Supervior.
func NewSupervisor() *Supervisor {
	return &Supervisor{
		scripts: map[string]*VM{},
	}
}

// Loop the supervisor to invoke timer events in any running scripts.
func (s *Supervisor) Loop() error {
	now := time.Now()
	for _, vm := range s.scripts {
		vm.TickTimer(now)
	}
	return nil
}

// InstallScripts loads scripts for all actors in the level.
func (s *Supervisor) InstallScripts(level *level.Level) error {
	for _, actor := range level.Actors {
		id := actor.ID()
		log.Debug("InstallScripts: load script from Actor %s", id)

		if _, ok := s.scripts[id]; ok {
			return fmt.Errorf("duplicate actor ID %s in level", id)
		}

		s.scripts[id] = NewVM(id)
		if err := s.scripts[id].RegisterLevelHooks(); err != nil {
			return err
		}
	}
	return nil
}

// To returns the VM for a named script.
func (s *Supervisor) To(name string) *VM {
	if vm, ok := s.scripts[name]; ok {
		return vm
	}

	// TODO: put this log back in, but add PLAYER script so it doesn't spam
	// the console for missing PLAYER.
	// log.Error("scripting.Supervisor.To(%s): no such VM but returning blank VM",
	// 	name,
	// )
	return NewVM(name)
}

// GetVM returns a script VM from the supervisor.
func (s *Supervisor) GetVM(name string) (*VM, error) {
	if vm, ok := s.scripts[name]; ok {
		return vm, nil
	}
	return nil, errors.New("not found")
}

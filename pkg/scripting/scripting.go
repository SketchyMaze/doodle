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

	// Global event handlers.
	onLevelExit func()
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
		if err := s.AddLevelScript(actor.ID()); err != nil {
			return err
		}
	}

	// Loop again to bridge channels together for linked VMs.
	for _, actor := range level.Actors {
		// Add linked actor IDs.
		if len(actor.Links) > 0 {
			// Bridge the links up.
			var thisVM = s.scripts[actor.ID()]
			for _, id := range actor.Links {
				// Assign this target actor's Inbound channel to the source
				// actor's array of Outbound channels.
				thisVM.Outbound = append(thisVM.Outbound, s.scripts[id].Inbound)
			}
		}
	}
	return nil
}

// AddLevelScript adds a script to the supervisor with level hooks.
func (s *Supervisor) AddLevelScript(id string) error {
	log.Debug("InstallScripts: load script from Actor %s", id)

	if _, ok := s.scripts[id]; ok {
		return fmt.Errorf("duplicate actor ID %s in level", id)
	}

	s.scripts[id] = NewVM(id)
	RegisterPublishHooks(s.scripts[id])
	RegisterEventHooks(s, s.scripts[id])
	if err := s.scripts[id].RegisterLevelHooks(); err != nil {
		return err
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

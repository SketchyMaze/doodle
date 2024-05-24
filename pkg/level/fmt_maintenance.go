package level

import "git.kirsle.net/SketchyMaze/doodle/pkg/log"

// Maintenance functions for the file format on disk.

// Vacuum runs any maintenance or migration tasks for the level at time of save.
//
// It will prune broken links between actors, or migrate internal data structures
// to optimize storage on disk of its binary data.
func (m *Level) Vacuum() error {
	if links := m.PruneLinks(); links > 0 {
		log.Debug("Vacuum: removed %d broken links between actors in this level.")
	}

	// Let the Chunker optimize accessor types.
	m.Chunker.OptimizeChunkerAccessors()

	return nil
}

// PruneLinks cleans up any Actor Links that can not be resolved in the
// level data. For example, if actors were linked in Edit Mode and one
// actor is deleted leaving a broken link.
//
// Returns the number of broken links pruned.
//
// This is called automatically in WriteFile.
func (m *Level) PruneLinks() int {
	var count int
	for id, actor := range m.Actors {
		var newLinks []string

		for _, linkID := range actor.Links {
			if _, ok := m.Actors[linkID]; !ok {
				log.Warn("Level.PruneLinks: actor %s (%s) was linked to unresolved actor %s",
					id,
					actor.Filename,
					linkID,
				)
				count++
				continue
			}
			newLinks = append(newLinks, linkID)
		}

		actor.Links = newLinks
	}
	return count
}

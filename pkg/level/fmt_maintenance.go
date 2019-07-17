package level

import "git.kirsle.net/apps/doodle/pkg/log"

// Maintenance functions for the file format on disk.

// PruneChunks cleans up any level chunks that have no pixel data.
func (m *Level) PruneChunks() int {
	var count int
	for coord, chunk := range m.Chunker.Chunks {
		if chunk.Len() == 0 {
			log.Info("PruneChunks: %d has no pixels", coord)
			delete(m.Chunker.Chunks, coord)
			count++
		}
	}
	return count
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

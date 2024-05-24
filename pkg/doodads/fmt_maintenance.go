package doodads

// Vacuum runs any maintenance or migration tasks for the level at time of save.
//
// It will prune broken links between actors, or migrate internal data structures
// to optimize storage on disk of its binary data.
func (m *Doodad) Vacuum() error {
	// Let the Chunker optimize accessor types.
	for _, layer := range m.Layers {
		layer.Chunker.OptimizeChunkerAccessors()
	}

	return nil
}

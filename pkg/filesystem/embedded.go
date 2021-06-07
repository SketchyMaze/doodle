package filesystem

// Embeddable file formats such as Levels or Doodads.
type Embeddable interface {
	GetFile(string) ([]byte, error)
}

/*
FindFileEmbedded searches for a file in a Level or Doodad's embedded filesystem,
before searching other places (as FindFile does) -- system paths and user paths.
*/
func FindFileEmbedded(filename string, em Embeddable) (string, error) {
	// Check in the embedded data.
	if _, err := em.GetFile(filename); err == nil {
		return filename, nil
	}

	// Not found in embedded, try the usual suspects.
	return FindFile(filename)
}

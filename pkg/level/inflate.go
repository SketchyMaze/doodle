package level

/*
Inflate the level from its compressed JSON form.

This is called as part of the LoadJSON function when the level is read
from disk. In the JSON format, pixels in the level refer to the palette
by its index number.

This function calls the following:

* Chunker.Inflate(Palette) to update references to the level's pixels to point
  to the Swatch entry.
* Actors.Inflate()
* Palette.Inflate() to load private instance values for the palette subsystem.
*/
func (l *Level) Inflate() {
	// Inflate the chunk metadata to map the pixels to their palette indexes.
	l.Chunker.Inflate(l.Palette)
	l.Actors.Inflate()

	// Inflate the private instance values.
	l.Palette.Inflate()
}

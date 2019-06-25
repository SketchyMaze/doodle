package level

import "fmt"

// PageType configures the bounds and wallpaper behavior of a Level.
type PageType int

// PageType values.
const (
	// Unbounded means the map can grow freely in any direction.
	// - Only the repeat texture is used for the wallpaper.
	Unbounded PageType = iota

	// NoNegativeSpace means the map is bounded at the top left edges.
	// - Can't scroll or visit any pixels in negative X,Y coordinates.
	// - Wallpaper shows the Corner at 0,0
	// - Wallpaper repeats the Top along the Y=0 plane
	// - Wallpaper repeats the Left along the X=0 plane
	// - The repeat texture fills the rest of the level.
	NoNegativeSpace

	// Bounded is the same as NoNegativeSpace but the level is imposing a
	// maximum cap on the width and height of the level.
	// - Can't scroll below X,Y origin at 0,0
	// - Can't scroll past the bounded width and height of the level
	Bounded

	// Bordered is like Bounded except the corner textures are wrapped
	// around the other edges of the level too.
	// - The wallpaper hoz mirrors Left along the X=Width plane
	// - The wallpaper vert mirrors Top along the Y=Width plane
	// - The wallpaper 180 rotates the Corner for opposite corners
	Bordered

	// If you add new PageType, also update the two functions below.
)

// String converts the PageType to a string label.
func (p PageType) String() string {
	switch p {
	case Unbounded:
		return "Unbounded"
	case NoNegativeSpace:
		return "NoNegativeSpace"
	case Bounded:
		return "Bounded"
	case Bordered:
		return "Bordered"
	}
	return fmt.Sprintf("PageType<%d>", p)
}

// PageTypeFromString returns a PageType from its string version.
func PageTypeFromString(name string) (PageType, bool) {
	// The min and max PageType value.
	for i := Unbounded; i <= Bordered; i++ {
		if name == i.String() {
			return PageType(i), true
		}
	}
	return 0, false
}

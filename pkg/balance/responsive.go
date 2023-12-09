package balance

/*
Responsive breakpoints and dimensions for Sketchy Maze.

Ideas for breakpoints (copying web CSS frameworks):
  - Mobile up to 768px
  - Tablet from 769px
  - Desktop from 1024px
  - Widescreen from 1216px
  - FullHD from 1408px
*/
const (
	// Title screen height needed for the main menu. Phones in landscape
	// mode will switch to the horizontal layout if less than this height.
	TitleScreenResponsiveHeight = 600

	BreakpointMobile     = 0    // 0-768
	BreakpointTablet     = 769  // from 769
	BreakpointDesktop    = 1024 // from 1024
	BreakpointWidescreen = 1216
	BreakpointFullHD     = 1408
)

// IsShortWide is a custom responsive breakpoint to mimic the mobile app in landscape mode like on a Pinephone.
//
// Parameters are the width and height of the application window (usually the screen if maximized).
//
// It is used on the MainScene to decide whether the main menu is drawn tall or wide.
func IsShortWide(width, height int) bool {
	return height < TitleScreenResponsiveHeight
}

func IsBreakpointTablet(width, height int) bool {
	return width >= BreakpointTablet
}

package balance

import (
	"os"
	"strconv"
	"strings"

	"git.kirsle.net/apps/doodle/render"
)

// Debug related variables that can toggle on or off certain features and
// overlays within the game.
var (
	/***************
	 * Visualizers *
	 ***************/

	// Debug overlay (FPS etc.) settings.
	DebugFontFilename       = "./fonts/DejaVuSans-Bold.ttf"
	DebugFontSize           = 15
	DebugLabelColor         = render.MustHexColor("#FF9900")
	DebugValueColor         = render.MustHexColor("#00CCFF")
	DebugStrokeDarken int32 = 80

	// Background color to use when exporting a drawing Chunk as a bitmap image
	// on disk. Default is white. Setting this to translucent yellow is a great
	// way to visualize the chunks loaded from cache on your screen.
	DebugChunkBitmapBackground = render.White // XXX: export $DEBUG_CHUNK_COLOR

	// Put a border around all Canvas widgets.
	DebugCanvasBorder = render.Invisible
	DebugCanvasLabel  = false // Tag the canvas with a label.
)

func init() {
	// Load values from environment variables.
	var config = map[string]interface{}{
		// Window size.
		"DOODLE_W": &Width,
		"DOODLE_H": &Height,

		// Tune some parameters. XXX: maybe dangerous at some point.
		"D_SCROLL_SPEED": &CanvasScrollSpeed,
		"D_DOODAD_SIZE":  &DoodadSize,

		// Shell settings.
		"D_SHELL_BG": &ShellBackgroundColor,
		"D_SHELL_FG": &ShellForegroundColor,
		"D_SHELL_PC": &ShellPromptColor,
		"D_SHELL_LN": &ShellHistoryLineCount,
		"D_SHELL_FS": &ShellFontSize,

		// Visualizers
		"DEBUG_CHUNK_COLOR":   &DebugChunkBitmapBackground,
		"DEBUG_CANVAS_BORDER": &DebugCanvasBorder,
		"DEBUG_CANVAS_LABEL":  &DebugCanvasLabel,
	}
	for name, value := range config {
		switch v := value.(type) {
		case *int:
			*v = IntEnv(name, *(v))
		case *bool:
			*v = BoolEnv(name, *(v))
		case *int32:
			*v = int32(IntEnv(name, int(*(v))))
		case *render.Color:
			*v = ColorEnv(name, *(v))
		}
	}

	// Debug all?
	if BoolEnv("DOODLE_DEBUG_ALL", false) {
		DebugChunkBitmapBackground = render.RGBA(255, 255, 0, 128)
		DebugCanvasBorder = render.Red
		DebugCanvasLabel = true
	}
}

// ColorEnv gets a color value from environment variable or returns a default.
// This will panic if the color is not valid, so only do this on startup time.
func ColorEnv(name string, v render.Color) render.Color {
	if color := os.Getenv(name); color != "" {
		return render.MustHexColor(color)
	}
	return v
}

// IntEnv gets an int value from environment variable or returns a default.
func IntEnv(name string, v int) int {
	if env := os.Getenv(name); env != "" {
		a, err := strconv.Atoi(env)
		if err != nil {
			panic(err)
		}
		return a
	}
	return v
}

// BoolEnv gets a bool from the environment with a default.
func BoolEnv(name string, v bool) bool {
	if env := os.Getenv(name); env != "" {
		switch strings.ToLower(env) {
		case "true", "t", "1", "on", "yes", "y":
			return true
		case "false", "f", "0", "off", "no", "n":
			return false
		}
	}
	return v
}

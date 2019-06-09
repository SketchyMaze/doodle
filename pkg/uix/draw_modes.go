package uix

// Tool is a draw mode for an editable Canvas.
type Tool int

// Draw modes for editable Canvas.
const (
	PencilTool Tool = iota // draw pixels where the mouse clicks
	ActorTool              // drag and move actors
	LinkTool
)

var toolNames = []string{
	"Pencil",
	"Doodad", // readable name for ActorTool
	"Link",
}

func (t Tool) String() string {
	return toolNames[t]
}

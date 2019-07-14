package drawtool

// Tool is a draw mode for an editable Canvas.
type Tool int

// Draw modes for editable Canvas.
const (
	PencilTool Tool = iota // draw pixels where the mouse clicks
	LineTool
	RectTool
	EllipseTool
	ActorTool // drag and move actors
	LinkTool
	EraserTool
)

var toolNames = []string{
	"Pencil",
	"Line",
	"Rectangle",
	"Ellipse",
	"Doodad", // readable name for ActorTool
	"Link",
	"Eraser",
}

func (t Tool) String() string {
	return toolNames[t]
}

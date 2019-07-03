package drawtool

// History manages a history of Strokes added to a drawing.
type History struct {
	limit int
	head  *HistoryElement // oldest history element, top of linked list
	tail  *HistoryElement // newest element added to history
}

// HistoryElement is a doubly linked list of stroke history.
type HistoryElement struct {
	stroke   *Stroke
	next     *HistoryElement
	previous *HistoryElement
}

// NewHistory initializes a History list.
func NewHistory(limit int) *History {
	return &History{
		limit: limit,
	}
}

// Reset clears the history.
func (h *History) Reset() {
	h.head = nil
	h.tail = nil
}

// Size returns the current size of the history list.
func (h *History) Size() int {
	var (
		size int
		node = h.head
	)

	for node != nil {
		size++
		node = node.next
	}

	return size
}

// Latest returns the tail of the history (the most recent stroke). If you had
// recently called Undo, the latest stroke may still have a 'next' stroke.
// Returns nil if there was no stroke in history.
func (h *History) Latest() *Stroke {
	if h.tail == nil {
		return nil
	}
	return h.tail.stroke
}

// Oldest returns the head of the history (the earliest stroke added). If the
// history size limit had been reached, the oldest stroke will creep along
// forward and not necessarily be the FIRST EVER stroke added.
func (h *History) Oldest() *Stroke {
	if h.head == nil {
		return nil
	}
	return h.head.stroke
}

// AddStroke adds a stroke to the history, becoming the new tail at the end
// of the history data.
func (h *History) AddStroke(s *Stroke) {
	var (
		elem = &HistoryElement{
			stroke: s,
		}
		tail = h.tail
	)

	// Make the current tail point to this one.
	if tail != nil {
		tail.next = elem
		elem.previous = tail
	}

	// First stroke of the history? Make it the head of the linked list.
	if h.head == nil {
		h.head = elem
	}

	h.tail = elem

	// Have we reached the history storage limit?
	var size = h.Size()
	if size > h.limit {
		var node = h.tail
		for i := 0; i < h.limit-1; i++ {
			if node.previous == nil {
				break
			}
			node = node.previous
		}
		h.head = node
		h.head.previous = nil
	}
}

// Undo steps back a step in the history. This sets the current tail to point
// to the "tail - 1" element, but doesn't change the link of that element to
// its future value yet; so that you can Redo it. But if you add a new stroke
// from this state, it will overwrite the tail.next and invalidate the old
// history that came after, starting a new branch of history from that point on.
//
// Returns false if the undo failed (no earlier node to move to).
func (h *History) Undo() bool {
	if h.tail == nil {
		return false
	}

	// if h.tail.previous == nil {
	// 	return false
	// }

	h.tail = h.tail.previous
	return true
}

// Redo advances forwards after a recent Undo. Note that if you added new strokes
// after an Undo, the new tail has no next node to move to and Redo returns false.
func (h *History) Redo() bool {
	if h.tail == nil || h.tail.next == nil {
		return false
	}
	h.tail = h.tail.next
	return true
}

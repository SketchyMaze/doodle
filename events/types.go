package events

// BoolFrameState holds boolean state between this frame and the previous.
type BoolFrameState struct {
	Now  bool
	Last bool
}

// Int32FrameState manages int32 state between this frame and the previous.
type Int32FrameState struct {
	Now  int32
	Last int32
}

// Push a bool state, copying the current Now value to Last.
func (bs *BoolFrameState) Push(v bool) {
	bs.Last = bs.Now
	bs.Now = v
}

// Push an int32 state, copying the current Now value to Last.
func (is *Int32FrameState) Push(v int32) {
	is.Last = is.Now
	is.Now = v
}

package events

// BoolTick holds boolean state between this frame and the previous.
type BoolTick struct {
	Now  bool
	Last bool
}

// Int32Tick manages int32 state between this frame and the previous.
type Int32Tick struct {
	Now  int32
	Last int32
}

// StringTick manages strings between this frame and the previous.
type StringTick struct {
	Now  string
	Last string
}

// Push a bool state, copying the current Now value to Last.
func (bs *BoolTick) Push(v bool) {
	bs.Last = bs.Now
	bs.Now = v
}

// Pressed returns true if the button was pressed THIS tick.
func (bs *BoolTick) Pressed() bool {
	return bs.Now && !bs.Last
}

// Read a bool state, resetting its value to false.
func (bs *BoolTick) Read() bool {
	now := bs.Now
	bs.Push(false)
	return now
}

// Push an int32 state, copying the current Now value to Last.
func (is *Int32Tick) Push(v int32) {
	is.Last = is.Now
	is.Now = v
}

// Push a string state.
func (s *StringTick) Push(v string) {
	s.Last = s.Now
	s.Now = v
}

// Read a string state, resetting its value.
func (s *StringTick) Read() string {
	now := s.Now
	s.Push("")
	return now
}

package ui

import (
	"fmt"
	"time"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/render/sdl"
)

// Target frames per second for the MainWindow to render at.
var (
	FPS = 60
)

// MainWindow is the parent window of a UI application.
type MainWindow struct {
	engine     render.Engine
	supervisor *Supervisor
	frame      *Frame
	w          int
	h          int
}

// NewMainWindow initializes the MainWindow. You should probably only have one
// of these per application.
func NewMainWindow(title string) (*MainWindow, error) {
	mw := &MainWindow{
		w:          800,
		h:          600,
		supervisor: NewSupervisor(),
	}

	mw.engine = sdl.New(
		title,
		mw.w,
		mw.h,
	)
	if err := mw.engine.Setup(); err != nil {
		return nil, err
	}

	// Add a default frame to the window.
	mw.frame = NewFrame("MainWindow Body")
	mw.frame.SetBackground(render.RGBA(0, 153, 255, 100))
	mw.Add(mw.frame)

	// Compute initial window size.
	mw.resized()

	return mw, nil
}

// Add a child widget to the window.
func (mw *MainWindow) Add(w Widget) {
	mw.supervisor.Add(w)
}

// Pack a child widget into the window's default frame.
func (mw *MainWindow) Pack(w Widget, pack Pack) {
	mw.Add(w)
	mw.frame.Pack(w, pack)
}

// resized handles the window being resized.
func (mw *MainWindow) resized() {
	mw.frame.Resize(render.Rect{
		W: int32(mw.w),
		H: int32(mw.h),
	})
}

// Present the window.
func (mw *MainWindow) Present() {
	mw.supervisor.Present(mw.engine)
}

// MainLoop starts the main event loop and blocks until there's an error.
func (mw *MainWindow) MainLoop() error {
	for true {
		if err := mw.Loop(); err != nil {
			return err
		}
	}
	return nil
}

// Loop does one loop of the UI.
func (mw *MainWindow) Loop() error {
	mw.engine.Clear(render.White)

	// Record how long this loop took.
	start := time.Now()

	// Poll for events.
	ev, err := mw.engine.Poll()
	if err != nil {
		return fmt.Errorf("event poll error: %s", err)
	}

	if ev.Resized.Now {
		w, h := mw.engine.WindowSize()
		if w != mw.w || h != mw.h {
			mw.w = w
			mw.h = h
			mw.resized()
		}
	}

	mw.frame.Compute(mw.engine)

	// Render the child widgets.
	mw.supervisor.Present(mw.engine)
	mw.engine.Present()

	// Delay to maintain target frames per second.
	var delay uint32
	var targetFPS = 1000 / FPS
	elapsed := time.Now().Sub(start) / time.Millisecond
	if targetFPS-int(elapsed) > 0 {
		delay = uint32(targetFPS - int(elapsed))
	}
	mw.engine.Delay(delay)

	return nil
}
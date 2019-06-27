// +build js,wasm

package main

import (
	"fmt"

	"syscall/js"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/render/canvas"
	doodle "git.kirsle.net/apps/doodle/pkg"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/log"
)

func main() {
	fmt.Printf("Hello world\n")
	// testRawCanvas()

	// Enable workarounds.
	balance.DisableChunkTextureCache = true
	js.Global().Get("sessionStorage").Call("clear")

	// HTML5 Canvas engine.
	engine, _ := canvas.New("canvas")
	engine.AddEventListeners()

	game := doodle.New(true, engine)
	game.SetupEngine()

	doodle.DebugOverlay = true

	// Manually inform Doodle of the canvas size since it can't control
	// the size on its own.
	w, h := engine.WindowSize()
	game.SetWindowSize(w, h)

	// game.Goto(&doodle.GUITestScene{})
	game.Goto(&doodle.EditorScene{})

	game.Run()
}

func testRawCanvas() {
	engine, _ := canvas.New("canvas")
	engine.SetTitle(
		fmt.Sprintf("%s v%s", branding.AppName, branding.Version),
	)
	fmt.Printf("Got engine: %+v\n", engine)
	engine.Clear(render.Green)

	for pt := range render.IterLine2(render.NewPoint(20, 20), render.NewPoint(300, 300)) {
		engine.DrawPoint(render.Red, pt)
	}

	engine.DrawLine(render.Blue, render.NewPoint(20, 300), render.NewPoint(300, 20))

	engine.DrawRect(render.Black, render.Rect{
		X: 5,
		Y: 5,
		W: 10,
		H: 10,
	})
	engine.DrawBox(render.White, render.Rect{
		X: 5,
		Y: 5,
		W: 10,
		H: 10,
	})

	engine.DrawBox(render.Purple, render.Rect{
		X: 25,
		Y: 5,
		W: 10,
		H: 10,
	})

	engine.DrawText(render.Text{
		Text:         "Hello world!",
		FontFilename: "DejaVuSans",
		Size:         14,
	}, render.NewPoint(400, 400))

	size, _ := engine.ComputeTextRect(render.Text{
		Text:         "Hello world! blah blah",
		FontFilename: "DejaVuSans",
		Size:         14,
	})
	log.Info("text rect: %+v", size)
	_ = engine
}

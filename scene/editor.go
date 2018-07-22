package scene

import "git.kirsle.net/apps/doodle/events"

// Editor is the drawing mode of the game where the user is clicking and
// dragging to draw pixels.
type Editor struct{}

func (s *Editor) String() string {
	return "Editor"
}

// Setup the scene.
func (s *Editor) Setup() error {
	return nil
}

// Loop the scene.
func (s *Editor) Loop(ev *events.State) error {
	// Taking a screenshot?
	if ev.ScreenshotKey.Pressed() {
		log.Info("Taking a screenshot")
		d.Screenshot()
		d.SaveLevel()
	}

	// Clear the canvas and fill it with white.
	d.renderer.SetDrawColor(255, 255, 255, 255)
	d.renderer.Clear()

	// Clicking? Log all the pixels while doing so.
	if ev.Button1.Now {
		pixel := Pixel{
			start: ev.Button1.Pressed(),
			x:     ev.CursorX.Now,
			y:     ev.CursorY.Now,
			dx:    ev.CursorX.Now,
			dy:    ev.CursorY.Now,
		}

		// Append unique new pixels.
		if len(pixelHistory) == 0 || pixelHistory[len(pixelHistory)-1] != pixel {
			// If not a start pixel, make the delta coord the previous one.
			if !pixel.start && len(pixelHistory) > 0 {
				prev := pixelHistory[len(pixelHistory)-1]
				pixel.dx = prev.x
				pixel.dy = prev.y
			}

			pixelHistory = append(pixelHistory, pixel)

			// Save in the pixel canvas map.
			d.canvas[pixel] = nil
		}
	}

	d.renderer.SetDrawColor(0, 0, 0, 255)
	for i, pixel := range pixelHistory {
		if !pixel.start && i > 0 {
			prev := pixelHistory[i-1]
			if prev.x == pixel.x && prev.y == pixel.y {
				d.renderer.DrawPoint(pixel.x, pixel.y)
			} else {
				d.renderer.DrawLine(
					pixel.x,
					pixel.y,
					prev.x,
					prev.y,
				)
			}
		}
		d.renderer.DrawPoint(pixel.x, pixel.y)
	}

	// Draw the FPS.
	d.DrawDebugOverlay()

	d.renderer.Present()

	return nil
}

// Destroy the scene.
func (s *Editor) Destroy() error {
	return nil
}

package doodle

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/campaign"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/apps/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// StoryScene manages the "Story Mode" menu selection screen.
type StoryScene struct {
	// Private variables.
	d       *Doodle
	running bool

	// Background wallpaper canvas.
	canvas *uix.Canvas

	// UI widgets.
	supervisor       *ui.Supervisor
	campaignFrame    *ui.Frame  // Select a Campaign screen
	levelSelectFrame *ui.Window // Select a level in the campaign screen

	// Pointer to the currently active frame.
	activeFrame *ui.Frame
}

// Name of the scene.
func (s *StoryScene) Name() string {
	return "Story"
}

// GotoStoryMenu initializes the story menu scene.
func (d *Doodle) GotoStoryMenu() {
	log.Info("Loading Story Scene")
	scene := &StoryScene{}
	d.Goto(scene)
}

// Setup the play scene.
func (s *StoryScene) Setup(d *Doodle) error {
	s.d = d

	// Set up the background wallpaper canvas.
	s.canvas = uix.NewCanvas(100, false)
	s.canvas.Resize(render.NewRect(d.width, d.height))
	s.canvas.LoadLevel(&level.Level{
		Chunker:   level.NewChunker(100),
		Palette:   level.NewPalette(),
		PageType:  level.Bounded,
		Wallpaper: "notebook.png",
	})

	s.supervisor = ui.NewSupervisor()

	// Set up the sub-screens of this scene.
	s.campaignFrame = s.setupCampaignFrame()
	s.levelSelectFrame = windows.NewLevelPackWindow(windows.LevelPack{
		Supervisor: s.supervisor,
		Engine:     d.Engine,

		OnPlayLevel: func(levelpack, filename string) {},
	})
	s.levelSelectFrame.Show()

	s.activeFrame = s.campaignFrame

	return nil
}

// setupCampaignFrame sets up the Campaign List screen.
func (s *StoryScene) setupCampaignFrame() *ui.Frame {
	var frame = ui.NewFrame("List Frame")
	frame.SetBackground(render.RGBA(0, 0, 255, 20))

	// Title label
	labelTitle := ui.NewLabel(ui.Label{
		Text: "Select a Story",
		Font: balance.TitleScreenFont,
	})
	labelTitle.Compute(s.d.Engine)
	frame.Place(labelTitle, ui.Place{
		Top:    120,
		Center: true,
	})

	// Buttons for campaign selection.
	{
		campaignFiles, err := campaign.List()
		if err != nil {
			log.Error("campaign.List: %s", err)
		}

		_ = campaignFiles
		// for _, file := range campaignFiles {
		//
		// }
	}

	frame.Resize(render.NewRect(s.d.width, s.d.height))
	frame.Compute(s.d.Engine)
	return frame
}

// Loop the story scene.
func (s *StoryScene) Loop(d *Doodle, ev *event.State) error {
	s.supervisor.Loop(ev)

	// Has the window been resized?
	if ev.WindowResized {
		w, h := d.Engine.WindowSize()
		if w != d.width || h != d.height {
			d.width = w
			d.height = h
			s.canvas.Resize(render.NewRect(d.width, d.height))
			s.activeFrame.Resize(render.NewRect(d.width, d.height))
			s.activeFrame.Compute(d.Engine)
			return nil
		}
	}

	return nil
}

// Draw the pixels on this frame.
func (s *StoryScene) Draw(d *Doodle) error {
	// Draw the background canvas.
	s.canvas.Present(d.Engine, render.Origin)

	// Draw the active screen.
	s.activeFrame.Present(d.Engine, render.Origin)

	s.supervisor.Present(d.Engine)

	return nil
}

// Destroy the scene.
func (s *StoryScene) Destroy() error {
	return nil
}

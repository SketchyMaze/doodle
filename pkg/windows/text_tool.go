package windows

import (
	"strconv"

	"git.kirsle.net/apps/doodle/assets"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	magicform "git.kirsle.net/apps/doodle/pkg/uix/magic-form"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// TextTool window.
type TextTool struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Callback when font settings are changed.
	OnChangeSettings func(font string, size int, message string)
}

// NewTextToolWindow initializes the window.
func NewTextToolWindow(cfg TextTool) *ui.Window {
	window := ui.NewWindow("Text Tool")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      330,
		Height:     170,
		Background: render.Grey,
	})

	// Text variables
	var (
		currentText = branding.AppName
		fontName    = balance.TextToolDefaultFont
		fontSize    = 16
	)

	// Get a listing of the available fonts.
	fonts, _ := assets.AssetDir("assets/fonts")
	var fontOption = []magicform.Option{}
	for _, font := range fonts {
		// Select the first font by default.
		if fontName == "" {
			fontName = font
		}

		fontOption = append(fontOption, magicform.Option{
			Label: font,
			Value: font,
		})
	}

	// Send the default config out.
	if cfg.OnChangeSettings != nil {
		cfg.OnChangeSettings(fontName, fontSize, currentText)
	}

	form := magicform.Form{
		Supervisor: cfg.Supervisor,
		Engine:     cfg.Engine,
		Vertical:   true,
		LabelWidth: 100,
		PadY:       2,
	}
	form.Create(window.ContentFrame(), []magicform.Field{
		{
			Label:       "Font Face:",
			Font:        balance.LabelFont,
			Options:     fontOption,
			SelectValue: fontName,
			OnSelect: func(v interface{}) {
				fontName = v.(string)
				if cfg.OnChangeSettings != nil {
					cfg.OnChangeSettings(fontName, fontSize, currentText)
				}
			},
		},
		{
			Label:       "Font Size:",
			Font:        balance.LabelFont,
			IntVariable: &fontSize,
			OnClick: func() {
				shmem.Prompt("Enter new font size: ", func(answer string) {
					if answer != "" {
						if i, err := strconv.Atoi(answer); err == nil {
							fontSize = i
							if cfg.OnChangeSettings != nil {
								cfg.OnChangeSettings(fontName, fontSize, currentText)
							}
						} else {
							shmem.FlashError("Not a valid font size: %s", answer)
						}
					}
				})
			},
		},
		{
			Label:        "Message:",
			Font:         balance.LabelFont,
			TextVariable: &currentText,
			OnClick: func() {
				shmem.PromptPre("Enter new message: ", currentText, func(answer string) {
					if answer != "" {
						currentText = answer
						if cfg.OnChangeSettings != nil {
							cfg.OnChangeSettings(fontName, fontSize, currentText)
						}
					}
				})
			},
		},
		{
			Label: "Be sure the Text Tool is selected, and click onto your\n" +
				"drawing to place this text onto it.",
			Font: balance.UIFont,
		},
	})

	return window
}

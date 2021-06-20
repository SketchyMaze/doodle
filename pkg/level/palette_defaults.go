package level

import (
	"git.kirsle.net/go/render"
)

// Some choice of palettes.
var (
	DefaultPaletteNames = []string{
		"Default",
		"Colored Pencil",
		"Blueprint",
	}

	DefaultPalettes = map[string]*Palette{
		"Default": {
			Swatches: []*Swatch{
				{
					Name:    "solid",
					Color:   render.MustHexColor("#777"),
					Solid:   true,
					Pattern: "noise.png",
				},
				{
					Name:    "decoration",
					Color:   render.MustHexColor("#CCC"),
					Pattern: "noise.png",
				},
				{
					Name:    "fire",
					Color:   render.Red,
					Fire:    true,
					Pattern: "marker.png",
				},
				{
					Name:    "water",
					Color:   render.MustHexColor("#09F"),
					Water:   true,
					Pattern: "ink.png",
				},
			},
		},

		"Colored Pencil": {
			Swatches: []*Swatch{
				{
					Name:    "grass",
					Color:   render.DarkGreen,
					Solid:   true,
					Pattern: "noise.png",
				},
				{
					Name:    "dirt",
					Color:   render.MustHexColor("#960"),
					Solid:   true,
					Pattern: "noise.png",
				},
				{
					Name:    "stone",
					Color:   render.Grey,
					Solid:   true,
					Pattern: "noise.png",
				},
				{
					Name:    "fire",
					Color:   render.Red,
					Fire:    true,
					Pattern: "marker.png",
				},
				{
					Name:    "water",
					Color:   render.RGBA(0, 153, 255, 255),
					Water:   true,
					Pattern: "ink.png",
				},
			},
		},

		"Blueprint": {
			Swatches: []*Swatch{
				{
					Name:    "solid",
					Color:   render.RGBA(254, 254, 254, 255),
					Solid:   true,
					Pattern: "noise.png",
				},
				{
					Name:    "decoration",
					Color:   render.Grey,
					Pattern: "noise.png",
				},
				{
					Name:    "fire",
					Color:   render.RGBA(255, 80, 0, 255),
					Fire:    true,
					Pattern: "marker.png",
				},
				{
					Name:    "water",
					Color:   render.RGBA(0, 153, 255, 255),
					Water:   true,
					Pattern: "ink.png",
				},
				{
					Name:    "electric",
					Color:   render.RGBA(255, 255, 0, 255),
					Solid:   true,
					Pattern: "marker.png",
				},
			},
		},
	}
)

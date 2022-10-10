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
		"Neon Bright",
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
					Pattern: "bubbles.png",
				},
				{
					Name:      "semisolid",
					Color:     render.MustHexColor("#AAA"),
					SemiSolid: true,
					Pattern:   "ink.png",
				},
				{
					Name:    "hint",
					Color:   render.MustHexColor("#F0F"),
					Pattern: "marker.png",
				},
			},
		},

		"Colored Pencil": {
			Swatches: []*Swatch{
				{
					Name:    "darkstone",
					Color:   render.MustHexColor("#777"),
					Pattern: "noise.png",
					Solid:   true,
				},
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
					Name:    "sandstone",
					Color:   render.RGBA(215, 114, 44, 255),
					Solid:   true,
					Pattern: "perlin-noise.png",
				},
				{
					Name:      "planks",
					Color:     render.MustHexColor("#B80"),
					SemiSolid: true,
					Pattern:   "marker.png",
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
					Pattern: "bubbles.png",
				},
				{
					Name:     "ice",
					Color:    render.MustHexColor("#099"),
					Solid:    true,
					Slippery: true,
					Pattern:  "marker.png",
				},
				{
					Name:    "hint",
					Color:   render.MustHexColor("#F0F"),
					Pattern: "marker.png",
				},
			},
		},

		"Neon Bright": {
			Swatches: []*Swatch{
				{
					Name:    "ground",
					Color:   render.MustHexColor("#FFE"),
					Solid:   true,
					Pattern: "noise.png",
				},
				{
					Name:    "grass green",
					Color:   render.Green,
					Solid:   true,
					Pattern: "noise.png",
				},
				{
					Name:    "fire",
					Color:   render.MustHexColor("#F90"),
					Pattern: "marker.png",
				},
				{
					Name:      "electricity",
					Color:     render.Yellow,
					SemiSolid: true,
					Pattern:   "perlin.png",
				},
				{
					Name:    "water",
					Color:   render.MustHexColor("#09F"),
					Pattern: "bubbles.png",
				},
				{
					Name:     "ice blue",
					Color:    render.MustHexColor("#0FF"),
					Solid:    true,
					Slippery: true,
					Pattern:  "marker.png",
				},
				{
					Name:    "hint",
					Color:   render.Magenta,
					Pattern: "marker.png",
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
					Pattern: "bubbles.png",
				},
				{
					Name:    "water",
					Color:   render.RGBA(0, 153, 255, 255),
					Water:   true,
					Pattern: "bubbles.png",
				},
				{
					Name:     "ice",
					Color:    render.MustHexColor("#FAF"),
					Solid:    true,
					Slippery: true,
					Pattern:  "marker.png",
				},
				{
					Name:      "electric",
					Color:     render.RGBA(255, 255, 0, 255),
					SemiSolid: true,
					Pattern:   "marker.png",
				},
				{
					Name:    "hint",
					Color:   render.MustHexColor("#F0F"),
					Pattern: "marker.png",
				},
			},
		},
	}
)

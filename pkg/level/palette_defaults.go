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
		"Default": &Palette{
			Swatches: []*Swatch{
				&Swatch{
					Name:  "solid",
					Color: render.Black,
					Solid: true,
				},
				&Swatch{
					Name:  "decoration",
					Color: render.Grey,
				},
				&Swatch{
					Name:  "fire",
					Color: render.Red,
					Fire:  true,
				},
				&Swatch{
					Name:  "water",
					Color: render.RGBA(0, 0, 255, 180),
					Water: true,
				},
			},
		},

		"Colored Pencil": &Palette{
			Swatches: []*Swatch{
				&Swatch{
					Name: "grass",
					Color: render.DarkGreen,
					Solid: true,
				},
				&Swatch{
					Name: "dirt",
					Color: render.RGBA(100, 64, 0, 255),
					Solid: true,
				},
				&Swatch{
					Name: "stone",
					Color: render.DarkGrey,
					Solid: true,
				},
				&Swatch{
					Name: "fire",
					Color: render.Red,
					Fire: true,
				},
				&Swatch{
					Name: "water",
					Color: render.RGBA(0, 153, 255, 255),
					Water: true,
				},
			},
		},

		"Blueprint": &Palette{
			Swatches: []*Swatch{
				&Swatch{
					Name:  "solid",
					Color: render.RGBA(254, 254, 254, 255),
					Solid: true,
				},
				&Swatch{
					Name:  "decoration",
					Color: render.Grey,
				},
				&Swatch{
					Name:  "fire",
					Color: render.RGBA(255, 80, 0, 255),
					Fire:  true,
				},
				&Swatch{
					Name:  "water",
					Color: render.RGBA(0, 153, 255, 255),
					Water: true,
				},
				&Swatch{
					Name:  "electric",
					Color: render.RGBA(255, 255, 0, 255),
					Solid: true,
				},
			},
		},
	}
)
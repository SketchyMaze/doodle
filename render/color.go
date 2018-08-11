package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var (
	// Regexps to parse hex color codes. Three formats are supported:
	// * reHexColor3 uses only 3 hex characters, like #F90
	// * reHexColor6 uses standard 6 characters, like #FF9900
	// * reHexColor8 is the standard 6 plus alpha channel, like #FF9900FF
	reHexColor3 = regexp.MustCompile(`^([A-Fa-f0-9])([A-Fa-f0-9])([A-Fa-f0-9])$`)
	reHexColor6 = regexp.MustCompile(`^([A-Fa-f0-9]{2})([A-Fa-f0-9]{2})([A-Fa-f0-9]{2})$`)
	reHexColor8 = regexp.MustCompile(`^([A-Fa-f0-9]{2})([A-Fa-f0-9]{2})([A-Fa-f0-9]{2})([A-Fa-f0-9]{2})$`)
)

// Color holds an RGBA color value.
type Color struct {
	Red   uint8
	Green uint8
	Blue  uint8
	Alpha uint8
}

// RGBA creates a new Color.
func RGBA(r, g, b, a uint8) Color {
	return Color{
		Red:   r,
		Green: g,
		Blue:  b,
		Alpha: a,
	}
}

// HexColor parses a color from hexadecimal code.
func HexColor(hex string) (Color, error) {
	c := Black // default color

	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	var m []string
	if len(hex) == 3 {
		m = reHexColor3.FindStringSubmatch(hex)
	} else if len(hex) == 6 {
		m = reHexColor6.FindStringSubmatch(hex)
	} else if len(hex) == 8 {
		m = reHexColor8.FindStringSubmatch(hex)
	} else {
		return c, errors.New("not a valid length for color code; only 3, 6 and 8 supported")
	}

	// Any luck?
	if m == nil {
		return c, errors.New("not a valid hex color code")
	}

	// Parse the color values. 16=base, 8=bit size
	red, _ := strconv.ParseUint(m[1], 16, 8)
	green, _ := strconv.ParseUint(m[2], 16, 8)
	blue, _ := strconv.ParseUint(m[3], 16, 8)

	// Alpha channel available?
	var alpha uint64 = 255
	if len(m) == 5 {
		alpha, _ = strconv.ParseUint(m[4], 16, 8)
	}

	c.Red = uint8(red)
	c.Green = uint8(green)
	c.Blue = uint8(blue)
	c.Alpha = uint8(alpha)
	return c, nil
}

func (c Color) String() string {
	return fmt.Sprintf(
		"Color<#%02x%02x%02x>",
		c.Red, c.Green, c.Blue,
	)
}

// MarshalJSON serializes the Color for JSON.
func (c Color) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(
		`"#%02x%02x%02x"`,
		c.Red, c.Green, c.Blue,
	)), nil
}

// UnmarshalJSON reloads the Color from JSON.
func (c *Color) UnmarshalJSON(b []byte) error {
	var hex string
	err := json.Unmarshal(b, &hex)
	if err != nil {
		return err
	}

	parsed, err := HexColor(hex)
	if err != nil {
		return err
	}

	c.Red = parsed.Red
	c.Blue = parsed.Blue
	c.Green = parsed.Green
	c.Alpha = parsed.Alpha
	return nil
}

// Add a relative color value to the color.
func (c Color) Add(r, g, b, a int32) Color {
	var (
		R = int32(c.Red) + r
		G = int32(c.Green) + g
		B = int32(c.Blue) + b
		A = int32(c.Alpha) + a
	)

	cap8 := func(v int32) uint8 {
		if v > 255 {
			v = 255
		} else if v < 0 {
			v = 0
		}
		return uint8(v)
	}

	return Color{
		Red:   cap8(R),
		Green: cap8(G),
		Blue:  cap8(B),
		Alpha: cap8(A),
	}
}

// Lighten a color value.
func (c Color) Lighten(v int32) Color {
	return c.Add(v, v, v, 0)
}

// Darken a color value.
func (c Color) Darken(v int32) Color {
	return c.Add(-v, -v, -v, 0)
}

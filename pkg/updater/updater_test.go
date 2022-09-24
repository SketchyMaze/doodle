package updater_test

import (
	"testing"

	"git.kirsle.net/SketchyMaze/doodle/pkg/updater"
)

// Test the semver logic.
func TestParseSemver(t *testing.T) {
	var tests = []struct {
		A      string // version strings
		B      string
		Expect bool // expect B is newer version than A
	}{
		{
			"0.1.0-alpha",
			"0.1.0-alpha",
			false,
		},
		{
			"0.1.0-alpha",
			"0.2.0-alpha",
			true,
		},
		{
			"0.2.0-alpha",
			"0.2.0-alpha",
			false,
		},
		{
			"0.2.0-alpha",
			"0.1.0-alpha",
			false,
		},
		{
			"0.2.0-alpha",
			"0.2.1-alpha",
			true,
		},
		{
			"0.1.0-alpha",
			"0.2.1-alpha",
			true,
		},
		{
			"0.2.1-alpha",
			"0.3.0",
			true,
		},
		{
			"0.3.0",
			"0.4.1",
			true,
		},
		{
			"0.1",
			"0.2.0",
			false,
		},
	}
	for i, test := range tests {
		result := updater.VersionInfo{
			LatestVersion: test.B,
		}.IsNewerVersionThan(test.A)
		if result != test.Expect {
			t.Errorf("Test %d: %s <> %s: expected %+v but got %+v",
				i, test.A, test.B, test.Expect, result,
			)
		}
	}
}

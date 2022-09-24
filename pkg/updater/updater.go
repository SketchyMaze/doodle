// Package updater checks for updates to Doodle.
package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
)

// VersionInfo holds the version.json data for self-update check.
type VersionInfo struct {
	LatestVersion string `json:"latestVersion"`
	DownloadURL   string `json:"downloadUrl"`
}

// IsNewerVersionThan checks if the version.json houses a newer
// game version than the currently running game's version string.
func (i VersionInfo) IsNewerVersionThan(versionString string) bool {
	// Parse the two semantic versions.
	curSemver, err := ParseSemver(versionString)
	if err != nil {
		log.Error("VersionInfo.IsNewerVersionThan: didn't parse semver: %s", err)
		return false
	}

	latestSemver, err := ParseSemver(i.LatestVersion)
	if err != nil {
		log.Error("VersionInfo.IsNewerVersionThan: didn't parse latest semver: %s", err)
		return false
	}

	// Check if there's a newer version.
	if latestSemver[Major] > curSemver[Major] {
		// We're a major version behind, like 0.x.x < 1.x.x
		return true
	} else if latestSemver[Major] == curSemver[Major] {
		// We're on the same major version, like 0.x.x
		// Check the minor version.
		if latestSemver[Minor] > curSemver[Minor] {
			return true
		} else if latestSemver[Minor] == curSemver[Minor] {
			// Same minor version, check patch version.
			if latestSemver[Patch] > curSemver[Patch] {
				return true
			}
		}
	}

	return false
}

// Last result of the update check, until forced to re-check.
var lastUpdate VersionInfo

// Check for new updates.
func Check() (VersionInfo, error) {
	var result VersionInfo

	// Return last cached check.
	if lastUpdate.LatestVersion != "" {
		return lastUpdate, nil
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", branding.UpdateCheckJSON, nil)
	if err != nil {
		return result, fmt.Errorf("updater.Check: HTTP error getting %s: %s", branding.UpdateCheckJSON, err)
	}
	req.Header.Add("User-Agent", branding.UserAgent())

	log.Debug("Checking for app updates")
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("updater.Check: HTTP error: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("updater.Check: unexpected HTTP status code %d", resp.StatusCode)
	}

	// Parse the JSON response.
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("updater.Check: JSON parse error: %s", err)
	}

	lastUpdate = result
	return result, nil
}

// CheckNow forces a re-check of the update info.
func CheckNow() (VersionInfo, error) {
	lastUpdate = VersionInfo{}
	return Check()
}

// ParseSemver parses a Project: Doodle semantic version number into its
// three integer parts. Historical versions of the game looked like
// either "0.1.0-alpha" or "0.7.0" and the given version string is
// expected to match that pattern.
func ParseSemver(versionString string) ([3]int, error) {
	var numbers [3]int
	result := semverRegexp.FindStringSubmatch(versionString)
	if result == nil {
		return numbers, fmt.Errorf("%s: didn't parse as a valid semver string", versionString)
	}

	// string to int helper
	var atoi = func(v string) int {
		a, _ := strconv.Atoi(v)
		return a
	}

	numbers = [3]int{
		atoi(result[1]),
		atoi(result[2]),
		atoi(result[3]),
	}
	return numbers, nil
}

var semverRegexp = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+).*`)

// Some semantic version indexing constants, to parse the result of
// ParseSemver with more semantic code.
const (
	Major int = iota
	Minor
	Patch
)

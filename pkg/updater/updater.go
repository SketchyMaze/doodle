// Package updater checks for updates to Doodle.
package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/log"
)

// VersionInfo holds the version.json data for self-update check.
type VersionInfo struct {
	LatestVersion string `json:"latestVersion"`
	DownloadURL   string `json:"downloadUrl"`
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

	log.Debug("Checking for app updates")

	resp, err := client.Get(branding.UpdateCheckJSON)
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

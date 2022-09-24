// +build !js

package native

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
)

// OpenURL opens a web browser to the given URL.
//
// On Linux this will look for xdg-open or try a few common browser names.
// On Windows this uses the ``start`` command.
// On MacOS this uses the ``open`` command.
func OpenURL(url string) {
	if runtime.GOOS == "windows" {
		go windowsOpenURL(url)
	} else if runtime.GOOS == "linux" {
		go linuxOpenURL(url)
	} else if runtime.GOOS == "darwin" {
		go macOpenURL(url)
	} else {
		log.Error("OpenURL: don't know how to open URLs")
	}
}

// OpenLocalURL opens a web browser to a local HTML path.
// It converts a relative path like "guidebook/index.html" to its absolute
// filesystem location.
func OpenLocalURL(path string) {
	abs, _ := filepath.Abs(path)
	url := fmt.Sprintf("file:///%s", abs)
	fmt.Printf("Open Local URL: %s\n", url)
	OpenURL(url)
}

func windowsOpenURL(url string) {
	_, err := exec.Command("start", url).Output()
	if err != nil {
		log.Error("native.windowsOpenURL(%s): %s", url, err)
	}
}

func macOpenURL(url string) {
	_, err := exec.Command("open", url).Output()
	if err != nil {
		log.Error("native.macOpenURL(%s): %s", url, err)
	}
}

func linuxOpenURL(url string) {
	// Commands to look for.
	var commands = []string{
		"xdg-open",
		"firefox",
		"google-chrome",
		"chromium-browser",
	}

	for _, command := range commands {
		log.Debug("OpenURL(linux): try %s %s", command, url)
		_, err := exec.Command(command, url).Output()
		if err == nil {
			return
		}
	}

	log.Error(
		"native.linuxOpenURL(%s): could not find browser executable, tried %+v",
		url, commands,
	)
}

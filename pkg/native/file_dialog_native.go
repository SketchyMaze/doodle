// +build !js

package native

import (
	"github.com/gen2brain/dlgs"
	"errors"
)

func init() {
	FileDialogsReady = true
}

// OpenFile invokes a native File Chooser dialog with the title
// and a set of file filters. The filters are a sequence of label
// and comma-separated file extensions.
//
// Example:
// OpenFile("Pick a file", "Images", "png,gif,jpg", "Audio", "mp3")
func OpenFile(title string, filter string) (string, error) {
	filename, ok, err := dlgs.File(title, filter, false)
	if err != nil {
		return "", err
	}

	if ok {
		return filename, nil
	}
	return "", errors.New("canceled")
}

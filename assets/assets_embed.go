// Package assets gets us off go-bindata by using Go 1.16 embed support.
//
// For Go 1.16 embed, this source file had to live inside the assets/ folder
// to embed the sub-files, so couldn't be under pkg/ like pkg/bindata/ was.
//
// Historically code referred to assets like "assets/fonts/DejaVuSans.ttf"
// but Go embed would just use "fonts/DejaVuSans.ttf" as that's what's relative
// to this source file.
//
// The functions in this module provide backwards compatibility by ignoring
// the "assets/" prefix sent by calling code.
package assets

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed *
var Embedded embed.FS

// AssetDir returns the list of embedded files at the directory name.
func AssetDir(name string) ([]string, error) {
	var result []string

	name = strings.TrimPrefix(name, "assets/")
	files, err := Embedded.ReadDir(name)
	if err != nil {
		return result, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		result = append(result, file.Name())
	}

	return result, nil
}

// Asset returns the byte data of an embedded asset.
func Asset(name string) ([]byte, error) {
	return Embedded.ReadFile(strings.TrimPrefix(name, "assets/"))
}

// AssetNames dumps the names of all embedded assets,
// with their legacy "assets/" prefix from go-bindata.
func AssetNames() []string {
	var result []string

	fs.WalkDir(Embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if d != nil && !d.IsDir() {
			result = append(result, "assets/"+path)
		}
		return nil
	})

	return result
}

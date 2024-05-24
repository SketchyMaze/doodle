//go:build doodad
// +build doodad

// Dummy version of assets_embed.go that doesn't embed any files.
// For the `doodad` tool.

package assets

import (
	"embed"
	"errors"
)

var Embedded embed.FS

var errNotEmbedded = errors.New("assets not embedded")

// AssetDir returns the list of embedded files at the directory name.
func AssetDir(name string) ([]string, error) {
	return nil, errNotEmbedded
}

// Asset returns the byte data of an embedded asset.
func Asset(name string) ([]byte, error) {
	return nil, errNotEmbedded
}

// AssetNames dumps the names of all embedded assets,
// with their legacy "assets/" prefix from go-bindata.
func AssetNames() []string {
	return nil
}

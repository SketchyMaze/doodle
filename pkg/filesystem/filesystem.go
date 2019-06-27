package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/bindata"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/userdir"
)

// Binary file format headers for Levels and Doodads.
//
// The header is 8 bytes long: "DOODLE" + file format version + file type number.
const (
	BinMagic            = "DOODLE"
	BinVersion    uint8 = 1 // version of the file format we support
	BinLevelType  uint8 = 1
	BinDoodadType uint8 = 2
)

// Paths to system-level assets bundled with the application.
var (
	SystemDoodadsPath = filepath.Join("assets", "doodads")
	SystemLevelsPath  = filepath.Join("assets", "levels")
)

// MakeHeader creates the binary file header.
func MakeHeader(filetype uint8) []byte {
	header := make([]byte, len(BinMagic)+2)
	for i := 0; i < len(BinMagic); i++ {
		header[i] = BinMagic[i]
	}

	header[len(header)-2] = BinVersion
	header[len(header)-1] = filetype

	return header
}

// ReadHeader reads and verifies a header from a filehandle.
func ReadHeader(filetype uint8, fh io.Reader) error {
	header := make([]byte, len(BinMagic)+2)
	_, err := fh.Read(header)
	if err != nil {
		return fmt.Errorf("ReadHeader: %s", err)
	}

	if string(header[:len(BinMagic)]) != BinMagic {
		return errors.New("not a doodle drawing (no magic number in header)")
	}

	// Verify the file format version and type.
	var (
		fileVersion = header[len(header)-2]
		fileType    = header[len(header)-1]
	)

	if fileVersion == 0 || fileVersion > BinVersion {
		return errors.New("binary format was created using a newer version of the game")
	} else if fileType != filetype {
		return errors.New("drawing type is not the type we expected")
	}

	return nil
}

/*
FindFile looks for a file (level or doodad) in a few places.

The filename should already have a ".level" or ".doodad" file extension. If
neither is given, the exact filename will be searched in all places.

1. Check in the files built into the program binary.
2. Check for system files in the binary's assets/ folder.
3. Check the user folders.

Returns the file path and an error if not found anywhere.
*/
func FindFile(filename string) (string, error) {
	var filetype string

	// Any hint on what type of file we're looking for?
	if strings.HasSuffix(filename, enum.LevelExt) {
		filetype = enum.LevelExt
	} else if strings.HasSuffix(filename, enum.DoodadExt) {
		filetype = enum.DoodadExt
	}

	// Search level directories.
	if filetype == enum.LevelExt || filetype == "" {
		// system levels
		candidate := filepath.Join(SystemLevelsPath, filename)

		// embedded system doodad?
		if _, err := bindata.Asset(candidate); err == nil {
			return candidate, nil
		}

		// WASM: can't check the filesystem. Let the caller go ahead and try
		// loading via ajax request.
		if runtime.GOOS == "js" {
			return candidate, nil
		}

		// external system level?
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			return candidate, nil
		}

		// user levels
		candidate = userdir.LevelPath(filename)
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			return candidate, nil
		}
	}

	// Search doodad directories.
	if filetype == enum.DoodadExt || filetype == "" {
		// system doodads path
		candidate := filepath.Join(SystemDoodadsPath, filename)

		// embedded system doodad?
		if _, err := bindata.Asset(candidate); err == nil {
			return candidate, nil
		}

		// WASM: can't check the filesystem. Let the caller go ahead and try
		// loading via ajax request.
		if runtime.GOOS == "js" {
			return candidate, nil
		}

		// external system doodad?
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			return candidate, nil
		}

		// user doodads
		candidate = userdir.DoodadPath(filename)
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			return candidate, nil
		}
	}

	return filename, errors.New("file not found")
}

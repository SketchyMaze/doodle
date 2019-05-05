package level

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"git.kirsle.net/apps/doodle/pkg/filesystem"
	"github.com/vmihailenco/msgpack"
)

// ToBinary serializes the level to binary format.
func (m *Level) ToBinary() ([]byte, error) {
	header := filesystem.MakeHeader(filesystem.BinLevelType)
	out := bytes.NewBuffer(header)
	encoder := msgpack.NewEncoder(out)
	err := encoder.Encode(m)
	return out.Bytes(), err
}

// WriteBinary writes a level to binary format on disk.
func (m *Level) WriteBinary(filename string) error {
	bin, err := m.ToBinary()
	if err != nil {
		return fmt.Errorf("Level.WriteBinary: encode error: %s", err)
	}

	err = ioutil.WriteFile(filename, bin, 0755)
	if err != nil {
		return fmt.Errorf("Level.WriteBinary: WriteFile error: %s", err)
	}

	return nil
}

// LoadBinary loads a map from binary file on disk.
func LoadBinary(filename string) (*Level, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	// Read and verify the file header from the binary format.
	err = filesystem.ReadHeader(filesystem.BinLevelType, fh)
	if err != nil {
		return nil, err
	}

	// Decode the file from disk.
	m := New()
	decoder := msgpack.NewDecoder(fh)
	err = decoder.Decode(&m)
	if err != nil {
		return m, fmt.Errorf("level.LoadBinary: decode error: %s", err)
	}

	// Fill in defaults.
	if m.Wallpaper == "" {
		m.Wallpaper = DefaultWallpaper
	}

	// Inflate the chunk metadata to map the pixels to their palette indexes.
	m.Chunker.Inflate(m.Palette)
	m.Actors.Inflate()

	// Inflate the private instance values.
	m.Palette.Inflate()
	return m, err
}

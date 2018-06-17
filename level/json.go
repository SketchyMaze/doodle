package level

import (
	"bytes"
	"encoding/json"
	"os"
)

// ToJSON serializes the level as JSON.
func (m *Level) ToJSON() ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(m)
	return out.Bytes(), err
}

// LoadJSON loads a map from JSON file.
func LoadJSON(filename string) (Level, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return Level{}, err
	}
	defer fh.Close()

	m := Level{}
	decoder := json.NewDecoder(fh)
	err = decoder.Decode(&m)
	return m, err
}

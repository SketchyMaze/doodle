// Package campaign contains types and functions for the single player campaigns.
package campaign

// Campaign structure for the JSON campaign files.
type Campaign struct {
	Version int     `json:"version"`
	Title   string  `json:"title"`
	Author  string  `json:"author"`
	Levels  []Level `json:"levels"`
}

// Level is the "levels" object of the JSON file.
type Level struct {
	Filename string `json:"filename"`
}

// LoadFile reads a campaign file from disk, checking a few locations.
// func LoadFile(filename string) (*Level, error) {
// 	// Search the system and user paths for this level.
// 	filename, err := filesystem.FindFile(filename)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Do we have the file in bindata?
// 	if jsonData, err := bindata.Asset(filename); err == nil {
// 		log.Info("loaded from embedded bindata")
// 		return FromJSON(filename, jsonData)
// 	}
//
// 	// WASM: try the file from localStorage or HTTP ajax request.
// 	if runtime.GOOS == "js" {
// 		if result, ok := wasm.GetSession(filename); ok {
// 			log.Info("recall level data from localStorage")
// 			return FromJSON(filename, []byte(result))
// 		}
//
// 		// Ajax request.
// 		jsonData, err := wasm.HTTPGet(filename)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		return FromJSON(filename, jsonData)
// 	}
//
// 	// Try the binary format.
// 	if level, err := LoadBinary(filename); err == nil {
// 		return level, nil
// 	} else {
// 		log.Warn(err.Error())
// 	}
//
// 	// Then the JSON format.
// 	if level, err := LoadJSON(filename); err == nil {
// 		return level, nil
// 	} else {
// 		log.Warn(err.Error())
// 	}
//
// 	return nil, errors.New("invalid file type")
// }

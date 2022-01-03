package savegame

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.kirsle.net/apps/doodle/pkg/usercfg"
	"git.kirsle.net/apps/doodle/pkg/userdir"
)

// SaveGame holds the user's progress thru level packs.
type SaveGame struct {
	LevelPacks map[string]*LevelPack `json:"levelPacks"`
}

// LevelPack holds savegame process for a level pack.
type LevelPack struct {
	Levels map[string]*Level `json:"levels"`
}

// Level holds high score information for a level.
type Level struct {
	Completed   bool           `json:"completed"`
	BestTime    *time.Duration `json:"bestTime"`
	PerfectTime *time.Duration `json:"perfectTime"`
}

// New creates a new SaveGame.
func New() *SaveGame {
	return &SaveGame{
		LevelPacks: map[string]*LevelPack{},
	}
}

// NewLevelPack initializes a LevelPack struct.
func NewLevelPack() *LevelPack {
	return &LevelPack{
		Levels: map[string]*Level{},
	}
}

// GetOrCreate the save game JSON. If the save file isn't found OR has an
// invalid checksum, it is created. Always returns a valid SaveGame struct
// and the error may communicate if there was a problem reading an existing file.
func GetOrCreate() (*SaveGame, error) {
	if sg, err := Load(); err == nil {
		return sg, nil
	} else {
		return New(), err
	}
}

// Load the save game JSON from the user's profile directory.
func Load() (*SaveGame, error) {
	fh, err := os.Open(userdir.SaveFile)
	if err != nil {
		return nil, err
	}

	// Read the checksum line.
	scanner := bufio.NewScanner(fh)
	scanner.Scan()
	var (
		checksum = scanner.Text()
		jsontext []byte
	)
	for scanner.Scan() {
		jsontext = append(jsontext, scanner.Bytes()...)
	}

	// Validate the checksum.
	if !verifyChecksum(jsontext, checksum) {
		return nil, errors.New("checksum error")
	}

	// Parse the JSON.
	var sg = New()
	err = json.Unmarshal(jsontext, sg)
	if err != nil {
		return nil, err
	}

	return sg, nil
}

// Save the savegame.json to disk.
func (sg *SaveGame) Save() error {
	// Encode to JSON.
	text, err := json.Marshal(sg)
	if err != nil {
		return err
	}

	// Create the checksum.
	checksum := makeChecksum(text)

	// Write the file.
	fh, err := os.Create(userdir.SaveFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	fh.Write([]byte(checksum))
	fh.Write([]byte{'\n'})
	fh.Write(text)

	return nil
}

// MarkCompleted is a helper function to mark a levelpack level completed.
// Parameters are the filename of the levelpack and the level therein.
// Extra path info except the base filename is stripped from both.
func (sg *SaveGame) MarkCompleted(levelpack, filename string) {
	lvl := sg.GetLevelScore(levelpack, filename)
	lvl.Completed = true
}

// NewHighScore may set a new highscore for a level.
//
// The level will be marked Completed and if the given score is better
// than the stored one it will update.
//
// Returns true if a new high score was logged.
func (sg *SaveGame) NewHighScore(levelpack, filename string, isPerfect bool, elapsed time.Duration) bool {
	levelpack = filepath.Base(levelpack)
	filename = filepath.Base(filename)

	score := sg.GetLevelScore(levelpack, filename)
	score.Completed = true
	var newHigh bool

	if isPerfect {
		if score.PerfectTime == nil || *score.PerfectTime > elapsed {
			score.PerfectTime = &elapsed
			newHigh = true
		}
	} else {
		if score.BestTime == nil || *score.BestTime > elapsed {
			score.BestTime = &elapsed
			newHigh = true
		}
	}

	if newHigh {
		if sg.LevelPacks[levelpack] == nil {
			sg.LevelPacks[levelpack] = NewLevelPack()
		}
		sg.LevelPacks[levelpack].Levels[filename] = score
	}

	return newHigh
}

// GetLevelScore finds or creates a default Level score.
func (sg *SaveGame) GetLevelScore(levelpack, filename string) *Level {
	levelpack = filepath.Base(levelpack)
	filename = filepath.Base(filename)

	if _, ok := sg.LevelPacks[levelpack]; !ok {
		sg.LevelPacks[levelpack] = NewLevelPack()
	}

	if row, ok := sg.LevelPacks[levelpack].Levels[filename]; ok {
		return row
	} else {
		row = &Level{}
		sg.LevelPacks[levelpack].Levels[filename] = row
		return row
	}
}

// CountCompleted returns the number of completed levels in a levelpack.
func (sg *SaveGame) CountCompleted(levelpack string) int {
	var count int
	levelpack = filepath.Base(levelpack)

	if lp, ok := sg.LevelPacks[levelpack]; ok {
		for _, lvl := range lp.Levels {
			if lvl.Completed {
				count++
			}
		}
	}

	return count
}

// FormatDuration pretty prints a time.Duration in MM:SS format.
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Millisecond)
	var (
		hour   = d / time.Hour
		minute = d / time.Minute
		second = d / time.Second
		ms     = fmt.Sprintf("%d", d/time.Millisecond%1000)
	)

	// Limit milliseconds to 2 digits.
	if len(ms) > 2 {
		ms = ms[:2]
	}

	return strings.TrimPrefix(
		fmt.Sprintf("%02d:%02d:%02d.%s", hour, minute, second, ms),
		"00:",
	)
}

// Hashing key that goes into the level's save data.
var secretKey = []byte(`Sc\x96R\x8e\xba\x96\x8e\x1fg\x01Q\xf5\xcbIX`)

func makeChecksum(jsontext []byte) string {
	h := sha1.New()
	h.Write(jsontext)
	h.Write(secretKey)
	h.Write(usercfg.Current.Entropy)
	return hex.EncodeToString(h.Sum(nil))
}

func verifyChecksum(jsontext []byte, checksum string) bool {
	expect := makeChecksum(jsontext)
	return expect == checksum
}

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

	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
)

// SaveGame holds the user's progress thru level packs.
type SaveGame struct {
	// DEPRECATED: savegame state spelled out by level packs and
	// filenames.
	LevelPacks map[string]*LevelPack `json:"levelPacks,omitempty"`

	// New data format: store high scores by level UUID. Adds a
	// nice layer of obfuscation + is more robust in case levels
	// move around between levelpacks, get renamed, etc. that
	// the user should be able to keep their high score.
	Levels map[string]*Level
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
		Levels:     map[string]*Level{},
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

/*
Migrate the savegame.json file to re-save it as its newest file format.

v0: we stored LevelPack filenames + level filenames to store high scores.
This was brittle in case a level gets moved into another levelpack later,
or either it or the levelpack is renamed.

v1: levels get UUID numbers and we store them by that. You can re-roll a
UUID in the level editor if you want to break high scores for your new
level version.
*/
func Migrate() error {
	sg, err := Load()
	if err != nil {
		return err
	}

	// Do we need to make any changes?
	var resave bool

	// Initialize new data structures.
	if sg.Levels == nil {
		sg.Levels = map[string]*Level{}
	}

	// Have any legacy LevelPack levels?
	if sg.LevelPacks != nil && len(sg.LevelPacks) > 0 {
		log.Info("Migrating savegame.json data to newer version")

		// See if we can track down a UUID for each level.
		for lpFilename, lpScore := range sg.LevelPacks {
			log.Info("SaveGame.Migrate: See levelpack %s", lpFilename)

			// Resolve the filename to this levelpack (on disk or bindata, etc)
			filename, err := filesystem.FindFile(lpFilename)
			if err != nil {
				log.Error("SaveGame.Migrate: Could not find levelpack %s: can't migrate high score", lpFilename)
				continue
			}

			// Find the levelpack.
			lp, err := levelpack.LoadFile(filename)
			if err != nil {
				log.Error("SaveGame.Migrate: Could not find levelpack %s: can't migrate high score", lpFilename)
				continue
			}

			// Search its levels for their UUIDs.
			for levelFilename, score := range lpScore.Levels {
				log.Info("SaveGame.Migrate: levelpack '%s' level '%s'", lp.Title, levelFilename)

				// Try and load this level.
				lvl, err := lp.GetLevel(levelFilename)
				if err != nil {
					log.Error("SaveGame.Migrate: could not load level '%s': %s", levelFilename, err)
					continue
				}

				// It has a UUID?
				if lvl.UUID == "" {
					log.Error("SaveGame.Migrate: level '%s' does not have a UUID, can not migrate savegame for it", levelFilename)
					continue
				}

				// Migrate!
				sg.Levels[lvl.UUID] = score
				delete(lpScore.Levels, levelFilename)
				resave = true
			}

			// Have we run out of levels?
			if len(lpScore.Levels) == 0 {
				log.Info("No more levels to migrate in levelpack '%s'!", lpFilename)
				delete(sg.LevelPacks, lpFilename)
				resave = true
			}
		}
	}

	if resave {
		log.Info("Resaving highscore.json in migration to newer file format")
		return sg.Save()
	}

	return nil
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
func (sg *SaveGame) MarkCompleted(levelpack, filename, uuid string) {
	lvl := sg.GetLevelScore(levelpack, filename, uuid)
	lvl.Completed = true
}

// NewHighScore may set a new highscore for a level.
//
// The level will be marked Completed and if the given score is better
// than the stored one it will update.
//
// Returns true if a new high score was logged.
func (sg *SaveGame) NewHighScore(levelpack, filename, uuid string, isPerfect bool, elapsed time.Duration, rules level.GameRule) bool {
	levelpack = filepath.Base(levelpack)
	filename = filepath.Base(filename)

	score := sg.GetLevelScore(levelpack, filename, uuid)
	score.Completed = true
	var newHigh bool

	if isPerfect {
		if score.PerfectTime == nil || *score.PerfectTime > elapsed {
			score.PerfectTime = &elapsed
			newHigh = true
		}
	} else {
		// GameRule: Survival (silver) - high score is based on longest time left alive rather
		// than fastest time completed.
		if rules.Survival {
			if score.BestTime == nil || *score.BestTime < elapsed {
				score.BestTime = &elapsed
				newHigh = true
			}
		} else {
			// Normally: fastest time is best time.
			if score.BestTime == nil || *score.BestTime > elapsed {
				score.BestTime = &elapsed
				newHigh = true
			}
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
func (sg *SaveGame) GetLevelScore(levelpack, filename, uuid string) *Level {
	// New format? Easy lookup by UUID.
	if uuid != "" && sg.Levels != nil {
		if row, ok := sg.Levels[uuid]; ok {
			return row
		}
	}

	// Old format: look it up by levelpack/filename.
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
func (sg *SaveGame) CountCompleted(levelpack *levelpack.LevelPack) int {
	var (
		count    int
		filename = filepath.Base(levelpack.Filename)
	)

	// Collect the level UUIDs for this levelpack.
	var uuids = map[string]interface{}{}
	for _, lvl := range levelpack.Levels {
		if lvl.UUID != "" {
			uuids[lvl.UUID] = nil
		}
	}

	// Count the new-style levels.
	if sg.Levels != nil {
		for uuid, lvl := range sg.Levels {
			if _, ok := uuids[uuid]; ok && lvl.Completed {
				count++
			}
		}
	}

	// Count the old-style levels.
	if sg.LevelPacks != nil {
		if lp, ok := sg.LevelPacks[filename]; ok {
			for _, lvl := range lp.Levels {
				if lvl.Completed {
					count++
				}
			}
		}
	}

	return count
}

// FormatDuration pretty prints a time.Duration in MM:SS format.
func FormatDuration(d time.Duration) string {
	var (
		millisecond = d.Milliseconds()
		second      = (millisecond / 1000) % 60
		minute      = (millisecond / (1000 * 60)) % 60
		hour        = (millisecond / (1000 * 60 * 60)) % 24
		ms          = fmt.Sprintf("%d", millisecond%1000)
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

package level

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/log"
)

/*
FileSystem embeds a map of files inside a parent drawing.

Old-style drawings this was a map of filenames to their data in the JSON.
New-style drawings this just holds the filenames and the data is read
from the zipfile on demand.
*/
type FileSystem struct {
	filemap map[string]File `json:"-"` // Legacy JSON format
	Zipfile *zip.Reader     `json:"-"` // New Zipfile format accessor
}

// File holds details about a file in the FileSystem.
type File struct {
	Data []byte `json:"data,omitempty"`
}

// NewFileSystem initializes the FileSystem struct.
func NewFileSystem() *FileSystem {
	return &FileSystem{
		filemap: map[string]File{},
	}
}

// Get a file from the FileSystem.
func (fs *FileSystem) Get(filename string) ([]byte, error) {
	if fs.filemap == nil {
		fs.filemap = map[string]File{}
	}

	// Legacy file map.
	if file, ok := fs.filemap[filename]; ok {
		if len(file.Data) > 0 {
			return file.Data, nil
		}
	}

	// Check in the zipfile.
	if fs.Zipfile != nil {
		file, err := fs.Zipfile.Open(filename)
		if err != nil {
			return []byte{}, fmt.Errorf("%s: not in zipfile: %s", filename, err)
		}

		bin, err := ioutil.ReadAll(file)
		if err != nil {
			return bin, fmt.Errorf("%s: couldn't read zipfile member: %s", filename, err)
		}

		return bin, nil
	}

	return []byte{}, fmt.Errorf("no such file")
}

// Set a file into the FileSystem. Note: it will go into the legacy map
// structure until the next save to disk, at which point queued files
// are committed to ZIP.
func (fs *FileSystem) Set(filename string, data []byte) {
	if fs.filemap == nil {
		fs.filemap = map[string]File{}
	}

	fs.filemap[filename] = File{
		Data: data,
	}
}

// Delete a file from the FileSystem. This will store zero bytes in the
// legacy file map structure to mark it for deletion. On the next save,
// filemap files with zero bytes skip the ZIP archive.
func (fs *FileSystem) Delete(filename string) {
	if fs.filemap == nil {
		fs.filemap = map[string]File{}
	}

	fs.filemap[filename] = File{
		Data: []byte{},
	}
}

// List files in the FileSystem, including the ZIP file.
//
// In the ZIP file, attachments are under the "assets/" prefix so this
// function won't mistakenly return chunks or level.json/doodad.json.
func (fs *FileSystem) List() []string {
	var (
		result = []string{}
		seen   = map[string]interface{}{}
	)

	// List the legacy or recently modified files first.
	if fs.filemap != nil {
		for filename := range fs.filemap {
			result = append(result, filename)
			seen[filename] = nil
		}
	}

	// List the zipfile members.
	if fs.Zipfile != nil {
		for _, file := range fs.Zipfile.File {
			if !strings.HasPrefix(file.Name, "assets/") {
				continue
			}

			if _, ok := seen[file.Name]; !ok {
				result = append(result, file.Name)
				seen[file.Name] = nil
			}
		}
	}

	sort.Strings(result)
	return result
}

// ListPrefix returns a list of files starting with the prefix.
func (fs *FileSystem) ListPrefix(prefix string) []string {
	var result = []string{}
	for _, name := range fs.List() {
		if strings.HasPrefix(name, prefix) {
			result = append(result, name)
		}
	}
	return result
}

// UnmarshalJSON reads in a FileSystem from its legacy JSON representation.
func (fs *FileSystem) UnmarshalJSON(text []byte) error {
	// Legacy format was a simple map[string]File.
	var legacy map[string]File
	err := json.Unmarshal(text, &legacy)
	if err != nil {
		return err
	}

	fs.filemap = legacy
	return nil
}

// MigrateZipfile is called on save to write attached files to the ZIP
// file format.
func (fs *FileSystem) MigrateZipfile(zf *zip.Writer) error {
	// Identify the files that we have marked for deletion.
	var (
		filesDeleted = map[string]interface{}{}
		filesZipped  = map[string]interface{}{}
	)
	if fs.filemap != nil {
		for filename, data := range fs.filemap {
			if len(data.Data) == 0 {
				log.Info("FileSystem.MigrateZipfile: %s has become empty, remove from zip", filename)
				filesDeleted[filename] = nil
			}
		}
	}

	// Copy all COLD STORED files from the old Zipfile into the new Zipfile
	// except for the ones marked for deletion OR the ones currently in the
	// warm cache which will be written next.
	if fs.Zipfile != nil {
		log.Info("FileSystem.MigrateZipfile: Copying files from old zip to new zip")
		for _, file := range fs.Zipfile.File {
			if !strings.HasPrefix(file.Name, "assets/") {
				continue
			}

			if _, ok := filesDeleted[file.Name]; ok {
				log.Debug("Skip copying attachment %s: was marked for deletion")
				continue
			}

			// Skip files currently in memory.
			if fs.filemap != nil {
				if _, ok := fs.filemap[file.Name]; ok {
					log.Debug("Skip copying attachment %s: one is loaded in memory")
					continue
				}
			}

			log.Debug("Copy zipfile attachment %s", file.Name)
			filesZipped[file.Name] = nil

			if err := zf.Copy(file); err != nil {
				return err
			}
		}
	}

	// Export currently warmed up files to ZIP, these will be ones that
	// were updated recently OR legacy files from an old level read.
	if fs.filemap != nil {
		log.Info("FileSystem.MigrateZipfile: has %d files in memory to write to ZIP", len(fs.filemap))
		for filename, file := range fs.filemap {
			if _, ok := filesZipped[filename]; ok {
				continue
			}

			writer, err := zf.Create(filename)
			if err != nil {
				return err
			}

			n, err := writer.Write(file.Data)
			if err != nil {
				return err
			}

			log.Debug("Exported file to zip: %s (%d bytes)", filename, n)
		}
	}

	return nil
}

////////////////
// Level class methods for its filesystem access
////////////////

// SetFile sets a file's data in the level.
func (l *Level) SetFile(filename string, data []byte) {
	l.Files.Set(filename, data)
}

// GetFile looks up an embedded file.
func (l *Level) GetFile(filename string) ([]byte, error) {
	if l.Files == nil {
		return []byte{}, errors.New("filesystem not initialized")
	}
	return l.Files.Get(filename)
}

// DeleteFile removes an embedded file.
func (l *Level) DeleteFile(filename string) bool {
	l.Files.Delete(filename)
	return true
}

// DeleteFiles removes all files beginning with the prefix.
func (l *Level) DeleteFiles(prefix string) int {
	var count int
	for _, filename := range l.Files.ListPrefix(prefix) {
		l.Files.Delete(filename)
		count++
	}
	return count
}

// ListFiles returns the list of all embedded file names, alphabetically.
func (l *Level) ListFiles() []string {
	if l == nil {
		log.Error("Level.ListFiles() was called on a nil Level??")
		return []string{}
	}

	if l.Files == nil {
		log.Error("Level(%s).ListFiles: FileSystem not initialized", l.Title)
		return []string{}
	}
	return l.Files.List()
}

// ListFilesAt returns the list of files having a common prefix.
func (l *Level) ListFilesAt(prefix string) []string {
	return l.Files.ListPrefix(prefix)
}

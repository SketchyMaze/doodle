package level

import (
	"errors"
	"sort"
	"strings"
)

// FileSystem embeds a map of files inside a parent drawing.
type FileSystem map[string]File

// File holds details about a file in the FileSystem.
type File struct {
	Data []byte `json:"data"`
}

// SetFile sets a file's data in the level.
func (l *Level) SetFile(filename string, data []byte) {
	if l.Files == nil {
		l.Files = map[string]File{}
	}

	l.Files[filename] = File{
		Data: data,
	}
}

// GetFile looks up an embedded file.
func (l *Level) GetFile(filename string) ([]byte, error) {
	if l.Files == nil {
		l.Files = map[string]File{}
	}

	if result, ok := l.Files[filename]; ok {
		return result.Data, nil
	}
	return []byte{}, errors.New("not found")
}

// DeleteFile removes an embedded file.
func (l *Level) DeleteFile(filename string) bool {
	if l.Files == nil {
		l.Files = map[string]File{}
	}

	if _, ok := l.Files[filename]; ok {
		delete(l.Files, filename)
		return true
	}
	return false
}

// ListFiles returns the list of all embedded file names, alphabetically.
func (l *Level) ListFiles() []string {
	var files []string

	if l.Files == nil {
		return files
	}

	for name := range l.Files {
		files = append(files, name)
	}

	sort.Strings(files)
	return files
}

// ListFilesAt returns the list of files having a common prefix.
func (l *Level) ListFilesAt(prefix string) []string {
	var (
		files = l.ListFiles()
		match = []string{}
	)
	for _, name := range files {
		if strings.HasPrefix(name, prefix) {
			match = append(match, name)
		}
	}
	return match
}

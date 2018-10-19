package level

// FileSystem embeds a map of files inside a parent drawing.
type FileSystem map[string]File

// File holds details about a file in the FileSystem.
type File struct {
	Data []byte `json:"data"`
}

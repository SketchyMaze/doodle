package native

/*
Native File Dialogs: Common Code
*/

// FileDialog common variables.
var (
	// This is set to True when a file dialog driver is available.
	// If false, a fallback uses the developer shell Prompt()
	// to ask for a file name.
	FileDialogsReady bool
)
// Package debugging contains useful methods for debugging the app, safely
// isolated from the rest of the app's packages.
package debugging

import (
	"fmt"
	"runtime"
	"strings"
)

// Configurable variables for the stack tracer functions.
var (
	// StackDepth is the depth that Callers() will crawl up the call stack. This
	// variable is configurable.
	StackDepth = 20

	// StopAt is the function name to stop the tracebacks at. Set to a blank
	// string to not stop and trace all the way up to `runtime.goexit` or
	// wherever.
	StopAt = "main.main"
)

// Minimum depth given to runtime.Caller() so that the call stacks will exclude
// the call to debugging.Caller() itself -- so this debug module won't debug its
// own function calls in the tracebacks.
const minDepth = 2

// Caller returns the filename and line number that called the calling
// function.
func Caller() string {
	if pc, file, no, ok := runtime.Caller(minDepth); ok {
		frames := runtime.CallersFrames([]uintptr{pc})
		frame, _ := frames.Next()
		if frame.Function != "" {
			return fmt.Sprintf("%s#%d: %s()",
				frame.File,
				frame.Line,
				frame.Function,
			)
		}
		return fmt.Sprintf("%s#%d",
			file,
			no,
		)
	}
	return "[no caller information]"
}

// Callers returns an array of all the callers of the current function.
func Callers() []string {
	var (
		callers []string
		pc      = make([]uintptr, StackDepth)
		count   = runtime.Callers(minDepth, pc)
	)
	pc = pc[:count] // only pass valid program counters to CallersFrames
	var frames = runtime.CallersFrames(pc)
	_ = frames

	// Loop to get frames of the call stack.
	for {
		frame, more := frames.Next()

		callers = append(callers, fmt.Sprintf("%s#%d: %s()",
			frame.File,
			frame.Line,
			frame.Function,
		))

		if StopAt != "" && frame.Function == StopAt {
			break
		}

		if !more {
			break
		}
	}

	return callers
}

// StringifyCallers pretty-prints the Callers as a single string with newlines.
func StringifyCallers() string {
	callers := Callers()
	var result []string
	for i, caller := range callers {
		if i == 0 {
			continue // StringifyCallers() would be the first row, skip it.
		}
		result = append(result, fmt.Sprintf("%d: %s", i, caller))
	}
	return strings.Join(result, "\n")
}

// PrintCallers prints the stringified callers directly to STDOUT.
func PrintCallers() {
	fmt.Println("Call stack (most recent/current function first):")
	for i, caller := range Callers() {
		if i == 0 {
			continue // PrintCallers() would be the first row, skip it.
		}
		fmt.Printf("%d: %s\n", i, caller)
	}
}

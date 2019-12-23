package log

import (
	"runtime"

	"git.kirsle.net/go/log"
)

// Logger is the public log.Logger object.
var Logger *log.Logger

func init() {
	Logger = log.GetLogger("doodle")
	Logger.Configure(&log.Config{
		Level:      log.InfoLevel,
		Theme:      log.DarkTheme,
		Colors:     log.ExtendedColor,
		TimeFormat: "2006-01-02 15:04:05.000000",
	})

	// TODO: Disable ANSI colors in logs on Windows.
	if runtime.GOOS == "windows" {
		Logger.Config.Colors = log.NoColor
	}
}

// Debug logger function.
func Debug(msg string, v ...interface{}) {
	Logger.Debug(msg, v...)
}

// Info logger function.
func Info(msg string, v ...interface{}) {
	Logger.Info(msg, v...)
}

// Warn logger function.
func Warn(msg string, v ...interface{}) {
	Logger.Warn(msg, v...)
}

// Error logger function.
func Error(msg string, v ...interface{}) {
	Logger.Error(msg, v...)
}

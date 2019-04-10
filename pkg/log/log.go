package log

import "github.com/kirsle/golog"

// Logger is the public golog.Logger object.
var Logger *golog.Logger

func init() {
	Logger = golog.GetLogger("doodle")
	Logger.Configure(&golog.Config{
		Level:      golog.DebugLevel,
		Theme:      golog.DarkTheme,
		Colors:     golog.ExtendedColor,
		TimeFormat: "2006-01-02 15:04:05.000000",
	})
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

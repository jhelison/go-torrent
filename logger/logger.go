package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// The default logger for the application
var Log zerolog.Logger

// init initializes the zerolog logger
// Since this application should be running locally,
// we can set as a production like environment
func init() {
	Log = zerolog.New(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		},
	).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()
}

// GetLogger just returns the project logger
func GetLogger() zerolog.Logger {
	return Log
}

// SetLogLevel update the current log level
func SetLogLevel(level string) {
	logLevel := zerolog.TraceLevel
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "err":
		logLevel = zerolog.ErrorLevel
	case "disabled":
		logLevel = zerolog.Disabled
	}

	Log = Log.Level(logLevel)
}

// Package logger provides a logger implementation using zerolog.
package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

// ZeroLogLogger is a logger implementation using zerolog.
type ZeroLogLogger struct {
	Log zerolog.Logger
}

// NewZeroLogLogger creates a new ZeroLogLogger instance with the specified log level.
// It configures the logger with console output, color-coded levels, and includes process ID in log entries.
func NewZeroLogLogger(level int) Logger {
	log := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04",
	}).
		Level(zerolog.Level(level)).
		With().
		Timestamp().
		Int("pid", os.Getpid()).
		Logger()

	return &ZeroLogLogger{Log: log}
}

// GetWriter returns an io.Writer that writes to the underlying zerolog logger.
// Useful for integrating with libraries that expect an io.Writer interface.
func (l *ZeroLogLogger) GetWriter() io.Writer {
	return l.Log
}

// Printf logs a formatted message at info level.
// It uses the same format string syntax as fmt.Printf.
func (l *ZeroLogLogger) Printf(format string, args ...any) {
	l.Log.Printf(format, args...)
}

// Error logs a message at error level with the provided arguments.
func (l *ZeroLogLogger) Error(args ...any) {
	l.Log.Error().Caller(1).Msg(fmt.Sprint(args...))
}

// Errorf logs a formatted message at error level.
func (l *ZeroLogLogger) Errorf(format string, args ...any) {
	l.Log.Error().Caller(1).Msgf(format, args...)
}

// Fatal logs a message at fatal level and terminates the program.
func (l *ZeroLogLogger) Fatal(args ...any) {
	l.Log.Fatal().Caller(1).Msg(fmt.Sprint(args...))
}

// Fatalf logs a formatted message at fatal level and terminates the program.
func (l *ZeroLogLogger) Fatalf(format string, args ...any) {
	l.Log.Fatal().Caller(1).Msgf(format, args...)
}

// Info logs a message at info level with the provided arguments.
func (l *ZeroLogLogger) Info(args ...any) {
	l.Log.Info().Caller(1).Msg(fmt.Sprint(args...))
}

// Infof logs a formatted message at info level.
func (l *ZeroLogLogger) Infof(format string, args ...any) {
	l.Log.Info().Caller(1).Msgf(format, args...)
}

// Warn logs a message at warning level with the provided arguments.
func (l *ZeroLogLogger) Warn(args ...any) {
	l.Log.Warn().Caller(1).Msg(fmt.Sprint(args...))
}

// Warnf logs a formatted message at warning level.
func (l *ZeroLogLogger) Warnf(format string, args ...any) {
	l.Log.Warn().Caller(1).Msgf(format, args...)
}

// Debug logs a message at debug level with the provided arguments.
func (l *ZeroLogLogger) Debug(args ...any) {
	l.Log.Debug().Caller(1).Msg(fmt.Sprint(args...))
}

// Debugf logs a formatted message at debug level.
func (l *ZeroLogLogger) Debugf(format string, args ...any) {
	l.Log.Debug().Caller(1).Msgf(format, args...)
}

// WithField returns a new logger instance with an additional field.
// If the value is an error, it will be handled specially by zerolog.
func (l *ZeroLogLogger) WithField(key string, value any) Logger {
	var log zerolog.Logger
	if err, ok := value.(error); ok {
		log = l.Log.With().AnErr(key, err).Logger()
	} else {
		log = l.Log.With().Any(key, value).Logger()
	}

	return &ZeroLogLogger{
		Log: log,
	}
}

// WithFields returns a new logger instance with multiple additional fields.
// Error values and slices of errors are handled specially by zerolog for better formatting.
func (l *ZeroLogLogger) WithFields(fields map[string]any) Logger {
	logCtx := l.Log.With()

	for k, v := range fields {
		switch val := v.(type) {
		case []error:
			logCtx = logCtx.Errs(k, val)
		case error:
			logCtx = logCtx.AnErr(k, val)
		default:
			logCtx = logCtx.Any(k, v)
		}
	}

	return &ZeroLogLogger{
		Log: logCtx.Logger(),
	}
}

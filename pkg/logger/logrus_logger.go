package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// LogrusLogger is a logger implementation using logrus.
type LogrusLogger struct {
	Log *logrus.Logger
}

// NewLogrusLogger creates a new LogrusLogger instance with the specified log level.
// It configures the logger with console output, color-coded levels, and includes process ID in log entries.
func NewLogrusLogger(level int) Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat:  "15:04",
		FullTimestamp:    true,
		ForceColors:      true,
		DisableTimestamp: false,
	})
	log.SetLevel(logrus.Level(level))
	log.WithFields(logrus.Fields{
		"pid": os.Getpid(),
	})

	return &LogrusLogger{Log: log}
}

// GetWriter returns an io.Writer that writes to the underlying logrus logger.
// Useful for integrating with libraries that expect an io.Writer interface.
func (l *LogrusLogger) GetWriter() io.Writer {
	return l.Log.Writer()
}

// Printf logs a formatted message at info level.
// It uses the same format string syntax as fmt.Printf.
func (l *LogrusLogger) Printf(format string, args ...any) {
	l.Log.Infof(format, args...)
}

// Debug logs a message at debug level with the provided arguments.
func (l *LogrusLogger) Debug(args ...any) {
	l.Log.Debug(args...)
}

// Debugf logs a formatted message at debug level.
func (l *LogrusLogger) Debugf(format string, args ...any) {
	l.Log.Debugf(format, args...)
}

// Info logs a message at info level with the provided arguments.
func (l *LogrusLogger) Info(args ...any) {
	l.Log.Info(args...)
}

// Infof logs a formatted message at info level.
func (l *LogrusLogger) Infof(format string, args ...any) {
	l.Log.Infof(format, args...)
}

// Warn logs a message at warning level with the provided arguments.
func (l *LogrusLogger) Warn(args ...any) {
	l.Log.Warn(args...)
}

// Warnf logs a formatted message at warning level.
func (l *LogrusLogger) Warnf(format string, args ...any) {
	l.Log.Warnf(format, args...)
}

// Error logs a message at error level with the provided arguments.
func (l *LogrusLogger) Error(args ...any) {
	l.Log.Error(args...)
}

// Errorf logs a formatted message at error level.
func (l *LogrusLogger) Errorf(format string, args ...any) {
	l.Log.Errorf(format, args...)
}

// Fatal logs a message at fatal level and terminates the program.
func (l *LogrusLogger) Fatal(args ...any) {
	l.Log.Fatal(args...)
}

// Fatalf logs a formatted message at fatal level and terminates the program.
func (l *LogrusLogger) Fatalf(format string, args ...any) {
	l.Log.Fatalf(format, args...)
}

// WithField returns a new logger instance with an additional field.
func (l *LogrusLogger) WithField(key string, value any) Logger {
	return &LogrusEntry{
		entry: l.Log.WithField(key, value),
	}
}

// WithFields returns a new logger instance with multiple additional fields.
func (l *LogrusLogger) WithFields(fields map[string]any) Logger {
	return &LogrusEntry{
		entry: l.Log.WithFields(fields),
	}
}

// LogrusEntry wraps a logrus.Entry to provide Logger interface methods.
type LogrusEntry struct {
	entry *logrus.Entry
}

// GetWriter returns an io.Writer that writes to the underlying logrus entry.
func (l *LogrusEntry) GetWriter() io.Writer {
	return l.entry.Writer()
}

// Printf logs a formatted message at info level using the logrus entry.
func (l *LogrusEntry) Printf(format string, args ...any) {
	l.entry.Printf(format, args...)
}

// Error logs a message at error level using the logrus entry.
func (l *LogrusEntry) Error(args ...any) {
	l.entry.Error(args...)
}

// Errorf logs a formatted message at error level using the logrus entry.
func (l *LogrusEntry) Errorf(format string, args ...any) {
	l.entry.Errorf(format, args...)
}

// Fatal logs a message at fatal level and terminates the program using the logrus entry.
func (l *LogrusEntry) Fatal(args ...any) {
	l.entry.Fatal(args...)
}

// Fatalf logs a formatted message at fatal level and terminates the program using the logrus entry.
func (l *LogrusEntry) Fatalf(format string, args ...any) {
	l.entry.Fatalf(format, args...)
}

// Info logs a message at info level using the logrus entry.
func (l *LogrusEntry) Info(args ...any) {
	l.entry.Info(args...)
}

// Infof logs a formatted message at info level using the logrus entry.
func (l *LogrusEntry) Infof(format string, args ...any) {
	l.entry.Infof(format, args...)
}

// Warn logs a message at warning level using the logrus entry.
func (l *LogrusEntry) Warn(args ...any) {
	l.entry.Warn(args...)
}

// Warnf logs a formatted message at warning level using the logrus entry.
func (l *LogrusEntry) Warnf(format string, args ...any) {
	l.entry.Warnf(format, args...)
}

// Debug logs a message at debug level using the logrus entry.
func (l *LogrusEntry) Debug(args ...any) {
	l.entry.Debug(args...)
}

// Debugf logs a formatted message at debug level using the logrus entry.
func (l *LogrusEntry) Debugf(format string, args ...any) {
	l.entry.Debugf(format, args...)
}

// WithField returns a new logger entry with an additional field.
func (l *LogrusEntry) WithField(key string, value any) (entry Logger) {
	entry = &LogrusEntry{l.entry.WithField(key, value)}

	return entry
}

// WithFields returns a new logger entry with multiple additional fields.
func (l *LogrusEntry) WithFields(args map[string]any) (entry Logger) {
	entry = &LogrusEntry{l.entry.WithFields(args)}

	return entry
}

// Package logger provides a Zap-based logger implementation.
package logger

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a logger implementation using Uber's Zap library.
type ZapLogger struct {
	Log *zap.Logger
}

// NewZapLogger creates a new ZapLogger instance with the specified log level.
// It configures the logger with console output, color-coded levels, and includes
// process ID in log entries.
func NewZapLogger(level int) Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.Level(level),
	)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(
			zap.Int("pid", os.Getpid()),
		),
	)

	return &ZapLogger{Log: logger}
}

// GetWriter returns an io.Writer that writes to the underlying zap logger.
// This is useful for integrating with libraries that expect an io.Writer interface.
func (l *ZapLogger) GetWriter() io.Writer {
	return zap.NewStdLog(l.Log).Writer()
}

// Printf logs a formatted message at info level.
// It uses the same format string syntax as fmt.Printf.
func (l *ZapLogger) Printf(format string, args ...any) {
	l.Log.Sugar().Infof(format, args...)
}

// Error logs a message at error level with the provided arguments.
func (l *ZapLogger) Error(args ...any) {
	l.Log.Sugar().Error(args...)
}

// Errorf logs a formatted message at error level.
func (l *ZapLogger) Errorf(format string, args ...any) {
	l.Log.Sugar().Errorf(format, args...)
}

// Fatal logs a message at fatal level and terminates the program.
func (l *ZapLogger) Fatal(args ...any) {
	l.Log.Sugar().Fatal(args...)
}

// Fatalf logs a formatted message at fatal level and terminates the program.
func (l *ZapLogger) Fatalf(format string, args ...any) {
	l.Log.Sugar().Fatalf(format, args...)
}

// Info logs a message at info level with the provided arguments.
func (l *ZapLogger) Info(args ...any) {
	l.Log.Sugar().Info(args...)
}

// Infof logs a formatted message at info level.
func (l *ZapLogger) Infof(format string, args ...any) {
	l.Log.Sugar().Infof(format, args...)
}

// Warn logs a message at warning level with the provided arguments.
func (l *ZapLogger) Warn(args ...any) {
	l.Log.Sugar().Warn(args...)
}

// Warnf logs a formatted message at warning level.
func (l *ZapLogger) Warnf(format string, args ...any) {
	l.Log.Sugar().Warnf(format, args...)
}

// Debug logs a message at debug level with the provided arguments.
func (l *ZapLogger) Debug(args ...any) {
	l.Log.Sugar().Debug(args...)
}

// Debugf logs a formatted message at debug level.
func (l *ZapLogger) Debugf(format string, args ...any) {
	l.Log.Sugar().Debugf(format, args...)
}

// WithField returns a new logger instance with an additional field.
// If the value is an error, it will be handled specially by zap.
func (l *ZapLogger) WithField(key string, value any) Logger {
	if err, ok := value.(error); ok {
		return &ZapLogger{Log: l.Log.With(zap.Error(err))}
	}

	return &ZapLogger{Log: l.Log.With(zap.Any(key, value))}
}

// WithFields returns a new logger instance with multiple additional fields.
// Error values and slices of errors are handled specially by zap for better formatting.
func (l *ZapLogger) WithFields(fields map[string]any) Logger {
	zapFields := make([]zap.Field, 0, len(fields))

	for k, v := range fields {
		switch val := v.(type) {
		case []error:
			zapFields = append(zapFields, zap.Errors(k, val))
		case error:
			zapFields = append(zapFields, zap.Error(val))
		default:
			zapFields = append(zapFields, zap.Any(k, v))
		}
	}

	return &ZapLogger{Log: l.Log.With(zapFields...)}
}

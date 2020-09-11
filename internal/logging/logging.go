package logging

import (
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

// Logger ...
type Logger struct {
	valid bool
}

// IsValid ...
func (l Logger) IsValid() bool {
	return l.valid
}

// Debug ...
func (Logger) Debug() *zerolog.Event {
	return zlog.Debug()
}

// Error ...
func (Logger) Error() *zerolog.Event {
	return zlog.Error()
}

// Info ...
func (Logger) Info() *zerolog.Event {
	return zlog.Info()
}

// Warn ...
func (Logger) Warn() *zerolog.Event {
	return zlog.Warn()
}

// Config represents the configuration necessary for this pkg.
type Config struct {
	Level string
}

// New initializes a logger.
func New(c Config) Logger {
	zerolog.SetGlobalLevel(level(c.Level))
	return Logger{
		valid: true,
	}
}

func level(l string) zerolog.Level {
	switch l {
	case "error":
		return zerolog.ErrorLevel
	case "warn":
		return zerolog.WarnLevel
	case "info":
		return zerolog.InfoLevel
	case "debug":
		fallthrough
	default:
		return zerolog.ErrorLevel
	}
}

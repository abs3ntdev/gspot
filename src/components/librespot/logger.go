package librespot

import (
	"fmt"
	"log/slog"

	golibrespot "github.com/devgianlu/go-librespot"
)

// SlogAdapter adapts *slog.Logger to the go-librespot Logger interface.
type SlogAdapter struct {
	log *slog.Logger
}

func NewSlogAdapter(log *slog.Logger) golibrespot.Logger {
	return &SlogAdapter{log: log.With("component", "librespot")}
}

func (a *SlogAdapter) Tracef(format string, args ...interface{}) {
	a.log.Debug(fmt.Sprintf(format, args...))
}

func (a *SlogAdapter) Debugf(format string, args ...interface{}) {
	a.log.Debug(fmt.Sprintf(format, args...))
}

func (a *SlogAdapter) Infof(format string, args ...interface{}) {
	a.log.Info(fmt.Sprintf(format, args...))
}

func (a *SlogAdapter) Warnf(format string, args ...interface{}) {
	a.log.Warn(fmt.Sprintf(format, args...))
}

func (a *SlogAdapter) Errorf(format string, args ...interface{}) {
	a.log.Error(fmt.Sprintf(format, args...))
}

func (a *SlogAdapter) Trace(args ...interface{}) {
	a.log.Debug(fmt.Sprint(args...))
}

func (a *SlogAdapter) Debug(args ...interface{}) {
	a.log.Debug(fmt.Sprint(args...))
}

func (a *SlogAdapter) Info(args ...interface{}) {
	a.log.Info(fmt.Sprint(args...))
}

func (a *SlogAdapter) Warn(args ...interface{}) {
	a.log.Warn(fmt.Sprint(args...))
}

func (a *SlogAdapter) Error(args ...interface{}) {
	a.log.Error(fmt.Sprint(args...))
}

func (a *SlogAdapter) WithField(key string, value interface{}) golibrespot.Logger {
	return &SlogAdapter{log: a.log.With(key, value)}
}

func (a *SlogAdapter) WithError(err error) golibrespot.Logger {
	return &SlogAdapter{log: a.log.With("error", err)}
}

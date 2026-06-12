package main

import (
	"fmt"

	"github.com/rs/zerolog"
)

// BadgerZerologAdapter routes Badger logs through Zerolog.
type BadgerZerologAdapter struct {
	Logger zerolog.Logger
}

// Errorf handles Badger Error logs
func (b *BadgerZerologAdapter) Errorf(format string, v ...interface{}) {
	b.Logger.Error().Msg(fmt.Sprintf(format, v...))
}

// Warningf handles Badger Warning logs
func (b *BadgerZerologAdapter) Warningf(format string, v ...interface{}) {
	b.Logger.Warn().Msg(fmt.Sprintf(format, v...))
}

// Infof handles Badger Info logs
func (b *BadgerZerologAdapter) Infof(format string, v ...interface{}) {
	b.Logger.Info().Msg(fmt.Sprintf(format, v...))
}

// Debugf handles Badger Debug logs (requires Debug level enabled)
func (b *BadgerZerologAdapter) Debugf(format string, v ...interface{}) {
	b.Logger.Debug().Msg(fmt.Sprintf(format, v...))
}

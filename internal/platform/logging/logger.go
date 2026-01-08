package logging

import (
	"os"

	"github.com/rs/zerolog"
)

// NewLogger creates a new structured logger
func NewLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "farohq-core-app").
		Logger()
}


package httpserver

import (
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog"
)

// CommonMiddleware sets up common HTTP middleware
func CommonMiddleware(logger zerolog.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		httplog.RequestLogger(logger),
		chimw.Recoverer,
		chimw.Timeout(60 * time.Second),
		chimw.RequestID,
		chimw.RealIP,
	}
}


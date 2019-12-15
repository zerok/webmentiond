package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

// HTTPError is a simple error wrapper that adds a HTTP status code.
type HTTPError struct {
	StatusCode int
	Err        error
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP error %d: %s", e.StatusCode, e.Err.Error())
	}
	return fmt.Sprintf("HTTP error %d", e.StatusCode)
}

func (srv *Server) sendError(ctx context.Context, w http.ResponseWriter, err error) {
	logger := zerolog.Ctx(ctx)
	statusCode := http.StatusInternalServerError
	var httpErr *HTTPError
	if ok := errors.As(err, &httpErr); ok {
		statusCode = httpErr.StatusCode
	}
	if statusCode >= 500 {
		logger.Error().Err(err).Msg("Processing failed")
	} else {
		logger.Debug().Err(err).Msg("Processing failed")
	}
	http.Error(w, "Error", statusCode)
}

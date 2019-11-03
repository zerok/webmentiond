package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zerok/webmentiond/pkg/server"
)

func TestReceiver(t *testing.T) {
	srv := server.New()
	req := httptest.NewRequest(http.MethodPost, "/receive", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

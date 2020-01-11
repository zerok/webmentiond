package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/mailer"
)

func TestAuthentication(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background())
	m := mailer.NewDummy()
	srv := New(func(c *Configuration) {
		c.Mailer = m
		c.Auth.AdminEmails = []string{"valid"}
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/request-login", bytes.NewBufferString(""))
	r.Header.Set("Content-type", "application/x-www-form-urlencoded")
	srv.ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusBadRequest)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/request-login", bytes.NewBufferString("email=invalid"))
	r.Header.Set("Content-type", "application/x-www-form-urlencoded")
	srv.ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusOK)
	require.Len(t, m.Messages, 0)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/request-login", bytes.NewBufferString("email=valid"))
	r.Header.Set("Content-type", "application/x-www-form-urlencoded")
	srv.ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusOK)
	require.Len(t, m.Messages, 1)

	token := srv.validToken["valid"]
	require.NotEmpty(t, token)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/authenticate?token="+token, nil)
	srv.ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusOK)
	require.Equal(t, w.Header().Get("Content-Type"), "application/jwt")
	jot := w.Body.Bytes()
	require.NotEmpty(t, jot)

	// If we run against the auth-middleware without a token, we
	// should get a 401 status:
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	srv.requireAuthMiddleware(nil).ServeHTTP(w, r)
	require.Equal(t, w.Code, http.StatusUnauthorized)

	// If we add a broken jwt, the middleware should fail with a
	// 401 status as if no token were present:
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Authorization", "Bearer BROKEN")
	srv.requireAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})).ServeHTTP(w, r.WithContext(ctx))
	require.Equal(t, w.Code, http.StatusUnauthorized)

	// If we add the correct token, the middleware should let us
	// pass:
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(jot)))
	srv.requireAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})).ServeHTTP(w, r.WithContext(ctx))
	require.Equal(t, w.Code, http.StatusOK)
}

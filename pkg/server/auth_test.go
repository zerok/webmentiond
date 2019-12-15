package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/mailer"
)

func TestAuthentication(t *testing.T) {
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
}

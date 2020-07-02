package server_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/server"
)

type conformanceEnvironment struct {
	Logger zerolog.Logger
	Ctx    context.Context
	DB     *sql.DB
	Srv    *server.Server
}

func (e *conformanceEnvironment) Destroy() {
	e.DB.Close()
}

func createConformanceEnvironment(t *testing.T) *conformanceEnvironment {
	t.Helper()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background())
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	require.NotNil(t, db)
	srv := server.New(func(c *server.Configuration) {
		c.Context = ctx
		c.Database = db
		c.MigrationsFolder = "./migrations"
		c.Receiver.TargetPolicy = server.RequestPolicyAllowHost("allowed.com")
		c.VerificationMaxRedirects = 1
	})
	require.NoError(t, srv.MigrateDatabase(ctx))
	e := conformanceEnvironment{
		DB:     db,
		Srv:    srv,
		Logger: logger,
		Ctx:    ctx,
	}
	return &e
}

func (e *conformanceEnvironment) SendMention(source string, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	data := url.Values{}
	data.Set("source", source)
	data.Set("target", target)
	req := httptest.NewRequest(http.MethodPost, "/receive", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e.Srv.ServeHTTP(w, req.WithContext(e.Ctx))
	return w
}

func (e *conformanceEnvironment) VerifyNextMention(t *testing.T) {
	verified, err := e.Srv.VerifyNextMention(e.Ctx)
	require.NoError(t, err)
	require.True(t, verified)
}

func TestConformance(t *testing.T) {
	t.Run("request-verification/url-validation", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		w := e.SendMention("<>", "https://allowed.com")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("request-verification/allowed-target", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		w := e.SendMention("https://source.com", "https://allowed.com")
		require.Equal(t, http.StatusAccepted, w.Code)

		w = e.SendMention("https://source.com", "https://not-allowed.com")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("request-verification/allowed-target-ignore-fragment", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		w := e.SendMention("https://source.com", "https://allowed.com/#fragment")
		require.Equal(t, http.StatusAccepted, w.Code)

		w = e.SendMention("https://source.com", "https://not-allowed.com/#fragment")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("webmention-validation/do-async", func(t *testing.T) {
		// Yes :-)
	})

	t.Run("webmention-validation/follow-at-least-one-redirect", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			e.Logger.Debug().Msg("Actual called")
			fmt.Fprintf(w, `<html><body><a href="https://allowed.com/">Target</a></body></html>`)
		})
		mux.Get("/redirect", func(w http.ResponseWriter, r *http.Request) {
			e.Logger.Debug().Msg("Redirect called")
			http.Redirect(w, r, "/actual", http.StatusTemporaryRedirect)
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/redirect", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "verified")
	})

	t.Run("webmention-validation/stop-after-n-redirects", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<html><body><a href="https://allowed.com/">Target</a></body></html>`)
		})
		mux.Get("/redirect-3", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/redirect-2", http.StatusTemporaryRedirect)
		})
		mux.Get("/redirect-2", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/redirect-1", http.StatusTemporaryRedirect)
		})
		mux.Get("/redirect-1", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/actual", http.StatusTemporaryRedirect)
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/redirect-2", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusInvalid)
	})

	t.Run("html-verification/reject-text", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<html><body>https://allowed.com/</body></html>`)
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusInvalid)
	})

	t.Run("html-verification/link-in-video", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<html><body><video><source src="https://allowed.com/"></video></body></html>`)
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusVerified)
	})

	t.Run("html-verification/link-in-audio", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<html><body><audio><source src="https://allowed.com/"></audio></body></html>`)
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusVerified)
	})

	t.Run("html-verification/reject-link-in-comment", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<html><body><!-- <a href="https://allowed.com/">Target</a> --></body></html>`)
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusInvalid)
	})

	t.Run("html-verification/link-in-img", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<html><body><img src="https://allowed.com/"></body></html>`)
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusVerified)
	})

	t.Run("update-tests/remove-if-link-missing", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		isPresent := true
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			if isPresent {
				fmt.Fprintf(w, `<html><body><img src="https://allowed.com/"></body></html>`)
			} else {
				fmt.Fprintf(w, `<html><body></body></html>`)
			}
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusVerified)

		isPresent = false
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusInvalid)
	})

	t.Run("update-tests/recognize-410", func(t *testing.T) {
		e := createConformanceEnvironment(t)
		defer e.Destroy()
		mux := chi.NewRouter()
		isPresent := true
		mux.Get("/actual", func(w http.ResponseWriter, r *http.Request) {
			if isPresent {
				fmt.Fprintf(w, `<html><body><img src="https://allowed.com/"></body></html>`)
			} else {
				http.Error(w, "Gone", http.StatusGone)
			}
		})
		src := httptest.NewServer(mux)
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusVerified)

		isPresent = false
		e.SendMention(src.URL+"/actual", "https://allowed.com/")
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", "new")
		e.VerifyNextMention(t)
		requireMentionWithStatus(t, e.Ctx, e.DB, "https://allowed.com/", server.MentionStatusInvalid)
	})
}

package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/server"
)

func TestVerify(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background())
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
	srv := server.New(func(c *server.Configuration) {
		c.Database = db
		c.MigrationsFolder = "./migrations"
	})
	require.NoError(t, srv.MigrateDatabase(ctx))

	// Now let's also launch a simple HTTP server that should act as source for
	// a mention.
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><head><title>title</title></head><body><a href=\"http://test.com\">target</a></body></html>")
	}))
	defer h.Close()
	createMention(t, db, "a", h.URL, "http://test.com")
	requireMentionStatus(t, db, "a", "new")
	srv.VerifyNextMention(ctx)
	requireMentionStatus(t, db, "a", "verified")
	requireMentionTitle(t, db, "a", "title")

	createMention(t, db, "b", h.URL, "http://unknown.com")
	requireMentionStatus(t, db, "b", "new")
	srv.VerifyNextMention(ctx)
	requireMentionStatus(t, db, "b", "invalid")
	requireMentionTitle(t, db, "b", "")
}

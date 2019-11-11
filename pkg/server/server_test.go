package server_test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/server"
)

func requireMentionWithStatus(t *testing.T, ctx context.Context, db *sql.DB, target string, status string) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	res, err := tx.QueryContext(ctx, "SELECT id, status, source from webmentions WHERE target = ?", target)
	require.NoError(t, err)
	defer res.Close()
	for res.Next() {
		var id string
		var stat string
		var source string
		err := res.Scan(&id, &stat, &source)
		require.NoError(t, err)
		if status == stat {
			return
		}
	}
	t.Fatalf("no mention with status %s found", status)
}

func TestReceiver(t *testing.T) {
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

	// If no body is sent, then a 400 error should be returned:
	req := httptest.NewRequest(http.MethodPost, "/receive", nil)
	w := httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	srv.ServeHTTP(w, req.WithContext(ctx))
	require.Equal(t, 400, w.Code)

	// Next, let's send a mention that should actually go through:
	w = httptest.NewRecorder()
	data := url.Values{}
	data.Set("source", "https://source.zerokspot.com")
	data.Set("target", "https://target.zerokspot.com")
	req = httptest.NewRequest(http.MethodPost, "/receive", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	srv.ServeHTTP(w, req.WithContext(ctx))
	require.Equal(t, 201, w.Code)
	requireMentionWithStatus(t, ctx, db, "https://target.zerokspot.com", "new")
}

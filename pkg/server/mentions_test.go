package server_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/zerok/webmentiond/pkg/server"
)

func setupDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	return db
}

func TestPagingMentions(t *testing.T) {
	ctx := context.Background()
	db := setupDatabase(t)
	defer db.Close()
	srv := server.New(func(c *server.Configuration) {
		c.Database = db
		c.MigrationsFolder = "migrations"
	})
	require.NoError(t, srv.MigrateDatabase(ctx))
	_, err := db.Exec("INSERT INTO webmentions (id, source, target, created_at) VALUES ('a', 'a', 'b', ?)", time.Now())
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO webmentions (id, source, target, created_at) VALUES ('b', 'b', 'c', ?)", time.Now())
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO webmentions (id, source, target, created_at) VALUES ('c', 'c', 'd', ?)", time.Now())
	require.NoError(t, err)

	// If we just send an authorized request, we should get a 401 back:
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/manage/mentions", nil)
	srv.ServeHTTP(w, r)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// An authorized request should get something back:
	var res server.PagedMentionList
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/manage/mentions?limit=1", nil)
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
	require.Equal(t, 3, res.Total)
	require.Len(t, res.Items, 1)
	require.Equal(t, "a", res.Items[0].ID)

	// Now, let's page through all the result-sets:
	require.NotEmpty(t, res.Next)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, res.Next, nil)
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
	require.Len(t, res.Items, 1)
	require.Equal(t, "b", res.Items[0].ID)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, res.Next, nil)
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
	require.Len(t, res.Items, 1)
	require.Equal(t, "c", res.Items[0].ID)
	require.Empty(t, res.Next)
}
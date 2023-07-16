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

func setupServer(t *testing.T, db *sql.DB) *server.Server {
	t.Helper()
	srv := server.New(func(c *server.Configuration) {
		c.Database = db
		c.MigrationsFolder = "migrations"
		c.ExposeMetrics = true
	})
	require.NoError(t, srv.MigrateDatabase(context.Background()))
	return srv
}

func createMention(t *testing.T, db *sql.DB, id, source, target string) {
	t.Helper()
	_, err := db.Exec("INSERT INTO webmentions (id, source, target, created_at) VALUES (?, ?, ?, ?)", id, source, target, time.Now())
	require.NoError(t, err)
}

func setMentionStatus(t *testing.T, db *sql.DB, id string, status string) {
	t.Helper()
	_, err := db.Exec("UPDATE webmentions SET status = ? WHERE id = ?", status, id)
	require.NoError(t, err)
}

func setMentionType(t *testing.T, db *sql.DB, id string, typ string) {
	t.Helper()
	_, err := db.Exec("UPDATE webmentions SET type = ? WHERE id = ?", typ, id)
	require.NoError(t, err)
}

func setMentionContent(t *testing.T, db *sql.DB, id string, content string) {
	t.Helper()
	_, err := db.Exec("UPDATE webmentions SET content = ? WHERE id = ?", content, id)
	require.NoError(t, err)
}

func setMentionTitle(t *testing.T, db *sql.DB, id string, title string) {
	t.Helper()
	_, err := db.Exec("UPDATE webmentions SET title = ? WHERE id = ?", title, id)
	require.NoError(t, err)
}

func requireMentionStatus(t *testing.T, db *sql.DB, id string, status string) {
	t.Helper()
	ctx := context.Background()
	var actual string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT status FROM webmentions WHERE id = ?", id).Scan(&actual))
	if actual != status {
		t.Errorf("Mention %s was expected to have status `%s` but has `%s` instead.", id, status, actual)
		t.Fail()
	}
}

func requireMentionNotExists(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	ctx := context.Background()
	var count int
	err := db.QueryRowContext(ctx, "SELECT count(id) FROM webmentions WHERE id = ?", id).Scan(&count)
	if sql.ErrNoRows == err {
		return
	}
	if count == 0 {
		return
	}
	t.Errorf("Mention %s should not exist but exists.", id)
	t.Fail()
}

func requireMentionTitle(t *testing.T, db *sql.DB, id string, title string) {
	t.Helper()
	ctx := context.Background()
	var actual string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT title FROM webmentions WHERE id = ?", id).Scan(&actual))
	if actual != title {
		t.Errorf("Mention %s was expected to have title `%s` but has `%s` instead.", id, title, actual)
		t.Fail()
	}
}

func TestPagingMentions(t *testing.T) {
	db := setupDatabase(t)
	defer db.Close()
	srv := setupServer(t, db)
	createMention(t, db, "a", "a", "b")
	setMentionTitle(t, db, "a", "titlea")
	setMentionContent(t, db, "a", "contenta")
	createMention(t, db, "b", "b", "c")
	createMention(t, db, "c", "c", "d")
	setMentionTitle(t, db, "c", "titlec")
	setMentionContent(t, db, "c", "contentc")

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
	require.Equal(t, "c", res.Items[0].ID)
	require.Equal(t, "titlec", res.Items[0].Title)
	require.Equal(t, "contentc", res.Items[0].Content)

	// Now, let's page through all the result-sets:
	require.NotEmpty(t, res.Next)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, res.Next, nil)
	res = server.PagedMentionList{}
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
	require.Len(t, res.Items, 1)
	require.Equal(t, "b", res.Items[0].ID)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, res.Next, nil)
	res = server.PagedMentionList{}
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
	require.Len(t, res.Items, 1)
	require.Equal(t, "a", res.Items[0].ID)
	require.Empty(t, res.Next)

	// For now, we don't have any verified mentions yet:
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/manage/mentions?status=verified", nil)
	res = server.PagedMentionList{}
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
	require.Equal(t, 0, res.Total)
	requireMetricValue(t, context.Background(), srv, "webmentiond_mentions{status=\verified\"}", 0)
}

func TestApprovingMention(t *testing.T) {
	db := setupDatabase(t)
	defer db.Close()
	srv := setupServer(t, db)
	createMention(t, db, "a", "a", "b")
	requireMentionStatus(t, db, "a", "new")

	r := httptest.NewRequest(http.MethodPost, "/manage/mentions/a/approve", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	requireMentionStatus(t, db, "a", "approved")

	// Calling approve on a non-existing object should return a 404 code
	r = httptest.NewRequest(http.MethodPost, "/manage/mentions/b/approve", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusNotFound, w.Code)
	requireMetricValue(t, context.Background(), srv, "webmentiond_mentions{status=\approved\"}", 1)
}

func TestRejectingMention(t *testing.T) {
	db := setupDatabase(t)
	defer db.Close()
	srv := setupServer(t, db)
	createMention(t, db, "a", "a", "b")
	requireMentionStatus(t, db, "a", "new")

	r := httptest.NewRequest(http.MethodPost, "/manage/mentions/a/reject", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	requireMentionStatus(t, db, "a", "rejected")

	// Calling reject on a non-existing object should return a 404 code
	r = httptest.NewRequest(http.MethodPost, "/manage/mentions/b/reject", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusNotFound, w.Code)
	requireMetricValue(t, context.Background(), srv, "webmentiond_mentions{status=\rejected\"}", 1)

	// If no ID is provided, return a 400 as the request is invalid:
	r = httptest.NewRequest(http.MethodPost, "/manage/mentions//reject", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletingMention(t *testing.T) {
	db := setupDatabase(t)
	defer db.Close()
	srv := setupServer(t, db)
	createMention(t, db, "a", "a", "b")
	requireMentionStatus(t, db, "a", "new")

	r := httptest.NewRequest(http.MethodDelete, "/manage/mentions/a", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r.WithContext(server.AuthorizeContext(r.Context())))
	require.Equal(t, http.StatusOK, w.Code)
	requireMentionNotExists(t, db, "a")
}

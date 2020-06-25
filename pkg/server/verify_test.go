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
	"github.com/zerok/webmentiond/pkg/mailer"
	"github.com/zerok/webmentiond/pkg/policies"
	"github.com/zerok/webmentiond/pkg/server"
)

func TestVerify(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background())
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
	dummymailer := mailer.NewDummy()
	srv := server.New(func(c *server.Configuration) {
		c.Database = db
		c.MigrationsFolder = "./migrations"
		c.NotifyOnVerification = false
		c.Mailer = dummymailer
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
	_, err = srv.VerifyNextMention(ctx)
	require.NoError(t, err)
	requireMentionStatus(t, db, "a", "verified")
	requireMentionTitle(t, db, "a", "title")
	require.Len(t, dummymailer.Messages, 0)

	createMention(t, db, "b", h.URL, "http://unknown.com")
	requireMentionCount(t, db, 2)
	requireMentionStatus(t, db, "b", "new")
	_, err = srv.VerifyNextMention(ctx)
	require.NoError(t, err)
	requireMentionStatus(t, db, "b", "invalid")
	requireMentionTitle(t, db, "b", "")
}

func TestVerifyNotification(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background())
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
	dummymailer := mailer.NewDummy()
	pols := policies.NewRegistry(policies.APPROVE)
	srv := server.New(func(c *server.Configuration) {
		c.Database = db
		c.MigrationsFolder = "./migrations"
		c.NotifyOnVerification = true
		c.Mailer = dummymailer
		c.MailFrom = "sender@test.com"
		c.Auth.AdminEmails = []string{"test@test.com"}
		c.PublicURL = "http://yoursite.com"
		c.Policies = pols
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
	_, err = srv.VerifyNextMention(ctx)
	require.NoError(t, err)
	requireMentionStatus(t, db, "a", "approved")
	require.Len(t, dummymailer.Messages, 1)
	m := dummymailer.Messages[0]
	require.Equal(t, "sender@test.com", m.From)
	require.Equal(t, []string{"test@test.com"}, m.To)
	require.Equal(t, "Mention verified", m.Subject)
	require.Equal(t, fmt.Sprintf("Source: <%s>\nTarget: <http://test.com>\nNew status: approved\n\nGo to <http://yoursite.com/ui/> for details.", h.URL), m.Body)
}

func requireMentionCount(t *testing.T, db *sql.DB, expected int) {
	var count int
	if err := db.QueryRow("SELECT count(*) FROM webmentions").Scan(&count); err != nil {
		t.Fatalf("Expected %d mentions but failed to count", expected)
	}
	if expected != count {
		t.Fatalf("Expected %d mentions but found %d", expected, count)
	}

}

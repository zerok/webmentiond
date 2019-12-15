package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/golang-migrate/migrate/v4"
	migrateDriver "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
	"github.com/zerok/webmentiond/pkg/mailer"
)

const webmentionStatusNew = "new"
const webmentionStatusVerified = "verified"
const webmentionStatusInvalid = "invalid"
const webmentionStatusAccepted = "accepted"
const webmentionStatusRejected = "rejected"

// Server implements the http.Handler interface and deals with
// inserting new webmentions into the database.
type Server struct {
	cfg             Configuration
	router          chi.Router
	validToken      map[string]string
	validTokenMutex sync.RWMutex
	mailer          mailer.Mailer
}

func New(configurators ...Configurator) *Server {
	cfg := Configuration{}
	for _, configurator := range configurators {
		configurator(&cfg)
	}
	srv := &Server{
		router:     chi.NewRouter(),
		cfg:        cfg,
		validToken: make(map[string]string),
		mailer:     cfg.Mailer,
	}
	if cfg.Context != nil {
		ctx := cfg.Context
		logger := zerolog.Ctx(ctx)
		srv.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(logger.WithContext(r.Context())))
			})
		})
	}
	srv.router.Post("/receive", srv.handleReceive)
	srv.router.Post("/request-login", srv.handleLogin)
	srv.router.Post("/authenticate", srv.handleAuthenticate)
	srv.router.Get("/get", srv.handleGet)
	return srv
}

// MigrateDatabase tries to update the underlying database to the
// latest version.
func (srv *Server) MigrateDatabase(ctx context.Context) error {
	driver, err := migrateDriver.WithInstance(srv.cfg.Database, &migrateDriver.Config{})
	if err != nil {
		return fmt.Errorf("failed to create sqlite3 driver for running migrations: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+srv.cfg.MigrationsFolder,
		"sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to prepare migrations: %w", err)
	}
	err = m.Up()
	if err == nil || err == migrate.ErrNoChange {
		return nil
	}
	return err
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.router.ServeHTTP(w, r)
}

type Mention struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	CreatedAt string `json:"created_at"`
	Status    string `json:"status,omitempty"`
}

// handleGet allows a website to get a list of all mentions stored for
// it in the database.
func (srv *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: err})
		return
	}
	target := r.Form.Get("target")
	if target == "" {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("no target specified")})
		return
	}
	tx, err := srv.cfg.Database.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "select id, source, created_at, status from webmentions where status = ? and target = ? order by created_at", webmentionStatusAccepted, target)
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	defer rows.Close()
	mentions := make([]Mention, 0, 10)
	for rows.Next() {
		m := Mention{}
		if err := rows.Scan(&m.ID, &m.Source, &m.CreatedAt, &m.Status); err != nil {
			srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
			return
		}
		mentions = append(mentions, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mentions)
}

package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-migrate/migrate/v4"
	migrateDriver "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/zerok/webmentiond/pkg/webmention"
)

const webmentionStatusNew = "new"
const webmentionStatusVerified = "verified"
const webmentionStatusInvalid = "invalid"
const webmentionStatusAccepted = "accepted"
const webmentionStatusRejected = "rejected"

// Server implements the http.Handler interface and deals with
// inserting new webmentions into the database.
type Server struct {
	cfg    Configuration
	router chi.Router
}

func New(configurators ...Configurator) *Server {
	cfg := Configuration{}
	for _, configurator := range configurators {
		configurator(&cfg)
	}
	srv := &Server{
		router: chi.NewRouter(),
		cfg:    cfg,
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

func (srv *Server) sendServerError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	logger := zerolog.Ctx(ctx)
	if status >= 500 {
		logger.Error().Err(err).Msg("Processing failed")
	}
	http.Error(w, "Error", status)
}

type Mention struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	CreatedAt string `json:"created_at"`
	Status    string `json:"status,omitempty"`
}

// handleGet allows a website to get a list of all mentions stored for
// it in the database.
func (srv *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		srv.sendServerError(ctx, w, http.StatusBadRequest, err)
		return
	}
	target := r.Form.Get("target")
	if target == "" {
		srv.sendServerError(ctx, w, http.StatusBadRequest, fmt.Errorf("no target specified"))
		return
	}
	tx, err := srv.cfg.Database.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		srv.sendServerError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "select id, source, created_at, status from webmentions where status = ? and target = ? order by created_at", webmentionStatusAccepted, target)
	if err != nil {
		srv.sendServerError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()
	mentions := make([]Mention, 0, 10)
	for rows.Next() {
		m := Mention{}
		if err := rows.Scan(&m.ID, &m.Source, &m.CreatedAt, &m.Status); err != nil {
			srv.sendServerError(ctx, w, http.StatusInternalServerError, err)
			return
		}
		mentions = append(mentions, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mentions)
}

// handleReceive adds a new mention to the database in the "new"
// state.
func (srv *Server) handleReceive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := webmention.ExtractMention(r)
	if err != nil {
		srv.sendServerError(ctx, w, http.StatusBadRequest, err)
		return
	}
	tx, err := srv.cfg.Database.BeginTx(ctx, nil)
	if err != nil {
		srv.sendServerError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	now := time.Now()
	id := xid.New()
	if _, err = tx.ExecContext(ctx, "insert into webmentions (id, source, target, created_at) VALUES (?, ?, ?, ?)", id.String(), m.Source, m.Target, now.Format(time.RFC3339)); err != nil {
		tx.Rollback()
		if e, ok := err.(sqlite3.Error); ok && e.Code == sqlite3.ErrConstraint {
			srv.sendServerError(ctx, w, http.StatusBadRequest, err)
		} else {
			srv.sendServerError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		srv.sendServerError(ctx, w, http.StatusInternalServerError, err)
	}
	w.WriteHeader(201)
}

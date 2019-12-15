package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/rs/xid"
	"github.com/zerok/webmentiond/pkg/webmention"
)

// handleReceive adds a new mention to the database in the "new"
// state.
func (srv *Server) handleReceive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := webmention.ExtractMention(r)
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: err})
		return
	}
	if srv.cfg.Receiver.TargetPolicy != nil {
		if !srv.cfg.Receiver.TargetPolicy(httptest.NewRequest(http.MethodGet, m.Target, nil)) {
			srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("target domain not allowed")})
			return
		}
	}
	tx, err := srv.cfg.Database.BeginTx(ctx, nil)
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	now := time.Now()
	id := xid.New()
	if _, err = tx.ExecContext(ctx, "insert into webmentions (id, source, target, created_at) VALUES (?, ?, ?, ?)", id.String(), m.Source, m.Target, now.Format(time.RFC3339)); err != nil {
		tx.Rollback()
		if e, ok := err.(sqlite3.Error); ok && e.Code == sqlite3.ErrConstraint {
			// TODO: The mention already exists. Set the
			// re-check pending status so that it gets
			// verified again at the next best occasion.
			w.WriteHeader(201)
			return
		}
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
	}
	w.WriteHeader(201)
}

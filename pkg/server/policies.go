package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type policy struct {
	ID         int    `json:"id"`
	URLPattern string `json:"url_pattern"`
	Weight     int    `json:"weight"`
	Policy     string `json:"policy"`
}

func (s *Server) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res := make([]policy, 0, 10)
	for _, p := range s.cfg.Policies.Policies() {
		res = append(res, policy{
			ID:         p.ID,
			URLPattern: p.URLPattern.String(),
			Weight:     p.Weight,
			Policy:     string(p.Policy),
		})
	}
	json.NewEncoder(w).Encode(res)
}

func (srv *Server) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	tx, err := srv.cfg.Database.BeginTx(ctx, nil)
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM url_policies WHERE id = ?", id)
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	if err := tx.Commit(); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	srv.reloadPolicies(ctx)
}

func (srv *Server) handleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := srv.cfg.Database.BeginTx(ctx, nil)
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	p := policy{}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusBadRequest, Err: err})
		return
	}
	if p.URLPattern == "" {
		srv.sendError(ctx, w, &HTTPError{Message: "No URL pattern provided", StatusCode: http.StatusBadRequest, Err: err})
		return
	}
	if p.Policy != "approve" {
		srv.sendError(ctx, w, &HTTPError{Message: "Unsupported policy provided", StatusCode: http.StatusBadRequest, Err: err})
		return
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO url_policies (url_pattern, policy, weight) VALUES (?,?,?)", p.URLPattern, p.Policy, p.Weight)
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	if err := tx.Commit(); err != nil {
		srv.sendError(ctx, w, &HTTPError{StatusCode: http.StatusInternalServerError, Err: err})
		return
	}
	srv.reloadPolicies(ctx)
}

package server

import (
	"encoding/json"
	"net/http"

	"github.com/zerok/webmentiond/pkg/webmention"
)

type SendRequest struct {
	Source string `json:"source"`
}

type SendResponseTargetStatus struct {
	URL      string `json:"url"`
	Endpoint string `json:"endpoint"`
	Error    string `json:"error"`
}

type SendResponse struct {
	Source  string                     `json:"source"`
	Targets []SendResponseTargetStatus `json:"targets"`
}

// handleSend sends a mention based on the given source.
func (srv *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := SendRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		srv.sendError(ctx, w, &HTTPError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	resp := SendResponse{
		Source:  req.Source,
		Targets: make([]SendResponseTargetStatus, 0, 5),
	}
	doc, err := webmention.DocumentFromURL(ctx, req.Source)
	if err != nil {
		srv.sendError(ctx, w, &HTTPError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	failed := false
	for _, target := range doc.ExternalLinks() {
		status := SendResponseTargetStatus{
			URL: target,
		}
		mention := webmention.Mention{
			Source: req.Source,
			Target: target,
		}
		disc := webmention.NewEndpointDiscoverer()
		ep, err := disc.DiscoverEndpoint(ctx, mention.Target)
		if err != nil {
			status.Error = err.Error()
			failed = true
			resp.Targets = append(resp.Targets, status)
			continue
		}
		if ep == "" {
			resp.Targets = append(resp.Targets, status)
			continue
		}
		sender := webmention.NewSender()
		status.Endpoint = ep
		if err := sender.Send(ctx, ep, mention); err != nil {
			failed = true
			status.Error = err.Error()
		}
		resp.Targets = append(resp.Targets, status)
	}

	if failed {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(&resp)
}

package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ArminDashti/nevarix-hub/internal/store"
)

// registerIngestAndReads wires the write path (agents and Managment post here) and
// the read path (Managment proxies dashboards to these routes).
func (s *Server) registerIngestAndReads(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/ingest/metrics", s.requireIngestToken(s.postIngestMetrics))
	mux.HandleFunc("POST /api/v1/ingest/softether", s.requireIngestToken(s.postIngestSoftEther))
	mux.HandleFunc("GET /api/v1/agents", s.getAgents)
	mux.HandleFunc("GET /api/v1/agents/latest", s.getLatestAgent)
	mux.HandleFunc("GET /api/v1/agents/{agentId}/history", s.getAgentHistory)
	mux.HandleFunc("GET /api/v1/softether/users", s.getSoftEtherUsers)
	mux.HandleFunc("GET /api/v1/softether/users/{username}/sessions", s.getSoftEtherUserSessions)
}

func (s *Server) requireIngestToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.IngestToken == "" || !bearerMatches(r.Header.Get("Authorization"), s.cfg.IngestToken) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func bearerMatches(header, token string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	got := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return subtle.ConstantTimeCompare([]byte(got), []byte(token)) == 1
}

// postIngestMetrics responds 200 only after the snapshot is committed, so the agent can
// delete its outbox row on any 2xx.
func (s *Server) postIngestMetrics(w http.ResponseWriter, r *http.Request) {
	var p store.AgentSnapshot
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.IngestAgent(r.Context(), p, s.cfg.HistoryPoints); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type softEtherIngestBody struct {
	Sessions []store.OnlineSession `json:"sessions"`
}

func (s *Server) postIngestSoftEther(w http.ResponseWriter, r *http.Request) {
	var body softEtherIngestBody
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.SyncOnlineSessions(r.Context(), body.Sessions); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.store.ListAgents(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"agents": agents})
}

// getLatestAgent returns {"agent": null} when no agent has reported yet.
func (s *Server) getLatestAgent(w http.ResponseWriter, r *http.Request) {
	latest, err := s.store.LatestAgent(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"agent": latest})
}

func (s *Server) getAgentHistory(w http.ResponseWriter, r *http.Request) {
	points, err := s.store.AgentHistory(r.Context(), r.PathValue("agentId"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"points": points})
}

func (s *Server) getSoftEtherUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUserStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (s *Server) getSoftEtherUserSessions(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	sessions, err := s.store.ListUserSessions(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"username": username, "sessions": sessions})
}

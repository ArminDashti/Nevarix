package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-hub/internal/config"
	"github.com/ArminDashti/nevarix-hub/internal/scan"
	"github.com/ArminDashti/nevarix-hub/internal/store"
)

type Server struct {
	cfg   config.Config
	store *store.Store
}

func New(cfg config.Config, st *store.Store) *Server {
	return &Server{cfg: cfg, store: st}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1/endpoints", s.getEndpoints)
	mux.HandleFunc("POST /api/v1/endpoints", s.postEndpoint)
	mux.HandleFunc("POST /api/v1/scan", s.postScan)
	mux.HandleFunc("GET /api/v1/settings", s.getSettings)
	mux.HandleFunc("PUT /api/v1/settings", s.putSettings)
	s.registerIngestAndReads(mux)
	return s.cors(mux)
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowed := range s.cfg.CORSOrigins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
				w.Header().Set("Vary", "Origin")
				break
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type endpointView struct {
	IP         string     `json:"ip,omitempty"`
	Hostname   string     `json:"hostname"`
	Source     string     `json:"source"`
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty"`
}

// getEndpoints merges scanned/added endpoints with agents that posted to this hub.
func (s *Server) getEndpoints(w http.ResponseWriter, r *http.Request) {
	scanned, err := s.store.Endpoints(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	agents, err := s.store.ListAgents(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	endpoints := []endpointView{}
	for _, e := range scanned {
		endpoints = append(endpoints, endpointView{IP: e.IP, Hostname: e.Hostname, Source: "scan"})
	}
	for _, a := range agents {
		seen := a.LastSeenAt
		endpoints = append(endpoints, endpointView{Hostname: a.Hostname, Source: "agent", LastSeenAt: &seen})
	}
	writeJSON(w, http.StatusOK, map[string]any{"endpoints": endpoints})
}

type addEndpointBody struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}

func (s *Server) postEndpoint(w http.ResponseWriter, r *http.Request) {
	var body addEndpointBody
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if net.ParseIP(body.IP) == nil {
		writeError(w, http.StatusBadRequest, "ip is not a valid address")
		return
	}
	if err := s.store.AddEndpoint(r.Context(), store.Endpoint{IP: body.IP, Hostname: body.Hostname}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

type scanBody struct {
	CIDR      string `json:"cidr"`
	Port      int    `json:"port"`
	TimeoutMs int    `json:"timeoutMs"`
}

func (s *Server) postScan(w http.ResponseWriter, r *http.Request) {
	var body scanBody
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defaults, err := s.store.Settings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if body.CIDR == "" {
		body.CIDR = defaults.DefaultCIDR
	}
	if body.Port == 0 {
		body.Port = defaults.AgentPort
	}
	if body.TimeoutMs == 0 {
		body.TimeoutMs = defaults.TimeoutMs
	}
	if err := validatePort(body.Port); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hosts, err := scan.Hosts(body.CIDR)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hits := scan.Run(r.Context(), hosts, body.Port, time.Duration(body.TimeoutMs)*time.Millisecond)
	log.Printf("scan %s port=%d hosts=%d hits=%d", body.CIDR, body.Port, len(hosts), len(hits))
	writeJSON(w, http.StatusOK, map[string]any{"scanned": len(hosts), "hits": hits})
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	st, err := s.store.Settings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *Server) putSettings(w http.ResponseWriter, r *http.Request) {
	var st store.Settings
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&st); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, _, err := net.ParseCIDR(st.DefaultCIDR); err != nil {
		writeError(w, http.StatusBadRequest, "defaultCidr is not a valid CIDR")
		return
	}
	if err := validatePort(st.AgentPort); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if st.TimeoutMs < 50 || st.TimeoutMs > 10000 {
		writeError(w, http.StatusBadRequest, "timeoutMs must be between 50 and 10000")
		return
	}
	if err := s.store.SaveSettings(r.Context(), st); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func validatePort(p int) error {
	if p < 1 || p > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": strings.TrimSpace(msg)})
}

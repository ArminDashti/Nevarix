package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Identity is what nevarix-hub reads from /health to confirm a host runs this agent.
type Identity struct {
	Agent    string `json:"agent"`
	AgentID  string `json:"agentId"`
	Hostname string `json:"hostname"`
}

func Handler(id Identity) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(id)
	})
	return mux
}

// Serve blocks until ctx is cancelled, then shuts the listener down.
func Serve(ctx context.Context, addr string, id Identity) error {
	srv := &http.Server{Addr: addr, Handler: Handler(id), ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-core/internal/paths"
)

// Agent is one registered nevarix-agent endpoint.
type Agent struct {
	Alias      string    `json:"alias"`
	Address    string    `json:"address"`    // host:port
	Protocol   string    `json:"protocol"`   // restful | grpc
	SecretKey  string    `json:"secret_key"` // ponytail: plain text in state file; upgrade to Windows DPAPI
	Enabled    bool      `json:"enabled"`
	LastScanAt time.Time `json:"last_scan_at,omitempty"`
	LastScanOK bool      `json:"last_scan_ok"`
}

// Stats counts scan activity across all agents.
type Stats struct {
	ScanRuns     int       `json:"scan_runs"`
	ScanFailures int       `json:"scan_failures"`
	LastScanAt   time.Time `json:"last_scan_at,omitempty"`
}

// State is the persisted core data: registered agents and counters.
type State struct {
	Agents map[string]Agent `json:"agents"`
	Stats  Stats            `json:"stats"`
}

// Path returns NEVARIX_CORE_STATE_PATH or core-state.json inside paths.DataDir().
func Path() string {
	if p := strings.TrimSpace(os.Getenv("NEVARIX_CORE_STATE_PATH")); p != "" {
		return p
	}
	return filepath.Join(paths.DataDir(), "core-state.json")
}

// Load reads state; a missing file is an empty state, not an error.
func Load(path string) (State, error) {
	st := State{Agents: map[string]Agent{}}
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal(raw, &st); err != nil {
		return st, fmt.Errorf("parse %s: %w", path, err)
	}
	if st.Agents == nil {
		st.Agents = map[string]Agent{}
	}
	return st, nil
}

// Save writes state atomically via a temp file so a crash cannot truncate it.
func Save(path string, st State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Aliases returns agent aliases sorted for stable output.
func (st State) Aliases() []string {
	return slices.Sorted(maps.Keys(st.Agents))
}

// EnabledCount returns how many agents are enabled.
func (st State) EnabledCount() int {
	n := 0
	for _, a := range st.Agents {
		if a.Enabled {
			n++
		}
	}
	return n
}

// ParseAddress validates "host:port" and returns it normalized.
func ParseAddress(s string) (string, error) {
	host, portStr, err := net.SplitHostPort(s)
	if err != nil || host == "" {
		return "", errors.New("expected address:port")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return "", errors.New("port must be 1-65535")
	}
	return net.JoinHostPort(host, portStr), nil
}

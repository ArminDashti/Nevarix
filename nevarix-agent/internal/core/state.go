package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Address is one nevarix-core endpoint saved under an alias.
type Address struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// State is the persisted Core connection data (no secrets).
type State struct {
	Active    string             `json:"active,omitempty"`
	Addresses map[string]Address `json:"addresses"`
}

// StatePath returns NEVARIX_CORE_STATE_PATH or the default data file.
func StatePath() string {
	if p := strings.TrimSpace(os.Getenv("NEVARIX_CORE_STATE_PATH")); p != "" {
		return p
	}
	return filepath.Join("data", "core-state.json")
}

// Load reads state; a missing file is an empty state, not an error.
func Load(path string) (State, error) {
	st := State{Addresses: map[string]Address{}}
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
	if st.Addresses == nil {
		st.Addresses = map[string]Address{}
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

// CoreURL returns the base URL of the active Core, if one is connected.
func (st State) CoreURL() (string, bool) {
	a, ok := st.Addresses[st.Active]
	if st.Active == "" || !ok {
		return "", false
	}
	return fmt.Sprintf("http://%s", net.JoinHostPort(a.Address, strconv.Itoa(a.Port))), true
}

// ParseEntry parses "alias:address:port" (address may be an IPv6 literal in [brackets]).
func ParseEntry(s string) (string, Address, error) {
	alias, rest, ok := strings.Cut(s, ":")
	if !ok || alias == "" {
		return "", Address{}, errors.New("expected alias:address:port")
	}
	a, err := ParseHostPort(rest)
	return alias, a, err
}

// ParseHostPort parses "address:port".
func ParseHostPort(s string) (Address, error) {
	host, portStr, err := net.SplitHostPort(s)
	if err != nil || host == "" {
		return Address{}, errors.New("expected address:port")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return Address{}, errors.New("port must be 1-65535")
	}
	return Address{Address: host, Port: port}, nil
}

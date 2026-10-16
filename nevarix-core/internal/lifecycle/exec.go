package lifecycle

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// run executes a command in the current directory, streaming its output.
func run(label string, env []string, name string, args ...string) error {
	fmt.Printf("==> %s\n", label)
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", label, err)
	}
	return nil
}

// capture returns trimmed stdout and whether the command exited 0.
func capture(env []string, name string, args ...string) (string, bool) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err == nil
}

func portFree(port int) bool {
	l, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return false
	}
	return l.Close() == nil
}

// probe reports whether the URL answers with any HTTP response.
func probe(url string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func waitUntil(check func() bool, seconds int) bool {
	for i := 0; i < seconds; i++ {
		if check() {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

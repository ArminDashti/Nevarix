//go:build windows

package service

import (
	"fmt"
	"os/exec"
	"strings"
)

// Name is the Windows service name the agent daemon is registered under.
const Name = "NevarixAgent"

const envScope = "run as administrator"

// Start, Stop, Restart and Status drive the Windows service through sc.exe (no extra dependency).
func Start() error  { return run("start") }
func Stop() error   { return run("stop") }
func Delete() error { return run("delete") }

// Create registers the service to run daemonBin automatically at boot.
func Create(daemonBin, _ string) error {
	out, err := exec.Command("sc.exe", "create", Name, "binPath=", daemonBin, "start=", "auto").CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc create %s (%s): %w: %s", Name, envScope, err, strings.TrimSpace(string(out)))
	}
	fmt.Println(strings.TrimSpace(string(out)))
	return nil
}

func Restart() error {
	_ = Stop()
	return Start()
}

func Status() (string, error) {
	out, err := exec.Command("sc.exe", "query", Name).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// PersistEnv stores machine-wide environment variables for the service process.
func PersistEnv(vars map[string]string) error {
	for key, val := range vars {
		out, err := exec.Command("setx", "/M", key, val).CombinedOutput()
		if err != nil {
			return fmt.Errorf("setx %s (%s): %w: %s", key, envScope, err, strings.TrimSpace(string(out)))
		}
	}
	return nil
}

func run(verb string) error {
	out, err := exec.Command("sc.exe", verb, Name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc %s %s: %w: %s", verb, Name, err, strings.TrimSpace(string(out)))
	}
	fmt.Println(strings.TrimSpace(string(out)))
	return nil
}

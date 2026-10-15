package service

import (
	"fmt"
	"os/exec"
	"strings"
)

// Name is the Windows service name the core daemon is registered under.
const Name = "NevarixCore"

// Start, Stop, Restart and Status drive the Windows service through sc.exe (no extra dependency).
func Start() error  { return run("start") }
func Stop() error   { return run("stop") }
func Delete() error { return run("delete") }

// Create registers the service to run binPath automatically at boot.
func Create(binPath string) error {
	out, err := exec.Command("sc.exe", "create", Name, "binPath=", binPath, "start=", "auto").CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc create %s: %w: %s", Name, err, strings.TrimSpace(string(out)))
	}
	fmt.Println(strings.TrimSpace(string(out)))
	return nil
}
func Restart() error {
	_ = Stop() // ponytail: ignores stop error when already stopped; upgrade to query state first
	return Start()
}

func Status() (string, error) {
	out, err := exec.Command("sc.exe", "query", Name).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func run(verb string) error {
	out, err := exec.Command("sc.exe", verb, Name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc %s %s: %w: %s", verb, Name, err, strings.TrimSpace(string(out)))
	}
	fmt.Println(strings.TrimSpace(string(out)))
	return nil
}

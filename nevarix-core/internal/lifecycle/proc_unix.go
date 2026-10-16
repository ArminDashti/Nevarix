//go:build !windows

package lifecycle

import (
	"errors"
	"os/exec"
	"syscall"
	"time"
)

// detach puts the child in its own session so it survives the CLI exiting.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func isAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// ponytail: TERM then KILL after 10s; pid reuse after exit is not guarded (ceiling: stale pid file).
func killTree(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	for i := 0; i < 100 && isAlive(pid); i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if isAlive(pid) {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	}
}

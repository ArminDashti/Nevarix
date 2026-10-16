//go:build windows

package lifecycle

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

const (
	detachedProcess       = 0x00000008
	createNewProcessGroup = 0x00000200
)

// detach starts the child outside this console so it survives the CLI exiting.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: detachedProcess | createNewProcessGroup,
		HideWindow:    true,
	}
}

func isAlive(pid int) bool {
	out, ok := capture(nil, "tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH", "/FO", "CSV")
	return ok && strings.Contains(out, `"`+strconv.Itoa(pid)+`"`)
}

// ponytail: pid reuse after exit is not guarded (ceiling: stale pid file). Same as the Node CLI.
func killTree(pid int) {
	_, _ = capture(nil, "taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
}

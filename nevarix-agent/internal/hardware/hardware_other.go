//go:build !linux && !windows

package hardware

import "github.com/ArminDashti/nevarix-agent/internal/sysreport"

func fillOS(h *sysreport.HardwareFull) {}

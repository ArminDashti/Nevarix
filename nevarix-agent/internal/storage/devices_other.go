//go:build !linux && !windows

package storage

import "github.com/ArminDashti/nevarix-agent/internal/sysreport"

func listDevices() []sysreport.DiskDevice { return nil }

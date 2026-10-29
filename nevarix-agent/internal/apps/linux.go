//go:build !windows

package apps

import (
	"os/exec"
	"strings"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

func scanApps() map[string]sysreport.AppInfo {
	out := map[string]sysreport.AppInfo{}
	// shortcut: only dpkg-based distros for now; add rpm/apk when needed
	if out2, err := exec.Command("dpkg-query", "-W", "-f", "${Package}\t${Version}\n").Output(); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out2)), "\n") {
			name, version, ok := strings.Cut(line, "\t")
			if !ok || name == "" {
				continue
			}
			out[name+"\x00"+version] = sysreport.AppInfo{
				Name: name, Version: version, Source: "dpkg",
			}
		}
	}
	return out
}

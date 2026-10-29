//go:build windows

package apps

import (
	"golang.org/x/sys/windows/registry"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

func scanApps() map[string]sysreport.AppInfo {
	out := map[string]sysreport.AppInfo{}
	roots := []struct {
		root      registry.Key
		path      string
		wow64Path string
	}{
		{registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Uninstall`, `Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall`, ""},
	}
	for _, r := range roots {
		paths := []string{r.path}
		if r.wow64Path != "" {
			paths = append(paths, r.wow64Path)
		}
		for _, p := range paths {
			key, err := registry.OpenKey(r.root, p, registry.ENUMERATE_SUB_KEYS)
			if err != nil {
				continue
			}
			names, _ := key.ReadSubKeyNames(-1)
			for _, sub := range names {
				sk, err := registry.OpenKey(r.root, p+`\`+sub, registry.QUERY_VALUE)
				if err != nil {
					continue
				}
				name, _, _ := sk.GetStringValue("DisplayName")
				if name == "" {
					sk.Close()
					continue
				}
				version, _, _ := sk.GetStringValue("DisplayVersion")
				publisher, _, _ := sk.GetStringValue("Publisher")
				date, _, _ := sk.GetStringValue("InstallDate")
				sk.Close()
				out[name+"\x00"+version] = sysreport.AppInfo{
					Name: name, Version: version, Publisher: publisher,
					InstallDate: date, Source: "registry",
				}
			}
			key.Close()
		}
	}
	return out
}

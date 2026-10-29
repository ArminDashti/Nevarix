// Package apps collects the installed-application inventory and diffs it
// between scans to emit app_added / app_removed events.
package apps

import (
	"sort"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

type appEntry = sysreport.AppInfo

// Collector keeps the previous scan's app list to detect changes.
type Collector struct {
	scan     func() map[string]sysreport.AppInfo
	previous map[string]sysreport.AppInfo
}

func New() *Collector {
	return &Collector{scan: scanApps}
}

// Collect returns the current app list and events for changes since the last call.
func (c *Collector) Collect() ([]sysreport.AppInfo, []sysreport.EventInfo) {
	current := c.scan()
	now := time.Now().UTC()
	var events []sysreport.EventInfo
	if c.previous != nil {
		for name, app := range current {
			if _, ok := c.previous[name]; !ok {
				events = append(events, sysreport.EventInfo{
					Kind: "app_added", Message: app.Name + " " + app.Version, OccurredAt: now,
				})
			}
		}
		for name, app := range c.previous {
			if _, ok := current[name]; !ok {
				events = append(events, sysreport.EventInfo{
					Kind: "app_removed", Message: name + " " + app.Version, OccurredAt: now,
				})
			}
		}
	}
	c.previous = make(map[string]sysreport.AppInfo, len(current))
	for name, app := range current {
		c.previous[name] = app
	}
	out := make([]sysreport.AppInfo, 0, len(current))
	for _, app := range current {
		out = append(out, app)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, events
}

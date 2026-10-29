package apps

import "testing"

func TestCollectEmitsAddedAndRemoved(t *testing.T) {
	c := New()
	c.scan = func() map[string]appEntry {
		return map[string]appEntry{"a": {Name: "A", Version: "1"}}
	}
	appList, events := c.Collect()
	if len(appList) != 1 || len(events) != 0 {
		t.Fatalf("first collect: apps=%d events=%d", len(appList), len(events))
	}
	c.scan = func() map[string]appEntry {
		return map[string]appEntry{"b": {Name: "B", Version: "2"}}
	}
	_, events = c.Collect()
	if len(events) != 2 {
		t.Fatalf("want 2 events, got %d: %+v", len(events), events)
	}
	kinds := map[string]int{}
	for _, e := range events {
		kinds[e.Kind]++
	}
	if kinds["app_added"] != 1 || kinds["app_removed"] != 1 {
		t.Fatalf("wrong kinds: %+v", kinds)
	}
}

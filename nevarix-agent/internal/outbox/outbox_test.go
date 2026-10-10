package outbox

import (
	"context"
	"path/filepath"
	"testing"
)

// Guarantee under test: a row survives until Ack, and Ack removes only that row.
func TestUnackedRowsSurvive(t *testing.T) {
	ctx := context.Background()
	box, err := Open(filepath.Join(t.TempDir(), "outbox.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer box.Close()

	for _, p := range []string{"a", "b"} {
		if err := box.Enqueue(ctx, []byte(p)); err != nil {
			t.Fatal(err)
		}
	}
	items, err := box.Pending(ctx, 10)
	if err != nil || len(items) != 2 || string(items[0].Payload) != "a" {
		t.Fatalf("want 2 ordered items, got %v err=%v", items, err)
	}
	if err := box.Ack(ctx, items[0].ID); err != nil {
		t.Fatal(err)
	}
	left, err := box.Pending(ctx, 10)
	if err != nil || len(left) != 1 || string(left[0].Payload) != "b" {
		t.Fatalf("want only b left, got %v err=%v", left, err)
	}
}

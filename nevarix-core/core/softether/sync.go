package softether

import (
	"context"
	"log"
	"time"

	"github.com/ArminDashti/nevarix-core/core/store"
)

// Sync polls SoftEther every interval and upserts the online set into the store.
// A failed poll leaves stored sessions untouched, so a SoftEther outage never closes them.
// ponytail: no retry or backoff; the next tick is the retry. Upgrade path: exponential backoff.
func Sync(ctx context.Context, interval time.Duration, client *Client, st *store.Store) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		pollCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		sessions, err := client.ListOnlineSessions(pollCtx)
		if err != nil {
			log.Printf("softether poll: %v", err)
		} else if err := st.SyncOnlineSessions(pollCtx, sessions); err != nil {
			log.Printf("softether sync: %v", err)
		}
		cancel()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

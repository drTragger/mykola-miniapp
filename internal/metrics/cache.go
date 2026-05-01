package metrics

import (
	"context"
	"log"
	"time"

	"github.com/drTragger/mykola-miniapp/internal/cache"
)

var snapshotCache = cache.New(Collect)

func StartBackgroundRefresh(ctx context.Context, interval time.Duration) {
	snapshotCache.Start(ctx, interval)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				resp, err := GetSnapshot()
				if err != nil {
					continue
				}

				collectedAt, _ := time.Parse(time.RFC3339, resp.CollectedAt)
				if collectedAt.IsZero() {
					collectedAt = time.Now()
				}

				if err := storeHistorySnapshot(resp, collectedAt); err != nil {
					log.Printf("metrics history store error: %v", err)
				}
			}
		}
	}()
}

func GetSnapshot() (Response, error) {
	return snapshotCache.Get()
}

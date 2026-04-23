package system

import (
	"context"
	"time"

	"github.com/drTragger/mykola-miniapp/internal/cache"
)

var snapshotCache = cache.New(Collect)

func StartBackgroundRefresh(ctx context.Context, interval time.Duration) {
	snapshotCache.Start(ctx, interval)
}

func GetSnapshot() (Response, error) {
	return snapshotCache.Get()
}

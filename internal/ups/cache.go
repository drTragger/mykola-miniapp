package ups

import (
	"log"
	"sync"
	"time"
)

var (
	cacheMu       sync.RWMutex
	cached        Response
	cacheReady    bool
	backgroundRun sync.Once
)

func StartBackgroundRefresh(interval time.Duration) {
	backgroundRun.Do(func() {
		refresh := func() {
			resp, err := Collect()
			if err != nil {
				log.Printf("ups refresh error: %v", err)
				return
			}

			cacheMu.Lock()
			cached = resp
			cacheReady = true
			cacheMu.Unlock()

			collectedAt, err := time.Parse(time.RFC3339, resp.CollectedAt)
			if err != nil {
				collectedAt = time.Now()
			}

			if err := storeHistorySnapshot(resp.Data, collectedAt); err != nil {
				log.Printf("ups history store error: %v", err)
			}
		}

		refresh()

		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for range ticker.C {
				refresh()
			}
		}()
	})
}

func GetSnapshot() (Response, error) {
	cacheMu.RLock()
	if cacheReady {
		resp := cached
		cacheMu.RUnlock()
		return resp, nil
	}
	cacheMu.RUnlock()

	resp, err := Collect()
	if err != nil {
		return Response{}, err
	}

	cacheMu.Lock()
	cached = resp
	cacheReady = true
	cacheMu.Unlock()

	collectedAt, parseErr := time.Parse(time.RFC3339, resp.CollectedAt)
	if parseErr != nil {
		collectedAt = time.Now()
	}

	if err := storeHistorySnapshot(resp.Data, collectedAt); err != nil {
		log.Printf("ups history store error: %v", err)
	}

	return resp, nil
}

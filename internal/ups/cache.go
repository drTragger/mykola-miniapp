package ups

import (
	"log"
	"sync"
	"time"
)

var (
	cacheMu          sync.RWMutex
	cached           Response
	cacheReady       bool
	backgroundRun    sync.Once
	lastSuccessAt    time.Time
	lastRefreshError string
	staleAfter       = 30 * time.Second
)

func StartBackgroundRefresh(interval time.Duration) {
	backgroundRun.Do(func() {
		refresh := func() {
			resp, err := Collect()
			if err != nil {
				cacheMu.Lock()
				lastRefreshError = err.Error()
				cacheMu.Unlock()

				log.Printf("ups refresh error: %v", err)
				return
			}

			collectedAt, err := time.Parse(time.RFC3339, resp.CollectedAt)
			if err != nil {
				collectedAt = time.Now()
				resp.CollectedAt = collectedAt.Format(time.RFC3339)
			}

			resp.Stale = false
			resp.LastSuccessAt = collectedAt.Format(time.RFC3339)

			cacheMu.Lock()
			cached = resp
			cacheReady = true
			lastSuccessAt = collectedAt
			lastRefreshError = ""
			cacheMu.Unlock()

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
		lastOK := lastSuccessAt
		lastErr := lastRefreshError
		cacheMu.RUnlock()

		if !lastOK.IsZero() {
			resp.LastSuccessAt = lastOK.Format(time.RFC3339)
			resp.Stale = time.Since(lastOK) > staleAfter
			if resp.Stale {
				if lastErr != "" {
					resp.Error = "UPS дані застаріли: " + lastErr
				} else {
					resp.Error = "UPS дані застаріли"
				}
			}
		}

		return resp, nil
	}
	cacheMu.RUnlock()

	resp, err := Collect()
	if err != nil {
		return Response{}, err
	}

	collectedAt, parseErr := time.Parse(time.RFC3339, resp.CollectedAt)
	if parseErr != nil {
		collectedAt = time.Now()
		resp.CollectedAt = collectedAt.Format(time.RFC3339)
	}

	resp.Stale = false
	resp.LastSuccessAt = collectedAt.Format(time.RFC3339)

	cacheMu.Lock()
	cached = resp
	cacheReady = true
	lastSuccessAt = collectedAt
	lastRefreshError = ""
	cacheMu.Unlock()

	if err := storeHistorySnapshot(resp.Data, collectedAt); err != nil {
		log.Printf("ups history store error: %v", err)
	}

	return resp, nil
}

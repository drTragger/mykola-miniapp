package metrics

import (
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
				return
			}

			cacheMu.Lock()
			cached = resp
			cacheReady = true
			cacheMu.Unlock()
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

	return resp, nil
}

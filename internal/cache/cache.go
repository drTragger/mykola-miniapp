package cache

import (
	"context"
	"sync"
	"time"
)

type Snapshot[T any] struct {
	mu     sync.RWMutex
	value  T
	ready  bool
	once   sync.Once
	fetch  func() (T, error)
	onHit  func(T)
	onMiss func(T)
}

func New[T any](fetch func() (T, error)) *Snapshot[T] {
	return &Snapshot[T]{fetch: fetch}
}

func (s *Snapshot[T]) OnStore(fn func(T)) *Snapshot[T] {
	s.onMiss = fn
	return s
}

func (s *Snapshot[T]) Start(ctx context.Context, interval time.Duration) {
	s.once.Do(func() {
		s.refresh()

		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					s.refresh()
				}
			}
		}()
	})
}

func (s *Snapshot[T]) refresh() {
	value, err := s.fetch()
	if err != nil {
		return
	}

	s.mu.Lock()
	s.value = value
	s.ready = true
	s.mu.Unlock()

	if s.onMiss != nil {
		s.onMiss(value)
	}
}

func (s *Snapshot[T]) Get() (T, error) {
	s.mu.RLock()
	if s.ready {
		value := s.value
		s.mu.RUnlock()
		return value, nil
	}
	s.mu.RUnlock()

	value, err := s.fetch()
	if err != nil {
		var zero T
		return zero, err
	}

	s.mu.Lock()
	s.value = value
	s.ready = true
	s.mu.Unlock()

	if s.onMiss != nil {
		s.onMiss(value)
	}

	return value, nil
}

func (s *Snapshot[T]) Store(value T) {
	s.mu.Lock()
	s.value = value
	s.ready = true
	s.mu.Unlock()
}

func (s *Snapshot[T]) Peek() (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value, s.ready
}

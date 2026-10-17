package zones

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

type mockProvider struct {
	mu    sync.Mutex
	calls int
	zones []string
	err   error
}

func (m *mockProvider) Zones(ctx context.Context) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	return m.zones, m.err
}

func (m *mockProvider) set(zones []string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.zones = zones
	m.err = err
}

func (m *mockProvider) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func TestCacheLoadsLazily(t *testing.T) {
	p := &mockProvider{zones: []string{"example.com."}}
	c := NewCache(p, time.Minute)
	if p.callCount() != 0 {
		t.Fatalf("expected no provider calls after NewCache, got %d", p.callCount())
	}
	zones, err := c.Zones(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !slices.Equal(zones, []string{"example.com."}) {
		t.Errorf("unexpected zones: %v", zones)
	}
	if p.callCount() != 1 {
		t.Errorf("expected 1 provider call, got %d", p.callCount())
	}
}

func TestCacheUsesCacheWithinTTL(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &mockProvider{zones: []string{"example.com."}}
		c := NewCache(p, time.Minute)
		_, _ = c.Zones(context.Background())
		time.Sleep(30 * time.Second)
		_, _ = c.Zones(context.Background())
		if p.callCount() != 1 {
			t.Errorf("expected 1 provider call, got %d", p.callCount())
		}
	})
}

func TestCacheRefreshesAfterTTL(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &mockProvider{zones: []string{"example.com."}}
		c := NewCache(p, time.Minute)
		_, _ = c.Zones(context.Background())
		p.set([]string{"example.net."}, nil)
		time.Sleep(time.Minute)
		zones, err := c.Zones(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !slices.Equal(zones, []string{"example.net."}) {
			t.Errorf("unexpected zones: %v", zones)
		}
		if p.callCount() != 2 {
			t.Errorf("expected 2 provider calls, got %d", p.callCount())
		}
	})
}

func TestCacheErrorWhenNeverLoaded(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		boom := errors.New("boom")
		p := &mockProvider{err: boom}
		c := NewCache(p, time.Minute)
		zones, err := c.Zones(context.Background())
		if !errors.Is(err, boom) {
			t.Fatalf("expected error %v, got %v", boom, err)
		}
		if len(zones) != 0 {
			t.Errorf("expected no zones, got %v", zones)
		}

		time.Sleep(failureBackoff - time.Second)
		_, err = c.Zones(context.Background())
		if !errors.Is(err, boom) {
			t.Errorf("expected cached error %v, got %v", boom, err)
		}
		if p.callCount() != 1 {
			t.Errorf("expected 1 provider call during backoff, got %d", p.callCount())
		}

		time.Sleep(time.Second)
		_, _ = c.Zones(context.Background())
		if p.callCount() != 2 {
			t.Errorf("expected 2 provider calls after backoff, got %d", p.callCount())
		}
	})
}

func TestCacheErrorKeepsStale(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		boom := errors.New("boom")
		p := &mockProvider{zones: []string{"example.com."}}
		c := NewCache(p, time.Minute)
		_, _ = c.Zones(context.Background())

		p.set(nil, boom)
		time.Sleep(time.Minute)
		zones, err := c.Zones(context.Background())
		if !errors.Is(err, boom) {
			t.Fatalf("expected error %v, got %v", boom, err)
		}
		if !slices.Equal(zones, []string{"example.com."}) {
			t.Errorf("expected stale zones, got %v", zones)
		}

		p.set([]string{"example.net."}, nil)
		time.Sleep(failureBackoff)
		zones, err = c.Zones(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !slices.Equal(zones, []string{"example.net."}) {
			t.Errorf("expected new zones, got %v", zones)
		}
	})
}

func TestCacheConcurrent(t *testing.T) {
	p := &mockProvider{zones: []string{"example.com."}}
	c := NewCache(p, time.Minute)
	var wg sync.WaitGroup
	results := make([][]string, 100)
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], _ = c.Zones(context.Background())
		}()
	}
	wg.Wait()
	if p.callCount() != 1 {
		t.Errorf("expected 1 provider call, got %d", p.callCount())
	}
	for i, zones := range results {
		if !slices.Equal(zones, []string{"example.com."}) {
			t.Errorf("result %d: unexpected zones %v", i, zones)
		}
	}
}

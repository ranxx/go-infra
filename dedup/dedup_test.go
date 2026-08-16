package dedup

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// mockClient implements Client for testing
type mockClient struct {
	mu   sync.Mutex
	keys map[string]bool
}

func newMockClient() *mockClient {
	return &mockClient{keys: make(map[string]bool)}
}

func (m *mockClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.keys[key] {
		return false, nil
	}
	m.keys[key] = true
	return true, nil
}

func TestNew(t *testing.T) {
	d := New(newMockClient(), 5*time.Second)
	if d == nil {
		t.Fatal("expected non-nil Dedup")
	}
	if d.ttl != 5*time.Second {
		t.Errorf("expected ttl 5s, got %v", d.ttl)
	}
}

func TestTryPublishFirstWin(t *testing.T) {
	d := New(newMockClient(), 5*time.Second)

	published := false
	first, err := d.TryPublish(context.Background(), "test:key", func() error {
		published = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !first {
		t.Error("expected true for first publish")
	}
	if !published {
		t.Error("expected publish function to be called")
	}
}

func TestTryPublishSecondSkipped(t *testing.T) {
	d := New(newMockClient(), 5*time.Second)

	published1 := false
	first1, _ := d.TryPublish(context.Background(), "test:key", func() error {
		published1 = true
		return nil
	})

	published2 := false
	first2, _ := d.TryPublish(context.Background(), "test:key", func() error {
		published2 = true
		return nil
	})

	if !first1 || !published1 {
		t.Error("first should publish")
	}
	if first2 {
		t.Error("second should NOT be first (duplicate)")
	}
	if published2 {
		t.Error("second publish function should not be called")
	}
}

func TestTryPublishDifferentKeys(t *testing.T) {
	d := New(newMockClient(), 5*time.Second)

	count := 0
	first1, _ := d.TryPublish(context.Background(), "key:1", func() error { count++; return nil })
	first2, _ := d.TryPublish(context.Background(), "key:2", func() error { count++; return nil })

	if !first1 || !first2 {
		t.Error("different keys should both be first")
	}
	if count != 2 {
		t.Errorf("expected 2 publishes, got %d", count)
	}
}

func TestTryPublishPublishError(t *testing.T) {
	d := New(newMockClient(), 5*time.Second)

	expectedErr := errors.New("publish failed")
	first, err := d.TryPublish(context.Background(), "test:key", func() error {
		return expectedErr
	})

	if !first {
		t.Error("dedup should report first (Redis SETNX succeeded)")
	}
	if err == nil {
		t.Error("expected publish error to be propagated")
	}
}

func TestTryPublishConcurrent(t *testing.T) {
	d := New(newMockClient(), 5*time.Second)

	var count int32
	var mu sync.Mutex
	firstWins := 0

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			first, _ := d.TryPublish(context.Background(), "concurrent:key", func() error {
				mu.Lock()
				count++
				mu.Unlock()
				return nil
			})
			if first {
				mu.Lock()
				firstWins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if count != 1 {
		t.Errorf("expected exactly 1 publish, got %d", count)
	}
	if firstWins != 1 {
		t.Errorf("expected exactly 1 first win, got %d", firstWins)
	}
}

func TestKey(t *testing.T) {
	ts := time.Unix(0, 1781331259909000000)
	k := Key("binance", "spot", "BTCUSDT", "trade", 6405763855, ts)

	expected := "tick:binance:spot:BTCUSDT:trade:6405763855:1781331259909000000"
	if k != expected {
		t.Errorf("expected %q, got %q", expected, k)
	}
}

func TestKeyString(t *testing.T) {
	k := KeyString("binance", "spot", "BTCUSDT", "depth", "1781331260001")
	expected := "tick:binance:spot:BTCUSDT:depth:1781331260001"
	if k != expected {
		t.Errorf("expected %q, got %q", expected, k)
	}
}

func TestKeyIsolation(t *testing.T) {
	ts := time.Unix(0, 1000000)
	// 不同交易所的相同 symbol 应有不同的去重键
	k1 := Key("binance", "spot", "BTCUSDT", "trade", 123, ts)
	k2 := Key("okx", "spot", "BTCUSDT", "trade", 123, ts)

	if k1 == k2 {
		t.Errorf("keys should differ across exchanges: %q", k1)
	}

	// 不同品种的相同 symbol 也应有不同的去重键
	k3 := Key("binance", "spot", "BTCUSDT", "trade", 123, ts)
	k4 := Key("binance", "futures", "BTCUSDT", "trade", 123, ts)
	if k3 == k4 {
		t.Errorf("keys should differ across instTypes: %q", k3)
	}
}

func TestClientInterface(t *testing.T) {
	// Verify that the go-infra/redis.RedisClient satisfies the dedup.Client interface
	// (compilation check)
	var _ Client = newMockClient()
}

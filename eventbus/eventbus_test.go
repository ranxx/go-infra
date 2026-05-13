package eventbus

import (
	"sync"
	"testing"
	"time"
)

type testEvent struct {
	ID   int
	Data string
}

type anotherEvent struct {
	Value float64
}

func TestEventBus_SubscribePublish(t *testing.T) {
	bus := New()

	ch := Subscribe[testEvent](bus, 10)

	event := testEvent{ID: 1, Data: "hello"}
	Publish(bus, event)

	select {
	case received := <-ch:
		if received.ID != 1 || received.Data != "hello" {
			t.Errorf("unexpected event: %+v", received)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := New()

	ch1 := Subscribe[testEvent](bus, 10)
	ch2 := Subscribe[testEvent](bus, 10)

	event := testEvent{ID: 42, Data: "broadcast"}
	Publish(bus, event)

	for i, ch := range []<-chan testEvent{ch1, ch2} {
		select {
		case received := <-ch:
			if received.ID != 42 {
				t.Errorf("subscriber %d: unexpected event ID: %d", i, received.ID)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d: timeout waiting for event", i)
		}
	}
}

func TestEventBus_TypeIsolation(t *testing.T) {
	bus := New()

	ch1 := Subscribe[testEvent](bus, 10)
	ch2 := Subscribe[anotherEvent](bus, 10)

	Publish(bus, testEvent{ID: 1, Data: "type1"})
	Publish(bus, anotherEvent{Value: 3.14})

	select {
	case received := <-ch1:
		if received.ID != 1 {
			t.Errorf("unexpected event from ch1: %+v", received)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for testEvent on ch1")
	}

	select {
	case received := <-ch2:
		if received.Value != 3.14 {
			t.Errorf("unexpected event from ch2: %+v", received)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for anotherEvent on ch2")
	}

	// ch1 should not receive anotherEvent
	select {
	case <-ch1:
		t.Error("ch1 should not receive anotherEvent")
	default:
	}
}

func TestEventBus_NonBlockingPublish(t *testing.T) {
	bus := New()

	// buffer size 1, fill it up
	ch := Subscribe[testEvent](bus, 1)

	Publish(bus, testEvent{ID: 1}) // fills buffer
	Publish(bus, testEvent{ID: 2}) // should not block (dropped)

	select {
	case received := <-ch:
		if received.ID != 1 {
			t.Errorf("expected first event, got: %+v", received)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := New()

	ch := Subscribe[testEvent](bus, 10000)

	var wg sync.WaitGroup
	count := 100

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			Publish(bus, testEvent{ID: n, Data: "concurrent"})
		}(i)
	}

	wg.Wait()

	received := make(map[int]bool)
	for i := 0; i < count; i++ {
		select {
		case event := <-ch:
			received[event.ID] = true
		case <-time.After(time.Second):
			t.Fatalf("timeout after receiving %d/%d events", len(received), count)
		}
	}

	if len(received) != count {
		t.Errorf("expected %d unique events, got %d", count, len(received))
	}
}

func TestEventBus_EmptyBus(t *testing.T) {
	bus := New()

	// publishing with no subscribers should not panic
	Publish(bus, testEvent{ID: 1, Data: "no subs"})

	// subscribing with zero buffer
	ch := Subscribe[testEvent](bus, 0)
	// publish to zero-buffer channel should not block
	Publish(bus, testEvent{ID: 2, Data: "zero buf"})

	select {
	case <-ch:
		t.Error("should not receive on zero-buffer channel (non-blocking publish)")
	default:
	}
}

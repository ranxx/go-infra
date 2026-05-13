package eventbus

import "sync"

// EventBus 类型安全的进程内事件总线
type EventBus struct {
	mu    sync.RWMutex
	chans []any
}

// New 创建 EventBus
func New() *EventBus {
	return &EventBus{
		chans: make([]any, 0),
	}
}

// Subscribe 订阅指定类型的事件，返回只读 channel
func Subscribe[T any](bus *EventBus, bufSize int) <-chan T {
	ch := make(chan T, bufSize)
	bus.mu.Lock()
	bus.chans = append(bus.chans, ch)
	bus.mu.Unlock()
	return ch
}

// Publish 发布事件到所有匹配类型的订阅者
// 非阻塞：缓冲区满时丢弃
func Publish[T any](bus *EventBus, event T) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	for _, raw := range bus.chans {
		if ch, ok := raw.(chan T); ok {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

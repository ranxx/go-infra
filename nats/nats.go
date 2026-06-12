package nats

import (
	"context"
	"fmt"
	"time"

	natsgo "github.com/nats-io/nats.go"
)

// ConnectWithContext 连接 NATS，支持 context 取消和超时
// ctx 用于整体超时控制；maxRetry 为单次连接失败后的重试次数
func ConnectWithContext(ctx context.Context, url string, maxRetry int) (*natsgo.Conn, error) {
	var nc *natsgo.Conn
	var err error

	for i := 0; i < maxRetry; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connect to NATS %s cancelled: %w", url, ctx.Err())
		default:
		}

		nc, err = natsgo.Connect(url,
			natsgo.Timeout(10*time.Second),
			natsgo.ReconnectWait(2*time.Second),
			natsgo.MaxReconnects(-1),
		)
		if err == nil {
			return nc, nil
		}

		delay := time.Duration(i+1) * time.Second
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connect to NATS %s cancelled: %w", url, ctx.Err())
		case <-time.After(delay):
		}
	}
	return nil, fmt.Errorf("connect to NATS %s failed after %d attempts: %w", url, maxRetry, err)
}

// Connect 连接 NATS（带重试），默认 2 分钟超时
func Connect(url string, maxRetry int) (*natsgo.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return ConnectWithContext(ctx, url, maxRetry)
}

// NewJetStream 创建 JetStream context 并确保 stream 存在
func NewJetStream(nc *natsgo.Conn, streamName string, subjects []string, maxAge time.Duration) (natsgo.JetStreamContext, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("create jetstream context: %w", err)
	}

	_, err = js.AddStream(&natsgo.StreamConfig{
		Name:     streamName,
		Subjects: subjects,
		Storage:  natsgo.FileStorage,
		MaxAge:   maxAge,
	})
	if err != nil {
		return js, nil // 流已存在，忽略
	}
	return js, nil
}

// Subscribe 创建 JetStream 持久订阅
func Subscribe(js natsgo.JetStreamContext, subject, durable string, handler natsgo.MsgHandler) (*natsgo.Subscription, error) {
	return js.Subscribe(subject, handler,
		natsgo.Durable(durable),
		natsgo.ManualAck(),
		natsgo.MaxDeliver(3),
	)
}

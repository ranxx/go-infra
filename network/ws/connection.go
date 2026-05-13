package ws

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ranxx/go-infra/logger"
	"github.com/ranxx/go-infra/network"
)

// Connection WebSocket 连接实现
type Connection struct {
	*network.BaseConnection
	conn    *websocket.Conn
	opts    *network.Options
	ctx     context.Context
	close   chan struct{}
	connMgr network.ConnectionManager
}

// NewConnection 创建新的 WebSocket 连接
func NewConnection(ctx context.Context, id string, conn *websocket.Conn, close chan struct{}, opts ...network.Option) *Connection {
	if close == nil {
		close = make(chan struct{})
	}

	options := network.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	baseConn := network.NewBaseConnection(id, options.MaxErrorCount)
	for k, v := range options.Metadata {
		baseConn.SetMetadata(k, v)
	}

	return &Connection{
		BaseConnection: baseConn,
		conn:           conn,
		opts:           &options,
		ctx:            ctx,
		close:          close,
	}
}

// GetRemoteAddr 获取远程地址
func (c *Connection) GetRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Send 发送消息
func (c *Connection) Send(data any) error {
	if c.IsClosed() {
		return fmt.Errorf("connection closed")
	}

	var msgBytes []byte
	var err error

	switch v := data.(type) {
	case []byte:
		msgBytes = v
	default:
		if c.opts != nil && c.opts.Coder != nil {
			msgBytes, err = c.opts.Coder.Marshal(data)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no coder configured")
		}
	}

	if err := c.conn.WriteMessage(websocket.BinaryMessage, msgBytes); err != nil {
		c.Close()
		return err
	}
	return nil
}

// Close 关闭连接
func (c *Connection) Close() error {
	if c.IsClosed() {
		return nil
	}

	c.MarkClosed()
	return c.conn.Close()
}

// SetConnectionManager 设置连接管理器
func (c *Connection) SetConnectionManager(mgr network.ConnectionManager) {
	c.connMgr = mgr
}

// StartHandling 开始处理连接的消息循环
func (c *Connection) StartHandling(ctx context.Context) {
	defer func() {
		if c.connMgr != nil {
			c.connMgr.RemoveConnection(c.GetID())
		}
		if c.opts != nil && c.opts.Handler != nil {
			c.opts.Handler.OnDisconnect(ctx, c, nil)
		}
		c.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.close:
			return
		default:
			if c.opts != nil && c.opts.ReadTimeout > 0 {
				c.conn.SetReadDeadline(time.Now().Add(c.opts.ReadTimeout))
			}
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
				) {
					return
				}

				errStr := err.Error()
				if strings.Contains(errStr, "use of closed network connection") ||
					strings.Contains(errStr, "connection reset by peer") ||
					strings.Contains(errStr, "broken pipe") {
					return
				}

				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					return
				}

				if c.AddErrorCount() {
					return
				}
				continue
			}

			var message any = data
			if c.opts != nil && c.opts.Coder != nil {
				msg, err := c.opts.Coder.Unmarshal(data)
				if err != nil {
					logger.GetFieldLogger().Errorf("消息解码失败: %v", err)
					if c.AddErrorCount() {
						return
					}
					continue
				}
				message = msg
			}

			if c.opts != nil && c.opts.Handler != nil {
				if err := c.opts.Handler.OnMessage(ctx, c, message); err != nil {
					if c.AddErrorCount() {
						return
					}
				}
			}
		}
	}
}

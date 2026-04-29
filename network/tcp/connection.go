package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/ranxx/go-infra/logger"
	"github.com/ranxx/go-infra/network"
)

// Connection TCP 连接实现
type Connection struct {
	*network.BaseConnection
	conn   net.Conn
	opts   *network.Options
	ctx    context.Context
	close  chan struct{}
	connMgr network.ConnectionManager
}

// NewConnection 创建新的 TCP 连接
func NewConnection(ctx context.Context, id string, conn net.Conn, close chan struct{}, opts ...network.Option) *Connection {
	if close == nil {
		close = make(chan struct{})
	}

	options := network.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &Connection{
		BaseConnection: network.NewBaseConnection(id, options.MaxErrorCount),
		conn:          conn,
		opts:          &options,
		ctx:           ctx,
		close:         close,
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

	if c.opts != nil && c.opts.Packer != nil {
		msgBytes, err = c.opts.Packer.Pack(msgBytes)
		if err != nil {
			return err
		}
	}

	_, err = c.conn.Write(msgBytes)
	if err != nil {
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

// StartHandling 开始处理连接的消息循环
func (c *Connection) StartHandling(ctx context.Context) {
	defer func() {
		if c.connMgr != nil {
			c.connMgr.RemoveConnection(c.GetID())
		}
		if handler := c.opts.Handler; handler != nil {
			handler.OnDisconnect(ctx, c, nil)
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
			data, err := c.readMessage()
			if err != nil {
				if err == io.EOF {
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

// readMessage 读取一条完整消息
func (c *Connection) readMessage() ([]byte, error) {
	maxMessageSize := int64(1024 * 1024)
	readTimeout := 60 * time.Second

	if c.opts != nil {
		if c.opts.MaxMessageSize > 0 {
			maxMessageSize = c.opts.MaxMessageSize
		}
		if c.opts.ReadTimeout > 0 {
			readTimeout = c.opts.ReadTimeout
		}
	}

	if readTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(readTimeout))
	}

	if c.opts == nil || c.opts.Packer == nil {
		buf := make([]byte, maxMessageSize)
		n, err := c.conn.Read(buf)
		if err != nil {
			return nil, err
		}
		return buf[:n], nil
	}

	headLen := c.opts.Packer.GetHeadLength()
	lengthBuf := make([]byte, headLen)
	if _, err := io.ReadFull(c.conn, lengthBuf); err != nil {
		return nil, err
	}

	length, err := c.opts.Packer.UnpackHead(lengthBuf)
	if err != nil {
		return nil, err
	}

	if length > uint32(maxMessageSize) {
		return nil, fmt.Errorf("message size %d exceeds limit %d", length, maxMessageSize)
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(c.conn, messageBuf)
	if err != nil {
		return nil, err
	}

	return messageBuf, nil
}

// SetConnectionManager 设置连接管理器
func (c *Connection) SetConnectionManager(mgr network.ConnectionManager) {
	c.connMgr = mgr
}

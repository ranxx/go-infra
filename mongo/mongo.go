package mongo

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ranxx/go-infra/proxy"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// mongoDialer 是 mongo-driver 需要的 ContextDialer 接口实现
type mongoDialer struct {
	proxy.Dialer
}

// DialContext 实现 mongo.ContextDialer 接口
func (d *mongoDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.Dial(network, addr)
}

// NewClient 创建 MongoDB 客户端，连接失败时返回 nil 但不报错
func NewClient(cfg *Config) (*mongo.Client, error) {
	// 如果 URI 为空，跳过连接
	if cfg.URI == "" {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	if cfg.Proxy {
		proxy.Wrap(func(dr proxy.Dialer) {
			clientOptions.SetDialer(&mongoDialer{Dialer: dr})
		})
	}
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	// Ping 检查连接
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return client, nil
}

// GetDatabase 根据名称获取数据库
func GetDatabase(client *mongo.Client, name string) *mongo.Database {
	if client == nil {
		return nil
	}
	return client.Database(name)
}

// NewMongoDatabase 创建 MongoDB 数据库实例
func NewMongoDatabase(client *mongo.Client, cfg *Config) *mongo.Database {
	return GetDatabase(client, cfg.Database)
}

var (
	globalMongoClient   *mongo.Client
	globalMongoDatabase *mongo.Database
	mongoOnce           sync.Once
)

// Init 初始化全局 MongoDB 连接（单例），应在服务启动时调用
func Init(cfg *Config) error {
	var initErr error
	mongoOnce.Do(func() {
		client, err := NewClient(cfg)
		if err != nil {
			initErr = err
			return
		}
		if client == nil {
			initErr = fmt.Errorf("mongo: failed to create client")
			return
		}
		globalMongoClient = client
		globalMongoDatabase = NewMongoDatabase(client, cfg)
	})
	return initErr
}

// Client 获取全局 MongoDB 客户端
func Client() *mongo.Client {
	return globalMongoClient
}

// Get 获取全局 MongoDB 数据库实例
func Get() *mongo.Database {
	return globalMongoDatabase
}

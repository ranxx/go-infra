package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ranxx/go-infra/elasticsearch"
	"github.com/ranxx/go-infra/logger"
	"github.com/ranxx/go-infra/mongo"
	"github.com/ranxx/go-infra/mysql"
	"github.com/ranxx/go-infra/redis"
	"gopkg.in/yaml.v3"
)

// Provider 定义配置提供者接口
type Provider interface {
	GetValue(key string) string
}

// Reloadable 定义支持热加载的接口
type Reloadable interface {
	Reload() error
}

// ApolloConfig Apollo 启动配置
type ApolloConfig struct {
	AppID     string `yaml:"appid"`
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
	IP        string `yaml:"ip"`
}

// ServerConfig 基础服务端口配置
type ServerConfig struct {
	Port int `json:"port"`
}

// Config 应用通用配置加载对象
type Config struct {
	Server        ServerConfig         `json:"server" yaml:"server"`
	GRPC          map[string]string    `json:"grpc" yaml:"grpc"`
	MySQL         mysql.Config         `json:"mysql" yaml:"mysql"`
	Redis         redis.Config         `json:"redis" yaml:"redis"`
	MongoDB       mongo.Config         `json:"mongodb" yaml:"mongodb"`
	Elasticsearch elasticsearch.Config `json:"elasticsearch" yaml:"elasticsearch"`
	Log           logger.Config        `json:"log" yaml:"log"`

	providers []Provider
	err       error
	mu        sync.RWMutex
	onReload  func(newConfig *Config)
}

// NewConfig 创建配置对象
func NewConfig(providers ...Provider) *Config {
	return &Config{
		providers: providers,
	}
}

// SetOnReload 设置配置重载时的回调函数
func (c *Config) SetOnReload(fn func(newConfig *Config)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onReload = fn
}

// TriggerReload 触发配置重载回调
func (c *Config) TriggerReload() {
	c.mu.RLock()
	fn := c.onReload
	c.mu.RUnlock()
	if fn != nil {
		fn(c)
	}
}

// LoadApolloConfig 加载本地 YAML 格式的 Apollo 启动参数
func LoadApolloConfig(env string) (*ApolloConfig, error) {
	if env == "" {
		env = "dev"
	}
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = filepath.Join("configs", fmt.Sprintf("app.%s.yaml", env))
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config file %s failed: %w", configPath, err)
	}
	var raw struct {
		Apollo ApolloConfig `yaml:"apollo"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}
	return &raw.Apollo, nil
}

// LoadByKey 从多个 Provider 按优先级加载配置（回退模式）
func LoadByKey(key string, ptr interface{}, providers ...Provider) error {
	if err := ApplyDefaults(ptr); err != nil {
		return fmt.Errorf("apply defaults for %s failed: %w", key, err)
	}

	for _, p := range providers {
		if p == nil {
			continue
		}
		content := p.GetValue(key)
		if content == "" {
			continue
		}
		if err := json.Unmarshal([]byte(content), ptr); err != nil {
			return fmt.Errorf("unmarshal %s failed: %w", key, err)
		}
		return nil
	}
	return nil
}

func (c *Config) getProviders(overrideP []Provider) []Provider {
	if len(overrideP) > 0 {
		return overrideP
	}
	return c.providers
}

// LoadGRPC 加载 gRPC 配置
func (c *Config) LoadGRPC(pps ...Provider) *Config {
	if c.err != nil {
		return c
	}
	c.err = LoadByKey("grpc", &c.GRPC, c.getProviders(pps)...)
	return c
}

// LoadMySQL 加载 MySQL 配置
func (c *Config) LoadMySQL(pps ...Provider) *Config {
	if c.err != nil {
		return c
	}
	c.err = LoadByKey("mysql", &c.MySQL, c.getProviders(pps)...)
	return c
}

// LoadRedis 加载 Redis 配置
func (c *Config) LoadRedis(pps ...Provider) *Config {
	if c.err != nil {
		return c
	}
	c.err = LoadByKey("redis", &c.Redis, c.getProviders(pps)...)
	return c
}

// LoadMongoDB 加载 MongoDB 配置
func (c *Config) LoadMongoDB(pps ...Provider) *Config {
	if c.err != nil {
		return c
	}
	c.err = LoadByKey("mongodb", &c.MongoDB, c.getProviders(pps)...)
	return c
}

// LoadElasticsearch 加载 Elasticsearch 配置
func (c *Config) LoadElasticsearch(pps ...Provider) *Config {
	if c.err != nil {
		return c
	}
	c.err = LoadByKey("elasticsearch", &c.Elasticsearch, c.getProviders(pps)...)
	return c
}

// LoadLog 加载日志配置
func (c *Config) LoadServer(pps ...Provider) *Config {
	if c.err != nil {
		return c
	}
	c.err = LoadByKey("server", &c.Server, c.getProviders(pps)...)
	return c
}

// LoadLog 加载日志配置
func (c *Config) LoadLog(pps ...Provider) *Config {
	if c.err != nil {
		return c
	}
	c.err = LoadByKey("log", &c.Log, c.getProviders(pps)...)
	return c
}

// LoadCustom 加载自定义配置项
func (c *Config) LoadCustom(key string, ptr interface{}, pps ...Provider) *Config {
	if c.err != nil {
		return c
	}
	c.err = LoadByKey(key, ptr, c.getProviders(pps)...)
	return c
}

func (c *Config) Error() error {
	return c.err
}

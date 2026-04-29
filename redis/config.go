package redis

// Config Redis 配置
type Config struct {
	Addr     string `json:"addr" yaml:"addr" default:"localhost:6379"` // Redis 地址
	Password string `json:"password" yaml:"password"`                  // Redis 密码
	DB       int    `json:"db" yaml:"db" default:"0"`                  // Redis 数据库
	Proxy    bool   `json:"proxy" yaml:"proxy"`
}

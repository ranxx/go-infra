package nats

// Config NATS 配置
type Config struct {
	URL    string       `json:"url" yaml:"url" default:"nats://localhost:4222"` // NATS 地址
	Prefix string       `json:"prefix" yaml:"prefix" default:"market"`          // 主题前缀
	Stream StreamConfig `json:"stream" yaml:"stream"`
}

// StreamConfig JetStream 流配置
type StreamConfig struct {
	Name     string   `json:"name" yaml:"name" default:"kline_events"` // JetStream 流名称
	Subjects []string `json:"subjects" yaml:"subjects"`                // 流订阅主题
	MaxAge   string   `json:"max_age" yaml:"max_age" default:"24h"`    // 消息最大保留时间
}

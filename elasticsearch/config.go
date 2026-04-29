package elasticsearch

// Config Elasticsearch 配置
type Config struct {
	URL                   string `json:"url" yaml:"url" default:"http://localhost:9200"` // ES 地址，支持逗号分隔的多个地址
	Index                 string `json:"index" yaml:"index"`
	RequestTimeoutSeconds int    `json:"request_timeout_seconds" yaml:"request_timeout_seconds" default:"5"`
	Proxy                 bool   `json:"proxy" yaml:"proxy" default:"false"`
}

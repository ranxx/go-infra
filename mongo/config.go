package mongo

// Config MongoDB 配置
type Config struct {
	URI      string `json:"uri" yaml:"uri" default:"mongodb://localhost:27017"` // MongoDB 连接 URI，例如 mongodb://user:pass@host:port/db
	Database string `json:"database" yaml:"database" default:"test"`            // 默认数据库名称
	DBPrefix string `json:"db_prefix" yaml:"db_prefix"`                         // 数据库名前缀，支持占位符 {env}，例如 "myapp_{env}"，在生产环境会替换为 "myapp_prod"
	Proxy    bool   `json:"proxy" yaml:"proxy"`
}

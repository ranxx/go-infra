package mysql

// Config MySQL 配置
type Config struct {
	DSN             string `json:"dsn" yaml:"dsn"`                                            // root:password@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local
	IdleConns       int    `json:"idle_conns" yaml:"idle_conns" default:"10"`                 // 空闲连接数
	MaxConns        int    `json:"max_conns" yaml:"max_conns" default:"100"`                  // 最大连接数
	MaxLifetime     int64  `json:"max_lifetime" yaml:"max_lifetime" default:"3600"`           // 连接最大生命周期 单位秒
	CreateBatchSize int    `json:"create_batch_size" yaml:"create_batch_size" default:"1000"` // 批量插入时每批次的大小
	Proxy           bool   `json:"proxy" yaml:"proxy"`
}

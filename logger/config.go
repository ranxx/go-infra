package logger

import "github.com/ranxx/go-infra/elasticsearch"

// Config 日志配置
type Config struct {
	Level       string `json:"level" yaml:"level" default:"debug"`                 // debug, info, warn, error
	Format      string `json:"format" yaml:"format" default:"text"`                // json, text
	ServiceName string `json:"service_name" yaml:"service_name" default:"unknown"` // 服务名称
}

// ElasticsearchConfig ES 配置（兼容旧 API，建议使用 elasticsearch.Config）
type ElasticsearchConfig = elasticsearch.Config

// Option 日志初始化选项
type Option struct {
	level       string
	format      string
	serviceName string
	esConfig    *ElasticsearchConfig
}

// Options 函数式选项类型
type Options func(*Option)

func defaultOptions() Option {
	return Option{
		level:       "debug",
		format:      "text",
		serviceName: "unknown",
	}
}

// WithOption 直接设置 Option（一次性设置所有字段）
func WithOption(opt Option) Options {
	return func(o *Option) {
		*o = opt
	}
}

// WithLevel 设置日志级别
func WithLevel(level string) Options {
	return func(o *Option) {
		o.level = level
	}
}

// WithFormat 设置日志格式
func WithFormat(format string) Options {
	return func(o *Option) {
		o.format = format
	}
}

// WithServiceName 设置服务名称
func WithServiceName(name string) Options {
	return func(o *Option) {
		o.serviceName = name
	}
}

// WithElasticConfig 设置 Elasticsearch 配置
func WithElasticConfig(cfg *ElasticsearchConfig) Options {
	return func(o *Option) {
		o.esConfig = cfg
	}
}

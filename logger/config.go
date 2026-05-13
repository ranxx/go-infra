package logger

import "github.com/ranxx/go-infra/elasticsearch"

// OutputMode 输出模式
type OutputMode int

const (
	OutputConsole OutputMode = iota // 仅控制台（默认）
	OutputFile                      // 仅文件
	OutputBoth                      // 控制台 + 文件
)

// Config 日志配置
type Config struct {
	Level       string `json:"level" yaml:"level" default:"debug"`                 // debug, info, warn, error
	Format      string `json:"format" yaml:"format" default:"text"`                // json, text
	ServiceName string `json:"service_name" yaml:"service_name" default:"unknown"` // 服务名称
	FilePath    string `json:"file_path" yaml:"file_path"`                        // 日志文件路径（空表示不写文件）
	OutputMode  OutputMode `json:"output_mode" yaml:"output_mode"`              // 输出模式
}

// ElasticsearchConfig ES 配置（兼容旧 API，建议使用 elasticsearch.Config）
type ElasticsearchConfig = elasticsearch.Config

// Option 日志初始化选项
type Option struct {
	level       string
	format      string
	serviceName string
	esConfig    *ElasticsearchConfig
	filePath    string     // 日志文件路径
	outputMode  OutputMode // 输出模式
}

// Options 函数式选项类型
type Options func(*Option)

func defaultOptions() Option {
	return Option{
		level:       "debug",
		format:      "text",
		serviceName: "unknown",
		outputMode:  OutputConsole,
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

// WithFilePath 设置日志文件路径
func WithFilePath(path string) Options {
	return func(o *Option) {
		o.filePath = path
	}
}

// WithOutputMode 设置输出模式（OutputConsole / OutputFile / OutputBoth）
func WithOutputMode(mode OutputMode) Options {
	return func(o *Option) {
		o.outputMode = mode
	}
}

// WithElasticConfig 设置 Elasticsearch 配置
func WithElasticConfig(cfg *ElasticsearchConfig) Options {
	return func(o *Option) {
		o.esConfig = cfg
	}
}

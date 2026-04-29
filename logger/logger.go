package logger

import (
	"os"

	"github.com/ranxx/go-infra/elasticsearch"

	"github.com/sirupsen/logrus"
)

var (
	logger      *logrus.Logger
	fieldLogger logrus.FieldLogger
)

// Init 初始化全局 logger，通过 options 传参
func Init(opts ...Options) *logrus.Logger {
	o := defaultOptions()

	for _, opt := range opts {
		opt(&o)
	}

	log := logrus.New()

	levelNum, err := logrus.ParseLevel(o.level)
	if err != nil {
		levelNum = logrus.InfoLevel
	}
	log.SetLevel(levelNum)

	if o.format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		log.SetFormatter(&TextFormatter{})
	}

	log.SetOutput(os.Stdout)

	// 若配置了 ES，则尝试初始化 ES Hook
	if o.esConfig != nil {
		hook, err := elasticsearch.NewHook(o.esConfig)
		if err != nil {
			log.WithField("service_name", o.serviceName).Warnf("failed to connect to Elasticsearch: %v, fallback to console logging", err)
		} else {
			log.AddHook(hook)
		}
	}

	logger = log
	fieldLogger = log.WithField("service_name", o.serviceName)

	return logger
}

// GetLogger 获取全局 logger
func GetLogger() *logrus.Logger {
	if logger == nil {
		return logrus.New()
	}
	return logger
}

// GetFieldLogger 获取全局 field logger（已包含 service_name）
func GetFieldLogger() logrus.FieldLogger {
	if fieldLogger != nil {
		return fieldLogger
	}
	return logrus.New().WithField("service_name", "unknown")
}

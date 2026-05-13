package logger

import (
	"io"
	"os"

	"github.com/ranxx/go-infra/elasticsearch"

	"github.com/sirupsen/logrus"
)

var (
	logger      *logrus.Logger
	fieldLogger logrus.FieldLogger
	logFile     *os.File // 持有文件句柄，供 Close 使用
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

	// 根据 OutputMode 决定写入目标
	writer, err := buildWriter(o)
	if err != nil {
		// 打开文件失败时回退到控制台并打印警告
		log.Warnf("logger: failed to open log file %q: %v, fallback to console", o.filePath, err)
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(writer)
	}

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

// buildWriter 根据配置构建 io.Writer
func buildWriter(o Option) (io.Writer, error) {
	switch o.outputMode {
	case OutputFile:
		if o.filePath == "" {
			return os.Stdout, nil // 没给路径，降级控制台
		}
		f, err := openLogFile(o.filePath)
		if err != nil {
			return nil, err
		}
		logFile = f
		return f, nil
	case OutputBoth:
		if o.filePath == "" {
			return os.Stdout, nil
		}
		f, err := openLogFile(o.filePath)
		if err != nil {
			return nil, err
		}
		logFile = f
		return io.MultiWriter(os.Stdout, f), nil
	default: // OutputConsole
		return os.Stdout, nil
	}
}

// openLogFile 以追加模式打开（或创建）日志文件
func openLogFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}

// Close 关闭日志文件（如有）
func Close() error {
	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
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

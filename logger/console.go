package logger

import (
	"os"
	"strings"

	"github.com/ranxx/go-infra/tracer"
	"github.com/sirupsen/logrus"
)

// TextFormatter 控制台友好的文本格式化器
type TextFormatter struct{}

// Format 实现 logrus.Formatter 接口
func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var timestamp string
	if t, ok := entry.Data["time"].(string); ok {
		timestamp = t
	} else {
		timestamp = entry.Time.Format("15:04:05.000")
	}

	level := entry.Level.String()
	msg := entry.Message

	// 基础输出
	output := timestamp + " [" + level + "] " + msg

	// 关键字段
	var fields []string
	if traceID, ok := entry.Data[tracer.GetTraceFieldName()].(string); ok && traceID != "" {
		fields = append(fields, tracer.GetTraceFieldName()+"="+traceID[:8])
	}
	if method, ok := entry.Data["method"].(string); ok {
		fields = append(fields, "method="+method)
	}
	if latency, ok := entry.Data["latency_ms"].(int64); ok {
		fields = append(fields, "latency="+itoa(int64(latency))+"ms")
	}
	if errMsg, ok := entry.Data["error"].(string); ok {
		fields = append(fields, "err="+truncate(errMsg, 80))
	}

	if len(fields) > 0 {
		output += " {" + join(fields, " ") + "}"
	}

	return []byte(output + "\n"), nil
}

// truncate 截断字符串
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// join 拼接字符串
func join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// itoa 简单的 int64 转字符串
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// SetupConsoleLogger 设置控制台日志
func SetupConsoleLogger(serviceName string, level string) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&TextFormatter{})

	return log
}

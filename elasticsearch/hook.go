package elasticsearch

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ranxx/go-infra/tracer"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

const (
	timeLayout      = "2006-01-02T15:04:05.000"
	timestampLayout = "2006-01-02T15:04:05.000Z0700"
)

// Hook 将日志写入 Elasticsearch。
type Hook struct {
	client       *elastic.Client
	indexPrefix  string
	levels       []logrus.Level
	ctxExtractor func(ctx context.Context) logrus.Fields
}

// NewHook 创建 Elasticsearch Hook。
func NewHook(cfg *Config) (*Hook, error) {
	client, err := NewRawClient(cfg)
	if err != nil {
		return nil, err
	}

	fieldName := tracer.GetTraceFieldName()
	return &Hook{
		client:      client,
		indexPrefix: cfg.Index,
		levels:      logrus.AllLevels,
		ctxExtractor: func(ctx context.Context) logrus.Fields {
			fields := logrus.Fields{}
			if traceID := tracer.GetTraceID(ctx); traceID != "" {
				fields[fieldName] = traceID
			}
			return fields
		},
	}, nil
}

// Levels 返回 Hook 生效级别。
func (hook *Hook) Levels() []logrus.Level {
	return hook.levels
}

// Fire 异步写入 Elasticsearch，避免阻塞业务流程。
func (hook *Hook) Fire(entry *logrus.Entry) error {
	doc := make(logrus.Fields, len(entry.Data)+6)
	for k, v := range entry.Data {
		doc[k] = v
	}

	loc := time.FixedZone("CST", 8*3600)
	tm := entry.Time.In(loc)
	doc["time"] = tm.Format(timeLayout)
	doc["@timestamp"] = tm.Format(timestampLayout)
	doc["level"] = entry.Level.String()
	doc["msg"] = entry.Message

	if entry.Context != nil {
		if fields := hook.ctxExtractor(entry.Context); fields != nil {
			for k, v := range fields {
				doc[k] = v
			}
		}
	}

	index := hook.indexPrefix + "-" + tm.Format("2006-01-02")

	go func(body logrus.Fields, targetIndex string) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := hook.client.Index().
			Index(targetIndex).
			BodyJson(body).
			Do(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ES Hook] failed to index log: %v\n", err)
		}
	}(doc, index)

	return nil
}

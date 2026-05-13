package topic

import (
	"encoding/json"
	"strings"

	"github.com/ranxx/go-infra/key"

	nats "github.com/nats-io/nats.go"
)

const (
	split = "."
)

// Register 注册
type Register interface {
	// 订阅
	// subj There are two type of wildcards: * for partial, and > for full.
	Subscribe(subj string, cb func(id int64, subj string, data []byte))
}

// Topic 分布式订阅发布
type Topic interface {
	// 订阅
	Subscribe(subj string, cb func(id int64, subj string, data []byte))
	// 发布消息
	Publish(id int64, subj string, data []byte) error
	// 发布消息
	PublishValue(id int64, subj string, val interface{}) error
}

type topic struct {
	ctl    *nats.Conn
	prefix string
	key    key.Keyer
}

type internalData struct {
	Id   int64
	Data []byte
}

type Options nats.Options
type Option func(*Options) error

var _topic_ *topic

// Init init topic
func Init(conf Config, key key.Keyer, options ...Option) Topic {
	if _topic_ != nil {
		return _topic_
	}
	conn, err := nats.Connect(conf.Host, func(o *nats.Options) error {
		for _, v := range options {
			if err := v((*Options)(o)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	_topic_ = &topic{ctl: conn, key: key, prefix: key.Key()}
	if len(_topic_.prefix) > 0 && !strings.HasSuffix(_topic_.prefix, split) {
		_topic_.prefix += split
	}
	return _topic_
}

// Get get topic
func Get() Topic {
	return _topic_
}

func (t *topic) Subscribe(subj string, cb func(id int64, subj string, data []byte)) {
	t.ctl.Subscribe(t.key.Key(subj), func(msg *nats.Msg) {
		_data := internalData{}
		json.Unmarshal(msg.Data, &_data)
		cb(_data.Id, strings.TrimPrefix(msg.Subject, t.prefix), _data.Data)
	})
}

func (t *topic) Publish(id int64, subj string, data []byte) error {
	_data := &internalData{
		Id:   id,
		Data: data,
	}
	data, _ = json.Marshal(_data)
	return t.ctl.Publish(t.key.Key(subj), data)
}

func (t *topic) PublishValue(id int64, subj string, val interface{}) error {
	data, _ := json.Marshal(val)
	_data := &internalData{
		Id:   id,
		Data: data,
	}
	data, _ = json.Marshal(_data)
	return t.ctl.Publish(t.key.Key(subj), data)
}

package grequests

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/ranxx/go-infra/logger"
	"github.com/sirupsen/logrus"

	"github.com/opentracing/opentracing-go"
)

// Encode ...
type Encode func(interface{}) ([]byte, error)

// Decode ...
type Decode func([]byte, interface{}) error

// DecodeWithCode ...
type DecodeWithCode func(int, []byte, interface{}) error

// HTTPDoer modules a upstream http client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Do ...
type Do func(req *http.Request) (*http.Response, error)

// Do ...
func (d Do) Do(req *http.Request) (*http.Response, error) {
	return d(req)
}

// RetryContinue ...
type RetryContinue func(i int, err error) bool

// Options 选项
type Options struct {
	OpenTracer             opentracing.Tracer
	Logger                 *logrus.Logger
	Retry                  uint           // default 1
	GlobalHeader           http.Header    // default Content-Type : application/json
	RequestEncode          Encode         // default json.Marshal
	ResponseDecode         Decode         // default json.Unmarshal
	ResponseDecodeWithCode DecodeWithCode // default json.Unmarshal
	Doer                   HTTPDoer       // http.Do
	RetryContinue          RetryContinue  // if retry > 1 effect, the return bool determines whether to continue
}

// DefaultOptions ...
func DefaultOptions() *Options {
	return &Options{
		OpenTracer:   nil,
		Logger:       logger.GetLogger(),
		Retry:        1,
		GlobalHeader: http.Header{"Content-Type": []string{"application/json"}},
		RequestEncode: func(i interface{}) ([]byte, error) {
			if i == nil {
				return []byte{}, nil
			}
			return json.Marshal(i)
		},
		ResponseDecode: func(b []byte, i interface{}) error {
			if i == nil {
				return nil
			}
			return json.Unmarshal(b, i)
		},
		Doer: http.DefaultClient,
		RetryContinue: func(i int, err error) bool {
			time.Sleep(time.Second * 1)
			return true
		},
	}
}

// Option ...
type Option func(*Options)

// WithOpenTracing 设置 tracer
func WithOpenTracing(tracer opentracing.Tracer) Option {
	return func(o *Options) {
		o.OpenTracer = tracer
	}
}

// WithRetry 设置 retry
func WithRetry(retry uint) Option {
	return func(o *Options) {
		o.Retry = retry
	}
}

// WithGlobalHeader 设置 header
func WithGlobalHeader(h http.Header) Option {
	return func(o *Options) {
		o.GlobalHeader = h
	}
}

// WithLogger 设置 logger
func WithLogger(l *logrus.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithRequestEncode 设置 request encode
func WithRequestEncode(encode Encode) Option {
	return func(o *Options) {
		o.RequestEncode = encode
	}
}

// WithResponseDecode 设置 response decode
func WithResponseDecode(decode Decode) Option {
	return func(o *Options) {
		o.ResponseDecode = decode
	}
}

// WithResponseDecodeWithCode 设置 response decode with status code
func WithResponseDecodeWithCode(decode DecodeWithCode) Option {
	return func(o *Options) {
		o.ResponseDecodeWithCode = decode
	}
}

// WithDoer http.do
func WithDoer(do HTTPDoer) Option {
	return func(o *Options) {
		o.Doer = do
	}
}

// WithRetryContinue ...
func WithRetryContinue(f RetryContinue) Option {
	return func(o *Options) {
		o.RetryContinue = f
	}
}

// Convert 转换
func Convert(v string, vv reflect.Value) {
	switch vv.Kind() {
	case reflect.String:
		vv.SetString(v)
	case reflect.Int:
		iii, _ := strconv.ParseInt(v, 10, 64)
		vv.SetInt(iii)
	case reflect.Uint8:
		iii, _ := strconv.ParseUint(v, 10, 64)
		vv.SetUint(iii)
	case reflect.Uint64:
		iii, _ := strconv.ParseUint(v, 10, 64)
		vv.SetUint(iii)
	default:
	}
}

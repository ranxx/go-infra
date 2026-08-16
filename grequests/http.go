package grequests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/opentracing/opentracing-go"
)

// HTTP ...
type HTTP struct {
	opt *Options
}

// NewHTTP new
func NewHTTP(opts ...Option) *HTTP {
	opt := DefaultOptions()
	for _, v := range opts {
		v(opt)
	}

	h := &HTTP{
		opt: opt,
	}
	return h
}

func (h *HTTP) do(ctx context.Context, method string, url string, header http.Header, req interface{}, out interface{}) error {
	// 合并header
	_header := make(http.Header)
	if header == nil {
		_header = h.opt.GlobalHeader
	} else {
		for k, vv := range h.opt.GlobalHeader {
			for _, v := range vv {
				_header.Add(k, v)
			}
		}
		for k, vv := range header {
			for _, v := range vv {
				_header.Add(k, v)
			}
		}
	}

	return h.retryDo(ctx, h.opt.Retry, method, url, _header, req, h.opt.RequestEncode, out, h.opt.ResponseDecode, h.opt.Doer, h.opt.OpenTracer, h.opt.Logger, h.opt.RetryContinue)
}

// Post post method
func (h *HTTP) Post(ctx context.Context, url string, header http.Header, req interface{}, out interface{}) error {
	return h.do(ctx, "POST", url, header, req, out)
}

// Put put method
func (h *HTTP) Put(ctx context.Context, url string, header http.Header, req interface{}, out interface{}) error {
	return h.do(ctx, "PUT", url, header, req, out)
}

// Get get method
func (h *HTTP) Get(ctx context.Context, url string, header http.Header, req interface{}, out interface{}) error {
	return h.do(ctx, "GET", url, header, req, out)
}

// Delete delete method
func (h *HTTP) Delete(ctx context.Context, url string, header http.Header, req interface{}, out interface{}) error {
	return h.do(ctx, "DELETE", url, header, req, out)
}

// HTTPDo custome method
func (h *HTTP) HTTPDo(ctx context.Context, method, url string, header http.Header, req interface{}, out interface{}) error {
	return h.do(ctx, method, url, header, req, out)
}

func defaultClient(tracer opentracing.Tracer) HTTPDoer {
	// if tracer != nil {
	// 	return clihttp.NewClient(tracer, clihttp.WithRequestLogThreshold(0), clihttp.WithResponseLogThreshold(0))
	// }
	return http.DefaultClient
}

func newRequest(ctx context.Context, logger *logrus.Logger, method string, url string, header http.Header, req interface{}, encode func(interface{}) ([]byte, error)) (*http.Request, error) {
	rdata, err := []byte{}, (error)(nil)
	switch req := req.(type) {
	case *[]byte:
		rdata = *req
	case []byte:
		rdata = req
	default:
		if encode != nil && req != nil {
			rdata, err = encode(req)
		}
	}
	if err != nil {
		return nil, err
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{"info": map[string]interface{}{"method": method, "url": url, "data": string(rdata), "header": fmt.Sprintf("%#v", header)}}).Info("requests")
	} else {
		log.Printf("method:%s url:%s header:%#v data:%s\n", method, url, header, string(rdata))
	}

	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(rdata))
	if err != nil {
		return nil, err
	}

	for k, values := range header {
		for _, v := range values {
			request.Header.Set(k, v)
		}
	}

	return request, nil
}

func (h *HTTP) do1(ctx context.Context, method, url string, header http.Header, req interface{}, encode Encode, out interface{}, decode Decode, doer HTTPDoer, tracer opentracing.Tracer, logger *logrus.Logger) error {
	if h.opt.ResponseDecodeWithCode != nil {
		return h.do2(ctx, method, url, header, req, encode, out, h.opt.ResponseDecodeWithCode, doer, tracer, logger)
	}

	request, err := newRequest(ctx, logger, method, url, header, req, encode)
	if err != nil {
		return err
	}

	if doer == nil {
		doer = defaultClient(tracer)
	}

	resp, err := doer.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if decode == nil {
		return nil
	}

	rdata, err := io.ReadAll(resp.Body)
	if err != nil {
		if logger != nil {
			logger.WithFields(logrus.Fields{"uri": request.URL.String(), "req": req, "resp": resp, "resp_body": string(rdata), "header": request.Header, "status": resp.StatusCode}).Error(err.Error())
		}
		return err
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{"uri": request.URL.String(), "req": req, "resp": resp, "resp_body": string(rdata), "header": request.Header, "status": resp.StatusCode}).Info("req finish")
	}
	return decode(rdata, out)
}

func (h *HTTP) do2(ctx context.Context, method, url string, header http.Header, req interface{}, encode Encode, out interface{}, decode DecodeWithCode, doer HTTPDoer, tracer opentracing.Tracer, logger *logrus.Logger) error {
	request, err := newRequest(ctx, logger, method, url, header, req, encode)
	if err != nil {
		return err
	}

	if doer == nil {
		doer = defaultClient(tracer)
	}

	resp, err := doer.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if decode == nil {
	}

	rdata, err := io.ReadAll(resp.Body)
	if err != nil {
		if logger != nil {
			logger.WithFields(logrus.Fields{"uri": request.URL.String(), "req": req, "resp": resp, "header": request.Header, "status": resp.StatusCode}).Error(err.Error())
		}
		return err
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{"uri": request.URL.String(), "req": req, "resp": resp, "header": request.Header, "status": resp.StatusCode}).Info("req finish")
	}
	return decode(resp.StatusCode, rdata, out)
}

func (h *HTTP) retryDo(ctx context.Context, retry uint, method string, url string, header http.Header, req interface{}, encode Encode, out interface{}, decode Decode, doer HTTPDoer, tracer opentracing.Tracer, logger *logrus.Logger, retryContinue RetryContinue) error {
	err := (error)(nil)
	for i := uint(0); i < retry; i++ {
		if err = h.do1(ctx, method, url, header, req, encode, out, decode, doer, tracer, logger); err == nil {
			return nil
		}
		if retry-i > 1 && !retryContinue(int(i), err) {
			return err
		}
	}
	return err
}

// Post ...
func Post(ctx context.Context, url string, header http.Header, req interface{}, out interface{}, opt ...Option) error {
	return NewHTTP(opt...).Post(ctx, url, header, req, out)
}

// Put ...
func Put(ctx context.Context, url string, header http.Header, req interface{}, out interface{}, opt ...Option) error {
	return NewHTTP(opt...).Put(ctx, url, header, req, out)
}

// Get ...
func Get(ctx context.Context, url string, header http.Header, req interface{}, out interface{}, opt ...Option) error {
	return NewHTTP(opt...).Get(ctx, url, header, req, out)
}

// Delete ...
func Delete(ctx context.Context, url string, header http.Header, req interface{}, out interface{}, opt ...Option) error {
	return NewHTTP(opt...).Delete(ctx, url, header, req, out)
}

// HTTPDo custome method
func HTTPDo(ctx context.Context, method, url string, header http.Header, req interface{}, out interface{}, opt ...Option) error {
	return NewHTTP(opt...).HTTPDo(ctx, method, url, header, req, out)
}

// EncodeURLParam ...
func EncodeURLParam(data map[string]interface{}) string {
	// query
	urlPath := "?"
	for key, value := range data {
		// 处理数组类型的值
		switch v := value.(type) {
		case []int64:
			for _, item := range v {
				urlPath = fmt.Sprintf("%s%s=%v&", urlPath, url.QueryEscape(key), url.QueryEscape(fmt.Sprintf("%v", item)))
			}
		case []string:
			for _, item := range v {
				urlPath = fmt.Sprintf("%s%s=%v&", urlPath, url.QueryEscape(key), url.QueryEscape(item))
			}
		case []interface{}:
			for _, item := range v {
				urlPath = fmt.Sprintf("%s%s=%s&", urlPath, url.QueryEscape(key), url.QueryEscape(fmt.Sprintf("%v", item)))
			}
		default:
			urlPath = fmt.Sprintf("%s%s=%s&", urlPath, url.QueryEscape(key), url.QueryEscape(fmt.Sprintf("%v", value)))
		}
	}

	urlPath = strings.TrimSuffix(strings.TrimSuffix(urlPath, "&"), "?")

	return urlPath
}

func JoinRequestParams(reqUrl string, reqData map[string]interface{}) string {
	urlPath := "?"
	for key, value := range reqData {
		// 处理数组类型的值
		switch v := value.(type) {
		case []int64:
			for _, item := range v {
				urlPath = fmt.Sprintf("%s%s=%v&", urlPath, url.QueryEscape(key), url.QueryEscape(fmt.Sprintf("%v", item)))
			}
		case []string:
			for _, item := range v {
				urlPath = fmt.Sprintf("%s%s=%s&", urlPath, url.QueryEscape(key), url.QueryEscape(item))
			}
		case []interface{}:
			for _, item := range v {
				urlPath = fmt.Sprintf("%s%s=%s&", urlPath, url.QueryEscape(key), url.QueryEscape(fmt.Sprintf("%v", item)))
			}
		default:
			urlPath = fmt.Sprintf("%s%s=%s&", urlPath, url.QueryEscape(key), url.QueryEscape(fmt.Sprintf("%v", value)))
		}
	}
	urlPath = strings.TrimSuffix(strings.TrimSuffix(urlPath, "&"), "?")
	return reqUrl + urlPath
}

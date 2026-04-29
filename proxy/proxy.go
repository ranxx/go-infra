package proxy

import (
	"net"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/proxy"
)

var (
	allProxyEnv = []string{"ALL_PROXY", "all_proxy"}
	noProxyEnv  = []string{"NO_PROXY", "no_proxy"}
)

// Dialer 是代理拨号器接口
type Dialer interface {
	// Dial 通过代理连接到指定地址
	Dial(network, addr string) (net.Conn, error)
}

// Wrap 从环境变量读取代理配置，并通过回调函数设置原始 dialer
// 如果未设置代理，回调不会执行
func Wrap(f func(dr Dialer)) bool {
	d := proxy.FromEnvironmentUsing(nil)
	if d == nil {
		return false
	}
	f(d)
	return true
}

// WrapHTTPProxy 从环境变量读取代理配置，并通过回调函数设置 HTTP 代理
// 用于 WebSocket 等需要 http.Proxy 格式的场景
// 如果未设置代理，回调不会执行
func WrapHTTPProxy(f func(*url.URL)) {
	allProxy := getAllProxyByEnv()
	if len(allProxy) == 0 {
		return
	}

	proxyURL, err := url.Parse(allProxy)
	if err != nil {
		return
	}

	f(proxyURL)
}

// NewHTTPProxyFunc 创建 http.ProxyFromRequest 格式的代理函数
func NewHTTPProxyFunc(proxyURL *url.URL) func(*http.Request) (*url.URL, error) {
	return func(*http.Request) (*url.URL, error) {
		return proxyURL, nil
	}
}

// WrapWithEnvKey 从环境变量读取代理配置，并通过回调函数设置
// 如果未设置代理，回调不会执行
func WrapWithEnvKey(envKey string, f func(dr Dialer)) {
	proxyURL := os.Getenv(envKey)
	if proxyURL == "" {
		return
	}

	pURL, err := url.Parse(proxyURL)
	if err != nil {
		return
	}

	socks5Dialer, err := proxy.FromURL(pURL, proxy.Direct)
	if err != nil {
		return
	}

	f(socks5Dialer)
}

func getAllProxyByEnv() string {
	for _, key := range allProxyEnv {
		if proxyURL := os.Getenv(key); proxyURL != "" {
			return proxyURL
		}
	}
	return ""
}

func getNoProxyByEnv() string {
	for _, key := range noProxyEnv {
		if proxyURL := os.Getenv(key); proxyURL != "" {
			return proxyURL
		}
	}
	return ""
}

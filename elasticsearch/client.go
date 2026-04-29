package elasticsearch

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ranxx/go-infra/proxy"

	"github.com/olivere/elastic/v7"
)

var (
	clientOnce sync.Once
	clientInst Client
)

// Client 是 Elasticsearch 通用客户端接口。
type Client interface {
	GetClient() *elastic.Client
	Index(ctx context.Context, index string, id string, doc interface{}) (*elastic.IndexResponse, error)
	Get(ctx context.Context, index string, id string) (*elastic.GetResult, error)
	Update(ctx context.Context, index string, id string, doc interface{}) (*elastic.UpdateResponse, error)
	Delete(ctx context.Context, index string, id string) (*elastic.DeleteResponse, error)
	Search(ctx context.Context, index string, query elastic.Query, from int, size int, sorters ...elastic.Sorter) (*elastic.SearchResult, error)
}

// ClientImpl 是 Client 的默认实现。
type ClientImpl struct {
	client       *elastic.Client
	defaultIndex string
}

// Init 初始化全局 ES 客户端（单例）。
func Init(cfg *Config) (Client, error) {
	var err error
	clientOnce.Do(func() {
		clientInst, err = NewClient(cfg)
	})
	return clientInst, err
}

// NewClient 创建新的 ES 客户端实例。
func NewClient(cfg *Config) (Client, error) {
	raw, err := NewRawClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ClientImpl{
		client:       raw,
		defaultIndex: cfg.Index,
	}, nil
}

// NewRawClient 创建底层 elastic.Client。
func NewRawClient(cfg *Config) (*elastic.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("elasticsearch config is nil")
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("elasticsearch url is empty")
	}

	httpClient := &http.Client{
		Timeout: resolveRequestTimeout(cfg),
	}

	if cfg.Proxy {
		httpClient.Transport = newTransport()
	}

	client, err := elastic.NewClient(
		elastic.SetURL(cfg.URL),
		elastic.SetSniff(false),
		elastic.SetHttpClient(httpClient),
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *ClientImpl) GetClient() *elastic.Client {
	return c.client
}

func (c *ClientImpl) Index(ctx context.Context, index string, id string, doc interface{}) (*elastic.IndexResponse, error) {
	idx, err := c.resolveIndex(index)
	if err != nil {
		return nil, err
	}
	service := c.client.Index().Index(idx).BodyJson(doc)
	if id != "" {
		service = service.Id(id)
	}
	return service.Do(withContext(ctx))
}

func (c *ClientImpl) Get(ctx context.Context, index string, id string) (*elastic.GetResult, error) {
	idx, err := c.resolveIndex(index)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, fmt.Errorf("document id is empty")
	}
	return c.client.Get().Index(idx).Id(id).Do(withContext(ctx))
}

func (c *ClientImpl) Update(ctx context.Context, index string, id string, doc interface{}) (*elastic.UpdateResponse, error) {
	idx, err := c.resolveIndex(index)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, fmt.Errorf("document id is empty")
	}
	return c.client.Update().Index(idx).Id(id).Doc(doc).Do(withContext(ctx))
}

func (c *ClientImpl) Delete(ctx context.Context, index string, id string) (*elastic.DeleteResponse, error) {
	idx, err := c.resolveIndex(index)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, fmt.Errorf("document id is empty")
	}
	return c.client.Delete().Index(idx).Id(id).Do(withContext(ctx))
}

func (c *ClientImpl) Search(ctx context.Context, index string, query elastic.Query, from int, size int, sorters ...elastic.Sorter) (*elastic.SearchResult, error) {
	idx, err := c.resolveIndex(index)
	if err != nil {
		return nil, err
	}
	if query == nil {
		query = elastic.NewMatchAllQuery()
	}
	service := c.client.Search().
		Index(idx).
		Query(query).
		From(from).
		Size(size)
	if len(sorters) > 0 {
		service = service.SortBy(sorters...)
	}
	return service.Do(withContext(ctx))
}

func (c *ClientImpl) resolveIndex(index string) (string, error) {
	if index != "" {
		return index, nil
	}
	if c.defaultIndex == "" {
		return "", fmt.Errorf("elasticsearch index is empty")
	}
	return c.defaultIndex, nil
}

func withContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}

func resolveRequestTimeout(cfg *Config) time.Duration {
	if cfg.RequestTimeoutSeconds <= 0 {
		return 5 * time.Second
	}
	return time.Duration(cfg.RequestTimeoutSeconds) * time.Second
}

func newTransport() *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var (
				conn net.Conn
				err  error
			)
			proxy.Wrap(func(dr proxy.Dialer) {
				conn, err = dr.Dial(network, addr)
			})
			if conn != nil || err != nil {
				return conn, err
			}
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}
}

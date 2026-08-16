package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Router HTTP 路由器，封装 Gin Engine 并提供 Swagger 文档支持
type Router struct {
	engine    *gin.Engine
	routes    []Route
	resources []Resource
	noRoute   HandlerFunc
	noMethod  HandlerFunc
}

// RouterOption 路由器选项
type RouterOption func(*Router)

// WithNoRoute 设置 404 处理器
func WithNoRoute(h HandlerFunc) RouterOption {
	return func(r *Router) { r.noRoute = h }
}

// WithNoMethod 设置 405 处理器
func WithNoMethod(h HandlerFunc) RouterOption {
	return func(r *Router) { r.noMethod = h }
}

// NewRouter 创建路由器
func NewRouter(opts ...RouterOption) *Router {
	r := &Router{
		engine: gin.New(),
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.noRoute != nil {
		r.engine.NoRoute(Wrap(r.noRoute))
	}
	if r.noMethod != nil {
		r.engine.NoMethod(Wrap(r.noMethod))
	}
	return r
}

// Engine 返回底层 Gin Engine（用于挂载中间件等）
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// Use 注册全局中间件
func (r *Router) Use(middleware ...gin.HandlerFunc) {
	r.engine.Use(middleware...)
}

// AddRoute 注册单条路由
func (r *Router) AddRoute(route Route) {
	handler := Wrap(route.Handler)
	switch route.Method {
	case "GET":
		r.engine.GET(route.Path, handler)
	case "POST":
		r.engine.POST(route.Path, handler)
	case "PUT":
		r.engine.PUT(route.Path, handler)
	case "DELETE":
		r.engine.DELETE(route.Path, handler)
	case "PATCH":
		r.engine.PATCH(route.Path, handler)
	case "HEAD":
		r.engine.HEAD(route.Path, handler)
	case "OPTIONS":
		r.engine.OPTIONS(route.Path, handler)
	default:
		r.engine.Handle(route.Method, route.Path, handler)
	}
	r.routes = append(r.routes, route)
}

// AddResource 注册资源组，自动拼接 Prefix + Route.Path
func (r *Router) AddResource(res Resource) {
	for i := range res.Routes {
		route := res.Routes[i]
		route.Path = res.Prefix + route.Path
		if len(route.Tags) == 0 {
			route.Tags = res.Tags
		}
		if route.Accept == "" {
			route.Accept = "json"
		}
		if route.Produce == "" {
			route.Produce = "json"
		}
		r.AddRoute(route)
	}
	r.resources = append(r.resources, res)
}

// GET 快捷注册 GET 路由
func (r *Router) GET(path string, handler HandlerFunc, opts ...RouteOption) {
	r.AddRoute(buildRoute("GET", path, handler, opts))
}

// POST 快捷注册 POST 路由
func (r *Router) POST(path string, handler HandlerFunc, opts ...RouteOption) {
	r.AddRoute(buildRoute("POST", path, handler, opts))
}

// PUT 快捷注册 PUT 路由
func (r *Router) PUT(path string, handler HandlerFunc, opts ...RouteOption) {
	r.AddRoute(buildRoute("PUT", path, handler, opts))
}

// DELETE 快捷注册 DELETE 路由
func (r *Router) DELETE(path string, handler HandlerFunc, opts ...RouteOption) {
	r.AddRoute(buildRoute("DELETE", path, handler, opts))
}

// PATCH 快捷注册 PATCH 路由
func (r *Router) PATCH(path string, handler HandlerFunc, opts ...RouteOption) {
	r.AddRoute(buildRoute("PATCH", path, handler, opts))
}

// RouteOption 路由选项函数
type RouteOption func(*Route)

// WithTags 设置 Swagger tags
func WithTags(tags ...string) RouteOption {
	return func(r *Route) { r.Tags = tags }
}

// WithSummary 设置摘要
func WithSummary(s string) RouteOption {
	return func(r *Route) { r.Summary = s }
}

// WithDescription 设置描述
func WithDescription(s string) RouteOption {
	return func(r *Route) { r.Description = s }
}

// WithParams 设置请求参数
func WithParams(params ...Param) RouteOption {
	return func(r *Route) { r.Params = params }
}

// WithSuccess 设置成功响应示例
func WithSuccess(statusCode int, desc string, schema interface{}) RouteOption {
	return func(r *Route) {
		r.Success = &ResponseExample{StatusCode: statusCode, Description: desc, Schema: schema}
	}
}

// WithSecurity 设置安全要求
func WithSecurity(security ...SecurityRequirement) RouteOption {
	return func(r *Route) { r.Security = security }
}

// WithDeprecated 标记为已废弃
func WithDeprecated() RouteOption {
	return func(r *Route) { r.Deprecated = true }
}

func buildRoute(method, path string, handler HandlerFunc, opts []RouteOption) Route {
	route := Route{
		Method:  method,
		Path:    path,
		Handler: handler,
		Accept:  "json",
		Produce: "json",
	}
	for _, opt := range opts {
		opt(&route)
	}
	return route
}

// Routes 返回已注册的所有路由（用于 Swagger 文档生成）
func (r *Router) Routes() []Route {
	return r.routes
}

// Run 启动 HTTP 服务
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

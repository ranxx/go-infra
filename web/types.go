// Package web 提供基于 Gin 的 HTTP 框架，支持统一响应格式、Swagger 文档自动生成。
//
// Handler 签名为 func(ctx *gin.Context) (interface{}, error)，
// 框架自动将返回值包装为 {code, msg, data} JSON 响应。
package web

import "github.com/gin-gonic/gin"

// HandlerFunc 统一 handler 签名：返回业务数据和 error。
// 框架自动将 data 包装到 ApiResponse.Data，error 转为错误响应。
type HandlerFunc func(ctx *gin.Context) (interface{}, error)

// ApiResponse 统一 JSON 响应格式
type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Route 定义一条 API 路由及其 Swagger 文档元数据
type Route struct {
	Method      string   // HTTP 方法：GET, POST, PUT, DELETE, PATCH
	Path        string   // 路径，如 /login、/:id
	Handler     HandlerFunc
	Tags        []string          // Swagger tags 分组
	Summary     string            // 简短摘要
	Description string            // 详细描述
	Accept      string            // 请求内容类型，默认 json
	Produce     string            // 响应内容类型，默认 json
	Params      []Param           // 请求参数
	Success     *ResponseExample  // 成功响应示例
	Security    []SecurityRequirement // 安全要求
	Deprecated  bool
}

// Param 描述一个请求参数（query, path, header, body, formData）
type Param struct {
	Name        string      // 参数名
	In          string      // 位置：query, path, header, body, formData
	Required    bool        // 是否必填
	Type        string      // 类型：string, integer, number, boolean, object, array
	Description string      // 描述
	Schema      interface{} // body 参数的 schema（传 struct 类型或实例）
}

// ResponseExample 描述一个响应
type ResponseExample struct {
	StatusCode  int         // HTTP 状态码
	Description string      // 描述
	Schema      interface{} // 响应数据类型（传 struct 类型或实例）
}

// SecurityRequirement Swagger 安全要求
type SecurityRequirement struct {
	Name   string   // 安全方案名，如 "ApiKeyAuth"
	Scopes []string // OAuth2 scope
}

// Resource 一组路由的集合，共享公共前缀和默认 tags
type Resource struct {
	Prefix      string   // 路径前缀，如 /api/v1/users
	Tags        []string // 默认 tags（Route 未指定 tags 时使用）
	Description string   // 资源组描述
	Routes      []Route  // 路由列表
}

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// SwaggerConfig Swagger 文档配置
type SwaggerConfig struct {
	Title       string   // API 标题
	Version     string   // API 版本
	Description string   // API 描述
	Servers     []string // 服务器地址列表
	BasePath    string   // Swagger UI 挂载路径，默认 /swagger/
}

// SwaggerSpec OpenAPI 3.0 规范结构
type SwaggerSpec struct {
	OpenAPI string              `json:"openapi"`
	Info    SwaggerInfo         `json:"info"`
	Servers []SwaggerServer     `json:"servers,omitempty"`
	Tags    []SwaggerTag        `json:"tags,omitempty"`
	Paths   map[string]PathItem `json:"paths"`
}

// SwaggerInfo 文档信息
type SwaggerInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// SwaggerServer 服务器信息
type SwaggerServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// SwaggerTag 标签
type SwaggerTag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// PathItem 路径项
type PathItem map[string]*Operation

// Operation 操作
type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
}

// Parameter 参数（OpenAPI 3.0）
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBody 请求体
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// Response 响应
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType 媒体类型
type MediaType struct {
	Schema  *Schema     `json:"schema,omitempty"`
	Example interface{} `json:"example,omitempty"`
}

// Schema JSON Schema 片段
type Schema struct {
	Type       string            `json:"type,omitempty"`
	Format     string            `json:"format,omitempty"`
	Ref        string            `json:"$ref,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Example    interface{}       `json:"example,omitempty"`
}

// typeRegistry 记录已注册的 schema 类型
var typeRegistry = make(map[string]reflect.Type)

// RegisterType 注册类型到 Swagger schemas（用于生成 $ref 引用）
func RegisterType(name string, t interface{}) {
	typeRegistry[name] = reflect.TypeOf(t)
}

// GenerateSpec 根据已注册路由生成 OpenAPI 3.0 规范
func (r *Router) GenerateSpec(cfg SwaggerConfig) *SwaggerSpec {
	if cfg.BasePath == "" {
		cfg.BasePath = "/swagger/"
	}

	spec := &SwaggerSpec{
		OpenAPI: "3.0.3",
		Info: SwaggerInfo{
			Title:       cfg.Title,
			Version:     cfg.Version,
			Description: cfg.Description,
		},
		Paths: make(map[string]PathItem),
	}

	// 服务器
	for _, s := range cfg.Servers {
		spec.Servers = append(spec.Servers, SwaggerServer{URL: s})
	}

	// 收集 tags
	tagSet := make(map[string]bool)
	for _, route := range r.routes {
		for _, t := range route.Tags {
			tagSet[t] = true
		}
	}
	for t := range tagSet {
		spec.Tags = append(spec.Tags, SwaggerTag{Name: t})
	}

	// 生成路径
	for _, route := range r.routes {
		path := swaggerPath(route.Path)
		if _, ok := spec.Paths[path]; !ok {
			spec.Paths[path] = make(PathItem)
		}
		spec.Paths[path][strings.ToLower(route.Method)] = buildOperation(route)
	}

	return spec
}

// swaggerPath 将 Gin 路径参数 :param 转为 Swagger 格式 {param}
func swaggerPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			parts[i] = "{" + p[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

// buildOperation 构造 OpenAPI Operation
func buildOperation(route Route) *Operation {
	opID := strings.ToLower(route.Method) + "_" + sanitizeOpID(route.Path)

	op := &Operation{
		Tags:        route.Tags,
		Summary:     route.Summary,
		Description: route.Description,
		OperationID: opID,
		Deprecated:  route.Deprecated,
		Responses:   make(map[string]Response),
	}

	// 参数
	for _, p := range route.Params {
		if p.In == "body" {
			op.RequestBody = buildRequestBody(p, route.Accept)
			continue
		}
		param := Parameter{
			Name:        p.Name,
			In:          p.In,
			Required:    p.Required,
			Description: p.Description,
			Schema:      buildParamSchema(p),
		}
		op.Parameters = append(op.Parameters, param)
	}

	// 从 path 中提取未在 Params 中声明的路径参数
	declaredParams := make(map[string]bool)
	for _, p := range op.Parameters {
		declaredParams[p.Name] = true
	}
	for _, part := range strings.Split(swaggerPath(route.Path), "/") {
		if strings.HasPrefix(part, "{") {
			name := part[1 : len(part)-1]
			if !declaredParams[name] {
				op.Parameters = append(op.Parameters, Parameter{
					Name:     name,
					In:       "path",
					Required: true,
					Schema:   &Schema{Type: "string"},
				})
			}
		}
	}

	// 成功响应
	if route.Success != nil {
		contentType := route.Produce
		if contentType == "" {
			contentType = "json"
		}
		mediaType := "application/" + contentType
		statusStr := fmt.Sprintf("%d", route.Success.StatusCode)

		resp := Response{
			Description: route.Success.Description,
		}

		schema := buildResponseSchema(route.Success.Schema)
		if schema != nil {
			// 包装在 ApiResponse 中
			resp.Content = map[string]MediaType{
				mediaType: {
					Schema: &Schema{
						Type: "object",
						Properties: map[string]Schema{
							"code": {Type: "integer", Example: 0},
							"msg":  {Type: "string", Example: "ok"},
							"data": *schema,
						},
					},
				},
			}
		}
		op.Responses[statusStr] = resp
	} else {
		// 默认 200 响应
		op.Responses["200"] = Response{Description: "成功返回"}
	}

	// 默认错误响应
	op.Responses["400"] = Response{Description: "请求参数错误"}
	op.Responses["500"] = Response{Description: "服务器内部错误"}

	// 安全要求
	for _, sec := range route.Security {
		op.Security = append(op.Security, map[string][]string{sec.Name: sec.Scopes})
	}

	return op
}

// buildParamSchema 根据 Param 构建 Schema
func buildParamSchema(p Param) *Schema {
	if p.Schema != nil {
		return schemaFromType(p.Schema)
	}
	if p.Type != "" {
		return &Schema{Type: p.Type}
	}
	return &Schema{Type: "string"}
}

// buildRequestBody 构建 RequestBody
func buildRequestBody(p Param, accept string) *RequestBody {
	if accept == "" {
		accept = "json"
	}

	schema := buildParamSchema(p)
	rb := &RequestBody{
		Description: p.Description,
		Required:    p.Required,
		Content: map[string]MediaType{
			"application/" + accept: {},
		},
	}
	if schema != nil {
		rb.Content["application/"+accept] = MediaType{Schema: schema}
	}
	return rb
}

// buildResponseSchema 根据类型构建响应 Schema
func buildResponseSchema(t interface{}) *Schema {
	if t == nil {
		return nil
	}
	return schemaFromType(t)
}

// schemaFromType 从 Go 类型反射生成 JSON Schema
func schemaFromType(t interface{}) *Schema {
	if t == nil {
		return nil
	}

	rt := reflect.TypeOf(t)

	// 如果是指针，取元素类型
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	// 特殊处理 map[string]interface{}
	if rt.Kind() == reflect.Map && rt.Key().Kind() == reflect.String {
		return &Schema{Type: "object"}
	}

	switch rt.Kind() {
	case reflect.Struct:
		return structToSchema(rt)
	case reflect.Slice, reflect.Array:
		elemSchema := schemaFromElemType(rt.Elem())
		return &Schema{
			Type:  "array",
			Items: elemSchema,
		}
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	default:
		return &Schema{Type: "object"}
	}
}

// schemaFromElemType 处理数组元素类型（包括指针类型）
func schemaFromElemType(rt reflect.Type) *Schema {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	switch rt.Kind() {
	case reflect.Struct:
		return structToSchema(rt)
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: schemaFromElemType(rt.Elem()),
		}
	default:
		return &Schema{Type: "object"}
	}
}

// structToSchema 将 Go struct 转为 JSON Schema
func structToSchema(rt reflect.Type) *Schema {
	s := &Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
	}

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}

		jsonName := jsonTagName(f)
		if jsonName == "-" {
			continue
		}

		fieldSchema := schemaFromType(reflect.New(f.Type).Elem().Interface())
		// 检查 json tag 中的 omitempty
		tag := f.Tag.Get("json")
		if strings.Contains(tag, ",string") {
			fieldSchema = &Schema{Type: "string"}
		}

		s.Properties[jsonName] = *fieldSchema

		// 检查 binding:"required"
		binding := f.Tag.Get("binding")
		if strings.Contains(binding, "required") {
			s.Required = append(s.Required, jsonName)
		}
	}

	return s
}

// jsonTagName 提取 json tag 中的字段名
func jsonTagName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	parts := strings.Split(tag, ",")
	if parts[0] == "" {
		return f.Name
	}
	return parts[0]
}

// sanitizeOpID 将路径转为合法的 operationId 片段
func sanitizeOpID(path string) string {
	s := strings.ReplaceAll(path, "/", "_")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.Trim(s, "_")
	return s
}

// ========================================
// Swagger JSON 与 UI 端点
// ========================================

// swaggerUIHTML Swagger UI 内嵌 HTML
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "%s",
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout",
        defaultModelsExpandDepth: -1,
        docExpansion: "list",
      });
    };
  </script>
</body>
</html>`

// ServeSwagger 注册 Swagger JSON 与 UI 端点
// specPath: Swagger JSON 的访问路径，如 "/swagger/doc.json"
// uiPath: Swagger UI 页面路径，如 "/swagger/index.html"
func (r *Router) ServeSwagger(specPath, uiPath string, cfg SwaggerConfig) {
	spec := r.GenerateSpec(cfg)
	specBytes, _ := json.Marshal(spec)

	// Swagger JSON 端点
	r.engine.GET(specPath, func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, string(specBytes))
	})

	// Swagger UI 页面
	html := fmt.Sprintf(swaggerUIHTML, specPath)
	r.engine.GET(uiPath, func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	})
}

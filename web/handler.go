package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Wrap 将 HandlerFunc 转为 gin.HandlerFunc。
// 自动处理：(data, nil) → 200 {code:0, msg:"ok", data:...}
//
//	(data, error) → 对应 HTTP 状态码 {code:..., msg:..., data:null}
func Wrap(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := h(c)
		if err != nil {
			statusCode, resp := ToApiResponse(err)
			c.JSON(statusCode, resp)
			return
		}
		c.JSON(http.StatusOK, ApiResponse{Code: 0, Msg: "ok", Data: data})
	}
}

// Success 快捷返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ApiResponse{Code: 0, Msg: "ok", Data: data})
}

// SuccessMsg 返回仅有消息的成功响应
func SuccessMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, ApiResponse{Code: 0, Msg: msg, Data: nil})
}

// Fail 快捷返回错误响应
func Fail(c *gin.Context, err error) {
	statusCode, resp := ToApiResponse(err)
	c.JSON(statusCode, resp)
}

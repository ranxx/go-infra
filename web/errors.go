package web

import (
	"fmt"
	"net/http"
)

// ErrCode 错误码
type ErrCode int

// 通用错误码
const (
	ErrBadRequest   ErrCode = 40000
	ErrUnauthorized ErrCode = 40100
	ErrForbidden    ErrCode = 40300
	ErrNotFound     ErrCode = 40400
	ErrConflict     ErrCode = 40900
	ErrInternal     ErrCode = 50000
	ErrUnavailable  ErrCode = 50300
)

// Error 带错误码和 HTTP 状态码的错误
type Error struct {
	Code       ErrCode `json:"code"`
	Msg        string  `json:"msg"`
	HTTPStatus int     `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

// NewError 创建业务错误
func NewError(code ErrCode, msg string) *Error {
	return &Error{
		Code:       code,
		Msg:        msg,
		HTTPStatus: httpStatusFromCode(code),
	}
}

// NewErrorWithStatus 创建带自定义 HTTP 状态码的错误
func NewErrorWithStatus(code ErrCode, httpStatus int, msg string) *Error {
	return &Error{
		Code:       code,
		Msg:        msg,
		HTTPStatus: httpStatus,
	}
}

// httpStatusFromCode 根据错误码段推断 HTTP 状态码
func httpStatusFromCode(code ErrCode) int {
	switch {
	case code >= 50000:
		return http.StatusInternalServerError
	case code >= 40000 && code < 40100:
		return http.StatusBadRequest
	case code >= 40100 && code < 40200:
		return http.StatusUnauthorized
	case code >= 40300 && code < 40400:
		return http.StatusForbidden
	case code >= 40400 && code < 40500:
		return http.StatusNotFound
	case code >= 40900 && code < 41000:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// ToApiResponse 将 error 转为 ApiResponse。非 *Error 类型按 500 内部错误处理。
func ToApiResponse(err error) (int, ApiResponse) {
	if e, ok := err.(*Error); ok {
		return e.HTTPStatus, ApiResponse{Code: int(e.Code), Msg: e.Msg, Data: nil}
	}
	return http.StatusInternalServerError, ApiResponse{Code: int(ErrInternal), Msg: err.Error(), Data: nil}
}

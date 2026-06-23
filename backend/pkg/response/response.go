// Package response 定义统一响应封包（ADR-009 §统一响应体）。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Body 统一响应体：{code, message, data, request_id}。
type Body struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
}

// requestID 从 context 中获取或临时生成 request_id。
func requestID(c *gin.Context) string {
	if v, ok := c.Get("request_id"); ok {
		if s, _ := v.(string); s != "" {
			return s
		}
	}
	return uuid.NewString()
}

// OK 返回成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:      0,
		Message:   "ok",
		Data:      data,
		RequestID: requestID(c),
	})
}

// RequestIDFromContext 从 context 中获取或临时生成 request_id（供中间件使用）。
func RequestIDFromContext(c *gin.Context) string {
	return requestID(c)
}

// Fail 返回业务失败响应（HTTP 状态保留 200，code 表达业务错误）。
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Body{
		Code:      code,
		Message:   msg,
		RequestID: requestID(c),
	})
}

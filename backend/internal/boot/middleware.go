package boot

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"student-system/pkg/response"
)

// requestIDMiddleware 为每个请求生成 / 透传 X-Request-ID（ADR-020）。
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Next()
	}
}

// corsMiddleware 简易 CORS 中间件（V1 开发期开放本地前端访问）。
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// utf8GuardMiddleware 守护所有 JSON 写入接口的 body 编码，杜绝中文/标点变 "?"。
//
// 背景（详见 .trae/skills/encoding-fix-zh/SKILL.md）：
//   - PowerShell 5 / Windows 老式 curl 默认按 GBK 把字符串塞进 HTTP body；
//   - 后端 c.ShouldBindJSON 按 UTF-8 反序列化 → 非法字节落入 string；
//   - 写入 SQLite TEXT 时这些非法字节被替换为 0x3F（'?'），形成不可逆脏数据。
//
// 策略：
//  1. 仅作用于 POST/PUT/PATCH 且 Content-Type 含 json 的请求；
//  2. 整体读取 body 后判断是否为合法 UTF-8；
//  3. 不合法时尝试用 GB18030 解码：
//     - 成功 → 重写 body 为 UTF-8 后放行（兼容老脚本）；
//     - 失败 → 返回 41000 错误，明确告知调用方使用 UTF-8。
//
// 性能影响：仅多一次 io.ReadAll；StudentHub 业务 body 体积都很小，可忽略。
func utf8GuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !needsBodyCheck(c.Request.Method) {
			c.Next()
			return
		}
		ct := c.GetHeader("Content-Type")
		if !strings.Contains(strings.ToLower(ct), "json") {
			c.Next()
			return
		}
		if c.Request.Body == nil {
			c.Next()
			return
		}
		raw, err := io.ReadAll(c.Request.Body)
		_ = c.Request.Body.Close()
		if err != nil {
			response.Fail(c, 41000, "读取请求体失败")
			c.Abort()
			return
		}
		if len(raw) == 0 {
			c.Request.Body = io.NopCloser(bytes.NewReader(raw))
			c.Next()
			return
		}
		if utf8.Valid(raw) {
			c.Request.Body = io.NopCloser(bytes.NewReader(raw))
			c.Request.ContentLength = int64(len(raw))
			c.Next()
			return
		}
		// 兼容 PowerShell 5 / GBK：尝试转换
		fixed, _, convErr := transform.Bytes(simplifiedchinese.GB18030.NewDecoder(), raw)
		if convErr != nil || !utf8.Valid(fixed) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":    41000,
				"message": "请求体不是合法 UTF-8 编码（请确保客户端以 UTF-8 发送 JSON；PowerShell 5 用户请改用 [Text.Encoding]::UTF8.GetBytes 或 PowerShell 7 / Postman）",
			})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(fixed))
		c.Request.ContentLength = int64(len(fixed))
		c.Next()
	}
}

func needsBodyCheck(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	}
	return false
}

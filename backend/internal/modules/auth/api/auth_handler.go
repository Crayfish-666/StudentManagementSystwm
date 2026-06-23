package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authjwt "student-system/internal/modules/auth/jwt"
	"student-system/internal/modules/auth/service"
	"student-system/pkg/response"
)

// 错误码段（docs/04 §3.2 + ADR-005 决策细化）。
const (
	codeRTMissing      = 1003 // 缺少 refresh_token / 未登录
	codeRTInvalid      = 1003 // refresh_token 解析失败
	codePasswordWeak   = 40002
	codeOldPwdWrong    = 40104
	codeRTRevoked      = 40103 // RT 已被吊销（黑名单命中 / token_version 失配 / 改密后）
	codeInternal       = 1500
)

// AuthHandler 认证接口处理器。
type AuthHandler struct {
	svc *service.AuthService
	jwt *authjwt.JWTManager
}

// NewAuthHandler 创建认证处理器。
func NewAuthHandler(svc *service.AuthService, jwt *authjwt.JWTManager) *AuthHandler {
	return &AuthHandler{svc: svc, jwt: jwt}
}

// Login 登录接口。POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "用户名和密码不能为空")
		return
	}

	pair, view, err := h.svc.Login(&req)
	if err != nil {
		response.Fail(c, 1001, err.Error())
		return
	}

	// 设置 refresh_token 为 HttpOnly Cookie（ADR-005）
	c.SetCookie(
		"refresh_token",
		pair.RefreshToken,
		7*24*3600, // 7 天
		"/api/v1/auth",
		"",
		false, // 生产环境应改为 true（HTTPS）
		true,  // HttpOnly
	)
	c.SetSameSite(http.SameSiteStrictMode)

	response.OK(c, gin.H{
		"access_token": pair.AccessToken,
		"token_type":   pair.TokenType,
		"expires_in":   pair.ExpiresIn,
		"user":         view,
	})
}

// Refresh 刷新 Token。POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	// 优先从 Cookie 取，其次从 body 取
	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.RefreshToken != "" {
			refreshToken = body.RefreshToken
		}
	}

	if refreshToken == "" {
		response.Fail(c, codeRTMissing, "缺少 refresh_token")
		return
	}

	pair, err := h.svc.Refresh(refreshToken)
	if err != nil {
		// 解析失败、jti 黑名单命中、token_version 失配 → 统一视为吊销
		// 区分错误码便于前端判定是否进入"强制登出"分支
		msg := err.Error()
		if isRevokedErr(msg) {
			response.Fail(c, codeRTRevoked, msg)
			return
		}
		response.Fail(c, codeRTInvalid, msg)
		return
	}

	// 更新 Cookie
	c.SetCookie(
		"refresh_token",
		pair.RefreshToken,
		7*24*3600,
		"/api/v1/auth",
		"",
		false,
		true,
	)

	response.OK(c, gin.H{
		"access_token": pair.AccessToken,
		"token_type":   pair.TokenType,
		"expires_in":   pair.ExpiresIn,
	})
}

// Logout 登出。POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 把当前 RT jti 加入黑名单
	if rt, _ := c.Cookie("refresh_token"); rt != "" {
		h.svc.Logout(rt)
	}

	// 清除 refresh_token Cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/api/v1/auth",
		"",
		false,
		true,
	)
	response.OK(c, gin.H{"message": "已登出"})
}

// Me 获取当前用户信息。GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		response.Fail(c, 1003, "未登录")
		return
	}

	view, err := h.svc.GetCurrentUser(uid.(int64))
	if err != nil {
		response.Fail(c, 1003, err.Error())
		return
	}

	response.OK(c, view)
}

// ChangePassword 修改当前用户密码。POST /api/v1/auth/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		response.Fail(c, 1003, "未登录")
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "旧密码和新密码不能为空")
		return
	}

	if err := h.svc.ChangePassword(uid.(int64), &req); err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "旧密码错误"):
			response.Fail(c, codeOldPwdWrong, msg)
		case strings.Contains(msg, "长度") || strings.Contains(msg, "字母") || strings.Contains(msg, "数字"):
			response.Fail(c, codePasswordWeak, msg)
		default:
			response.Fail(c, codeInternal, msg)
		}
		return
	}

	// 改密成功后，当前 RT 也一并吊销（防同设备旧 RT 继续可用）
	if rt, _ := c.Cookie("refresh_token"); rt != "" {
		h.svc.Logout(rt)
		c.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", false, true)
	}

	response.OK(c, gin.H{"message": "密码已更新，请重新登录"})
}

// RegisterRoutes 注册认证路由（公开部分，/me、/password 在 boot 中注册到受保护路由组）。
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
	}
}

// RegisterProtectedRoutes 注册需要登录态的认证路由。
func (h *AuthHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.GET("/me", h.Me)
		auth.POST("/password", h.ChangePassword)
	}
}

// isRevokedErr 判断是否属于 RT 吊销类错误。
func isRevokedErr(msg string) bool {
	return strings.Contains(msg, "已被吊销")
}

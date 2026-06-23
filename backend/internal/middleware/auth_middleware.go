package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	authjwt "student-system/internal/modules/auth/jwt"
	"student-system/pkg/cachex"
	"student-system/pkg/response"
)

// TokenVersionLookup 轻量查 user.token_version（中间件用）。
// 解耦于具体 repository，便于单测 mock。
type TokenVersionLookup interface {
	LookupTokenVersion(uid int64) (int, error)
}

// JWTAuth JWT 认证中间件（ADR-005）。
// 1) 解析 Bearer Token；
// 2) 校验 token_version（带 cachex 缓存 60s，避免每请求查 DB）；
// 3) 失配 → 40103 RT_REVOKED。
func JWTAuth(jwtManager *authjwt.JWTManager, lookup TokenVersionLookup) gin.HandlerFunc {
	// tv 缓存：key = uid (string)，value = int token_version，TTL 60s
	tvCache := cachex.New(2048, 60*time.Second)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Body{
				Code:      1003,
				Message:   "未登录",
				RequestID: response.RequestIDFromContext(c),
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Body{
				Code:      1003,
				Message:   "Token 格式错误",
				RequestID: response.RequestIDFromContext(c),
			})
			return
		}

		claims, err := jwtManager.ParseAccess(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Body{
				Code:      40101,
				Message:   "Token 失效，请重新登录",
				RequestID: response.RequestIDFromContext(c),
			})
			return
		}

		// 校验 token_version：先查缓存，miss 查 DB
		cached, ok := tvCache.Get(tvKey(claims.UID))
		var currentTV int
		if !ok {
			tv, lookupErr := lookup.LookupTokenVersion(claims.UID)
			if lookupErr != nil {
				abortRevoked(c, "账号不存在或已被吊销")
				return
			}
			tvCache.SetWithTTL(tvKey(claims.UID), tv, 60*time.Second)
			currentTV = tv
		} else {
			// 类型断言失败按 0 处理，触发下方"失配"分支
			if v, isInt := cached.(int); isInt {
				currentTV = v
			}
		}

		if currentTV != claims.TokenVersion {
			// 缓存里不是 int 或值失配 → 回源一次（避免冷启动零值误判）
			tv, lookupErr := lookup.LookupTokenVersion(claims.UID)
			if lookupErr != nil || tv != claims.TokenVersion {
				abortRevoked(c, "Token 已被吊销，请重新登录")
				return
			}
			tvCache.SetWithTTL(tvKey(claims.UID), tv, 60*time.Second)
		}

		// 将用户信息注入 context
		c.Set("uid", claims.UID)
		c.Set("user_name", claims.Name)
		c.Set("user_roles", claims.Roles)
		c.Set("claims", claims)
		c.Next()
	}
}

func tvKey(uid int64) string {
	return "uid:" + strconv.FormatInt(uid, 10)
}

func abortRevoked(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, response.Body{
		Code:      40103,
		Message:   msg,
		RequestID: response.RequestIDFromContext(c),
	})
}

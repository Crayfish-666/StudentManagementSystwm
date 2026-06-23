package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"student-system/pkg/response"
)

// RequireRoles 角色校验中间件：检查当前用户是否拥有指定角色之一。
func RequireRoles(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Body{
				Code:      1004,
				Message:   "无权限",
				RequestID: response.RequestIDFromContext(c),
			})
			return
		}

		rolesList, ok := userRoles.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Body{
				Code:      1004,
				Message:   "无权限",
				RequestID: response.RequestIDFromContext(c),
			})
			return
		}

		for _, r := range rolesList {
			if _, found := roleSet[r]; found {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, response.Body{
			Code:      1004,
			Message:   "无权限",
			RequestID: response.RequestIDFromContext(c),
		})
	}
}

// Authorize 资源+动作级权限校验中间件（S03 升级版）。
// 当前基于角色实现，后续可替换为 Casbin 引擎（ADR-006）。
// 校验逻辑：用户角色列表中任一角色在 policy 中对 (resource, action) 有权限则放行。
func Authorize(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Body{
				Code:      1004,
				Message:   "无权限",
				RequestID: response.RequestIDFromContext(c),
			})
			return
		}

		rolesList, ok := userRoles.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Body{
				Code:      1004,
				Message:   "无权限",
				RequestID: response.RequestIDFromContext(c),
			})
			return
		}

		if checkPolicy(rolesList, resource, action) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, response.Body{
			Code:      1004,
			Message:   "无权限",
			RequestID: response.RequestIDFromContext(c),
		})
	}
}

// policyEntry 策略条目：(role, resource, action) → allow。
type policyEntry struct {
	Role     string
	Resource string
	Action   string
}

// builtinPolicies 内置策略表。
// 后续可迁移至 Casbin + DB 存储（ADR-006）。
var builtinPolicies = []policyEntry{
	// 系统管理
	{"R-SY-ADMIN", "sys:dict", "read"},
	{"R-SY-ADMIN", "sys:dict", "create"},
	{"R-SY-ADMIN", "sys:dict", "update"},
	{"R-SY-ADMIN", "sys:dict", "delete"},
	{"R-SY-ADMIN", "sys:menu", "read"},
	{"R-SY-ADMIN", "sys:user", "read"},
	{"R-SY-ADMIN", "sys:user", "create"},
	{"R-SY-ADMIN", "sys:user", "update"},
	{"R-SY-ADMIN", "sys:user", "delete"},
	{"R-SY-ADMIN", "sys:role", "read"},
	{"R-SY-ADMIN", "sys:role", "create"},
	{"R-SY-ADMIN", "sys:role", "update"},
	{"R-SY-ADMIN", "sys:role", "delete"},
	// 校级管理员全部权限
	{"R-SY-ADMIN", "*", "*"},
	{"R-SY-LEAGUE", "*", "read"},
	{"R-SY-AFFAIRS", "*", "read"},
	// 字典读取对所有登录用户开放
	{"R-STU-NORM", "sys:dict", "read"},
	{"R-STU-LEAGUE", "sys:dict", "read"},
	{"R-STU-ASSOC", "sys:dict", "read"},
	{"R-STU-COMMUNITY", "sys:dict", "read"},
	{"R-COL-LEAGUE", "sys:dict", "read"},
	{"R-COL-COUN", "sys:dict", "read"},
	{"R-COL-TUTOR", "sys:dict", "read"},
	// 菜单读取对所有登录用户开放
	{"R-STU-NORM", "sys:menu", "read"},
	{"R-STU-LEAGUE", "sys:menu", "read"},
	{"R-STU-ASSOC", "sys:menu", "read"},
	{"R-STU-COMMUNITY", "sys:menu", "read"},
	{"R-COL-LEAGUE", "sys:menu", "read"},
	{"R-COL-COUN", "sys:menu", "read"},
	{"R-COL-TUTOR", "sys:menu", "read"},
	{"R-SY-LEAGUE", "sys:menu", "read"},
	{"R-SY-AFFAIRS", "sys:menu", "read"},
}

// checkPolicy 检查策略是否允许。
func checkPolicy(roles []string, resource, action string) bool {
	for _, role := range roles {
		for _, p := range builtinPolicies {
			if p.Role == role {
				if (p.Resource == "*" || p.Resource == resource) &&
					(p.Action == "*" || p.Action == action) {
					return true
				}
			}
		}
	}
	return false
}

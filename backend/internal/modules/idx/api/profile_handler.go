package api

import (
	"github.com/gin-gonic/gin"

	"student-system/internal/modules/idx/service"
	"student-system/pkg/response"
)

// ProfileHandler 学生画像接口处理器。
type ProfileHandler struct {
	svc *service.StudentService
}

// NewProfileHandler 创建画像处理器。
func NewProfileHandler(svc *service.StudentService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

// Me 获取当前登录学生的画像。GET /api/v1/idx/profile/me
func (h *ProfileHandler) Me(c *gin.Context) {
	// 从 JWT 中间件注入的 user context 获取 user_id
	userID, exists := c.Get("uid")
	if !exists {
		response.Fail(c, 1401, "未获取到用户信息")
		return
	}

	uid, ok := userID.(int64)
	if !ok {
		// 尝试 float64 转 int64（JSON 数字默认解析为 float64）
		if f, ok := userID.(float64); ok {
			uid = int64(f)
		} else {
			response.Fail(c, 1401, "用户 ID 类型错误")
			return
		}
	}

	student, err := h.svc.GetProfileByUserID(uid)
	if err != nil {
		response.Fail(c, 1404, "未找到关联的学生信息")
		return
	}

	response.OK(c, student)
}

// RegisterRoutes 注册画像路由。
func (h *ProfileHandler) RegisterRoutes(rg *gin.RouterGroup) {
	profile := rg.Group("/idx/profile")
	{
		profile.GET("/me", h.Me)
	}
}

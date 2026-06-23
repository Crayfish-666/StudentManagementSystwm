package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	sysservice "student-system/internal/modules/sys/service"
	"student-system/pkg/response"
)

// UserHandler 用户/角色管理 HTTP 接口。
// 对齐 SRD §5.2：
//   GET    /sys/users                       列表
//   GET    /sys/users/:id                   详情
//   POST   /sys/users                       新建
//   PUT    /sys/users/:id                   更新基本信息
//   DELETE /sys/users/:id                   软删
//   POST   /sys/users/:id/reset-password    重置密码
//   POST   /sys/users/:id/lock              锁定
//   POST   /sys/users/:id/unlock            解锁
//   POST   /sys/users/:id/disable           禁用
//   POST   /sys/users/:id/enable            启用
//   GET    /sys/roles                       角色列表
//   POST   /sys/users/:id/roles             分配角色（覆盖式）
//   DELETE /sys/users/:id/roles/:rid        撤销单个角色
type UserHandler struct {
	svc *sysservice.UserService
}

// NewUserHandler 创建处理器。
func NewUserHandler(svc *sysservice.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// operatorID 提取当前操作人 ID。
func operatorID(c *gin.Context) int64 {
	if v, ok := c.Get("uid"); ok {
		if id, ok := v.(int64); ok {
			return id
		}
		if id, ok := v.(int); ok {
			return int64(id)
		}
	}
	return 0
}

// handleBizError 统一处理业务错误。
func handleBizError(c *gin.Context, err error) {
	if be, ok := sysservice.AsBizError(err); ok {
		response.Fail(c, be.Code, be.Message)
		return
	}
	response.Fail(c, 1500, "服务异常: "+err.Error())
}

// ============ 用户管理 ============

// ListUsers 用户列表。GET /api/v1/sys/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.svc.List(sysservice.ListParams{
		Keyword:  c.Query("keyword"),
		Status:   c.Query("status"),
		RoleCode: c.Query("role_code"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, result)
}

// GetUser 用户详情。GET /api/v1/sys/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	user, err := h.svc.Get(id)
	if err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, user)
}

// CreateUser 新建用户。POST /api/v1/sys/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req sysservice.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}
	user, err := h.svc.Create(&req, operatorID(c))
	if err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, user)
}

// UpdateUser 更新用户基本信息。PUT /api/v1/sys/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	var req sysservice.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误: "+err.Error())
		return
	}
	user, err := h.svc.Update(id, &req, operatorID(c))
	if err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, user)
}

// DeleteUser 软删用户。DELETE /api/v1/sys/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	if err := h.svc.Delete(id, operatorID(c)); err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ResetPassword 重置密码。POST /api/v1/sys/users/:id/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.ResetPassword(id, req.NewPassword, operatorID(c)); err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

// LockUser 锁定用户。POST /api/v1/sys/users/:id/lock
func (h *UserHandler) LockUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	if err := h.svc.Lock(id, operatorID(c)); err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, gin.H{"id": id, "status": "locked"})
}

// UnlockUser 解锁用户。POST /api/v1/sys/users/:id/unlock
func (h *UserHandler) UnlockUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	if err := h.svc.Unlock(id, operatorID(c)); err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, gin.H{"id": id, "status": "active"})
}

// DisableUser 禁用用户。POST /api/v1/sys/users/:id/disable
func (h *UserHandler) DisableUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	if err := h.svc.Disable(id, operatorID(c)); err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, gin.H{"id": id, "status": "disabled"})
}

// EnableUser 启用用户。POST /api/v1/sys/users/:id/enable
func (h *UserHandler) EnableUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	if err := h.svc.Enable(id, operatorID(c)); err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, gin.H{"id": id, "status": "active"})
}

// ============ 角色管理 ============

// ListRoles 角色列表。GET /api/v1/sys/roles
func (h *UserHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles()
	if err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, gin.H{"items": roles, "total": len(roles)})
}

// AssignRoles 分配角色（覆盖式）。POST /api/v1/sys/users/:id/roles
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	var req struct {
		RoleIDs []int64 `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误: "+err.Error())
		return
	}
	user, err := h.svc.AssignRoles(id, req.RoleIDs, operatorID(c))
	if err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, user)
}

// RevokeRole 撤销单个角色。DELETE /api/v1/sys/users/:id/roles/:rid
func (h *UserHandler) RevokeRole(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的用户 ID")
		return
	}
	rid, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的角色 ID")
		return
	}
	user, err := h.svc.RevokeRole(uid, rid)
	if err != nil {
		handleBizError(c, err)
		return
	}
	response.OK(c, user)
}

// RegisterRoutes 注册用户/角色管理路由。
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	users := rg.Group("/sys/users")
	{
		users.GET("", adminOnly, h.ListUsers)
		users.POST("", adminOnly, h.CreateUser)
		users.GET("/:id", adminOnly, h.GetUser)
		users.PUT("/:id", adminOnly, h.UpdateUser)
		users.DELETE("/:id", adminOnly, h.DeleteUser)
		users.POST("/:id/reset-password", adminOnly, h.ResetPassword)
		users.POST("/:id/lock", adminOnly, h.LockUser)
		users.POST("/:id/unlock", adminOnly, h.UnlockUser)
		users.POST("/:id/disable", adminOnly, h.DisableUser)
		users.POST("/:id/enable", adminOnly, h.EnableUser)
		users.POST("/:id/roles", adminOnly, h.AssignRoles)
		users.DELETE("/:id/roles/:rid", adminOnly, h.RevokeRole)
	}

	roles := rg.Group("/sys/roles")
	{
		roles.GET("", adminOnly, h.ListRoles)
	}
}

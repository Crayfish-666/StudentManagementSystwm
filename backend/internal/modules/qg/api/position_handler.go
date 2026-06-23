package api

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/qg/service"
	"student-system/pkg/response"
)

// PositionHandler 岗位+申请接口处理器。
type PositionHandler struct {
	svc *service.PositionService
}

// NewPositionHandler 创建岗位处理器。
func NewPositionHandler(svc *service.PositionService) *PositionHandler {
	return &PositionHandler{svc: svc}
}

// List 分页查询岗位列表。GET /api/v1/qg/positions
func (h *PositionHandler) List(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	deptType := c.Query("dept_type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(keyword, deptType, status, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询岗位列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取岗位详情。GET /api/v1/qg/positions/:id
func (h *PositionHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的岗位 ID")
		return
	}

	pos, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, pos)
}

// Create 创建岗位。POST /api/v1/qg/positions
func (h *PositionHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	pos, err := h.svc.Create(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, pos)
}

// Submit 提交岗位审批。POST /api/v1/qg/positions/:id/submit
func (h *PositionHandler) Submit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的岗位 ID")
		return
	}

	pos, err := h.svc.Submit(id)
	if err != nil {
		msg := err.Error()
		code := 1409
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, pos)
}

// Approve 审批通过岗位。POST /api/v1/qg/positions/:id/approve
func (h *PositionHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的岗位 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req struct {
		Level string `json:"level" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	pos, err := h.svc.Approve(id, userID, req.Level)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "岗位不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, pos)
}

// Reject 审批驳回岗位。POST /api/v1/qg/positions/:id/reject
func (h *PositionHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的岗位 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req struct {
		Opinion string `json:"opinion"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	pos, err := h.svc.Reject(id, userID, req.Opinion)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "岗位不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, pos)
}

// Delete 删除岗位。DELETE /api/v1/qg/positions/:id
func (h *PositionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的岗位 ID")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// Apply 学生申请岗位。POST /api/v1/qg/applies
func (h *PositionHandler) Apply(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.ApplyPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	// 从上下文获取 studentID（需要通过中间件或查询获取）
	var studentID int64
	if v := c.Query("student_id"); v != "" {
		studentID, _ = strconv.ParseInt(v, 10, 64)
	}

	apply, err := h.svc.Apply(userID, studentID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "必须先完成困难认定才能申请岗位" || msg == "困难认定等级为'不困难'，无法申请岗位" {
			code = 40301
		} else if msg == "岗位不存在" {
			code = 1404
		} else if msg == "当前状态不允许申请" {
			code = 1409
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, apply)
}

// AcceptApply 录用申请。POST /api/v1/qg/applies/:id/accept
func (h *PositionHandler) AcceptApply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	apply, err := h.svc.AcceptApply(id, userID)
	if err != nil {
		msg := err.Error()
		code := 1409
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, apply)
}

// ConfirmApply 学生确认录用。POST /api/v1/qg/applies/:id/confirm
func (h *PositionHandler) ConfirmApply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	apply, err := h.svc.ConfirmApply(id)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "申请记录不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, apply)
}

// Onboard 上岗。POST /api/v1/qg/applies/:id/onboard
func (h *PositionHandler) Onboard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	apply, err := h.svc.Onboard(id)
	if err != nil {
		msg := err.Error()
		code := 1409
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, apply)
}

// GetApply 获取申请详情。GET /api/v1/qg/applies/:id
func (h *PositionHandler) GetApply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	apply, err := h.svc.GetApply(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, apply)
}

// RegisterRoutes 注册岗位相关路由。
func (h *PositionHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	qg := rg.Group("/qg")
	{
		// 岗位管理
		qg.GET("/positions", h.List)
		qg.GET("/positions/:id", h.Get)
		qg.POST("/positions", h.Create)
		qg.POST("/positions/:id/submit", h.Submit)
		qg.POST("/positions/:id/approve", h.Approve)
		qg.POST("/positions/:id/reject", h.Reject)
		qg.DELETE("/positions/:id", h.Delete)

		// 岗位申请
		qg.POST("/applies", h.Apply)
		qg.GET("/applies/:id", h.GetApply)
		qg.POST("/applies/:id/accept", h.AcceptApply)
		qg.POST("/applies/:id/confirm", h.ConfirmApply)
		qg.POST("/applies/:id/onboard", h.Onboard)
	}
}

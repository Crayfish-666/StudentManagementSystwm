package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/ty/service"
	"student-system/pkg/response"
)

// RosterHandler 团员花名册接口处理器。
type RosterHandler struct {
	svc *service.RosterService
}

// NewRosterHandler 创建团员花名册处理器。
func NewRosterHandler(svc *service.RosterService) *RosterHandler {
	return &RosterHandler{svc: svc}
}

// List 列表查询团员。GET /api/v1/ty/members
func (h *RosterHandler) List(c *gin.Context) {
	var branchID int64
	if v := c.Query("branch_id"); v != "" {
		branchID, _ = strconv.ParseInt(v, 10, 64)
	}
	status := c.Query("status")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(branchID, status, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询团员列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取团员详情。GET /api/v1/ty/members/:id
func (h *RosterHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的团员 ID")
		return
	}

	member, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, "团员记录不存在")
		return
	}
	response.OK(c, member)
}

// Update 编辑团员信息。PATCH /api/v1/ty/members/:id
func (h *RosterHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的团员 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	member, err := h.svc.Update(id, &req, userID)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "团员记录不存在":
			code = 1404
		case "团员证号已存在，错误码:2640":
			code = 2640
		case "仅在册团员可执行该操作":
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, member)
}

// TransferOut 团员转出。POST /api/v1/ty/members/:id/transfer-out
func (h *RosterHandler) TransferOut(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的团员 ID")
		return
	}

	userID, name, role, ip, ua := actorFromCtx(c)

	var req service.TransferOutRequest
	_ = c.ShouldBindJSON(&req)

	member, err := h.svc.TransferOut(id, userID, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "团员记录不存在":
			code = 1404
		case "仅在册团员可执行转出操作":
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, member)
}

// Overtime 超龄离团。POST /api/v1/ty/members/:id/overtime
func (h *RosterHandler) Overtime(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的团员 ID")
		return
	}

	userID, name, role, ip, ua := actorFromCtx(c)

	member, err := h.svc.Overtime(id, userID, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "团员记录不存在":
			code = 1404
		case "仅在册团员可执行超龄离团操作":
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, member)
}

// Archive 归档（保留5年）。POST /api/v1/ty/members/:id/archive
func (h *RosterHandler) Archive(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的团员 ID")
		return
	}

	userID, name, role, ip, ua := actorFromCtx(c)

	member, err := h.svc.Archive(id, userID, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "团员记录不存在":
			code = 1404
		case "该团员已归档":
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, member)
}

// CountByStatus 按状态统计团员数量。GET /api/v1/ty/members/count
func (h *RosterHandler) CountByStatus(c *gin.Context) {
	status := c.Query("status")

	count, err := h.svc.CountByStatus(status)
	if err != nil {
		response.Fail(c, 1500, "统计失败")
		return
	}
	response.OK(c, gin.H{"count": count})
}

// RegisterRoutes 注册团员花名册相关路由。
func (h *RosterHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	ty := rg.Group("/ty")
	{
		// 团员花名册 CRUD + 特殊操作
		ty.GET("/members", h.List)
		ty.GET("/members/:id", h.Get)
		ty.PATCH("/members/:id", h.Update)
		ty.GET("/members/count", h.CountByStatus)

		// 特殊状态变更操作
		ty.POST("/members/:id/transfer-out", h.TransferOut)
		ty.POST("/members/:id/overtime", h.Overtime)
		ty.POST("/members/:id/archive", h.Archive)
	}
}

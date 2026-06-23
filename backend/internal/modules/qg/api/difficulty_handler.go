package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/qg/service"
	"student-system/pkg/response"
)

// DifficultyHandler 困难认定接口处理器。
type DifficultyHandler struct {
	svc *service.DifficultyService
}

// NewDifficultyHandler 创建困难认定处理器。
func NewDifficultyHandler(svc *service.DifficultyService) *DifficultyHandler {
	return &DifficultyHandler{svc: svc}
}

// List 分页查询困难认定列表。GET /api/v1/qg/difficulty-certs
func (h *DifficultyHandler) List(c *gin.Context) {
	level := c.Query("level")
	status := c.Query("status")
	var studentID int64
	if v := c.Query("student_id"); v != "" {
		studentID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(level, status, studentID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询困难认定列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取困难认定详情。GET /api/v1/qg/difficulty-certs/:id
func (h *DifficultyHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的认定 ID")
		return
	}

	cert, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, cert)
}

// Create 创建困难认定。POST /api/v1/qg/difficulty-certs
func (h *DifficultyHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateDifficultyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	cert, err := h.svc.Create(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, cert)
}

// Submit 提交困难认定审批。POST /api/v1/qg/difficulty-certs/:id/submit
func (h *DifficultyHandler) Submit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的认定 ID")
		return
	}

	cert, err := h.svc.Submit(id)
	if err != nil {
		msg := err.Error()
		code := 1409
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, cert)
}

// Approve 审批通过困难认定。POST /api/v1/qg/difficulty-certs/:id/approve
func (h *DifficultyHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的认定 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.ApproveDifficultyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	cert, err := h.svc.Approve(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1409
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, cert)
}

// Reject 审批驳回困难认定。POST /api/v1/qg/difficulty-certs/:id/reject
func (h *DifficultyHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的认定 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.ApproveDifficultyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	cert, err := h.svc.Reject(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1409
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, cert)
}

// Delete 删除困难认定。DELETE /api/v1/qg/difficulty-certs/:id
func (h *DifficultyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的认定 ID")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// RegisterRoutes 注册困难认定相关路由。
func (h *DifficultyHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	qg := rg.Group("/qg")
	{
		qg.GET("/difficulty-certs", h.List)
		qg.GET("/difficulty-certs/:id", h.Get)
		qg.POST("/difficulty-certs", h.Create)
		qg.POST("/difficulty-certs/:id/submit", h.Submit)
		qg.POST("/difficulty-certs/:id/approve", h.Approve)
		qg.POST("/difficulty-certs/:id/reject", h.Reject)
		qg.DELETE("/difficulty-certs/:id", h.Delete)
	}
}

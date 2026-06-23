package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/qg/service"
	"student-system/pkg/response"
)

// AssessmentHandler 考核+薪酬接口处理器。
type AssessmentHandler struct {
	svc *service.AssessmentService
}

// NewAssessmentHandler 创建考核薪酬处理器。
func NewAssessmentHandler(svc *service.AssessmentService) *AssessmentHandler {
	return &AssessmentHandler{svc: svc}
}

// CreateAssessment 创建月度考核。POST /api/v1/qg/monthly-assessments
func (h *AssessmentHandler) CreateAssessment(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	assess, err := h.svc.CreateAssessment(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "考核已存在" {
			code = 40905
		} else if msg == "申请不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, assess)
}

// ListAssess 查询月度考核列表。GET /api/v1/qg/monthly-assessments
func (h *AssessmentHandler) ListAssess(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))
	var applyID int64
	if v := c.Query("apply_id"); v != "" {
		applyID, _ = strconv.ParseInt(v, 10, 64)
	}
	positionTitle := c.Query("position_title")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListAssess(year, month, applyID, positionTitle, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询月度考核列表失败")
		return
	}
	response.OK(c, result)
}

// GetAssess 获取月度考核详情。GET /api/v1/qg/monthly-assessments/:id
func (h *AssessmentHandler) GetAssess(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的考核 ID")
		return
	}

	assess, err := h.svc.GetAssess(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, assess)
}

// ComputePayroll 计算薪酬。POST /api/v1/qg/payrolls/compute
func (h *AssessmentHandler) ComputePayroll(c *gin.Context) {
	var req struct {
		ApplyID int64 `json:"apply_id" binding:"required"`
		Year    int   `json:"year" binding:"required"`
		Month   int   `json:"month" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	payroll, err := h.svc.ComputePayroll(req.ApplyID, req.Year, req.Month)
	if err != nil {
		msg := err.Error()
		code := 1500
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, payroll)
}

// ListPayroll 查询薪酬列表。GET /api/v1/qg/payrolls
func (h *AssessmentHandler) ListPayroll(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))
	status := c.Query("status")
	positionTitle := c.Query("position_title")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListPayroll(year, month, status, positionTitle, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询薪酬列表失败")
		return
	}
	response.OK(c, result)
}

// GetPayroll 获取薪酬详情。GET /api/v1/qg/payrolls/:id
func (h *AssessmentHandler) GetPayroll(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的薪酬 ID")
		return
	}

	payroll, err := h.svc.GetPayroll(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, payroll)
}

// ReviewPayroll 审核薪酬。POST /api/v1/qg/payrolls/:id/review
func (h *AssessmentHandler) ReviewPayroll(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的薪酬 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	payroll, err := h.svc.ReviewPayroll(id, userID)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "薪酬不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, payroll)
}

// PayPayroll 发放薪酬。POST /api/v1/qg/payrolls/:id/pay
func (h *AssessmentHandler) PayPayroll(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的薪酬 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	payroll, err := h.svc.PayPayroll(id, userID)
	if err != nil {
		msg := err.Error()
		code := 1409
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, payroll)
}

// RegisterRoutes 注册考核薪酬相关路由。
func (h *AssessmentHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	qg := rg.Group("/qg")
	{
		// 月度考核
		qg.POST("/monthly-assessments", h.CreateAssessment)
		qg.GET("/monthly-assessments", h.ListAssess)
		qg.GET("/monthly-assessments/:id", h.GetAssess)

		// 薪酬
		qg.POST("/payrolls/compute", h.ComputePayroll)
		qg.GET("/payrolls", h.ListPayroll)
		qg.GET("/payrolls/:id", h.GetPayroll)
		qg.POST("/payrolls/:id/review", h.ReviewPayroll)
		qg.POST("/payrolls/:id/pay", h.PayPayroll)
	}
}

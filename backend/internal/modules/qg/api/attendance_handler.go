package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/qg/service"
	"student-system/pkg/response"
)

// AttendanceHandler 工时打卡接口处理器。
type AttendanceHandler struct {
	svc *service.AttendanceService
}

// NewAttendanceHandler 创建工时打卡处理器。
func NewAttendanceHandler(svc *service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

// List 查询打卡记录列表。GET /api/v1/qg/attendances
func (h *AttendanceHandler) List(c *gin.Context) {
	var applyID, studentID int64
	if v := c.Query("apply_id"); v != "" {
		applyID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := c.Query("student_id"); v != "" {
		studentID, _ = strconv.ParseInt(v, 10, 64)
	}
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	positionTitle := c.Query("position_title")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(applyID, studentID, positionTitle, dateFrom, dateTo, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询打卡记录失败")
		return
	}
	response.OK(c, result)
}

// ClockIn 上班打卡。POST /api/v1/qg/attendances/clock-in
func (h *AttendanceHandler) ClockIn(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.ClockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	// 从查询参数获取 studentID
	var studentID int64
	if v := c.Query("student_id"); v != "" {
		studentID, _ = strconv.ParseInt(v, 10, 64)
	}

	attend, err := h.svc.ClockIn(userID, studentID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, attend)
}

// ClockOut 下班打卡。POST /api/v1/qg/attendances/:id/clock-out
func (h *AttendanceHandler) ClockOut(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的打卡记录 ID")
		return
	}

	attend, err := h.svc.ClockOut(id)
	if err != nil {
		msg := err.Error()
		code := 1409
		if bizErr, ok := err.(*service.BizError); ok {
			code = bizErr.Code
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, attend)
}

// MonthlySummary 月度工时汇总。GET /api/v1/qg/attendances/monthly-summary
func (h *AttendanceHandler) MonthlySummary(c *gin.Context) {
	var studentID int64
	if v := c.Query("student_id"); v != "" {
		studentID, _ = strconv.ParseInt(v, 10, 64)
	}
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))

	result, err := h.svc.MonthlySummary(studentID, year, month)
	if err != nil {
		response.Fail(c, 1500, "查询月度工时汇总失败")
		return
	}
	response.OK(c, result)
}

// Delete 删除打卡记录。DELETE /api/v1/qg/attendances/:id
func (h *AttendanceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的打卡记录 ID")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// RegisterRoutes 注册工时打卡相关路由。
func (h *AttendanceHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	qg := rg.Group("/qg")
	{
		qg.GET("/attendances", h.List)
		qg.POST("/attendances/clock-in", h.ClockIn)
		qg.POST("/attendances/:id/clock-out", h.ClockOut)
		qg.GET("/attendances/monthly-summary", h.MonthlySummary)
		qg.DELETE("/attendances/:id", h.Delete)
	}
}

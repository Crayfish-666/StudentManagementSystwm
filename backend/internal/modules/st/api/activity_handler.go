package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/st/service"
	"student-system/pkg/response"
)

// ActivityHandler 活动接口处理器。
type ActivityHandler struct {
	svc *service.ActivityService
}

// NewActivityHandler 创建活动处理器。
func NewActivityHandler(svc *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

// List 分页查询活动列表。GET /api/v1/st/activities
func (h *ActivityHandler) List(c *gin.Context) {
	status := c.Query("status")
	var associationID int64
	if v := c.Query("association_id"); v != "" {
		associationID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(associationID, status, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询活动列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取活动详情。GET /api/v1/st/activities/:id
func (h *ActivityHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	act, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, "活动不存在")
		return
	}
	response.OK(c, act)
}

// Create 创建活动（保存为 S0 草稿）。POST /api/v1/st/activities
func (h *ActivityHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	act, err := h.svc.Create(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "活动等级必须为 A/B/C/D" || msg == "开始时间格式错误，请使用 RFC3339 格式" || msg == "结束时间格式错误，请使用 RFC3339 格式" {
			code = 40001
		} else if msg == "A/B 级活动必须上传应急预案" {
			code = 40002
		} else if msg == "结束时间必须晚于开始时间" {
			code = 40002
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, act)
}

// Update 更新活动（仅 S0 草稿状态可改）。PUT /api/v1/st/activities/:id
func (h *ActivityHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	act, err := h.svc.Update(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅草稿状态可修改" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, act)
}

// Submit 提交活动（S0 → S1）。POST /api/v1/st/activities/:id/submit
func (h *ActivityHandler) Submit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	act, err := h.svc.Submit(id, userID, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅草稿状态可提交" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, act)
}

// SoftDelete 软删除活动。DELETE /api/v1/st/activities/:id
func (h *ActivityHandler) SoftDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	if err := h.svc.SoftDelete(id, userID); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// Withdraw 撤回活动（S1 → S0）。POST /api/v1/st/activities/:id/withdraw
func (h *ActivityHandler) Withdraw(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	act, err := h.svc.Withdraw(id, userID, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅待审状态可撤回" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, act)
}

// Approve 审批活动。POST /api/v1/st/activities/:id/approve
func (h *ActivityHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	var req service.ApproveActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	act, err := h.svc.Approve(id, userID, &req, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "活动不存在":
			code = 1404
		case "活动状态不允许审批":
			code = 40901
		case "无该步骤审批权限":
			code = 40301
		case "该步骤已通过，请勿重复审批", "该步骤已审批，请勿重复操作":
			code = 40902
		case "前置审批步骤尚未通过":
			code = 40901
		case "驳回意见至少 30 字":
			code = 40001
		case "审批步骤编号超出范围":
			code = 40001
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, act)
}

// ListApprovals 审批记录列表。GET /api/v1/st/activities/:id/approvals
func (h *ActivityHandler) ListApprovals(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	views, err := h.svc.ListApprovals(id)
	if err != nil {
		response.Fail(c, 1500, "查询审批记录失败")
		return
	}
	response.OK(c, views)
}

// Timeline 事件时间线。GET /api/v1/st/activities/:id/timeline
func (h *ActivityHandler) Timeline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	entries, err := h.svc.Timeline(id)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "活动不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, entries)
}

// Checkin 签到。POST /api/v1/st/activities/:id/checkin
func (h *ActivityHandler) Checkin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	// 获取学生 ID：优先使用请求体中的 student_id（社长代签），否则用当前登录者关联
	var req service.CheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}
	sid := req.StudentID
	if sid == 0 {
		studentID, _ := c.Get("student_id")
		sid, _ = studentID.(int64)
		if sid == 0 {
			sid = userID
		}
	}

	view, err := h.svc.Checkin(id, sid, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "活动不存在":
			code = 1404
		case "活动未通过审批，不可签到":
			code = 40901
		case "签到尚未开始（活动开始前30分钟开放）":
			code = 40901
		case "已签到，不可重复签到":
			code = 40902
		case "GPS签到需提供有效经纬度":
			code = 40001
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// ListCheckins 签到列表。GET /api/v1/st/activities/:id/checkins
func (h *ActivityHandler) ListCheckins(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListCheckins(id, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询签到列表失败")
		return
	}
	response.OK(c, result)
}

// SubmitSummary 提交活动总结。POST /api/v1/st/activities/:id/summary
func (h *ActivityHandler) SubmitSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.SubmitSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	view, err := h.svc.SubmitSummary(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "活动不存在":
			code = 1404
		case "活动未通过审批，不可提交总结":
			code = 40901
		case "活动尚未结束，不可提交总结":
			code = 40901
		case "该活动已提交总结，不可重复提交":
			code = 40902
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// GetSummary 获取活动总结。GET /api/v1/st/activities/:id/summary
func (h *ActivityHandler) GetSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的活动 ID")
		return
	}

	view, err := h.svc.GetSummary(id)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "活动总结不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// RegisterRoutes 注册活动相关路由。
func (h *ActivityHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	st := rg.Group("/st")
	{
		st.GET("/activities", h.List)
		st.GET("/activities/:id", h.Get)
		st.POST("/activities", h.Create)
		st.PUT("/activities/:id", h.Update)
		st.POST("/activities/:id/submit", h.Submit)
		st.DELETE("/activities/:id", h.SoftDelete)

		// 审批流
		st.POST("/activities/:id/withdraw", h.Withdraw)
		st.POST("/activities/:id/approve", h.Approve)
		st.GET("/activities/:id/approvals", h.ListApprovals)
		st.GET("/activities/:id/timeline", h.Timeline)

		// 签到
		st.POST("/activities/:id/checkin", h.Checkin)
		st.GET("/activities/:id/checkins", h.ListCheckins)

		// 总结
		st.POST("/activities/:id/summary", h.SubmitSummary)
		st.GET("/activities/:id/summary", h.GetSummary)
	}
}

// toString 安全地将 interface{} 转为 string。
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/ty/service"
	"student-system/pkg/response"
)

// ProbationaryHandler 预备期/转正接口处理器。
type ProbationaryHandler struct {
	svc *service.ProbationaryService
}

// NewProbationaryHandler 创建预备期/转正处理器。
func NewProbationaryHandler(svc *service.ProbationaryService) *ProbationaryHandler {
	return &ProbationaryHandler{svc: svc}
}

// ---- 预备期考察记录接口 ----

// CreateProbationaryRecord 创建预备期考察记录。POST /api/v1/ty/probationary-records
func (h *ProbationaryHandler) CreateProbationaryRecord(c *gin.Context) {
	userID, name, role, ip, ua := actorFromCtx(c)

	var req service.CreateProbationaryRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	record, err := h.svc.CreateProbationaryRecord(userID, &req, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "季度必须在1-4之间":
			code = 40001
		case "考察总结须 ≥ 100字":
			code = 40001
		case "该季度已存在考察记录，不可重复创建":
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, record)
}

// GetProbationaryRecord 获取预备期考察记录详情。GET /api/v1/ty/probationary-records/:id
func (h *ProbationaryHandler) GetProbationaryRecord(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的记录 ID")
		return
	}

	record, err := h.svc.GetProbationaryRecordByID(id)
	if err != nil {
		response.Fail(c, 1404, "预备期考察记录不存在")
		return
	}
	response.OK(c, record)
}

// ListProbationaryRecords 列表查询预备期考察记录。GET /api/v1/ty/probationary-records
// 可选 query：application_id（按申请ID过滤）、page、page_size。
// application_id 缺省时返回全部记录，供"转正流程管理"列表页使用。
func (h *ProbationaryHandler) ListProbationaryRecords(c *gin.Context) {
	var appID int64
	var appPtr *int64
	if v := c.Query("application_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			appID = parsed
			appPtr = &appID
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListProbationaryRecords(appPtr, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询预备期考察记录失败")
		return
	}
	response.OK(c, result)
}

// ---- 转正大会接口 ----

// CreateProbationaryMeeting 创建转正大会。POST /api/v1/ty/probationary-meetings
func (h *ProbationaryHandler) CreateProbationaryMeeting(c *gin.Context) {
	userID, name, role, ip, ua := actorFromCtx(c)

	var req service.CreateProbationaryMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	meeting, err := h.svc.CreateProbationaryMeeting(userID, &req, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "入团申请不存在":
			code = 1404
		case "未找到团员花名册记录，错误码:2630",
			"预备期未满1年，错误码:2630":
			code = 2630
		case "实到人数不足应到人数的2/3",
			"赞成票数不满足要求，须超过实到人数的一半":
			code = 2620
		case "无效的决策值，必须是 pass/reject", "会议时间格式错误":
			code = 40001
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, meeting)
}

// GetProbationaryMeeting 获取转正大会详情。GET /api/v1/ty/probationary-meetings/:id
func (h *ProbationaryHandler) GetProbationaryMeeting(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的转正大会 ID")
		return
	}

	meeting, err := h.svc.GetProbationaryMeetingByID(id)
	if err != nil {
		response.Fail(c, 1404, "转正大会不存在")
		return
	}
	response.OK(c, meeting)
}

// ListProbationaryMeetings 列表查询转正大会。GET /api/v1/ty/probationary-meetings
// 可选 query：application_id（按申请ID过滤）、page、page_size。
// application_id 缺省时返回全部记录，供"转正流程管理"列表页使用。
func (h *ProbationaryHandler) ListProbationaryMeetings(c *gin.Context) {
	var appID int64
	var appPtr *int64
	if v := c.Query("application_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			appID = parsed
			appPtr = &appID
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListProbationaryMeetings(appPtr, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询转正大会列表失败")
		return
	}
	response.OK(c, result)
}

// RegisterRoutes 注册预备期/转正相关路由。
func (h *ProbationaryHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	ty := rg.Group("/ty")
	{
		// 预备期考察记录
		ty.POST("/probationary-records", h.CreateProbationaryRecord)
		ty.GET("/probationary-records/:id", h.GetProbationaryRecord)
		ty.GET("/probationary-records", h.ListProbationaryRecords)

		// 转正大会
		ty.POST("/probationary-meetings", h.CreateProbationaryMeeting)
		ty.GET("/probationary-meetings/:id", h.GetProbationaryMeeting)
		ty.GET("/probationary-meetings", h.ListProbationaryMeetings)
	}
}

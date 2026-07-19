package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/ty/service"
	"student-system/pkg/response"
)

// CultivationHandler 培养考察接口处理器。
type CultivationHandler struct {
	svc *service.CultivationService
}

// NewCultivationHandler 创建培养考察处理器。
func NewCultivationHandler(svc *service.CultivationService) *CultivationHandler {
	return &CultivationHandler{svc: svc}
}

// ==================== 培养联系人接口 ====================

// AssignMentor 分配培养联系人。POST /api/v1/ty/cultivation-links
// AssignMentors 批量分配 2 位培养联系人（PRD §4.3.4）。POST /api/v1/ty/cultivation-links
//
// 请求体：
//
//	{
//	  "application_id": 88,
//	  "start_at": "2026-06-24",
//	  "mentors": [
//	    { "mentor_student_id": 12, "mentor_type": "league_member" },
//	    { "mentor_student_id": 13, "mentor_type": "league_member" }
//	  ]
//	}
//
// 响应：data 为两位联系人的视图数组。
func (h *CultivationHandler) AssignMentors(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.AssignMentorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	result, err := h.svc.AssignMentors(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch {
		case msg == "入团申请不存在":
			code = 1404
		case contains(msg, "已存在在任培养联系人"):
			code = 2540
		case contains(msg, "培养联系人数量须为 2 位"):
			code = 2541
		case contains(msg, "两位培养联系人不能为同一人"):
			code = 2542
		case contains(msg, "类型无效") || contains(msg, "类型与"):
			code = 2543
		case contains(msg, "不存在") || contains(msg, "民主党派"):
			code = 2544
		case contains(msg, "优先从中选任") || contains(msg, "优先从支部团员"):
			code = 2545
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// EndMentor 结束培养关系。POST /api/v1/ty/cultivation-links/:id/end
func (h *CultivationHandler) EndMentor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的培养联系人 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	result, err := h.svc.EndMentor(id, userID)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "培养联系人记录不存在" {
			code = 1404
		} else if contains(msg, "已结束") {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// ListLinks 查询培养联系人列表。GET /api/v1/ty/cultivation-links?application_id=xxx
func (h *CultivationHandler) ListLinks(c *gin.Context) {
	var applicationID int64
	if v := c.Query("application_id"); v != "" {
		applicationID, _ = strconv.ParseInt(v, 10, 64)
	}

	result, err := h.svc.ListLinks(applicationID)
	if err != nil {
		response.Fail(c, 1500, "查询培养联系人失败")
		return
	}
	response.OK(c, result)
}

// ==================== 培养记录接口 ====================

// CreateRecord 创建培养记录。POST /api/v1/ty/cultivation-records
func (h *CultivationHandler) CreateRecord(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	result, err := h.svc.CreateRecord(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch {
		case contains(msg, "摘要不足 50 字"):
			code = 2510
		case contains(msg, "该月份的培养记录已存在"):
			code = 2511
		case contains(msg, "成绩须在 0-100"):
			code = 2520
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// ListRecords 查询培养记录列表。GET /api/v1/ty/cultivation-records?application_id=xxx
func (h *CultivationHandler) ListRecords(c *gin.Context) {
	var applicationID int64
	if v := c.Query("application_id"); v != "" {
		applicationID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListRecords(applicationID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询培养记录失败")
		return
	}
	response.OK(c, result)
}

// ==================== 团课记录接口 ====================

// CreateCourse 创建团课记录。POST /api/v1/ty/course-records
func (h *CultivationHandler) CreateCourse(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	result, err := h.svc.CreateCourse(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if contains(msg, "团课成绩须在 0-100") {
			code = 2520
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// ListCourses 查询团课列表。GET /api/v1/ty/course-records?student_id=xxx
func (h *CultivationHandler) ListCourses(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var studentID int64
	if v := c.Query("student_id"); v != "" {
		studentID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListCourses(userID, studentID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询团课记录失败")
		return
	}
	response.OK(c, result)
}

// UpdatePassStatus 标记结业。PUT /api/v1/ty/course-records/:id/pass
func (h *CultivationHandler) UpdatePassStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的团课记录 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	result, err := h.svc.UpdatePassStatus(id, userID)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "团课记录不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// UpdateCourse 更新团课记录（成绩、证书编号等）。PUT /api/v1/ty/course-records/:id
func (h *CultivationHandler) UpdateCourse(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的团课记录 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	result, err := h.svc.UpdateCourse(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "团课记录不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// ==================== 思想汇报接口 ====================

// CreateReport 提交思想汇报。POST /api/v1/ty/thought-reports
func (h *CultivationHandler) CreateReport(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	result, err := h.svc.CreateReport(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch {
		case contains(msg, "不足 1000 字"):
			code = 2530
		case contains(msg, "该季度的思想汇报已提交"):
			code = 2531
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// ListReports 查询思想汇报列表。GET /api/v1/ty/thought-reports?application_id=xxx
//
// 数据范围隔离：
//   - 学生仅能查看自己的思想汇报
//   - 辅导员仅能查看本专业学生的思想汇报
//   - 校/院系管理员可查看全部
func (h *CultivationHandler) ListReports(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var applicationID int64
	if v := c.Query("application_id"); v != "" {
		applicationID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListReports(userID, applicationID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询思想汇报失败")
		return
	}
	response.OK(c, result)
}

// GetReport 思想汇报详情。GET /api/v1/ty/thought-reports/:id
func (h *CultivationHandler) GetReport(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的思想汇报 ID")
		return
	}

	result, err := h.svc.GetReport(id)
	if err != nil {
		response.Fail(c, 1404, "思想汇报不存在")
		return
	}
	response.OK(c, result)
}

// RegisterRoutes 注册培养考察相关路由。
func (h *CultivationHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	ty := rg.Group("/ty")
	{
		// 培养联系人
		ty.POST("/cultivation-links", h.AssignMentors)
		ty.GET("/cultivation-links", h.ListLinks)
		ty.POST("/cultivation-links/:id/end", h.EndMentor)

		// 培养记录
		ty.POST("/cultivation-records", h.CreateRecord)
		ty.GET("/cultivation-records", h.ListRecords)

		// 团课记录
		ty.POST("/course-records", h.CreateCourse)
		ty.GET("/course-records", h.ListCourses)
		ty.PUT("/course-records/:id", h.UpdateCourse)
		ty.PUT("/course-records/:id/pass", h.UpdatePassStatus)

		// 思想汇报
		ty.POST("/thought-reports", h.CreateReport)
		ty.GET("/thought-reports", h.ListReports)
		ty.GET("/thought-reports/:id", h.GetReport)
	}
}



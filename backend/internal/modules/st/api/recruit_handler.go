// Package api 实现 ST 模块招新 HTTP 接口。
//
// 设计依据：docs/04 §7.3。
// 路由：
//   GET    /st/recruit-plans                        列表
//   GET    /st/recruit-plans/:id                    详情
//   POST   /st/recruit-plans                        创建（S0）
//   PUT    /st/recruit-plans/:id                    更新（仅 S0）
//   POST   /st/recruit-plans/:id/submit             提交（S0 → S1）
//   POST   /st/recruit-plans/:id/withdraw           撤回（S1 → S0）
//   POST   /st/recruit-plans/:id/approve            审批通过（S1 → S3）
//   POST   /st/recruit-plans/:id/reject             驳回（S1 → S4）
//   POST   /st/recruit-plans/:id/publish            发布（S3 保持 + 设 result_deadline）
//   POST   /st/recruit-plans/:id/finish             提前结束招新（仅 S3 + 未结束可用）
//
//   GET    /st/recruit-applies                      申请列表
//   POST   /st/recruit-applies                      学生投递
//   POST   /st/recruit-applies/:id/result           录入面试结果
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/st/service"
	"student-system/pkg/response"
)

// RecruitHandler 招新接口处理器。
type RecruitHandler struct {
	svc *service.RecruitService
}

// NewRecruitHandler 创建招新处理器。
func NewRecruitHandler(svc *service.RecruitService) *RecruitHandler {
	return &RecruitHandler{svc: svc}
}

// ---- 招新计划 ----

// ListPlans 分页查询招新计划。GET /api/v1/st/recruit-plans
func (h *RecruitHandler) ListPlans(c *gin.Context) {
	var associationID int64
	if v := c.Query("association_id"); v != "" {
		associationID, _ = strconv.ParseInt(v, 10, 64)
	}
	status := c.Query("status")
	academicYear := c.Query("academic_year")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListPlans(associationID, status, academicYear, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询招新计划失败")
		return
	}
	response.OK(c, result)
}

// GetPlan 查询招新计划详情。GET /api/v1/st/recruit-plans/:id
func (h *RecruitHandler) GetPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新计划 ID")
		return
	}
	view, err := h.svc.GetPlan(id)
	if err != nil {
		response.Fail(c, 1404, "招新计划不存在")
		return
	}
	response.OK(c, view)
}

// CreatePlan 创建招新计划。POST /api/v1/st/recruit-plans
func (h *RecruitHandler) CreatePlan(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	var req service.CreateRecruitPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}
	view, err := h.svc.CreatePlan(userID, &req, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "社团不存在", "招新计划不存在":
			code = 1404
		case "招新季节必须为 autumn(秋) 或 spring(春)", "目标人数必须大于 0", "面试时间格式错误，请使用 RFC3339 格式":
			code = 40001
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// UpdatePlan 更新招新计划。PUT /api/v1/st/recruit-plans/:id
func (h *RecruitHandler) UpdatePlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新计划 ID")
		return
	}
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.UpdateRecruitPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}
	view, err := h.svc.UpdatePlan(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅草稿状态可修改" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// SubmitPlan 提交招新计划。POST /api/v1/st/recruit-plans/:id/submit
func (h *RecruitHandler) SubmitPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新计划 ID")
		return
	}
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	view, err := h.svc.SubmitPlan(id, userID, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅草稿状态可提交" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// WithdrawPlan 撤回招新计划。POST /api/v1/st/recruit-plans/:id/withdraw
func (h *RecruitHandler) WithdrawPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新计划 ID")
		return
	}
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	view, err := h.svc.WithdrawPlan(id, userID, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅待审状态可撤回" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// ApprovePlan 审批通过招新计划。POST /api/v1/st/recruit-plans/:id/approve
func (h *RecruitHandler) ApprovePlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新计划 ID")
		return
	}
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	view, err := h.svc.ApprovePlan(id, userID, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "招新计划不存在":
			code = 1404
		case "仅待审状态可审批":
			code = 40901
		case "无招新计划审批权限":
			code = 40301
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// RejectPlan 驳回招新计划。POST /api/v1/st/recruit-plans/:id/reject
func (h *RecruitHandler) RejectPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新计划 ID")
		return
	}
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	var req struct {
		Opinion string `json:"opinion"`
	}
	_ = c.ShouldBindJSON(&req)

	view, err := h.svc.RejectPlan(id, userID, req.Opinion, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "招新计划不存在":
			code = 1404
		case "仅待审状态可驳回":
			code = 40901
		case "无招新计划审批权限":
			code = 40301
		case "驳回意见至少 10 字":
			code = 40001
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// PublishPlan 发布招新计划。POST /api/v1/st/recruit-plans/:id/publish
func (h *RecruitHandler) PublishPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新计划 ID")
		return
	}
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	view, err := h.svc.PublishPlan(id, userID, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅已通过审批的计划可发布" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// FinishPlan 提前结束招新（仅 S3 + 未结束 状态可用，操作不可逆）。POST /api/v1/st/recruit-plans/:id/finish
func (h *RecruitHandler) FinishPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新计划 ID")
		return
	}
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	var req service.FinishRecruitPlanRequest
	_ = c.ShouldBindJSON(&req) // body 可选

	view, err := h.svc.FinishPlan(id, userID, &req, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "招新计划不存在":
			code = 1404
		case "仅已通过审批的招新计划可结束", "该招新计划已结束，不可重复结束":
			code = 40901
		case "无招新计划结束权限":
			code = 40301
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// ---- 招新申请 ----

// ListApplies 分页查询招新申请。GET /api/v1/st/recruit-applies
// 默认按当前登录用户关联的 student_id 过滤（学生视角）。
// 仅当显式传 ?scope=all 时才返回全量（管理员视角）。
// 单独传 ?student_id=xxx 会覆盖默认过滤，用于管理员指定学生查询。
func (h *RecruitHandler) ListApplies(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var planID int64
	if v := c.Query("plan_id"); v != "" {
		planID, _ = strconv.ParseInt(v, 10, 64)
	}

	var studentID int64
	scope := c.Query("scope")
	if scope == "all" {
		// 显式请求全量：忽略默认过滤；可再用 ?student_id=xxx 指定单一学生
		if v := c.Query("student_id"); v != "" {
			studentID, _ = strconv.ParseInt(v, 10, 64)
		}
	} else {
		// 默认按当前用户的学生身份过滤（无论该账号是否还兼有其他角色）
		sid, _ := h.svc.GetStudentIDByUserID(userID)
		if sid > 0 {
			studentID = sid
		}
		// 仅当用户未关联学生身份且未传 scope=all 时才不按 student 过滤（教师账号无 student_id）
	}

	result := c.Query("result")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	out, err := h.svc.ListApplies(planID, studentID, result, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询招新申请失败")
		return
	}
	response.OK(c, out)
}

// CreateApply 学生投递招新申请。POST /api/v1/st/recruit-applies
func (h *RecruitHandler) CreateApply(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	// 从登录用户读取关联的 student_id（学生身份）
	studentID, _ := h.svc.GetStudentIDByUserID(userID)
	if studentID == 0 {
		response.Fail(c, 1404, "当前用户未关联学生身份，无法投递招新申请")
		return
	}
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	var req service.CreateRecruitApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	view, err := h.svc.CreateApply(userID, studentID, &req, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "招新计划不存在":
			code = 1404
		case "学生不存在":
			code = 1404
		case "该招新计划未发布，不可投递":
			code = 40901
		case "您已投递过该招新计划，不可重复投递":
			code = 40902
		}
		// 学年 3 社团上限 → 42230（按 docs/04 §7.3 末尾约定）
		if len(msg) >= 8 && msg[:8] == "您本学年" {
			code = 42230
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// SubmitApplyResult 录入面试结果。POST /api/v1/st/recruit-applies/:id/result
func (h *RecruitHandler) SubmitApplyResult(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的招新申请 ID")
		return
	}
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	actorName, _ := c.Get("nickname")
	actorRole, _ := c.Get("role_code")
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	var req service.SubmitApplyResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	view, err := h.svc.SubmitApplyResult(id, userID, &req, toString(actorName), toString(actorRole), ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "招新申请不存在", "招新计划不存在":
			code = 1404
		case "该申请已录入结果，不可重复操作":
			code = 40902
		case "结果必须为 accepted 或 rejected":
			code = 40001
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, view)
}

// RegisterRoutes 注册招新相关路由。
// 学生可访问：查询招新计划、投递招新申请、查询自己的申请；
// 管理端：创建/更新/审批/发布计划、录入面试结果（需 admin 中间件）。
func (h *RecruitHandler) RegisterRoutes(rg *gin.RouterGroup, admin gin.HandlerFunc) {
	st := rg.Group("/st")
	{
		// 学生可访问（登录即可）
		st.GET("/recruit-plans", h.ListPlans)
		st.GET("/recruit-plans/:id", h.GetPlan)
		st.GET("/recruit-applies", h.ListApplies)
		st.POST("/recruit-applies", h.CreateApply)

		// 管理端（仅 R-SY-ADMIN）
		st.POST("/recruit-plans", admin, h.CreatePlan)
		st.PUT("/recruit-plans/:id", admin, h.UpdatePlan)
		st.POST("/recruit-plans/:id/submit", admin, h.SubmitPlan)
		st.POST("/recruit-plans/:id/withdraw", admin, h.WithdrawPlan)
		st.POST("/recruit-plans/:id/approve", admin, h.ApprovePlan)
		st.POST("/recruit-plans/:id/reject", admin, h.RejectPlan)
		st.POST("/recruit-plans/:id/publish", admin, h.PublishPlan)
		st.POST("/recruit-plans/:id/finish", admin, h.FinishPlan)
		st.POST("/recruit-applies/:id/result", admin, h.SubmitApplyResult)
	}
}

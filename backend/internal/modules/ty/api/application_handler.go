package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/ty/service"
	"student-system/pkg/response"
)

// ApplicationHandler 入团申请接口处理器。
type ApplicationHandler struct {
	svc *service.ApplicationService
}

// NewApplicationHandler 创建入团申请处理器。
func NewApplicationHandler(svc *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

// List 分页查询入团申请列表。GET /api/v1/ty/applications
func (h *ApplicationHandler) List(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	status := c.Query("status")
	var studentID, collegeID int64
	if v := c.Query("student_id"); v != "" {
		studentID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := c.Query("college_id"); v != "" {
		collegeID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(userID, status, studentID, collegeID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询入团申请列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取入团申请详情。GET /api/v1/ty/applications/:id
func (h *ApplicationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	app, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, "申请不存在")
		return
	}
	response.OK(c, app)
}

// Create 创建入团申请（保存为 S0 草稿）。POST /api/v1/ty/applications
func (h *ApplicationHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	app, err := h.svc.Create(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "思想政治表现自述字数须 ≥ 500" {
			code = 2401
		} else if msg == "申请人年龄超出 14-28 周岁范围" {
			code = 2402
		} else if msg == "已存在审批中申请，请勿重复提交" {
			code = 2403
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, app)
}

// Update 更新入团申请（仅 S0 状态可改）。PUT /api/v1/ty/applications/:id
func (h *ApplicationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	app, err := h.svc.Update(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "思想政治表现自述字数须 ≥ 500" {
			code = 2401
		} else if msg == "仅草稿状态可修改" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, app)
}

// Submit 提交入团申请（S0 → S1）。POST /api/v1/ty/applications/:id/submit
func (h *ApplicationHandler) Submit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	userID, name, role, ip, ua := actorFromCtx(c)

	app, err := h.svc.Submit(id, userID, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "思想政治表现自述字数须 ≥ 500" {
			code = 2401
		} else if msg == "申请人年龄超出 14-28 周岁范围" {
			code = 2402
		} else if msg == "已存在审批中申请，请勿重复提交" {
			code = 2403
		} else if msg == "仅草稿状态可提交" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, app)
}

// Withdraw 撤回入团申请（S1 → S0）。POST /api/v1/ty/applications/:id/withdraw
func (h *ApplicationHandler) Withdraw(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	userID, name, role, ip, ua := actorFromCtx(c)

	var req service.WithdrawRequest
	_ = c.ShouldBindJSON(&req)

	app, err := h.svc.Withdraw(id, userID, req.Reason, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅待审状态可撤回" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, app)
}

// SoftDelete 软删除入团申请（仅 S0/S4）。DELETE /api/v1/ty/applications/:id
func (h *ApplicationHandler) SoftDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	if err := h.svc.SoftDelete(id, userID); err != nil {
		msg := err.Error()
		code := 1500
		if msg == "仅草稿或驳回状态可删除" {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, gin.H{"id": id})
}

// Approve 三级审批通过/驳回端点。POST /api/v1/ty/applications/:id/approve
func (h *ApplicationHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	userID, name, role, ip, ua := actorFromCtx(c)

	var req service.ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	app, err := h.svc.Approve(id, userID, &req, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "无该步骤审批权限", "仅可审批本院系申请", "仅可审批本专业学生的申请":
			code = 1004
		case "辅导员初审尚未通过，院系不可复核",
			"院系复核尚未通过，校级不可终审",
			"该步骤已通过，请勿重复审批":
			code = 40901
		case "无效的审批步骤", "无效的审批结果", "审批意见至少 5 字", "无法解析审批动作":
			code = 40001
		case "申请不存在":
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, app)
}

// ListApprovals 列出审批记录。GET /api/v1/ty/applications/:id/approvals
func (h *ApplicationHandler) ListApprovals(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}
	records, err := h.svc.ListApprovals(id)
	if err != nil {
		response.Fail(c, 1500, "查询审批记录失败")
		return
	}
	response.OK(c, gin.H{"records": records})
}

// Timeline 返回事件流时间线。GET /api/v1/ty/applications/:id/timeline
func (h *ApplicationHandler) Timeline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}
	entries, err := h.svc.Timeline(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, gin.H{"entries": entries})
}

// Pending 列出当前用户待办审批。GET /api/v1/ty/approvals/pending
func (h *ApplicationHandler) Pending(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListPending(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询待办失败")
		return
	}
	response.OK(c, result)
}

// DevelopmentTrack 查询某学生的团员发展全流程轨迹。GET /api/v1/ty/students/:id/development-track
func (h *ApplicationHandler) DevelopmentTrack(c *gin.Context) {
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的学生 ID")
		return
	}

	result, err := h.svc.DevelopmentTrack(studentID)
	if err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, result)
}

// ListBranches 获取团支部列表（下拉选择用）。GET /api/v1/ty/branches
func (h *ApplicationHandler) ListBranches(c *gin.Context) {
	var collegeID int64
	if v := c.Query("college_id"); v != "" {
		collegeID, _ = strconv.ParseInt(v, 10, 64)
	}

	branches, err := h.svc.ListBranches(collegeID)
	if err != nil {
		response.Fail(c, 1500, "查询团支部列表失败")
		return
	}
	response.OK(c, branches)
}

// RegisterRoutes 注册入团申请相关路由。
func (h *ApplicationHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	ty := rg.Group("/ty")
	{
		ty.GET("/branches", h.ListBranches)
		ty.GET("/applications", h.List)
		ty.GET("/applications/:id", h.Get)
		ty.GET("/applications/:id/approvals", h.ListApprovals)
		ty.GET("/applications/:id/timeline", h.Timeline)

		// 学生可操作：创建/更新/提交/撤回/删除
		ty.POST("/applications", h.Create)
		ty.PUT("/applications/:id", h.Update)
		ty.POST("/applications/:id/submit", h.Submit)
		ty.POST("/applications/:id/withdraw", h.Withdraw)
		ty.DELETE("/applications/:id", h.SoftDelete)

		// 三级审批
		ty.POST("/applications/:id/approve", h.Approve)
		ty.GET("/approvals/pending", h.Pending)

		// 发展轨迹
		ty.GET("/students/:id/development-track", h.DevelopmentTrack)
	}
}

// actorFromCtx 从 gin 上下文提取 (uid, name, primaryRole, ip, ua)。
func actorFromCtx(c *gin.Context) (int64, string, string, string, string) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	nameRaw, _ := c.Get("user_name")
	name, _ := nameRaw.(string)
	rolesRaw, _ := c.Get("user_roles")
	roleStr := ""
	if rs, ok := rolesRaw.([]string); ok && len(rs) > 0 {
		roleStr = rs[0]
	}
	return userID, name, roleStr, c.ClientIP(), c.GetHeader("User-Agent")
}

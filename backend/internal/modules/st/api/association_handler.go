package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/st/service"
	"student-system/pkg/response"
)

// AssociationHandler 社团接口处理器。
type AssociationHandler struct {
	svc *service.AssociationService
}

// NewAssociationHandler 创建社团处理器。
func NewAssociationHandler(svc *service.AssociationService) *AssociationHandler {
	return &AssociationHandler{svc: svc}
}

// List 分页查询社团列表。GET /api/v1/st/associations
func (h *AssociationHandler) List(c *gin.Context) {
	status := c.Query("status")
	keyword := c.Query("keyword")
	var collegeID int64
	if v := c.Query("college_id"); v != "" {
		collegeID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(status, collegeID, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询社团列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取社团详情。GET /api/v1/st/associations/:id
func (h *AssociationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的社团 ID")
		return
	}

	assoc, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, "社团不存在")
		return
	}
	response.OK(c, assoc)
}

// Create 创建社团。POST /api/v1/st/associations
func (h *AssociationHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateAssociationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	assoc, err := h.svc.Create(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "发起人须 5-20 名" {
			code = 3401
		} else if msg == "同名社团已存在" {
			code = 40904
		} else if msg == "指导教师同期最多指导 3 个社团" {
			code = 42220
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, assoc)
}

// Update 更新社团。PUT /api/v1/st/associations/:id
func (h *AssociationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的社团 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.UpdateAssociationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	assoc, err := h.svc.Update(id, userID, &req)
	if err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, assoc)
}

// SoftDelete 软删除社团。DELETE /api/v1/st/associations/:id
func (h *AssociationHandler) SoftDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的社团 ID")
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

// ListFounders 查询社团发起人。GET /api/v1/st/associations/:id/founders
func (h *AssociationHandler) ListFounders(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的社团 ID")
		return
	}

	founders, err := h.svc.ListFounders(id)
	if err != nil {
		response.Fail(c, 1500, "查询发起人列表失败")
		return
	}
	response.OK(c, gin.H{"items": founders})
}

// ListMembers 查询社团成员。GET /api/v1/st/associations/:id/members
func (h *AssociationHandler) ListMembers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的社团 ID")
		return
	}

	members, err := h.svc.ListMembers(id)
	if err != nil {
		response.Fail(c, 1500, "查询成员列表失败")
		return
	}
	response.OK(c, gin.H{"items": members})
}

// ListUsers 查询用户列表(指导教师下拉用,仅教职工)。GET /api/v1/st/users
func (h *AssociationHandler) ListUsers(c *gin.Context) {
	users, err := h.svc.ListUsers()
	if err != nil {
		response.Fail(c, 1500, "查询用户列表失败")
		return
	}
	response.OK(c, gin.H{"items": users})
}

// ListStudents 查询学生列表(社长下拉用)。GET /api/v1/st/students
func (h *AssociationHandler) ListStudents(c *gin.Context) {
	students, err := h.svc.ListStudents()
	if err != nil {
		response.Fail(c, 1500, "查询学生列表失败")
		return
	}
	response.OK(c, gin.H{"items": students})
}

// RegisterRoutes 注册社团相关路由。
func (h *AssociationHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	st := rg.Group("/st")
	{
		st.GET("/associations", h.List)
		st.GET("/associations/:id", h.Get)
		st.POST("/associations", h.Create)
		st.PUT("/associations/:id", h.Update)
		st.DELETE("/associations/:id", h.SoftDelete)
		st.GET("/associations/:id/founders", h.ListFounders)
		st.GET("/associations/:id/members", h.ListMembers)
		st.GET("/users", h.ListUsers)
		st.GET("/students", h.ListStudents)
	}
}

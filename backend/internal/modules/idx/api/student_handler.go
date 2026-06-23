package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/idx/service"
	"student-system/pkg/response"
)

// StudentHandler 学生接口处理器。
type StudentHandler struct {
	svc *service.StudentService
}

// NewStudentHandler 创建学生处理器。
func NewStudentHandler(svc *service.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

// List 分页查询学生列表。GET /api/v1/idx/students
func (h *StudentHandler) List(c *gin.Context) {
	var collegeID, classID int64
	if v := c.Query("college_id"); v != "" {
		collegeID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := c.Query("class_id"); v != "" {
		classID, _ = strconv.ParseInt(v, 10, 64)
	}
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(collegeID, classID, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询学生列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取学生详情。GET /api/v1/idx/students/:id
func (h *StudentHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的学生 ID")
		return
	}

	student, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, "学生不存在")
		return
	}
	response.OK(c, student)
}

// Create 创建学生。POST /api/v1/idx/students
func (h *StudentHandler) Create(c *gin.Context) {
	var req service.CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	student, err := h.svc.Create(&req)
	if err != nil {
		response.Fail(c, 1409, err.Error())
		return
	}
	response.OK(c, student)
}

// Update 更新学生。PUT /api/v1/idx/students/:id
func (h *StudentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的学生 ID")
		return
	}

	var req service.UpdateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	student, err := h.svc.Update(id, &req)
	if err != nil {
		response.Fail(c, 1500, "更新学生失败: "+err.Error())
		return
	}
	response.OK(c, student)
}

// SoftDelete 软删除学生。DELETE /api/v1/idx/students/:id
func (h *StudentHandler) SoftDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的学生 ID")
		return
	}

	if err := h.svc.SoftDelete(id); err != nil {
		response.Fail(c, 1500, "删除学生失败")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// Import 批量导入学生（CSV）。POST /api/v1/idx/students/import
func (h *StudentHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, 40001, "请上传 CSV 文件")
		return
	}
	defer file.Close()

	result, err := h.svc.ImportCSV(file)
	if err != nil {
		response.Fail(c, 1500, "导入失败: "+err.Error())
		return
	}
	response.OK(c, result)
}

// OrgTree 获取组织树。GET /api/v1/idx/org-tree
func (h *StudentHandler) OrgTree(c *gin.Context) {
	tree, err := h.svc.BuildOrgTree()
	if err != nil {
		response.Fail(c, 1500, "获取组织树失败")
		return
	}
	response.OK(c, gin.H{"tree": tree})
}

// RegisterRoutes 注册学生相关路由。
func (h *StudentHandler) RegisterRoutes(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	students := rg.Group("/idx")
	{
		students.GET("/students", h.List)
		students.GET("/students/:id", h.Get)
		students.GET("/org-tree", h.OrgTree)

		// 管理类接口
		students.POST("/students", adminOnly, h.Create)
		students.PUT("/students/:id", adminOnly, h.Update)
		students.DELETE("/students/:id", adminOnly, h.SoftDelete)
		students.POST("/students/import", adminOnly, h.Import)
	}
}

package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"student-system/internal/models"
	"student-system/pkg/response"
)

// OrgHandler 院系/专业/班级 CRUD 接口处理器。
type OrgHandler struct {
	db *gorm.DB
}

// NewOrgHandler 创建组织管理处理器。
func NewOrgHandler(db *gorm.DB) *OrgHandler {
	return &OrgHandler{db: db}
}

// ============ 院系 CRUD ============

// ListColleges 院系列表。GET /api/v1/sys/colleges
func (h *OrgHandler) ListColleges(c *gin.Context) {
	var colleges []models.SysCollege
	if err := h.db.Where("is_deleted = 0").Order("id ASC").Find(&colleges).Error; err != nil {
		response.Fail(c, 1500, "查询院系失败")
		return
	}
	response.OK(c, gin.H{"items": colleges, "total": len(colleges)})
}

// CreateCollege 新增院系。POST /api/v1/sys/colleges
func (h *OrgHandler) CreateCollege(c *gin.Context) {
	var req struct {
		Code   string `json:"code" binding:"required"`
		Name   string `json:"name" binding:"required"`
		NameEn string `json:"name_en"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整")
		return
	}

	college := models.SysCollege{
		Code:   req.Code,
		Name:   req.Name,
		NameEn: req.NameEn,
	}
	if err := h.db.Create(&college).Error; err != nil {
		response.Fail(c, 1409, "院系代码已存在或创建失败")
		return
	}
	response.OK(c, college)
}

// UpdateCollege 更新院系。PUT /api/v1/sys/colleges/:id
func (h *OrgHandler) UpdateCollege(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的 ID")
		return
	}

	var college models.SysCollege
	if err := h.db.Where("id = ? AND is_deleted = 0", id).First(&college).Error; err != nil {
		response.Fail(c, 1404, "院系不存在")
		return
	}

	var req struct {
		Code   *string `json:"code"`
		Name   *string `json:"name"`
		NameEn *string `json:"name_en"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.NameEn != nil {
		updates["name_en"] = *req.NameEn
	}

	if len(updates) > 0 {
		h.db.Model(&college).Updates(updates)
	}
	h.db.Where("id = ? AND is_deleted = 0", id).First(&college)
	response.OK(c, college)
}

// DeleteCollege 删除院系。DELETE /api/v1/sys/colleges/:id
func (h *OrgHandler) DeleteCollege(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的 ID")
		return
	}

	// 检查是否有关联专业
	var count int64
	h.db.Model(&models.SysMajor{}).Where("college_id = ? AND is_deleted = 0", id).Count(&count)
	if count > 0 {
		response.Fail(c, 1409, "该院系下存在专业，无法删除")
		return
	}

	if err := h.db.Model(&models.SysCollege{}).Where("id = ?", id).Update("is_deleted", 1).Error; err != nil {
		response.Fail(c, 1500, "删除失败")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ============ 专业 CRUD ============

// ListMajors 专业列表。GET /api/v1/sys/majors
func (h *OrgHandler) ListMajors(c *gin.Context) {
	query := h.db.Where("is_deleted = 0")
	if collegeID := c.Query("college_id"); collegeID != "" {
		query = query.Where("college_id = ?", collegeID)
	}

	var majors []models.SysMajor
	if err := query.Order("id ASC").Find(&majors).Error; err != nil {
		response.Fail(c, 1500, "查询专业失败")
		return
	}
	response.OK(c, gin.H{"items": majors, "total": len(majors)})
}

// CreateMajor 新增专业。POST /api/v1/sys/majors
func (h *OrgHandler) CreateMajor(c *gin.Context) {
	var req struct {
		CollegeID int64  `json:"college_id" binding:"required"`
		Code      string `json:"code" binding:"required"`
		Name      string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整")
		return
	}

	major := models.SysMajor{
		CollegeID: req.CollegeID,
		Code:      req.Code,
		Name:      req.Name,
	}
	if err := h.db.Create(&major).Error; err != nil {
		response.Fail(c, 1409, "专业代码已存在或创建失败")
		return
	}
	response.OK(c, major)
}

// UpdateMajor 更新专业。PUT /api/v1/sys/majors/:id
func (h *OrgHandler) UpdateMajor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的 ID")
		return
	}

	var major models.SysMajor
	if err := h.db.Where("id = ? AND is_deleted = 0", id).First(&major).Error; err != nil {
		response.Fail(c, 1404, "专业不存在")
		return
	}

	var req struct {
		CollegeID *int64  `json:"college_id"`
		Code      *string `json:"code"`
		Name      *string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.CollegeID != nil {
		updates["college_id"] = *req.CollegeID
	}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if len(updates) > 0 {
		h.db.Model(&major).Updates(updates)
	}
	h.db.Where("id = ? AND is_deleted = 0", id).First(&major)
	response.OK(c, major)
}

// DeleteMajor 删除专业。DELETE /api/v1/sys/majors/:id
func (h *OrgHandler) DeleteMajor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的 ID")
		return
	}

	var count int64
	h.db.Model(&models.IdxClass{}).Where("major_id = ? AND is_deleted = 0", id).Count(&count)
	if count > 0 {
		response.Fail(c, 1409, "该专业下存在班级，无法删除")
		return
	}

	if err := h.db.Model(&models.SysMajor{}).Where("id = ?", id).Update("is_deleted", 1).Error; err != nil {
		response.Fail(c, 1500, "删除失败")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ============ 班级 CRUD ============

// ListClasses 班级列表。GET /api/v1/sys/classes
func (h *OrgHandler) ListClasses(c *gin.Context) {
	query := h.db.Where("is_deleted = 0")
	if majorID := c.Query("major_id"); majorID != "" {
		query = query.Where("major_id = ?", majorID)
	}

	var classes []models.IdxClass
	if err := query.Order("id ASC").Find(&classes).Error; err != nil {
		response.Fail(c, 1500, "查询班级失败")
		return
	}
	response.OK(c, gin.H{"items": classes, "total": len(classes)})
}

// CreateClass 新增班级。POST /api/v1/sys/classes
func (h *OrgHandler) CreateClass(c *gin.Context) {
	var req struct {
		MajorID     int64  `json:"major_id" binding:"required"`
		Grade       int    `json:"grade" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Name        string `json:"name" binding:"required"`
		CounselorID *int64 `json:"counselor_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整")
		return
	}

	class := models.IdxClass{
		MajorID:     req.MajorID,
		Grade:       req.Grade,
		Code:        req.Code,
		Name:        req.Name,
		CounselorID: req.CounselorID,
	}
	if err := h.db.Create(&class).Error; err != nil {
		response.Fail(c, 1409, "班级代码已存在或创建失败")
		return
	}
	response.OK(c, class)
}

// UpdateClass 更新班级。PUT /api/v1/sys/classes/:id
func (h *OrgHandler) UpdateClass(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的 ID")
		return
	}

	var class models.IdxClass
	if err := h.db.Where("id = ? AND is_deleted = 0", id).First(&class).Error; err != nil {
		response.Fail(c, 1404, "班级不存在")
		return
	}

	var req struct {
		MajorID     *int64  `json:"major_id"`
		Grade       *int    `json:"grade"`
		Code        *string `json:"code"`
		Name        *string `json:"name"`
		CounselorID *int64  `json:"counselor_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.MajorID != nil {
		updates["major_id"] = *req.MajorID
	}
	if req.Grade != nil {
		updates["grade"] = *req.Grade
	}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.CounselorID != nil {
		updates["counselor_id"] = *req.CounselorID
	}

	if len(updates) > 0 {
		h.db.Model(&class).Updates(updates)
	}
	h.db.Where("id = ? AND is_deleted = 0", id).First(&class)
	response.OK(c, class)
}

// DeleteClass 删除班级。DELETE /api/v1/sys/classes/:id
func (h *OrgHandler) DeleteClass(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的 ID")
		return
	}

	if err := h.db.Model(&models.IdxClass{}).Where("id = ?", id).Update("is_deleted", 1).Error; err != nil {
		response.Fail(c, 1500, "删除失败")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// RegisterRoutes 注册组织管理路由。
func (h *OrgHandler) RegisterRoutes(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	// 院系
	colleges := rg.Group("/sys/colleges")
	{
		colleges.GET("", h.ListColleges)
		colleges.POST("", adminOnly, h.CreateCollege)
		colleges.PUT("/:id", adminOnly, h.UpdateCollege)
		colleges.DELETE("/:id", adminOnly, h.DeleteCollege)
	}

	// 专业
	majors := rg.Group("/sys/majors")
	{
		majors.GET("", h.ListMajors)
		majors.POST("", adminOnly, h.CreateMajor)
		majors.PUT("/:id", adminOnly, h.UpdateMajor)
		majors.DELETE("/:id", adminOnly, h.DeleteMajor)
	}

	// 班级
	classes := rg.Group("/sys/classes")
	{
		classes.GET("", h.ListClasses)
		classes.POST("", adminOnly, h.CreateClass)
		classes.PUT("/:id", adminOnly, h.UpdateClass)
		classes.DELETE("/:id", adminOnly, h.DeleteClass)
	}
}

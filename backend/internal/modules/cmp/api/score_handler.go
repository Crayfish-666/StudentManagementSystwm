// Package cmp 综合素质量化 API 处理器。
package api

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	cmpservice "student-system/internal/modules/cmp/service"
	"student-system/pkg/response"
)

// ScoreHandler 综合分查询/重算 API 处理器。
type ScoreHandler struct {
	svc *cmpservice.ScoreService
}

// NewScoreHandler 创建综合分处理器。
func NewScoreHandler(svc *cmpservice.ScoreService) *ScoreHandler {
	return &ScoreHandler{svc: svc}
}

// MyScore 学生本人综合分。GET /api/v1/cmp/scores/me
func (h *ScoreHandler) MyScore(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)
	academicYear := c.Query("term")

	view, err := h.svc.MyScore(userID, academicYear)
	if err != nil {
		code := 1500
		if err.Error() == "当前账户未关联学生身份" {
			code = 1404
		}
		response.Fail(c, code, err.Error())
		return
	}
	response.OK(c, view)
}

// List 排行/列表。GET /api/v1/cmp/scores
func (h *ScoreHandler) List(c *gin.Context) {
	academicYear := c.Query("term")
	var collegeID, classID int64
	if v := c.Query("college_id"); v != "" {
		collegeID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := c.Query("class_id"); v != "" {
		classID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(academicYear, collegeID, classID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询综合分列表失败")
		return
	}
	response.OK(c, result)
}

// Get 详情。GET /api/v1/cmp/scores/:student_id
func (h *ScoreHandler) Get(c *gin.Context) {
	studentID, err := strconv.ParseInt(c.Param("student_id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的学生 ID")
		return
	}
	academicYear := c.Query("term")
	view, err := h.svc.Get(studentID, academicYear)
	if err != nil {
		response.Fail(c, 1404, "综合分记录不存在")
		return
	}
	response.OK(c, view)
}

// RecomputeOne 手动重算。POST /api/v1/cmp/scores/:student_id/recompute
func (h *ScoreHandler) RecomputeOne(c *gin.Context) {
	studentID, err := strconv.ParseInt(c.Param("student_id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的学生 ID")
		return
	}
	academicYear := c.Query("term")
	view, err := h.svc.RecomputeOne(c.Request.Context(), studentID, academicYear)
	if err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, view)
}

// RecomputeBatch 批量重算。POST /api/v1/cmp/scores/compute
func (h *ScoreHandler) RecomputeBatch(c *gin.Context) {
	var req struct {
		CollegeID    int64  `json:"college_id"`
		AcademicYear string `json:"term"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// body 可为空：仅依赖 query
		req.AcademicYear = c.Query("term")
		if v := c.Query("college_id"); v != "" {
			req.CollegeID, _ = strconv.ParseInt(v, 10, 64)
		}
	}

	count, err := h.svc.RecomputeBatch(context.Background(), req.CollegeID, req.AcademicYear)
	if err != nil {
		response.Fail(c, 1500, "批量重算失败")
		return
	}
	response.OK(c, gin.H{
		"recomputed_count": count,
		"academic_year":    req.AcademicYear,
	})
}

// RegisterRoutes 注册综合分路由。
func (h *ScoreHandler) RegisterRoutes(rg *gin.RouterGroup) {
	cmp := rg.Group("/cmp")
	{
		cmp.GET("/scores/me", h.MyScore)
		cmp.GET("/scores", h.List)
		cmp.GET("/scores/:student_id", h.Get)
		cmp.POST("/scores/:student_id/recompute", h.RecomputeOne)
		cmp.POST("/scores/compute", h.RecomputeBatch)
	}
}

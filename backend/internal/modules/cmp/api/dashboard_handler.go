// Package cmp 综合看板 API 处理器。
package api

import (
	"github.com/gin-gonic/gin"

	cmpservice "student-system/internal/modules/cmp/service"
	"student-system/pkg/response"
)

// DashboardHandler 综合看板 API 处理器。
type DashboardHandler struct {
	svc *cmpservice.DashboardService
}

// NewDashboardHandler 创建看板处理器。
func NewDashboardHandler(svc *cmpservice.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// KPI 关键 KPI。GET /api/v1/cmp/dashboard/kpi
func (h *DashboardHandler) KPI(c *gin.Context) {
	academicYear := c.Query("term")
	view, err := h.svc.KPI(academicYear)
	if err != nil {
		response.Fail(c, 1500, "查询 KPI 失败")
		return
	}
	response.OK(c, view)
}

// Trends 趋势图。GET /api/v1/cmp/dashboard/trends
func (h *DashboardHandler) Trends(c *gin.Context) {
	metric := c.DefaultQuery("metric", "ty_pass_rate")
	rangeKey := c.DefaultQuery("range", "12m")
	points, err := h.svc.Trends(metric, rangeKey)
	if err != nil {
		response.Fail(c, 1500, "查询趋势失败")
		return
	}
	response.OK(c, gin.H{
		"metric": metric,
		"range":  rangeKey,
		"points": points,
	})
}

// Distribution 分布。GET /api/v1/cmp/dashboard/distribution
func (h *DashboardHandler) Distribution(c *gin.Context) {
	dim := c.DefaultQuery("dim", "college")
	academicYear := c.Query("term")
	buckets, err := h.svc.Distribution(dim, academicYear)
	if err != nil {
		response.Fail(c, 1500, "查询分布失败")
		return
	}
	response.OK(c, gin.H{
		"dim":     dim,
		"buckets": buckets,
	})
}

// ActiveAssocByCollege 活跃社团按院系分布。
func (h *DashboardHandler) ActiveAssocByCollege(c *gin.Context) {
	buckets, err := h.svc.ActiveAssocByCollege()
	if err != nil {
		response.Fail(c, 1500, "查询活跃社团失败")
		return
	}
	response.OK(c, gin.H{"buckets": buckets})
}

// IncidentLevelDistribution 事件等级分布。
func (h *DashboardHandler) IncidentLevelDistribution(c *gin.Context) {
	buckets, err := h.svc.IncidentLevelDistribution()
	if err != nil {
		response.Fail(c, 1500, "查询事件等级分布失败")
		return
	}
	response.OK(c, gin.H{"buckets": buckets})
}

// RegisterRoutes 注册看板路由。
func (h *DashboardHandler) RegisterRoutes(rg *gin.RouterGroup) {
	d := rg.Group("/cmp/dashboard")
	{
		d.GET("/kpi", h.KPI)
		d.GET("/trends", h.Trends)
		d.GET("/distribution", h.Distribution)
		d.GET("/active-assoc-by-college", h.ActiveAssocByCollege)
		d.GET("/incident-level", h.IncidentLevelDistribution)
	}
}

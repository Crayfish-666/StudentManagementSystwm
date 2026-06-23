// Package api 工作台统计 API 处理器。
package api

import (
	"github.com/gin-gonic/gin"
	"student-system/internal/modules/dashboard/service"
	"student-system/pkg/response"
)

// DashboardHandler 工作台 API 处理器。
type DashboardHandler struct {
	svc *service.DashboardService
}

// NewDashboardHandler 创建工作台处理器。
func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Overview 工作台概览。GET /api/v1/dashboard/overview
func (h *DashboardHandler) Overview(c *gin.Context) {
	uid, _ := c.Get("uid")
	overview, err := h.svc.Overview(uid)
	if err != nil {
		response.Fail(c, 1500, "查询工作台概览失败")
		return
	}
	response.OK(c, overview)
}

// RegisterRoutes 注册路由。
func (h *DashboardHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/dashboard")
	{
		g.GET("/overview", h.Overview)
	}
}

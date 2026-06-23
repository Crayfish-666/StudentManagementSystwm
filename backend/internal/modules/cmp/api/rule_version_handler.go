// Package cmp 综合素质量化规则版本 API 处理器。
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	cmpservice "student-system/internal/modules/cmp/service"
	"student-system/pkg/response"
)

// RuleVersionHandler 规则版本 API 处理器。
type RuleVersionHandler struct {
	svc *cmpservice.RuleVersionService
}

// NewRuleVersionHandler 创建规则版本处理器。
func NewRuleVersionHandler(svc *cmpservice.RuleVersionService) *RuleVersionHandler {
	return &RuleVersionHandler{svc: svc}
}

// List 列出全部版本。GET /api/v1/cmp/rule-versions
func (h *RuleVersionHandler) List(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		response.Fail(c, 1500, "查询规则版本失败")
		return
	}
	response.OK(c, gin.H{"items": items, "total": len(items)})
}

// Create 新建规则版本。POST /api/v1/cmp/rule-versions
func (h *RuleVersionHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req cmpservice.CreateRuleVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}
	view, err := h.svc.Create(userID, &req)
	if err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, view)
}

// Activate 激活。POST /api/v1/cmp/rule-versions/:id/activate
func (h *RuleVersionHandler) Activate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的规则版本 ID")
		return
	}
	if err := h.svc.Activate(id); err != nil {
		response.Fail(c, 1500, "激活规则版本失败")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// RegisterRoutes 注册规则版本路由。
func (h *RuleVersionHandler) RegisterRoutes(rg *gin.RouterGroup, adminOnly ...gin.HandlerFunc) {
	mw := []gin.HandlerFunc{}
	if len(adminOnly) > 0 {
		mw = append(mw, adminOnly...)
	}
	r := rg.Group("/cmp/rule-versions", mw...)
	{
		r.GET("", h.List)
		r.POST("", h.Create)
		r.POST("/:id/activate", h.Activate)
	}
}

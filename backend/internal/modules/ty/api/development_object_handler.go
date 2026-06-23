package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/ty/service"
	"student-system/pkg/response"
)

// DevelopmentObjectHandler 发展对象接口处理器。
type DevelopmentObjectHandler struct {
	svc *service.DevelopmentObjectService
}

// NewDevelopmentObjectHandler 创建发展对象处理器。
func NewDevelopmentObjectHandler(svc *service.DevelopmentObjectService) *DevelopmentObjectHandler {
	return &DevelopmentObjectHandler{svc: svc}
}

// Create 提交发展对象申请。POST /api/v1/ty/development-objects
func (h *DevelopmentObjectHandler) Create(c *gin.Context) {
	userID, name, _, _, _ := actorFromCtx(c)

	var req service.CreateDevelopmentObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	result, err := h.svc.Submit(userID, &req, name, "", "", "")
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "入团申请不存在":
			code = 1404
		default:
			if contains(msg, "错误码:2601") {
				code = 2601
			} else if contains(msg, "错误码:2602") {
				code = 2602
			} else if contains(msg, "错误码:2603") {
				code = 2603
			}
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// List 发展对象列表。GET /api/v1/ty/development-objects
func (h *DevelopmentObjectHandler) List(c *gin.Context) {
	status := c.Query("status")
	var collegeID int64
	if v := c.Query("college_id"); v != "" {
		collegeID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(status, collegeID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询发展对象列表失败")
		return
	}
	response.OK(c, result)
}

// Get 发展对象详情。GET /api/v1/ty/development-objects/:id
func (h *DevelopmentObjectHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的发展对象 ID")
		return
	}

	result, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, "发展对象记录不存在")
		return
	}
	response.OK(c, result)
}

// Publicize 设置公示期。POST /api/v1/ty/development-objects/:id:publicize
func (h *DevelopmentObjectHandler) Publicize(c *gin.Context) {
	userID, _, _, _, _ := actorFromCtx(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的发展对象 ID")
		return
	}

	var req service.PublicizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.Publicize(id, userID, &req, "", "", "", "")
	if err != nil {
		msg := err.Error()
		code := 1500
		if contains(msg, "仅待审状态可设置公示") {
			code = 40901
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// Approve 审批发展对象。POST /api/v1/ty/development-objects/:id:approve
func (h *DevelopmentObjectHandler) Approve(c *gin.Context) {
	userID, _, _, _, _ := actorFromCtx(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的发展对象 ID")
		return
	}

	var req service.ApproveDevelopmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.Approve(id, userID, &req, "", "", "", "")
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "发展对象记录不存在":
			code = 1404
		default:
			if contains(msg, "该步骤已审批通过") || contains(msg, "团支部大会尚未通过") || contains(msg, "院系复核尚未通过") || contains(msg, "校级不可终审") {
				code = 40901
			}
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// RegisterRoutes 注册发展对象相关路由。
func (h *DevelopmentObjectHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	ty := rg.Group("/ty")
	{
		ty.GET("/development-objects", h.List)
		ty.GET("/development-objects/:id", h.Get)
		ty.POST("/development-objects", h.Create)
		ty.POST("/development-objects/:id/publicize", h.Publicize)
		ty.POST("/development-objects/:id/approve", h.Approve)
	}
}

package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/sq/service"
	"student-system/pkg/response"
)

// InspectionHandler 巡查接口处理器。
type InspectionHandler struct {
	svc *service.InspectionService
}

// NewInspectionHandler 创建巡查处理器。
func NewInspectionHandler(svc *service.InspectionService) *InspectionHandler {
	return &InspectionHandler{svc: svc}
}

// List 分页查询巡查列表。GET /api/v1/sq/inspections
func (h *InspectionHandler) List(c *gin.Context) {
	inspectionType := c.Query("inspection_type")
	var buildingID int64
	if v := c.Query("building_id"); v != "" {
		buildingID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(inspectionType, buildingID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询巡查列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取巡查详情。GET /api/v1/sq/inspections/:id
func (h *InspectionHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的巡查 ID")
		return
	}

	insp, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, insp)
}

// Create 创建巡查记录。POST /api/v1/sq/inspections
func (h *InspectionHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	insp, err := h.svc.Create(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "楼栋不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, insp)
}

// Delete 删除巡查记录。DELETE /api/v1/sq/inspections/:id
func (h *InspectionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的巡查 ID")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// RegisterRoutes 注册巡查相关路由。
func (h *InspectionHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	sq := rg.Group("/sq")
	{
		sq.GET("/inspections", h.List)
		sq.GET("/inspections/:id", h.Get)
		sq.POST("/inspections", h.Create)
		sq.DELETE("/inspections/:id", h.Delete)
	}
}

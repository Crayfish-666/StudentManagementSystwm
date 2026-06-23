package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/sq/service"
	"student-system/pkg/response"
)

// IncidentHandler 异常事件接口处理器。
type IncidentHandler struct {
	svc *service.IncidentService
}

// NewIncidentHandler 创建事件处理器。
func NewIncidentHandler(svc *service.IncidentService) *IncidentHandler {
	return &IncidentHandler{svc: svc}
}

// List 分页查询事件列表。GET /api/v1/sq/incidents
func (h *IncidentHandler) List(c *gin.Context) {
	incidentLevel := c.Query("incident_level")
	status := c.Query("status")
	var buildingID int64
	if v := c.Query("building_id"); v != "" {
		buildingID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(incidentLevel, status, buildingID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询事件列表失败")
		return
	}
	response.OK(c, result)
}

// Get 获取事件详情。GET /api/v1/sq/incidents/:id
func (h *IncidentHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的事件 ID")
		return
	}

	inc, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, inc)
}

// Create 上报事件。POST /api/v1/sq/incidents
func (h *IncidentHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	inc, err := h.svc.Create(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		if msg == "楼栋不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, inc)
}

// Handle 处置事件。POST /api/v1/sq/incidents/:id/handle
func (h *IncidentHandler) Handle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的事件 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.HandleIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	inc, err := h.svc.Handle(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "事件不存在" {
			code = 1404
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, inc)
}

// Close 结案。POST /api/v1/sq/incidents/:id/close
func (h *IncidentHandler) Close(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的事件 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CloseIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	inc, err := h.svc.Close(id, userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1409
		if msg == "事件不存在" {
			code = 1404
		} else if msg == "L4 级事件必须由教师结案" {
			code = 40310
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, inc)
}

// Delete 删除事件。DELETE /api/v1/sq/incidents/:id
func (h *IncidentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的事件 ID")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// RegisterRoutes 注册事件相关路由。
func (h *IncidentHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	sq := rg.Group("/sq")
	{
		sq.GET("/incidents", h.List)
		sq.GET("/incidents/:id", h.Get)
		sq.POST("/incidents", h.Create)
		sq.POST("/incidents/:id/handle", h.Handle)
		sq.POST("/incidents/:id/close", h.Close)
		sq.DELETE("/incidents/:id", h.Delete)
	}
}

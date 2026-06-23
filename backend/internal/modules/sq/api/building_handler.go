package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/sq/service"
	"student-system/pkg/response"
)

// BuildingHandler 楼栋/楼层/寝室接口处理器。
type BuildingHandler struct {
	svc *service.BuildingService
}

// NewBuildingHandler 创建楼栋处理器。
func NewBuildingHandler(svc *service.BuildingService) *BuildingHandler {
	return &BuildingHandler{svc: svc}
}

// GetBuildingTree 获取楼栋树形结构。GET /api/v1/sq/buildings/tree
func (h *BuildingHandler) GetBuildingTree(c *gin.Context) {
	tree, err := h.svc.GetBuildingTree()
	if err != nil {
		response.Fail(c, 1500, "查询楼栋树失败")
		return
	}
	response.OK(c, gin.H{"items": tree})
}

// ListBuildings 查询楼栋列表。GET /api/v1/sq/buildings
func (h *BuildingHandler) ListBuildings(c *gin.Context) {
	buildings, err := h.svc.ListBuildings()
	if err != nil {
		response.Fail(c, 1500, "查询楼栋列表失败")
		return
	}
	response.OK(c, gin.H{"items": buildings})
}

// GetBuilding 获取楼栋详情。GET /api/v1/sq/buildings/:id
func (h *BuildingHandler) GetBuilding(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的楼栋 ID")
		return
	}

	building, err := h.svc.GetBuilding(id)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, building)
}

// CreateBuilding 创建楼栋。POST /api/v1/sq/buildings
func (h *BuildingHandler) CreateBuilding(c *gin.Context) {
	var req service.CreateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	building, err := h.svc.CreateBuilding(&req)
	if err != nil {
		response.Fail(c, 1409, err.Error())
		return
	}
	response.OK(c, building)
}

// UpdateBuilding 更新楼栋。PUT /api/v1/sq/buildings/:id
func (h *BuildingHandler) UpdateBuilding(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的楼栋 ID")
		return
	}

	var req service.UpdateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	building, err := h.svc.UpdateBuilding(id, &req)
	if err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, building)
}

// DeleteBuilding 删除楼栋。DELETE /api/v1/sq/buildings/:id
func (h *BuildingHandler) DeleteBuilding(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的楼栋 ID")
		return
	}

	if err := h.svc.DeleteBuilding(id); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ListFloors 查询楼层列表。GET /api/v1/sq/buildings/:id/floors
func (h *BuildingHandler) ListFloors(c *gin.Context) {
	buildingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的楼栋 ID")
		return
	}

	floors, err := h.svc.ListFloors(buildingID)
	if err != nil {
		response.Fail(c, 1500, "查询楼层列表失败")
		return
	}
	response.OK(c, gin.H{"items": floors})
}

// CreateFloor 创建楼层。POST /api/v1/sq/floors
func (h *BuildingHandler) CreateFloor(c *gin.Context) {
	var req service.CreateFloorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	floor, err := h.svc.CreateFloor(&req)
	if err != nil {
		response.Fail(c, 1409, err.Error())
		return
	}
	response.OK(c, floor)
}

// DeleteFloor 删除楼层。DELETE /api/v1/sq/floors/:id
func (h *BuildingHandler) DeleteFloor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的楼层 ID")
		return
	}

	if err := h.svc.DeleteFloor(id); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ListRooms 查询寝室列表。GET /api/v1/sq/buildings/:id/rooms
func (h *BuildingHandler) ListRooms(c *gin.Context) {
	buildingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的楼栋 ID")
		return
	}

	rooms, err := h.svc.ListRooms(buildingID)
	if err != nil {
		response.Fail(c, 1500, "查询寝室列表失败")
		return
	}
	response.OK(c, gin.H{"items": rooms})
}

// CreateRoom 创建寝室。POST /api/v1/sq/rooms
func (h *BuildingHandler) CreateRoom(c *gin.Context) {
	var req service.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	room, err := h.svc.CreateRoom(&req)
	if err != nil {
		response.Fail(c, 1409, err.Error())
		return
	}
	response.OK(c, room)
}

// DeleteRoom 删除寝室。DELETE /api/v1/sq/rooms/:id
func (h *BuildingHandler) DeleteRoom(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的寝室 ID")
		return
	}

	if err := h.svc.DeleteRoom(id); err != nil {
		response.Fail(c, 1500, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// GetRoomMembers 查询寝室入住成员。GET /api/v1/sq/rooms/:id/members
func (h *BuildingHandler) GetRoomMembers(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的寝室 ID")
		return
	}

	members, err := h.svc.GetRoomMembers(roomID)
	if err != nil {
		response.Fail(c, 1500, "查询入住成员失败")
		return
	}
	response.OK(c, gin.H{"items": members})
}

// RegisterRoutes 注册楼栋相关路由。
func (h *BuildingHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	sq := rg.Group("/sq")
	{
		// 楼栋
		sq.GET("/buildings/tree", h.GetBuildingTree)
		sq.GET("/buildings", h.ListBuildings)
		sq.GET("/buildings/:id", h.GetBuilding)
		sq.POST("/buildings", h.CreateBuilding)
		sq.PUT("/buildings/:id", h.UpdateBuilding)
		sq.DELETE("/buildings/:id", h.DeleteBuilding)

		// 楼层
		sq.GET("/buildings/:id/floors", h.ListFloors)
		sq.POST("/floors", h.CreateFloor)
		sq.DELETE("/floors/:id", h.DeleteFloor)

		// 寝室
		sq.GET("/buildings/:id/rooms", h.ListRooms)
		sq.POST("/rooms", h.CreateRoom)
		sq.DELETE("/rooms/:id", h.DeleteRoom)
		sq.GET("/rooms/:id/members", h.GetRoomMembers)
	}
}

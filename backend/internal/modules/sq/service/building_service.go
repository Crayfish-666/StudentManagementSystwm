package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
	"student-system/internal/modules/sq/repository"
)

// BuildingService 楼栋/楼层/寝室业务服务层。
type BuildingService struct {
	repo *repository.BuildingRepository
	db   *gorm.DB
}

// NewBuildingService 创建楼栋服务。
func NewBuildingService(repo *repository.BuildingRepository, db *gorm.DB) *BuildingService {
	return &BuildingService{repo: repo, db: db}
}

// ---- DTO ----

// BuildingTreeNode 楼栋树节点。
type BuildingTreeNode struct {
	ID       int64             `json:"id"`
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	Type     string            `json:"type"` // building / floor / room
	FloorNo  *int              `json:"floor_no,omitempty"`
	RoomNo   *string           `json:"room_no,omitempty"`
	BedCount *int              `json:"bed_count,omitempty"`
	Children []BuildingTreeNode `json:"children,omitempty"`
}

// BuildingView 楼栋视图。
type BuildingView struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	FloorCount  int    `json:"floor_count"`
	TutorUserID *int64 `json:"tutor_user_id,omitempty"`
	TutorName   string `json:"tutor_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// FloorView 楼层视图。
type FloorView struct {
	ID                   int64  `json:"id"`
	BuildingID           int64  `json:"building_id"`
	FloorNo              int    `json:"floor_no"`
	FloorLeaderStudentID *int64 `json:"floor_leader_student_id,omitempty"`
	FloorLeaderName      string `json:"floor_leader_name"`
	CreatedAt            string `json:"created_at"`
}

// RoomView 寝室视图。
type RoomView struct {
	ID                  int64  `json:"id"`
	BuildingID          int64  `json:"building_id"`
	FloorID             int64  `json:"floor_id"`
	RoomNo              string `json:"room_no"`
	BedCount            int    `json:"bed_count"`
	RoomLeaderStudentID *int64 `json:"room_leader_student_id,omitempty"`
	RoomLeaderName      string `json:"room_leader_name"`
	CreatedAt           string `json:"created_at"`
}

// CreateBuildingRequest 创建楼栋请求。
type CreateBuildingRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	FloorCount  int    `json:"floor_count"`
	TutorUserID *int64 `json:"tutor_user_id"`
}

// UpdateBuildingRequest 更新楼栋请求。
type UpdateBuildingRequest struct {
	Code        *string `json:"code"`
	Name        *string `json:"name"`
	FloorCount  *int    `json:"floor_count"`
	TutorUserID *int64  `json:"tutor_user_id"`
}

// CreateFloorRequest 创建楼层请求。
type CreateFloorRequest struct {
	BuildingID           int64  `json:"building_id" binding:"required"`
	FloorNo              int    `json:"floor_no" binding:"required"`
	FloorLeaderStudentID *int64 `json:"floor_leader_student_id"`
}

// CreateRoomRequest 创建寝室请求。
type CreateRoomRequest struct {
	BuildingID           int64  `json:"building_id" binding:"required"`
	FloorID              int64  `json:"floor_id" binding:"required"`
	RoomNo               string `json:"room_no" binding:"required"`
	BedCount             int    `json:"bed_count"`
	RoomLeaderStudentID *int64 `json:"room_leader_student_id"`
}

// ---- 业务方法 ----

// GetBuildingTree 获取楼栋树形结构。
func (s *BuildingService) GetBuildingTree() ([]BuildingTreeNode, error) {
	buildings, err := s.repo.ListBuildings()
	if err != nil {
		return nil, err
	}

	nodes := make([]BuildingTreeNode, 0, len(buildings))
	for _, b := range buildings {
		bNode := BuildingTreeNode{
			ID:   b.ID,
			Code: b.Code,
			Name: b.Name,
			Type: "building",
		}

		// 加载楼层
		floors, err := s.repo.ListFloorsByBuilding(b.ID)
		if err != nil {
			floors = nil
		}

		floorNodes := make([]BuildingTreeNode, 0, len(floors))
		for _, f := range floors {
			fNode := BuildingTreeNode{
				ID:      f.ID,
				Name:    fmt.Sprintf("%d\u5c42", f.FloorNo),
				Type:    "floor",
				FloorNo: &f.FloorNo,
			}

			// 加载寝室
			rooms, err := s.repo.ListRoomsByFloor(f.ID)
			if err != nil {
				rooms = nil
			}

			roomNodes := make([]BuildingTreeNode, 0, len(rooms))
			for _, r := range rooms {
				rNode := BuildingTreeNode{
					ID:       r.ID,
					Name:     r.RoomNo,
					Type:     "room",
					RoomNo:   &r.RoomNo,
					BedCount: &r.BedCount,
				}
				roomNodes = append(roomNodes, rNode)
			}
			fNode.Children = roomNodes
			floorNodes = append(floorNodes, fNode)
		}
		bNode.Children = floorNodes
		nodes = append(nodes, bNode)
	}

	return nodes, nil
}

// ListBuildings 查询楼栋列表。
func (s *BuildingService) ListBuildings() ([]BuildingView, error) {
	buildings, err := s.repo.ListBuildings()
	if err != nil {
		return nil, err
	}

	views := make([]BuildingView, 0, len(buildings))
	for _, b := range buildings {
		v := BuildingView{
			ID:          b.ID,
			Code:        b.Code,
			Name:        b.Name,
			FloorCount:  b.FloorCount,
			TutorUserID: b.TutorUserID,
			CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
		}
		if b.TutorUserID != nil {
			if user, err := s.repo.GetUserByID(*b.TutorUserID); err == nil {
				v.TutorName = user.DisplayName
			}
		}
		views = append(views, v)
	}
	return views, nil
}

// GetBuilding 获取楼栋详情。
func (s *BuildingService) GetBuilding(id int64) (*BuildingView, error) {
	b, err := s.repo.GetBuilding(id)
	if err != nil {
		return nil, fmt.Errorf("楼栋不存在")
	}
	v := BuildingView{
		ID:          b.ID,
		Code:        b.Code,
		Name:        b.Name,
		FloorCount:  b.FloorCount,
		TutorUserID: b.TutorUserID,
		CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if b.TutorUserID != nil {
		if user, err := s.repo.GetUserByID(*b.TutorUserID); err == nil {
			v.TutorName = user.DisplayName
		}
	}
	return &v, nil
}

// CreateBuilding 创建楼栋。
func (s *BuildingService) CreateBuilding(req *CreateBuildingRequest) (*BuildingView, error) {
	// 校验 code 唯一
	count, err := s.repo.CountBuildingByCode(req.Code, 0)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("楼栋编码已存在")
	}

	bedCount := req.FloorCount
	if bedCount < 0 {
		bedCount = 0
	}

	b := &models.IdxDormBuilding{
		Code:        req.Code,
		Name:        req.Name,
		FloorCount:  bedCount,
		TutorUserID: req.TutorUserID,
	}
	if err := s.repo.CreateBuilding(b); err != nil {
		return nil, err
	}
	return s.GetBuilding(b.ID)
}

// UpdateBuilding 更新楼栋。
func (s *BuildingService) UpdateBuilding(id int64, req *UpdateBuildingRequest) (*BuildingView, error) {
	b, err := s.repo.GetBuilding(id)
	if err != nil {
		return nil, fmt.Errorf("楼栋不存在")
	}

	if req.Code != nil {
		count, err := s.repo.CountBuildingByCode(*req.Code, id)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("楼栋编码已存在")
		}
		b.Code = *req.Code
	}
	if req.Name != nil {
		b.Name = *req.Name
	}
	if req.FloorCount != nil {
		b.FloorCount = *req.FloorCount
	}
	if req.TutorUserID != nil {
		b.TutorUserID = req.TutorUserID
	}

	if err := s.repo.UpdateBuilding(b); err != nil {
		return nil, err
	}
	return s.GetBuilding(b.ID)
}

// DeleteBuilding 删除楼栋。
func (s *BuildingService) DeleteBuilding(id int64) error {
	return s.repo.SoftDeleteBuilding(id)
}

// ListFloors 查询楼层列表。
func (s *BuildingService) ListFloors(buildingID int64) ([]FloorView, error) {
	floors, err := s.repo.ListFloorsByBuilding(buildingID)
	if err != nil {
		return nil, err
	}

	views := make([]FloorView, 0, len(floors))
	for _, f := range floors {
		v := FloorView{
			ID:                   f.ID,
			BuildingID:           f.BuildingID,
			FloorNo:              f.FloorNo,
			FloorLeaderStudentID: f.FloorLeaderStudentID,
			CreatedAt:            f.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		}
		if f.FloorLeaderStudentID != nil {
			if student, err := s.repo.GetStudentByID(*f.FloorLeaderStudentID); err == nil {
				v.FloorLeaderName = student.Name
			}
		}
		views = append(views, v)
	}
	return views, nil
}

// CreateFloor 创建楼层。
func (s *BuildingService) CreateFloor(req *CreateFloorRequest) (*FloorView, error) {
	// 校验楼栋存在
	if _, err := s.repo.GetBuilding(req.BuildingID); err != nil {
		return nil, fmt.Errorf("楼栋不存在")
	}

	f := &models.IdxDormFloor{
		BuildingID:           req.BuildingID,
		FloorNo:              req.FloorNo,
		FloorLeaderStudentID: req.FloorLeaderStudentID,
	}
	if err := s.repo.CreateFloor(f); err != nil {
		return nil, err
	}

	// 更新楼栋楼层计数
	s.updateBuildingFloorCount(req.BuildingID)

	floors, _ := s.repo.ListFloorsByBuilding(req.BuildingID)
	for _, fl := range floors {
		if fl.FloorNo == req.FloorNo {
			v := FloorView{
				ID:                   fl.ID,
				BuildingID:           fl.BuildingID,
				FloorNo:              fl.FloorNo,
				FloorLeaderStudentID: fl.FloorLeaderStudentID,
				CreatedAt:            fl.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			}
			if fl.FloorLeaderStudentID != nil {
				if student, err := s.repo.GetStudentByID(*fl.FloorLeaderStudentID); err == nil {
					v.FloorLeaderName = student.Name
				}
			}
			return &v, nil
		}
	}
	return nil, fmt.Errorf("创建楼层失败")
}

// DeleteFloor 删除楼层。
func (s *BuildingService) DeleteFloor(id int64) error {
	f, err := s.repo.GetFloor(id)
	if err != nil {
		return fmt.Errorf("楼层不存在")
	}
	if err := s.repo.SoftDeleteFloor(id); err != nil {
		return err
	}
	s.updateBuildingFloorCount(f.BuildingID)
	return nil
}

// ListRooms 查询寝室列表。
func (s *BuildingService) ListRooms(buildingID int64) ([]RoomView, error) {
	rooms, err := s.repo.ListRoomsByBuilding(buildingID)
	if err != nil {
		return nil, err
	}

	views := make([]RoomView, 0, len(rooms))
	for _, r := range rooms {
		v := RoomView{
			ID:                  r.ID,
			BuildingID:          r.BuildingID,
			FloorID:             r.FloorID,
			RoomNo:              r.RoomNo,
			BedCount:            r.BedCount,
			RoomLeaderStudentID: r.RoomLeaderStudentID,
			CreatedAt:           r.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		}
		if r.RoomLeaderStudentID != nil {
			if student, err := s.repo.GetStudentByID(*r.RoomLeaderStudentID); err == nil {
				v.RoomLeaderName = student.Name
			}
		}
		views = append(views, v)
	}
	return views, nil
}

// CreateRoom 创建寝室。
func (s *BuildingService) CreateRoom(req *CreateRoomRequest) (*RoomView, error) {
	// 校验楼栋存在
	if _, err := s.repo.GetBuilding(req.BuildingID); err != nil {
		return nil, fmt.Errorf("楼栋不存在")
	}

	// 校验楼层存在
	if _, err := s.repo.GetFloor(req.FloorID); err != nil {
		return nil, fmt.Errorf("楼层不存在")
	}

	// 校验房间号唯一
	count, err := s.repo.CountRoomByNo(req.BuildingID, req.RoomNo, 0)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("该楼栋下房间号已存在")
	}

	bedCount := req.BedCount
	if bedCount <= 0 {
		bedCount = 4
	}

	room := &models.IdxDormRoom{
		BuildingID:          req.BuildingID,
		FloorID:             req.FloorID,
		RoomNo:              req.RoomNo,
		BedCount:            bedCount,
		RoomLeaderStudentID: req.RoomLeaderStudentID,
	}
	if err := s.repo.CreateRoom(room); err != nil {
		return nil, err
	}

	// 自动创建床位
	for i := 1; i <= bedCount; i++ {
		bed := &models.IdxDormBed{
			RoomID: room.ID,
			BedNo:  fmt.Sprintf("%d", i),
		}
		_ = s.repo.CreateBed(bed)
	}

	v := RoomView{
		ID:                  room.ID,
		BuildingID:          room.BuildingID,
		FloorID:             room.FloorID,
		RoomNo:              room.RoomNo,
		BedCount:            room.BedCount,
		RoomLeaderStudentID: room.RoomLeaderStudentID,
		CreatedAt:           room.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if room.RoomLeaderStudentID != nil {
		if student, err := s.repo.GetStudentByID(*room.RoomLeaderStudentID); err == nil {
			v.RoomLeaderName = student.Name
		}
	}
	return &v, nil
}

// DeleteRoom 删除寝室。
func (s *BuildingService) DeleteRoom(id int64) error {
	return s.repo.SoftDeleteRoom(id)
}

// GetRoomMembers 查询寝室入住成员。
func (s *BuildingService) GetRoomMembers(roomID int64) ([]map[string]interface{}, error) {
	beds, err := s.repo.ListRoomMembers(roomID)
	if err != nil {
		return nil, err
	}

	members := make([]map[string]interface{}, 0, len(beds))
	for _, bed := range beds {
		m := map[string]interface{}{
			"bed_id":    bed.ID,
			"bed_no":    bed.BedNo,
			"student_id": bed.OccupantStudentID,
			"move_in_at": nil,
		}
		if bed.MoveInAt != nil {
			m["move_in_at"] = bed.MoveInAt.Format("2006-01-02")
		}
		if bed.OccupantStudentID != nil {
			if student, err := s.repo.GetStudentByID(*bed.OccupantStudentID); err == nil {
				m["student_no"] = student.StudentNo
				m["student_name"] = student.Name
			}
		}
		members = append(members, m)
	}
	return members, nil
}

// ---- 内部方法 ----

// updateBuildingFloorCount 更新楼栋楼层计数。
func (s *BuildingService) updateBuildingFloorCount(buildingID int64) {
	floors, err := s.repo.ListFloorsByBuilding(buildingID)
	if err != nil {
		return
	}
	b, err := s.repo.GetBuilding(buildingID)
	if err != nil {
		return
	}
	b.FloorCount = len(floors)
	_ = s.repo.UpdateBuilding(b)
}

// formatTime 格式化时间。
func formatTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05+08:00")
}

package service

import (
	"fmt"

	"gorm.io/gorm"

	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/sq/repository"
)

// InspectionService 巡查业务服务层。
type InspectionService struct {
	repo *repository.InspectionRepository
	db   *gorm.DB
}

// NewInspectionService 创建巡查服务。
func NewInspectionService(repo *repository.InspectionRepository, db *gorm.DB) *InspectionService {
	return &InspectionService{repo: repo, db: db}
}

// ---- DTO ----

// InspectionListResult 巡查列表结果。
type InspectionListResult struct {
	Items    []InspectionView `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// InspectionView 巡查视图。
type InspectionView struct {
	ID              int64           `json:"id"`
	BizNo           string          `json:"biz_no"`
	InspectionType  string          `json:"inspection_type"`
	InspectionTypeText string       `json:"inspection_type_text"`
	BuildingID      int64           `json:"building_id"`
	BuildingName    string          `json:"building_name"`
	FloorID         *int64          `json:"floor_id,omitempty"`
	FloorNo         *int            `json:"floor_no,omitempty"`
	RoomID          *int64          `json:"room_id,omitempty"`
	RoomNo          *string         `json:"room_no,omitempty"`
	InspectorUserID int64           `json:"inspector_user_id"`
	InspectorName   string          `json:"inspector_name"`
	InspectedAt     string          `json:"inspected_at"`
	Score           *int            `json:"score,omitempty"`
	Summary         string          `json:"summary"`
	Status          string          `json:"status"`
	StatusText      string          `json:"status_text"`
	Deductions      []DeductionView `json:"deductions,omitempty"`
	CreatedAt       string          `json:"created_at"`
}

// DeductionView 扣分项视图。
type DeductionView struct {
	ID          int64  `json:"id"`
	Item        string `json:"item"`
	Deduction   int    `json:"deduction"`
	PhotoFileID *int64 `json:"photo_file_id,omitempty"`
}

// CreateInspectionRequest 创建巡查请求。
type CreateInspectionRequest struct {
	InspectionType  string            `json:"inspection_type" binding:"required"`
	BuildingID      int64             `json:"building_id" binding:"required"`
	FloorID         *int64            `json:"floor_id"`
	RoomID          *int64            `json:"room_id"`
	InspectedAt     string            `json:"inspected_at" binding:"required"`
	Score           *int              `json:"score"`
	Summary         string            `json:"summary"`
	Deductions      []DeductionInput  `json:"deductions"`
}

// DeductionInput 扣分项输入。
type DeductionInput struct {
	Item        string `json:"item" binding:"required"`
	Deduction   int    `json:"deduction" binding:"required"`
	PhotoFileID *int64 `json:"photo_file_id"`
}

// ---- 状态映射 ----

var inspectionTypeTextMap = map[string]string{
	"hygiene":     "卫生巡查",
	"late_return": "晚归检查",
	"appliance":   "违规电器",
	"safety":      "安全隐患",
	"fire_lane":   "消防通道",
}

var inspectionStatusTextMap = map[string]string{
	"draft":    "草稿",
	"submitted": "已提交",
}

// ---- 业务方法 ----

// List 分页查询巡查列表。
func (s *InspectionService) List(inspectionType string, buildingID int64, page, pageSize int) (*InspectionListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	inspections, total, err := s.repo.List(inspectionType, buildingID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]InspectionView, 0, len(inspections))
	for _, insp := range inspections {
		v := s.toView(insp)
		items = append(items, v)
	}

	return &InspectionListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取巡查详情（含扣分项）。
func (s *InspectionService) Get(id int64) (*InspectionView, error) {
	insp, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("巡查记录不存在")
	}

	v := s.toView(*insp)

	// 加载扣分项
	deductions, err := s.repo.ListDeductionsByInspection(id)
	if err != nil {
		deductions = nil
	}
	v.Deductions = make([]DeductionView, 0, len(deductions))
	for _, d := range deductions {
		v.Deductions = append(v.Deductions, DeductionView{
			ID:          d.ID,
			Item:        d.Item,
			Deduction:   d.Deduction,
			PhotoFileID: d.PhotoFileID,
		})
	}

	return &v, nil
}

// Create 创建巡查记录。
func (s *InspectionService) Create(userID int64, req *CreateInspectionRequest) (*InspectionView, error) {
	// 校验楼栋存在
	if _, err := s.repo.GetBuildingByID(req.BuildingID); err != nil {
		return nil, fmt.Errorf("楼栋不存在")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "SQ")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// 解析巡查时间
	inspectedAt, err := parseTime(req.InspectedAt)
	if err != nil {
		return nil, fmt.Errorf("巡查时间格式错误")
	}

	// 计算总分（如有扣分项）
	score := req.Score
	if len(req.Deductions) > 0 {
		totalDeduction := 0
		for _, d := range req.Deductions {
			totalDeduction += d.Deduction
		}
		calculated := 100 - totalDeduction
		if calculated < 0 {
			calculated = 0
		}
		score = &calculated
	}

	insp := &models.SqInspection{
		BizNo:           bizNo,
		InspectionType:  req.InspectionType,
		BuildingID:      req.BuildingID,
		FloorID:         req.FloorID,
		RoomID:          req.RoomID,
		InspectorUserID: userID,
		InspectedAt:     inspectedAt,
		Score:           score,
		Summary:         req.Summary,
		Status:          "submitted",
	}

	// 事务：创建巡查 + 扣分项
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(insp).Error; err != nil {
			return err
		}
		for _, d := range req.Deductions {
			deduction := &models.SqInspectionDeduction{
				InspectionID: insp.ID,
				Item:         d.Item,
				Deduction:    d.Deduction,
				PhotoFileID:  d.PhotoFileID,
			}
			if err := tx.Create(deduction).Error; err != nil {
				return fmt.Errorf("创建扣分项失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.Get(insp.ID)
}

// Delete 删除巡查记录。
func (s *InspectionService) Delete(id int64) error {
	return s.repo.SoftDelete(id)
}

// ---- 内部方法 ----

func (s *InspectionService) toView(insp models.SqInspection) InspectionView {
	v := InspectionView{
		ID:               insp.ID,
		BizNo:            insp.BizNo,
		InspectionType:   insp.InspectionType,
		InspectionTypeText: inspectionTypeTextMap[insp.InspectionType],
		BuildingID:       insp.BuildingID,
		FloorID:          insp.FloorID,
		RoomID:           insp.RoomID,
		InspectorUserID:  insp.InspectorUserID,
		InspectedAt:      insp.InspectedAt.Format("2006-01-02T15:04:05+08:00"),
		Score:            insp.Score,
		Summary:          insp.Summary,
		Status:           insp.Status,
		StatusText:       inspectionStatusTextMap[insp.Status],
		CreatedAt:        insp.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 加载楼栋名称
	if b, err := s.repo.GetBuildingByID(insp.BuildingID); err == nil {
		v.BuildingName = b.Name
	}

	// 加载楼层号
	if insp.FloorID != nil {
		if f, err := s.repo.GetFloorByID(*insp.FloorID); err == nil {
			v.FloorNo = &f.FloorNo
		}
	}

	// 加载房间号
	if insp.RoomID != nil {
		if r, err := s.repo.GetRoomByID(*insp.RoomID); err == nil {
			v.RoomNo = &r.RoomNo
		}
	}

	// 加载巡查人姓名
	if u, err := s.repo.GetUserByID(insp.InspectorUserID); err == nil {
		v.InspectorName = u.DisplayName
	}

	return v
}

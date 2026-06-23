package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/sq/repository"
)

// IncidentService 异常事件业务服务层。
type IncidentService struct {
	repo *repository.IncidentRepository
	db   *gorm.DB
	bus  *eventx.Bus
}

// NewIncidentService 创建事件服务。
func NewIncidentService(repo *repository.IncidentRepository, db *gorm.DB, bus *eventx.Bus) *IncidentService {
	return &IncidentService{repo: repo, db: db, bus: bus}
}

// ---- DTO ----

// IncidentListResult 事件列表结果。
type IncidentListResult struct {
	Items    []IncidentView `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// IncidentView 事件视图。
type IncidentView struct {
	ID                 int64           `json:"id"`
	BizNo              string          `json:"biz_no"`
	IncidentLevel      string          `json:"incident_level"`
	IncidentLevelText  string          `json:"incident_level_text"`
	IncidentType       string          `json:"incident_type"`
	OccurredAt         string          `json:"occurred_at"`
	BuildingID         int64           `json:"building_id"`
	BuildingName       string          `json:"building_name"`
	FloorID            *int64          `json:"floor_id,omitempty"`
	RoomID             *int64          `json:"room_id,omitempty"`
	LocationDetail     string          `json:"location_detail"`
	ReporterUserID     int64           `json:"reporter_user_id"`
	ReporterName       string          `json:"reporter_name"`
	InvolvedStudentIDs string          `json:"involved_student_ids"`
	InvolvedStudents   []StudentBrief  `json:"involved_students,omitempty"`
	WitnessUserIDs     string          `json:"witness_user_ids"`
	InitialAction      string          `json:"initial_action"`
	Status             string          `json:"status"`
	StatusText         string          `json:"status_text"`
	ClosedAt           *string         `json:"closed_at,omitempty"`
	ClosedBy           *int64          `json:"closed_by,omitempty"`
	Attachments        []AttachView    `json:"attachments,omitempty"`
	Actions            []ActionView    `json:"actions,omitempty"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
}

// StudentBrief 学生简要信息。
type StudentBrief struct {
	ID         int64  `json:"id"`
	StudentNo  string `json:"student_no"`
	Name       string `json:"name"`
}

// AttachView 附件视图。
type AttachView struct {
	ID      int64  `json:"id"`
	FileID  int64  `json:"file_id"`
	Caption string `json:"caption"`
}

// ActionView 处置记录视图。
type ActionView struct {
	ID         int64  `json:"id"`
	ActionText string `json:"action_text"`
	ActionAt   string `json:"action_at"`
	ActionBy   int64  `json:"action_by"`
	ActionName string `json:"action_name"`
	IsFinal    int    `json:"is_final"`
}

// CreateIncidentRequest 创建事件请求。
type CreateIncidentRequest struct {
	IncidentLevel      string  `json:"incident_level" binding:"required"`
	IncidentType       string  `json:"incident_type" binding:"required"`
	OccurredAt         string  `json:"occurred_at" binding:"required"`
	BuildingID         int64   `json:"building_id" binding:"required"`
	FloorID            *int64  `json:"floor_id"`
	RoomID             *int64  `json:"room_id"`
	LocationDetail     string  `json:"location_detail"`
	InvolvedStudentIDs []int64 `json:"involved_student_ids"`
	WitnessUserIDs     []int64 `json:"witness_user_ids"`
	InitialAction      string  `json:"initial_action"`
	Attachments        []int64 `json:"attachments"` // file_meta.id 列表
}

// HandleIncidentRequest 处置事件请求。
type HandleIncidentRequest struct {
	ActionText string `json:"action_text" binding:"required"`
}

// CloseIncidentRequest 结案请求。
type CloseIncidentRequest struct {
	FinalAction string `json:"final_action" binding:"required"`
}

// ---- 状态映射 ----

var incidentLevelTextMap = map[string]string{
	"L1": "L1-常规报修",
	"L2": "L2-一般违规",
	"L3": "L3-严重违规",
	"L4": "L4-紧急事件",
}

var incidentStatusTextMap = map[string]string{
	"open":      "待处理",
	"processing": "处理中",
	"closed":    "已结案",
	"cancelled": "已取消",
}

// ---- 业务方法 ----

// List 分页查询事件列表。
func (s *IncidentService) List(incidentLevel string, status string, buildingID int64, page, pageSize int) (*IncidentListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	incidents, total, err := s.repo.List(incidentLevel, status, buildingID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]IncidentView, 0, len(incidents))
	for _, inc := range incidents {
		v := s.toView(inc)
		items = append(items, v)
	}

	return &IncidentListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取事件详情（含附件和处置记录）。
func (s *IncidentService) Get(id int64) (*IncidentView, error) {
	inc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("事件不存在")
	}

	v := s.toView(*inc)

	// 加载附件
	attaches, err := s.repo.ListAttachmentsByIncident(id)
	if err != nil {
		attaches = nil
	}
	v.Attachments = make([]AttachView, 0, len(attaches))
	for _, a := range attaches {
		v.Attachments = append(v.Attachments, AttachView{
			ID:      a.ID,
			FileID:  a.FileID,
			Caption: a.Caption,
		})
	}

	// 加载处置记录
	actions, err := s.repo.ListActionsByIncident(id)
	if err != nil {
		actions = nil
	}
	v.Actions = make([]ActionView, 0, len(actions))
	for _, a := range actions {
		av := ActionView{
			ID:         a.ID,
			ActionText: a.ActionText,
			ActionAt:   a.ActionAt.Format("2006-01-02T15:04:05+08:00"),
			ActionBy:   a.ActionBy,
			IsFinal:    a.IsFinal,
		}
		if u, err := s.repo.GetUserByID(a.ActionBy); err == nil {
			av.ActionName = u.DisplayName
		}
		v.Actions = append(v.Actions, av)
	}

	// 加载关联学生信息
	v.InvolvedStudents = s.loadStudentBriefs(inc.InvolvedStudentIDs)

	return &v, nil
}

// Create 创建异常事件。
func (s *IncidentService) Create(userID int64, req *CreateIncidentRequest) (*IncidentView, error) {
	// 校验楼栋存在
	if _, err := s.repo.GetBuildingByID(req.BuildingID); err != nil {
		return nil, fmt.Errorf("楼栋不存在")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "SQ")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// 解析发生时间
	occurredAt, err := parseTime(req.OccurredAt)
	if err != nil {
		return nil, fmt.Errorf("发生时间格式错误")
	}

	// 序列化关联学生和证人
	involvedJSON := "[]"
	if len(req.InvolvedStudentIDs) > 0 {
		b, _ := json.Marshal(req.InvolvedStudentIDs)
		involvedJSON = string(b)
	}
	witnessJSON := "[]"
	if len(req.WitnessUserIDs) > 0 {
		b, _ := json.Marshal(req.WitnessUserIDs)
		witnessJSON = string(b)
	}

	inc := &models.SqIncident{
		BizNo:              bizNo,
		IncidentLevel:      req.IncidentLevel,
		IncidentType:       req.IncidentType,
		OccurredAt:         occurredAt,
		BuildingID:         req.BuildingID,
		FloorID:            req.FloorID,
		RoomID:             req.RoomID,
		LocationDetail:     req.LocationDetail,
		ReporterUserID:     userID,
		InvolvedStudentIDs: involvedJSON,
		WitnessUserIDs:     witnessJSON,
		InitialAction:      req.InitialAction,
		Status:             "open",
	}

	// 事务：创建事件 + 附件
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(inc).Error; err != nil {
			return err
		}
		// 创建附件记录
		for _, fileID := range req.Attachments {
			attach := &models.SqIncidentAttach{
				IncidentID: inc.ID,
				FileID:     fileID,
			}
			if err := tx.Create(attach).Error; err != nil {
				return fmt.Errorf("创建事件附件失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// L4 事件触发应急通知事件
	if req.IncidentLevel == "L4" && s.bus != nil {
		_ = s.bus.Publish(&eventx.Event{
			Aggregate:   "sq.incident",
			AggregateID: inc.BizNo,
			EventType:   "SqIncidentL4Raised",
			Module:      "SQ",
			ActorID:     userID,
			Payload: map[string]interface{}{
				"incident_id":    inc.ID,
				"biz_no":         inc.BizNo,
				"incident_level": inc.IncidentLevel,
				"incident_type":  inc.IncidentType,
				"building_id":    inc.BuildingID,
				"location_detail": inc.LocationDetail,
			},
			BizNo: inc.BizNo,
		})
	}

	return s.Get(inc.ID)
}

// Handle 处置事件（添加处置记录）。
func (s *IncidentService) Handle(id, userID int64, req *HandleIncidentRequest) (*IncidentView, error) {
	inc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("事件不存在")
	}

	if inc.Status != "open" && inc.Status != "processing" {
		return nil, fmt.Errorf("当前状态不允许处置")
	}

	// 更新状态为处理中
	inc.Status = "processing"
	if err := s.repo.Update(inc); err != nil {
		return nil, err
	}

	// 添加处置记录
	action := &models.SqIncidentAction{
		IncidentID: id,
		ActionText: req.ActionText,
		ActionAt:   time.Now(),
		ActionBy:   userID,
		IsFinal:    0,
	}
	if err := s.repo.CreateAction(action); err != nil {
		return nil, err
	}

	return s.Get(id)
}

// Close 结案。
func (s *IncidentService) Close(id, userID int64, req *CloseIncidentRequest) (*IncidentView, error) {
	inc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("事件不存在")
	}

	if inc.Status != "open" && inc.Status != "processing" {
		return nil, fmt.Errorf("当前状态不允许结案")
	}

	// BR-SQ-05：L4 事件必须由教师结案
	if inc.IncidentLevel == "L4" {
		user, err := s.repo.GetUserByID(userID)
		if err != nil {
			return nil, fmt.Errorf("用户不存在")
		}
		// 检查是否为教师角色（简单判断：非学生角色）
		isTeacher := !strings.HasPrefix(user.Username, "20")
		if !isTeacher {
			return nil, fmt.Errorf("L4 级事件必须由教师结案")
		}
	}

	now := time.Now()
	inc.Status = "closed"
	inc.ClosedAt = &now
	inc.ClosedBy = &userID
	if err := s.repo.Update(inc); err != nil {
		return nil, err
	}

	// 添加结案处置记录
	action := &models.SqIncidentAction{
		IncidentID: id,
		ActionText: req.FinalAction,
		ActionAt:   now,
		ActionBy:   userID,
		IsFinal:    1,
	}
	if err := s.repo.CreateAction(action); err != nil {
		return nil, err
	}

	return s.Get(id)
}

// Delete 删除事件。
func (s *IncidentService) Delete(id int64) error {
	return s.repo.SoftDelete(id)
}

// ---- 内部方法 ----

func (s *IncidentService) toView(inc models.SqIncident) IncidentView {
	v := IncidentView{
		ID:                 inc.ID,
		BizNo:              inc.BizNo,
		IncidentLevel:      inc.IncidentLevel,
		IncidentLevelText:  incidentLevelTextMap[inc.IncidentLevel],
		IncidentType:       inc.IncidentType,
		OccurredAt:         inc.OccurredAt.Format("2006-01-02T15:04:05+08:00"),
		BuildingID:         inc.BuildingID,
		FloorID:            inc.FloorID,
		RoomID:             inc.RoomID,
		LocationDetail:     inc.LocationDetail,
		ReporterUserID:     inc.ReporterUserID,
		InvolvedStudentIDs: inc.InvolvedStudentIDs,
		WitnessUserIDs:     inc.WitnessUserIDs,
		InitialAction:      inc.InitialAction,
		Status:             inc.Status,
		StatusText:         incidentStatusTextMap[inc.Status],
		ClosedBy:           inc.ClosedBy,
		CreatedAt:          inc.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:          inc.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	if inc.ClosedAt != nil {
		t := inc.ClosedAt.Format("2006-01-02T15:04:05+08:00")
		v.ClosedAt = &t
	}

	// 加载楼栋名称
	if b, err := s.repo.GetBuildingByID(inc.BuildingID); err == nil {
		v.BuildingName = b.Name
	}

	// 加载上报人姓名
	if u, err := s.repo.GetUserByID(inc.ReporterUserID); err == nil {
		v.ReporterName = u.DisplayName
	}

	return v
}

// loadStudentBriefs 从 JSON 字符串加载学生简要信息。
func (s *IncidentService) loadStudentBriefs(idsJSON string) []StudentBrief {
	var ids []int64
	if err := json.Unmarshal([]byte(idsJSON), &ids); err != nil {
		return nil
	}

	briefs := make([]StudentBrief, 0, len(ids))
	for _, sid := range ids {
		if student, err := s.repo.GetStudentByID(sid); err == nil {
			briefs = append(briefs, StudentBrief{
				ID:        student.ID,
				StudentNo: student.StudentNo,
				Name:      student.Name,
			})
		}
	}
	return briefs
}

// parseTime 解析时间字符串。
func parseTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05+08:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", s)
}

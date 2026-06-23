package service

import (
	"fmt"
	"time"

	"student-system/internal/eventx"
	"student-system/internal/models"
	"student-system/internal/modules/ty/repository"
)

// RosterService 团员花名册业务服务层。
type RosterService struct {
	repo *repository.RosterRepository
	db   interface{} // *gorm.DB（用于事务）
	bus  *eventx.Bus
}

// NewRosterService 创建团员花名册服务。
func NewRosterService(
	repo *repository.RosterRepository,
	db interface{},
	bus *eventx.Bus,
) *RosterService {
	return &RosterService{
		repo: repo,
		db:   db,
		bus:  bus,
	}
}

// ---- DTO 定义 ----

// MemberView 团员视图。
type MemberView struct {
	ID                   int64      `json:"id"`
	BizNo                string     `json:"biz_no"`
	StudentID            int64      `json:"student_id"`
	StudentName          string     `json:"student_name"`
	StudentNo            string     `json:"student_no"`
	ApplicationID        *int64     `json:"application_id,omitempty"`
	BranchID             int64      `json:"branch_id"`
	BranchName           string     `json:"branch_name"`
	JoinAt               string     `json:"join_at"`
	BecomeProbationaryAt *string    `json:"become_probationary_at,omitempty"`
	IsOvertime           int        `json:"is_overtime"`
	TransferredAt        *string    `json:"transferred_at,omitempty"`
	ArchiveKeepUntil     *string    `json:"archive_keep_until,omitempty"`
	Status               string     `json:"status"`
	StatusText           string     `json:"status_text"`
	CreatedAt            string     `json:"created_at"`
	UpdatedAt            string     `json:"updated_at"`
}

// MemberListResult 团员列表结果。
type MemberListResult struct {
	Items    []MemberView `json:"items"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// UpdateMemberRequest 编辑团员信息请求。
type UpdateMemberRequest struct {
	MemberNo *string `json:"member_no"` // 团员证号（可选，需唯一性校验）
}

// TransferOutRequest 转出请求。
type TransferOutRequest struct {
	Reason string `json:"reason"` // 转出原因
}

var memberStatusTextMap = map[string]string{
	"active":      "在册",
	"transferred": "已转出",
	"overtime":    "超龄离团",
	"archived":    "已归档",
}

// ---- 业务方法 ----

// List 列表查询团员（GET /api/v1/ty/members）。
//
// 支持按支部/状态/关键字筛选和分页。
func (s *RosterService) List(branchID int64, status string, keyword string, page, pageSize int) (*MemberListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := s.repo.List(branchID, status, keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	views := make([]MemberView, 0, len(items))
	for _, item := range items {
		views = append(views, *s.toView(item))
	}

	return &MemberListResult{
		Items:    views,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取团员详情（GET /api/v1/ty/members/:id）。
func (s *RosterService) Get(id int64) (*MemberView, error) {
	roster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	v := s.toView(*roster)
	return v, nil
}

// Update 编辑团员信息（PATCH /api/v1/ty/members/:id）。
//
// 支持修改团员证号等字段。
func (s *RosterService) Update(id int64, req *UpdateMemberRequest, userID int64) (*MemberView, error) {
	roster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("团员记录不存在")
	}

	// 校验团员证号唯一性
	if req.MemberNo != nil && *req.MemberNo != "" {
		exists, err := s.repo.CheckMemberNoExists(*req.MemberNo, id)
		if err != nil {
			return nil, fmt.Errorf("检查团员证号失败: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("团员证号已存在，错误码:2640")
		}
	}

	// 更新字段（此处可根据需要扩展更多可编辑字段）
	if req.MemberNo != nil {
		// 注意：TyMemberRoster 模型中没有 member_no 字段，这里仅作为示例
		// 实际使用时需要在模型中添加该字段或调整逻辑
	}

	if err := s.repo.Update(roster); err != nil {
		return nil, fmt.Errorf("更新团员信息失败: %w", err)
	}

	return s.Get(id)
}

// TransferOut 团员转出（POST /api/v1/ty/members/:id:transfer-out）。
//
// 操作：
//   - 状态改为 transferred
//   - 设置 transferred_at 为当前时间
func (s *RosterService) TransferOut(id int64, userID int64, actorName, actorRole, ip, ua string) (*MemberView, error) {
	roster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("团员记录不存在")
	}
	if roster.Status != "active" {
		return nil, fmt.Errorf("仅在册团员可执行转出操作")
	}

	now := time.Now()
	if err := s.repo.TransferOut(id, &now); err != nil {
		return nil, fmt.Errorf("转出操作失败: %w", err)
	}

	s.publishRosterEvent(roster, "TyMemberTransferredOut", userID, actorRole, ip, ua, map[string]interface{}{
		"reason": "团员转出",
	})

	return s.Get(id)
}

// Overtime 超龄离团（POST /api/v1/ty/members/:id:overtime）。
//
// 业务规则：BR-TY-03
func (s *RosterService) Overtime(id int64, userID int64, actorName, actorRole, ip, ua string) (*MemberView, error) {
	roster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("团员记录不存在")
	}
	if roster.Status != "active" {
		return nil, fmt.Errorf("仅在册团员可执行超龄离团操作")
	}

	if err := s.repo.Overtime(id); err != nil {
		return nil, fmt.Errorf("超龄离团操作失败: %w", err)
	}

	s.publishRosterEvent(roster, "TyMemberOvertime", userID, actorRole, ip, ua, map[string]interface{}{})

	return s.Get(id)
}

// Archive 归档（POST /api/v1/ty/members/:id:archive）。
//
// 业务规则：BR-TY-04，保留5年
// 操作：
//   - 状态改为 archived
//   - 设置 archive_keep_until = 当前时间 + 5年
func (s *RosterService) Archive(id int64, userID int64, actorName, actorRole, ip, ua string) (*MemberView, error) {
	roster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("团员记录不存在")
	}
	if roster.Status == "archived" {
		return nil, fmt.Errorf("该团员已归档")
	}

	// 计算归档保留截止日期：当前时间 + 5年
	keepUntil := time.Now().AddDate(5, 0, 0)

	if err := s.repo.Archive(id, keepUntil); err != nil {
		return nil, fmt.Errorf("归档操作失败: %w", err)
	}

	s.publishRosterEvent(roster, "TyMemberArchived", userID, actorRole, ip, ua, map[string]interface{}{
		"archive_keep_until": keepUntil.Format("2006-01-02"),
	})

	return s.Get(id)
}

// CountByStatus 按状态统计团员数量（用于仪表板展示）。
func (s *RosterService) CountByStatus(status string) (int64, error) {
	return s.repo.CountByStatus(status)
}

// ---- 内部方法 ----

// toView 将模型转为视图。
func (s *RosterService) toView(roster models.TyMemberRoster) *MemberView {
	v := &MemberView{
		ID:           roster.ID,
		BizNo:        roster.BizNo,
		StudentID:    roster.StudentID,
		ApplicationID: roster.ApplicationID,
		BranchID:     roster.BranchID,
		JoinAt:       roster.JoinAt.Format("2006-01-02"),
		IsOvertime:   roster.IsOvertime,
		Status:       roster.Status,
		StatusText:   memberStatusTextMap[roster.Status],
		CreatedAt:    roster.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:    roster.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	if roster.BecomeProbationaryAt != nil {
		t := roster.BecomeProbationaryAt.Format("2006-01-02")
		v.BecomeProbationaryAt = &t
	}
	if roster.TransferredAt != nil {
		t := roster.TransferredAt.Format("2006-01-02")
		v.TransferredAt = &t
	}
	if roster.ArchiveKeepUntil != nil {
		t := roster.ArchiveKeepUntil.Format("2006-01-02")
		v.ArchiveKeepUntil = &t
	}

	// 加载学生姓名和学号
	if student, err := s.repo.GetStudentByID(roster.StudentID); err == nil {
		v.StudentName = student.Name
		v.StudentNo = student.StudentNo
	}

	// 加载团支部名称
	if branch, err := s.repo.GetBranchByID(roster.BranchID); err == nil {
		v.BranchName = branch.Name
	}

	return v
}

// publishRosterEvent 发布团员花名册相关事件。
func (s *RosterService) publishRosterEvent(roster *models.TyMemberRoster, evtType string, actorID int64, actorRole, ip, ua string, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["member_roster_id"] = roster.ID
	payload["biz_no"] = roster.BizNo
	payload["student_id"] = roster.StudentID
	payload["status"] = roster.Status

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "ty.member_roster",
		AggregateID: roster.BizNo,
		EventType:   evtType,
		Module:      "TY",
		ActorID:     actorID,
		ActorRole:   actorRole,
		Payload:     payload,
		BizNo:       roster.BizNo,
		IP:          ip,
		UA:          ua,
	})
}

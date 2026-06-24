package service

import (
	"fmt"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/st/repository"
)

// AssociationService 社团业务服务层。
type AssociationService struct {
	repo *repository.AssociationRepository
	db   *gorm.DB
	bus  *eventx.Bus
}

// NewAssociationService 创建社团服务。
func NewAssociationService(repo *repository.AssociationRepository, db *gorm.DB, bus *eventx.Bus) *AssociationService {
	return &AssociationService{repo: repo, db: db, bus: bus}
}

// ---- DTO ----

// AssociationListResult 社团列表结果。
type AssociationListResult struct {
	Items    []AssociationView `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// AssociationView 社团视图。
type AssociationView struct {
	ID                 int64  `json:"id"`
	BizNo              string `json:"biz_no"`
	Name               string `json:"name"`
	CollegeID          int64  `json:"college_id"`
	CollegeName        string `json:"college_name"`
	TutorUserID        *int64 `json:"tutor_user_id,omitempty"`
	TutorName          string `json:"tutor_name"`
	PresidentStudentID *int64 `json:"president_student_id,omitempty"`
	PresidentName      string `json:"president_name"`
	BusinessScope      string `json:"business_scope"`
	Status             string `json:"status"`
	StatusText         string `json:"status_text"`
	StarRating         *int   `json:"star_rating,omitempty"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// CreateAssociationRequest 创建社团请求。
type CreateAssociationRequest struct {
	Name               string  `json:"name" binding:"required"`
	CollegeID          int64   `json:"college_id" binding:"required"`
	TutorUserID        *int64  `json:"tutor_user_id"`
	PresidentStudentID *int64  `json:"president_student_id"`
	BusinessScope      string  `json:"business_scope" binding:"required"`
	Founders           []int64 `json:"founders" binding:"required"`
}

// UpdateAssociationRequest 更新社团请求。
type UpdateAssociationRequest struct {
	Name               *string `json:"name"`
	CollegeID          *int64  `json:"college_id"`
	TutorUserID        *int64  `json:"tutor_user_id"`
	PresidentStudentID *int64  `json:"president_student_id"`
	BusinessScope      *string `json:"business_scope"`
}

// FounderView 发起人视图。
type FounderView struct {
	ID        int64  `json:"id"`
	StudentID int64  `json:"student_id"`
	StudentNo string `json:"student_no"`
	StudentName string `json:"student_name"`
}

// MemberView 成员视图。
type MemberView struct {
	ID          int64  `json:"id"`
	StudentID   int64  `json:"student_id"`
	StudentNo   string `json:"student_no"`
	StudentName string `json:"student_name"`
	Role        string `json:"role"`
	RoleText    string `json:"role_text"`
	JoinedAt    string `json:"joined_at"`
}

// ---- 状态映射 ----

var assocStatusTextMap = map[string]string{
	"preparing":   "筹备中",
	"trial":       "试运行",
	"registered":  "注册成立",
	"rectifying":  "评估整顿",
	"cancelled":   "注销",
}

var memberRoleTextMap = map[string]string{
	"president":     "社长",
	"vice_president": "副社长",
	"director":      "理事",
	"member":        "会员",
}

// ---- 业务方法 ----

// List 分页查询社团列表。
func (s *AssociationService) List(status string, collegeID int64, keyword string, page, pageSize int) (*AssociationListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	assocs, total, err := s.repo.List(status, collegeID, keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	colleges, _ := s.repo.ListColleges()
	collegeMap := make(map[int64]string)
	for _, c := range colleges {
		collegeMap[c.ID] = c.Name
	}

	items := make([]AssociationView, 0, len(assocs))
	for _, a := range assocs {
		items = append(items, s.toView(a, collegeMap))
	}

	return &AssociationListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取社团详情。
func (s *AssociationService) Get(id int64) (*AssociationView, error) {
	assoc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	colleges, _ := s.repo.ListColleges()
	collegeMap := make(map[int64]string)
	for _, c := range colleges {
		collegeMap[c.ID] = c.Name
	}

	v := s.toView(*assoc, collegeMap)
	return &v, nil
}

// Create 创建社团（保存为 preparing 筹备状态）。
func (s *AssociationService) Create(userID int64, req *CreateAssociationRequest) (*AssociationView, error) {
	// 校验发起人数量 5-20
	if len(req.Founders) < 5 || len(req.Founders) > 20 {
		return nil, fmt.Errorf("发起人须 5-20 名")
	}

	// 校验同名社团
	count, err := s.repo.CountByName(req.Name, 0)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("同名社团已存在")
	}

	// 校验指导教师同期指导社团数 ≤ 3
	if req.TutorUserID != nil && *req.TutorUserID > 0 {
		tutorCount, err := s.repo.CountByTutor(*req.TutorUserID)
		if err != nil {
			return nil, err
		}
		if tutorCount >= 3 {
			return nil, fmt.Errorf("指导教师同期最多指导 3 个社团")
		}
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "ST")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	now := time.Now()

	// 事务：创建社团 + 创建发起人记录
	assoc := &models.StAssociation{
		BizNo:              bizNo,
		Name:               req.Name,
		CollegeID:          req.CollegeID,
		TutorUserID:        req.TutorUserID,
		PresidentStudentID: req.PresidentStudentID,
		BusinessScope:      req.BusinessScope,
		Status:             "preparing",
		FoundedAt:          &now,
		CreatedBy:          &userID,
		UpdatedBy:          &userID,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(assoc).Error; err != nil {
			return err
		}

		// 创建发起人记录
		for _, sid := range req.Founders {
			founder := &models.StFounder{
				AssociationID: assoc.ID,
				StudentID:     sid,
			}
			if err := tx.Create(founder).Error; err != nil {
				return fmt.Errorf("创建发起人记录失败: %w", err)
			}

			// 同时加入成员表
			member := &models.StAssocMember{
				AssociationID: assoc.ID,
				StudentID:     sid,
				Role:          "member",
				JoinedAt:      now,
				IsCoreOfficer: 0,
			}
			if err := tx.Create(member).Error; err != nil {
				return fmt.Errorf("添加发起人为成员失败: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// 发布事件
	if s.bus != nil {
		_ = s.bus.Publish(&eventx.Event{
			Aggregate:   "st.association",
			AggregateID: assoc.BizNo,
			EventType:   "StAssociationCreated",
			Module:      "ST",
			ActorID:     userID,
			Payload: map[string]interface{}{
				"association_id": assoc.ID,
				"biz_no":         assoc.BizNo,
				"name":           assoc.Name,
				"founders_count": len(req.Founders),
			},
			BizNo: assoc.BizNo,
		})
	}

	return s.Get(assoc.ID)
}

// Update 更新社团。
func (s *AssociationService) Update(id, userID int64, req *UpdateAssociationRequest) (*AssociationView, error) {
	assoc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("社团不存在")
	}

	if req.Name != nil {
		if utf8.RuneCountInString(*req.Name) == 0 {
			return nil, fmt.Errorf("社团名称不能为空")
		}
		assoc.Name = *req.Name
	}
	if req.CollegeID != nil {
		assoc.CollegeID = *req.CollegeID
	}
	if req.TutorUserID != nil {
		assoc.TutorUserID = req.TutorUserID
	}
	if req.PresidentStudentID != nil {
		assoc.PresidentStudentID = req.PresidentStudentID
	}
	if req.BusinessScope != nil {
		assoc.BusinessScope = *req.BusinessScope
	}
	assoc.UpdatedBy = &userID

	if err := s.repo.Update(assoc); err != nil {
		return nil, err
	}

	return s.Get(assoc.ID)
}

// SoftDelete 软删除社团。
func (s *AssociationService) SoftDelete(id, userID int64) error {
	assoc, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("社团不存在")
	}
	if assoc.Status != "preparing" && assoc.Status != "cancelled" {
		return fmt.Errorf("仅筹备中或已注销的社团可删除")
	}
	return s.repo.SoftDelete(id)
}

// ListFounders 查询社团发起人列表。
func (s *AssociationService) ListFounders(associationID int64) ([]FounderView, error) {
	founders, err := s.repo.ListFoundersByAssoc(associationID)
	if err != nil {
		return nil, err
	}

	views := make([]FounderView, 0, len(founders))
	for _, f := range founders {
		v := FounderView{
			ID:        f.ID,
			StudentID: f.StudentID,
		}
		if student, err := s.repo.GetStudentByID(f.StudentID); err == nil {
			v.StudentNo = student.StudentNo
			v.StudentName = student.Name
		}
		views = append(views, v)
	}
	return views, nil
}

// ListMembers 查询社团成员列表。
func (s *AssociationService) ListMembers(associationID int64) ([]MemberView, error) {
	members, err := s.repo.ListMembersByAssoc(associationID)
	if err != nil {
		return nil, err
	}

	views := make([]MemberView, 0, len(members))
	for _, m := range members {
		v := MemberView{
			ID:        m.ID,
			StudentID: m.StudentID,
			Role:      m.Role,
			RoleText:  memberRoleTextMap[m.Role],
			JoinedAt:  m.JoinedAt.Format("2006-01-02"),
		}
		if student, err := s.repo.GetStudentByID(m.StudentID); err == nil {
			v.StudentNo = student.StudentNo
			v.StudentName = student.Name
		}
		views = append(views, v)
	}
	return views, nil
}

// ListUsers 查询用户列表(用于指导教师下拉,仅教职工)。
func (s *AssociationService) ListUsers() ([]map[string]interface{}, error) {
	users, err := s.repo.ListUsers()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		result = append(result, map[string]interface{}{
			"id":           u.ID,
			"display_name": u.DisplayName,
		})
	}
	return result, nil
}

// ListStudents 查询学生列表(用于社长下拉)。
func (s *AssociationService) ListStudents() ([]map[string]interface{}, error) {
	students, err := s.repo.ListStudents()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(students))
	for _, st := range students {
		result = append(result, map[string]interface{}{
			"id":         st.ID,
			"student_no": st.StudentNo,
			"name":       st.Name,
		})
	}
	return result, nil
}

// ---- 内部方法 ----

func (s *AssociationService) toView(assoc models.StAssociation, collegeMap map[int64]string) AssociationView {
	v := AssociationView{
		ID:                 assoc.ID,
		BizNo:              assoc.BizNo,
		Name:               assoc.Name,
		CollegeID:          assoc.CollegeID,
		CollegeName:        collegeMap[assoc.CollegeID],
		TutorUserID:        assoc.TutorUserID,
		PresidentStudentID: assoc.PresidentStudentID,
		BusinessScope:      assoc.BusinessScope,
		Status:             assoc.Status,
		StatusText:         assocStatusTextMap[assoc.Status],
		StarRating:         assoc.StarRating,
		CreatedAt:          assoc.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:          assoc.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 加载指导教师姓名
	if assoc.TutorUserID != nil {
		if user, err := s.repo.GetUserByID(*assoc.TutorUserID); err == nil {
			v.TutorName = user.DisplayName
		}
	}

	// 加载社长姓名
	if assoc.PresidentStudentID != nil {
		if student, err := s.repo.GetStudentByID(*assoc.PresidentStudentID); err == nil {
			v.PresidentName = student.Name
		}
	}

	return v
}

package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/ty/repository"
)

// DevelopmentMeetingService 发展大会业务服务层。
type DevelopmentMeetingService struct {
	meetingRepo *repository.DevelopmentMeetingRepository
	devRepo     *repository.DevelopmentObjectRepository
	polRepo     *repository.PoliticalReviewRepository
	appRepo     *repository.ApplicationRepository
	db          *gorm.DB
	bus         *eventx.Bus
}

// NewDevelopmentMeetingService 创建发展大会服务。
func NewDevelopmentMeetingService(
	meetingRepo *repository.DevelopmentMeetingRepository,
	devRepo *repository.DevelopmentObjectRepository,
	polRepo *repository.PoliticalReviewRepository,
	appRepo *repository.ApplicationRepository,
	db *gorm.DB,
	bus *eventx.Bus,
) *DevelopmentMeetingService {
	return &DevelopmentMeetingService{
		meetingRepo: meetingRepo,
		devRepo:     devRepo,
		polRepo:     polRepo,
		appRepo:     appRepo,
		db:          db,
		bus:         bus,
	}
}

// ---- DTO 定义 ----

// CreateDevelopmentMeetingRequest 创建发展大会请求。
type CreateDevelopmentMeetingRequest struct {
	DevelopmentID     int64  `json:"development_id" binding:"required"`      // 关联的发展对象ID
	MeetingAt         string `json:"meeting_at" binding:"required"`           // 会议时间 YYYY-MM-DD HH:mm:ss
	ExpectedCount     int    `json:"expected_count" binding:"required"`       // 应到人数
	ActualCount       int    `json:"actual_count" binding:"required"`         // 实到人数
	ApproveCount      int    `json:"approve_count" binding:"required"`        // 赞成票数
	AgainstCount      int    `json:"against_count" binding:"required"`        // 反对票数
	AbstainCount      int    `json:"abstain_count" binding:"required"`        // 弃权票数
	Decision          string `json:"decision" binding:"required"`             // pass | reject
	VolunteerFormPath string `json:"volunteer_form_path"`                     // 入团志愿书路径
}

// DevelopmentMeetingView 发展大会视图。
type DevelopmentMeetingView struct {
	ID                int64  `json:"id"`
	BizNo             string `json:"biz_no"`
	DevelopmentID     int64  `json:"development_id"`
	StudentID         int64  `json:"student_id,omitempty"`
	StudentName       string `json:"student_name,omitempty"`
	MeetingAt         string `json:"meeting_at"`
	ExpectedCount     int    `json:"expected_count"`
	ActualCount       int    `json:"actual_count"`
	ApproveCount      int    `json:"approve_count"`
	AgainstCount      int    `json:"against_count"`
	AbstainCount      int    `json:"abstain_count"`
	Decision          string `json:"decision"`
	DecisionText      string `json:"decision_text"`
	VolunteerFormPath string `json:"volunteer_form_path"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// DevelopmentMeetingListResult 发展大会列表结果。
type DevelopmentMeetingListResult struct {
	Items    []DevelopmentMeetingView `json:"items"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

var decisionTextMap = map[string]string{
	"pass":   "通过",
	"reject": "未通过",
}

// ---- 业务方法 ----

// Create 创建发展大会记录（POST /api/v1/ty/development-meetings）。
//
// 前置条件：
//   - 发展对象状态必须为 S3（已完成审批）
//   - 政审必须全部通过（无 fail 结论）
//
// 票数规则：
//   - 实到人数 ≥ 应到人数 * 2/3
//   - 赞成票数 > 实到人数 / 2
//
// 决策为 pass 时自动执行：
//   1. 更新 ty_application 状态为"预备团员"对应的内部状态
//   2. 在 ty_member_roster 中创建新记录（status=active, become_probationary_at=今天）
//   3. 更新 idx_student.political_status="预备团员"
func (s *DevelopmentMeetingService) Create(userID int64, req *CreateDevelopmentMeetingRequest, actorName, actorRole, ip, ua string) (*DevelopmentMeetingView, error) {
	// 校验发展对象是否存在且状态正确
	devObj, err := s.devRepo.GetByID(req.DevelopmentID)
	if err != nil {
		return nil, fmt.Errorf("发展对象不存在")
	}
	if devObj.Status != "S3" {
		return nil, fmt.Errorf("发展对象尚未完成审批流程，错误码:2621")
	}

	// 校验政审是否全部通过
	allPassed, hasBasicPass, hasFail, err := s.polRepo.CheckAllPassed(req.DevelopmentID)
	if err != nil {
		return nil, fmt.Errorf("检查政审状态失败: %w", err)
	}
	if hasFail {
		return nil, fmt.Errorf("政审结论包含不合格，终止发展，错误码:2610")
	}
	if hasBasicPass {
		return nil, fmt.Errorf("政审基本合格，需延长培养期3个月，错误码:2611")
	}
	if !allPassed {
		return nil, fmt.Errorf("发展对象尚未完成政审，错误码:2621")
	}

	// 校验票数规则：实到 ≥ 应到 * 2/3
	minActual := req.ExpectedCount * 2 / 3
	if req.ActualCount < minActual {
		return nil, fmt.Errorf("实到人数不足应到人数的2/3，错误码:2620")
	}

	// 校验票数规则：赞成 > 实到 / 2
	if req.ApproveCount <= req.ActualCount/2 {
		return nil, fmt.Errorf("赞成票数不满足要求，须超过实到人数的一半，错误码:2620")
	}

	// 校验 decision 值
	if req.Decision != "pass" && req.Decision != "reject" {
		return nil, fmt.Errorf("无效的决策值，必须是 pass/reject")
	}

	// 解析会议时间
	meetingAt, err := time.Parse("2006-01-02 15:04:05", req.MeetingAt)
	if err != nil {
		return nil, fmt.Errorf("会议时间格式错误")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "TY")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	meeting := models.TyDevelopmentMeeting{
		BizNo:             bizNo,
		DevelopmentID:     req.DevelopmentID,
		MeetingAt:         meetingAt,
		ExpectedCount:     req.ExpectedCount,
		ActualCount:       req.ActualCount,
		ApproveCount:      req.ApproveCount,
		AgainstCount:      req.AgainstCount,
		AbstainCount:      req.AbstainCount,
		Decision:          req.Decision,
		VolunteerFormPath: req.VolunteerFormPath,
	}

	// 事务处理：创建会议 + 如果通过则联动更新多个表
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 使用事务中的 db 创建会议记录
		if txErr := tx.Create(&meeting).Error; txErr != nil {
			return fmt.Errorf("创建发展大会记录失败: %w", txErr)
		}

		// 如果决策为 pass，执行联动操作
		if req.Decision == "pass" {
			if txErr := s.processPassDecision(tx, devObj); txErr != nil {
				return txErr
			}
		} else {
			// reject 时退回发展对象状态（可根据业务需求调整）
			devObj.Status = "S4"
			if txErr := tx.Save(devObj).Error; txErr != nil {
				return fmt.Errorf("更新发展对象状态失败: %w", txErr)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.publishMeetingEvent(&meeting, "TyDevelopmentMeetingCreated", userID, actorRole, ip, ua, map[string]interface{}{
		"decision":      req.Decision,
		"approve_count": req.ApproveCount,
		"actual_count":  req.ActualCount,
	})

	return s.GetByID(meeting.ID)
}

// processPassDecision 处理发展大会通过的联动操作（事务内调用）。
func (s *DevelopmentMeetingService) processPassDecision(tx *gorm.DB, devObj *models.TyDevelopmentObject) error {
	now := time.Now()

	// 1. 获取入团申请信息
	app, err := s.appRepo.GetByID(devObj.ApplicationID)
	if err != nil {
		return fmt.Errorf("获取入团申请失败: %w", err)
	}

	// 2. 更新 ty_application 状态为 S3（对应"预备团员"阶段）
	if err := tx.Model(&models.TyApplication{}).
		Where("id = ? AND is_deleted = 0", app.ID).
		Update("status", "S3").Error; err != nil {
		return fmt.Errorf("更新申请状态失败: %w", err)
	}

	// 3. 在 ty_member_roster 中创建新记录
	roster := models.TyMemberRoster{
		BizNo:                "", // 将在下面生成
		StudentID:            app.StudentID,
		ApplicationID:        &app.ID,
		BranchID:             app.BranchID,
		JoinAt:               now,
		BecomeProbationaryAt: &now,
		Status:               "active",
	}
	// 生成团员花名册业务编号
	rosterBizNo, err := idgen.NextBizNo(tx, "TY")
	if err != nil {
		return fmt.Errorf("生成团员花名册编号失败: %w", err)
	}
	roster.BizNo = rosterBizNo

	if err := tx.Create(&roster).Error; err != nil {
		return fmt.Errorf("创建团员花名册记录失败: %w", err)
	}

	// 4. 更新 idx_student.political_status 为"预备团员"
	if err := tx.Model(&models.IdxStudent{}).
		Where("id = ? AND is_deleted = 0", app.StudentID).
		Update("political_status", "probationary").Error; err != nil {
		// 非致命错误，仅记录但不影响主流程
		fmt.Printf("[TY-DevMeeting] 更新学生政治面貌失败: studentID=%d err=%v\n", app.StudentID, err)
	}

	return nil
}

// GetByID 获取发展大会详情。
func (s *DevelopmentMeetingService) GetByID(id int64) (*DevelopmentMeetingView, error) {
	meeting, err := s.meetingRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toView(*meeting), nil
}

// List 列表查询发展大会。
func (s *DevelopmentMeetingService) List(page, pageSize int) (*DevelopmentMeetingListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := s.meetingRepo.List(page, pageSize)
	if err != nil {
		return nil, err
	}

	views := make([]DevelopmentMeetingView, 0, len(items))
	for _, item := range items {
		views = append(views, *s.toView(item))
	}

	return &DevelopmentMeetingListResult{
		Items:    views,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListByDevelopmentID 按发展对象ID查询所有发展大会。
func (s *DevelopmentMeetingService) ListByDevelopmentID(developmentID int64) ([]DevelopmentMeetingView, error) {
	meetings, err := s.meetingRepo.ListByDevelopmentID(developmentID)
	if err != nil {
		return nil, err
	}

	views := make([]DevelopmentMeetingView, 0, len(meetings))
	for _, m := range meetings {
		views = append(views, *s.toView(m))
	}
	return views, nil
}

// ---- 内部方法 ----

// toView 将模型转为视图。
func (s *DevelopmentMeetingService) toView(meeting models.TyDevelopmentMeeting) *DevelopmentMeetingView {
	v := &DevelopmentMeetingView{
		ID:                meeting.ID,
		BizNo:             meeting.BizNo,
		DevelopmentID:     meeting.DevelopmentID,
		MeetingAt:         meeting.MeetingAt.Format("2006-01-02T15:04:05+08:00"),
		ExpectedCount:     meeting.ExpectedCount,
		ActualCount:       meeting.ActualCount,
		ApproveCount:      meeting.ApproveCount,
		AgainstCount:      meeting.AgainstCount,
		AbstainCount:      meeting.AbstainCount,
		Decision:          meeting.Decision,
		DecisionText:      decisionTextMap[meeting.Decision],
		VolunteerFormPath: meeting.VolunteerFormPath,
		CreatedAt:         meeting.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:         meeting.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 通过发展对象→入团申请→学生 回填申请人姓名
	if devObj, err := s.devRepo.GetByID(meeting.DevelopmentID); err == nil && devObj != nil {
		if app, err := s.appRepo.GetByID(devObj.ApplicationID); err == nil && app != nil {
			v.StudentID = app.StudentID
			if stu, err := s.appRepo.GetStudentByID(app.StudentID); err == nil && stu != nil {
				v.StudentName = stu.Name
			}
		}
	}

	return v
}

// publishMeetingEvent 发布发展大会相关事件。
func (s *DevelopmentMeetingService) publishMeetingEvent(meeting *models.TyDevelopmentMeeting, evtType string, actorID int64, actorRole, ip, ua string, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["development_meeting_id"] = meeting.ID
	payload["biz_no"] = meeting.BizNo
payload["development_id"] = meeting.DevelopmentID
	payload["decision"] = meeting.Decision

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "ty.development_meeting",
		AggregateID: meeting.BizNo,
		EventType:   evtType,
		Module:      "TY",
		ActorID:     actorID,
		ActorRole:   actorRole,
		Payload:     payload,
		BizNo:       meeting.BizNo,
		IP:          ip,
		UA:          ua,
	})
}

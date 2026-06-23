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

// RecommendationService 推优大会业务服务层。
type RecommendationService struct {
	repo *repository.RecommendationRepository
	appRepo *repository.ApplicationRepository
	db   *gorm.DB
	bus  *eventx.Bus
}

// NewRecommendationService 创建推优大会服务。
func NewRecommendationService(
	repo *repository.RecommendationRepository,
	appRepo *repository.ApplicationRepository,
	db *gorm.DB,
	bus *eventx.Bus,
) *RecommendationService {
	return &RecommendationService{
		repo:   repo,
		appRepo: appRepo,
		db:     db,
		bus:    bus,
	}
}

// ---- DTO 定义 ----

// CreateMeetingRequest 创建推优大会请求。
type CreateMeetingRequest struct {
	ApplicationID  int64  `json:"application_id" binding:"required"`
	MeetingAt      string `json:"meeting_at" binding:"required"`
	Location       string `json:"location" binding:"required"`
	ExpectedCount  int    `json:"expected_count" binding:"required"`
	ActualCount    int    `json:"actual_count" binding:"required"`
	PhotoOverallID int64  `json:"photo_overall_id" binding:"required"`
	PhotoVoteID    int64  `json:"photo_vote_id" binding:"required"`
	Decision       string `json:"decision" binding:"required"` // pass / reject
	DecisionReason string `json:"decision_reason"`
	ApproveCount   int    `json:"approve_count" binding:"min=0"`
	AgainstCount   int    `json:"against_count" binding:"min=0"`
	AbstainCount   int    `json:"abstain_count" binding:"min=0"`
}

// MeetingListResult 推优大会列表结果。
type MeetingListResult struct {
	Items    []MeetingView `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// MeetingView 推优大会视图。
type MeetingView struct {
	ID              int64      `json:"id"`
	BizNo           string     `json:"biz_no"`
	ApplicationID   int64      `json:"application_id"`
	StudentName     string     `json:"student_name"`
	MeetingAt       string     `json:"meeting_at"`
	Location        string     `json:"location"`
	ExpectedCount   int        `json:"expected_count"`
	ActualCount     int        `json:"actual_count"`
	PhotoOverallID  *int64     `json:"photo_overall_id,omitempty"`
	PhotoVoteID     *int64     `json:"photo_vote_id,omitempty"`
	Decision        string     `json:"decision"`
	DecisionReason  string     `json:"decision_reason"`
	CreatedAt       string     `json:"created_at"`
	Vote            *VoteView  `json:"vote,omitempty"`
}

// VoteView 投票视图。
type VoteView struct {
	ID           int64 `json:"id"`
	MeetingID    int64 `json:"meeting_id"`
	ApproveCount int   `json:"approve_count"`
	AgainstCount int   `json:"against_count"`
	AbstainCount int   `json:"abstain_count"`
}

// ---- 业务方法 ----

// List 分页查询推优大会列表。支持按 branch_id 过滤（通过申请单关联）。
func (s *RecommendationService) List(branchID int64, page, pageSize int) (*MeetingListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	meetings, total, err := s.repo.ListByBranchID(branchID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]MeetingView, 0, len(meetings))
	for _, m := range meetings {
		v := s.toMeetingView(m)
		items = append(items, v)
	}

	return &MeetingListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取推优大会详情（含投票信息）。
func (s *RecommendationService) Get(id int64) (*MeetingView, error) {
	meeting, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("推优大会不存在")
	}

	v := s.toMeetingView(*meeting)

	// 加载投票信息
	votes, _ := s.repo.GetVotesByMeetingID(meeting.ID)
	if len(votes) > 0 {
		v.Vote = &VoteView{
			ID:           votes[0].ID,
			MeetingID:    votes[0].MeetingID,
			ApproveCount: votes[0].ApproveCount,
			AgainstCount: votes[0].AgainstCount,
			AbstainCount: votes[0].AbstainCount,
		}
	}

	return &v, nil
}

// GetByApplication 按申请ID查询推优大会。
func (s *RecommendationService) GetByApplication(applicationID int64) (*MeetingView, error) {
	meeting, vote, err := s.repo.GetByApplicationID(applicationID)
	if err != nil {
		return nil, fmt.Errorf("未找到该申请的推优大会记录")
	}

	v := s.toMeetingView(*meeting)
	if vote != nil {
		v.Vote = &VoteView{
			ID:           vote.ID,
			MeetingID:    vote.MeetingID,
			ApproveCount: vote.ApproveCount,
			AgainstCount: vote.AgainstCount,
			AbstainCount: vote.AbstainCount,
		}
	}

	return &v, nil
}

// Create 创建推优大会（含校验）。
//
// 校验规则：
//   - 申请必须处于 S3 状态（推优池）
//   - actual_count >= expected_count * 2/3 （到会率≥2/3）
//   - photo_overall_id 和 photo_vote_id 必须存在（BR-TY-02）
//   - approve_count > actual_count / 2 才允许 decision='pass'
//   - 同一申请人3个月内不得重复推优
func (s *RecommendationService) Create(userID int64, req *CreateMeetingRequest) (*MeetingView, error) {
	// 校验申请状态：必须为 S3（推优池）
	app, err := s.appRepo.GetByID(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("入团申请不存在")
	}
	if app.Status != "S3" {
		return nil, fmt.Errorf("仅 S3 状态的申请可召开推优大会，当前状态：%s", app.Status)
	}

	// 校验到会率 ≥ 2/3
	minAttendee := req.ExpectedCount * 2 / 3
	if req.ActualCount < minAttendee {
		return nil, fmt.Errorf("到会人数不足预期人数的 2/3（%d/%d）", req.ActualCount, minAttendee)
	}

	// 校验会议照片必须上传
	if req.PhotoOverallID <= 0 || req.PhotoVoteID <= 0 {
		return nil, fmt.Errorf("必须上传会议全景照和投票现场照")
	}

	// 校验赞成票过半才允许 decision=pass
	if req.Decision == "pass" && req.ApproveCount <= req.ActualCount/2 {
		return nil, fmt.Errorf("赞成票数未超过到会人数的一半，不可通过")
	}

	// 校验同一申请人3个月内不得重复推优
	hasRecent, err := s.repo.CheckRecentRecommendation(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("检查近期推优记录失败: %w", err)
	}
	if hasRecent {
		return nil, fmt.Errorf("该申请人 3 个月内已进行过推优，请勿重复操作")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "TY-REC")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// 解析会议时间
	meetingAt, err := parseTime(req.MeetingAt)
	if err != nil {
		return nil, fmt.Errorf("会议时间格式错误")
	}

	photoOverallID := req.PhotoOverallID
	photoVoteID := req.PhotoVoteID
	recorderUserID := userID

	meeting := models.TyRecommendationMeeting{
		BizNo:          bizNo,
		ApplicationID:  req.ApplicationID,
		MeetingAt:      meetingAt,
		Location:       req.Location,
		ExpectedCount:  req.ExpectedCount,
		ActualCount:    req.ActualCount,
		PhotoOverallID: &photoOverallID,
		PhotoVoteID:    &photoVoteID,
		Decision:       req.Decision,
		DecisionReason: req.DecisionReason,
		RecorderUserID: &recorderUserID,
	}

	// 事务：创建推优大会 + 投票记录
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&meeting).Error; err != nil {
			return fmt.Errorf("创建推优大会记录失败: %w", err)
		}

		vote := models.TyRecommendationVote{
			MeetingID:     meeting.ID,
			ApplicationID: req.ApplicationID,
			ApproveCount:  req.ApproveCount,
			AgainstCount:  req.AgainstCount,
			AbstainCount:  req.AbstainCount,
		}
		if err := tx.Create(&vote).Error; err != nil {
			return fmt.Errorf("创建投票记录失败: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 发布事件
	s.publishMeetingEvent(&meeting, "TyRecommendationMeetingCreated", userID, map[string]interface{}{
		"decision":       req.Decision,
		"actual_count":   req.ActualCount,
		"approve_count":  req.ApproveCount,
	})

	return s.Get(meeting.ID)
}

// toMeetingView 将模型转为视图。
func (s *RecommendationService) toMeetingView(m models.TyRecommendationMeeting) MeetingView {
	v := MeetingView{
		ID:              m.ID,
		BizNo:           m.BizNo,
		ApplicationID:   m.ApplicationID,
		MeetingAt:       m.MeetingAt.Format("2006-01-02T15:04:05+08:00"),
		Location:        m.Location,
		ExpectedCount:   m.ExpectedCount,
		ActualCount:     m.ActualCount,
		PhotoOverallID:  m.PhotoOverallID,
		PhotoVoteID:     m.PhotoVoteID,
		Decision:        m.Decision,
		DecisionReason:  m.DecisionReason,
		CreatedAt:       m.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 加载学生姓名（通过申请单关联）
	if app, err := s.appRepo.GetByID(m.ApplicationID); err == nil {
		if student, err := s.appRepo.GetStudentByID(app.StudentID); err == nil {
			v.StudentName = student.Name
		}
	}

	return v
}

// publishMeetingEvent 发布推优大会事件。
func (s *RecommendationService) publishMeetingEvent(meeting *models.TyRecommendationMeeting, evtType string, actorID int64, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["meeting_id"] = meeting.ID
	payload["biz_no"] = meeting.BizNo
	payload["application_id"] = meeting.ApplicationID
	payload["decision"] = meeting.Decision

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "ty.recommendation_meeting",
		AggregateID: meeting.BizNo,
		EventType:   evtType,
		Module:      "TY",
		ActorID:     actorID,
		Payload:     payload,
		BizNo:       meeting.BizNo,
	})
}

// parseTime 解析时间字符串（支持 RFC3339 和日期格式）。
func parseTime(t string) (time.Time, error) {
	// 尝试 RFC3339 格式
	if v, err := time.Parse(time.RFC3339, t); err == nil {
		return v, nil
	}
	// 尝试日期格式
	return time.Parse("2006-01-02 15:04:05", t)
}

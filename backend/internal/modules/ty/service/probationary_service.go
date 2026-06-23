package service

import (
	"fmt"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/ty/repository"
)

// ProbationaryService 预备期/转正业务服务层。
type ProbationaryService struct {
	repo    *repository.ProbationaryRepository
	appRepo *repository.ApplicationRepository
	db      *gorm.DB
	bus     *eventx.Bus
}

// NewProbationaryService 创建预备期服务。
func NewProbationaryService(
	repo *repository.ProbationaryRepository,
	appRepo *repository.ApplicationRepository,
	db *gorm.DB,
	bus *eventx.Bus,
) *ProbationaryService {
	return &ProbationaryService{
		repo:    repo,
		appRepo: appRepo,
		db:      db,
		bus:     bus,
	}
}

// ---- DTO 定义 ----

// CreateProbationaryRecordRequest 创建预备期考察记录请求。
type CreateProbationaryRecordRequest struct {
	ApplicationID int64  `json:"application_id" binding:"required"` // 关联的入团申请ID
	RecordYear    int    `json:"record_year" binding:"required"`    // 年份
	RecordQuarter int    `json:"record_quarter" binding:"required"` // 季度（1-4）
	Summary       string `json:"summary" binding:"required"`        // 考察总结 ≥100字
}

// ProbationaryRecordView 预备期考察记录视图。
type ProbationaryRecordView struct {
	ID            int64  `json:"id"`
	ApplicationID int64  `json:"application_id"`
	StudentID     int64  `json:"student_id,omitempty"`
	StudentName   string `json:"student_name,omitempty"`
	StudentNo     string `json:"student_no,omitempty"`
	RecordYear    int    `json:"record_year"`
	RecordQuarter int    `json:"record_quarter"`
	Summary       string `json:"summary"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// CreateProbationaryMeetingRequest 创建转正大会请求。
type CreateProbationaryMeetingRequest struct {
	ApplicationID       int64  `json:"application_id" binding:"required"`         // 关联的入团申请ID
	SelfApplicationPath string `json:"self_application_path"`                     // 转正申请书路径
	MeetingAt           string `json:"meeting_at" binding:"required"`             // 会议时间 YYYY-MM-DD HH:mm:ss
	ExpectedCount       int    `json:"expected_count" binding:"required"`          // 应到人数
	ActualCount         int    `json:"actual_count" binding:"required"`            // 实到人数
	ApproveCount        int    `json:"approve_count" binding:"required"`           // 赞成票数
	Decision            string `json:"decision" binding:"required"`               // pass | reject
}

// ProbationaryMeetingView 转正大会视图。
type ProbationaryMeetingView struct {
	ID                  int64      `json:"id"`
	BizNo               string     `json:"biz_no"`
	ApplicationID       int64      `json:"application_id"`
	StudentID           int64      `json:"student_id,omitempty"`
	StudentName         string     `json:"student_name,omitempty"`
	StudentNo           string     `json:"student_no,omitempty"`
	SelfApplicationPath string     `json:"self_application_path"`
	MeetingAt           string     `json:"meeting_at"`
	ExpectedCount       int        `json:"expected_count"`
	ActualCount         int        `json:"actual_count"`
	ApproveCount        int        `json:"approve_count"`
	Decision            string     `json:"decision"`
	DecisionText        string     `json:"decision_text"`
	FormalJoinAt        *string    `json:"formal_join_at,omitempty"`
	CreatedAt           string     `json:"created_at"`
	UpdatedAt           string     `json:"updated_at"`
}

// ProbationaryRecordListResult 预备期考察列表分页结果。
type ProbationaryRecordListResult struct {
	Items    []ProbationaryRecordView `json:"items"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

// ProbationaryMeetingListResult 转正大会列表分页结果。
type ProbationaryMeetingListResult struct {
	Items    []ProbationaryMeetingView `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// ---- 状态映射 ----

var probationaryDecisionTextMap = map[string]string{
	"pass":   "通过",
	"reject": "未通过",
}

// ---- 业务方法 ----

// CreateProbationaryRecord 创建预备期考察记录（POST /api/v1/ty/probationary-records）。
//
// 校验规则：
//   - summary ≥ 100字
//   - 每季度仅允许1条记录（unique约束：application_id + year + quarter）
func (s *ProbationaryService) CreateProbationaryRecord(userID int64, req *CreateProbationaryRecordRequest, actorName, actorRole, ip, ua string) (*ProbationaryRecordView, error) {
	// 校验季度范围
	if req.RecordQuarter < 1 || req.RecordQuarter > 4 {
		return nil, fmt.Errorf("季度必须在1-4之间")
	}

	// 校验总结字数
	if utf8.RuneCountInString(req.Summary) < 100 {
		return nil, fmt.Errorf("考察总结须 ≥ 100字")
	}

	// 检查唯一性：每季度仅1条
	exists, err := s.repo.CheckQuarterlyRecordExists(req.ApplicationID, req.RecordYear, req.RecordQuarter)
	if err != nil {
		return nil, fmt.Errorf("检查季度记录失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("该季度已存在考察记录，不可重复创建")
	}

	record := models.TyProbationaryRecord{
		ApplicationID: req.ApplicationID,
		RecordYear:    req.RecordYear,
		RecordQuarter: req.RecordQuarter,
		Summary:       req.Summary,
	}

	if err := s.repo.CreateProbationaryRecord(&record); err != nil {
		return nil, fmt.Errorf("创建预备期考察记录失败: %w", err)
	}

	s.publishProbEvent("TyProbationaryRecordCreated", userID, actorRole, ip, ua, map[string]interface{}{
		"application_id": req.ApplicationID,
		"record_year":    req.RecordYear,
		"record_quarter": req.RecordQuarter,
	})

	return s.GetProbationaryRecordByID(record.ID)
}

// GetProbationaryRecordByID 获取预备期考察记录详情。
func (s *ProbationaryService) GetProbationaryRecordByID(id int64) (*ProbationaryRecordView, error) {
	record, err := s.repo.GetProbationaryRecordByID(id)
	if err != nil {
		return nil, err
	}
	return s.recordToView(*record), nil
}

// ListProbationaryRecordsByApplicationID 按申请ID查询所有考察记录。
func (s *ProbationaryService) ListProbationaryRecordsByApplicationID(applicationID int64) ([]ProbationaryRecordView, error) {
	records, err := s.repo.ListProbationaryRecordsByApplicationID(applicationID)
	if err != nil {
		return nil, err
	}

	views := make([]ProbationaryRecordView, 0, len(records))
	for _, r := range records {
		views = append(views, *s.recordToView(r))
	}
	return views, nil
}

// ListProbationaryRecords 分页查询预备期考察记录，applicationID 可选。
// applicationID 为 nil 时返回全部；page/pageSize 由调用方约束（≥1）。
func (s *ProbationaryService) ListProbationaryRecords(applicationID *int64, page, pageSize int) (*ProbationaryRecordListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := s.repo.ListProbationaryRecords(applicationID, page, pageSize)
	if err != nil {
		return nil, err
	}

	views := make([]ProbationaryRecordView, 0, len(records))
	for _, r := range records {
		views = append(views, *s.recordToView(r))
	}
	return &ProbationaryRecordListResult{
		Items:    views,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// CreateProbationaryMeeting 创建转正大会（POST /api/v1/ty/probationary-meetings）。
//
// 前置条件：
//   - 预备期满1年（become_probationary_at 距今 ≥ 365天）
//
// 校验规则：
//   - self_application_path 对应的转正申请书 ≥ 800字（此处仅校验路径非空，实际字数校验由前端或文件服务完成）
//   - 票数规则：actual >= expected*2/3, approve > actual/2
//
// 决策为 pass 时自动执行：
//   1. 更新 ty_member_roster.status 保持 active，设置 formal_join_at=今天
//   2. 更新 idx_student.political_status="正式团员" (member)
func (s *ProbationaryService) CreateProbationaryMeeting(userID int64, req *CreateProbationaryMeetingRequest, actorName, actorRole, ip, ua string) (*ProbationaryMeetingView, error) {
	// 获取入团申请信息以获取 student_id
	appRepo := repository.NewApplicationRepository(s.db)
	app, err := appRepo.GetByID(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("入团申请不存在")
	}

	// 获取团员花名册记录以检查预备期满时间
	roster, err := s.repo.GetMemberRosterByStudentID(app.StudentID)
	if err != nil {
		return nil, fmt.Errorf("未找到团员花名册记录，错误码:2630")
	}

	// 校验预备期满1年
	if roster.BecomeProbationaryAt == nil {
		return nil, fmt.Errorf("预备期开始时间未设置，错误码:2630")
	}
	now := time.Now()
	probationDuration := now.Sub(*roster.BecomeProbationaryAt).Hours() / 24
	if probationDuration < 365 {
		return nil, fmt.Errorf("预备期未满1年，错误码:2630")
	}

	// 校验票数规则：实到 ≥ 应到 * 2/3
	minActual := req.ExpectedCount * 2 / 3
	if req.ActualCount < minActual {
		return nil, fmt.Errorf("实到人数不足应到人数的2/3")
	}

	// 校验票数规则：赞成 > 实到 / 2
	if req.ApproveCount <= req.ActualCount/2 {
		return nil, fmt.Errorf("赞成票数不满足要求，须超过实到人数的一半")
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

	meeting := models.TyProbationaryMeeting{
		BizNo:               bizNo,
		ApplicationID:       req.ApplicationID,
		SelfApplicationPath: req.SelfApplicationPath,
		MeetingAt:           meetingAt,
		ExpectedCount:       req.ExpectedCount,
		ActualCount:         req.ActualCount,
		ApproveCount:        req.ApproveCount,
		Decision:            req.Decision,
	}

	// 事务处理：创建会议 + 如果通过则联动更新多个表 + 写入审批记录
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if txErr := tx.Create(&meeting).Error; txErr != nil {
			return fmt.Errorf("创建转正大会记录失败: %w", txErr)
		}

		if req.Decision == "pass" {
			if txErr := s.processPassDecision(tx, roster); txErr != nil {
				return txErr
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 写入审批记录：转正大会结果作为审批记录
	rec := &models.TyApprovalRecord{
		ApplicationID: req.ApplicationID,
		Module:        "probationary",
		TargetID:      meeting.ID,
		Step:          "meeting",
		ApproverID:    userID,
		ApproverName:  actorName,
		ApproverRole:  actorRole,
		Result:        map[string]string{"pass": "approve", "reject": "reject"}[req.Decision],
		Opinion:       fmt.Sprintf("转正大会：应到%d人，实到%d人，赞成%d票，决议：%s", req.ExpectedCount, req.ActualCount, req.ApproveCount, req.Decision),
		FromStatus:    "probationary",
		ToStatus:      map[string]string{"pass": "formal", "reject": "probationary"}[req.Decision],
		IP:            ip,
	}
	_ = s.appRepo.CreateApprovalRecord(rec)

	s.publishProbEvent("TyProbationaryMeetingCreated", userID, actorRole, ip, ua, map[string]interface{}{
		"application_id": req.ApplicationID,
		"decision":       req.Decision,
	})

	return s.GetProbationaryMeetingByID(meeting.ID)
}

// processPassDecision 处理转正大会通过的联动操作（事务内调用）。
func (s *ProbationaryService) processPassDecision(tx *gorm.DB, roster *models.TyMemberRoster) error {
	now := time.Now()

	// 1. 更新 ty_member_roster：保持 status=active，设置 formal_join_at=今天
	if err := tx.Model(&models.TyMemberRoster{}).
		Where("id = ? AND is_deleted = 0", roster.ID).
		Updates(map[string]interface{}{
			"formal_join_at": now,
		}).Error; err != nil {
		return fmt.Errorf("更新团员花名册失败: %w", err)
	}

	// 2. 更新 idx_student.political_status 为"正式团员"(member)
	if err := tx.Model(&models.IdxStudent{}).
		Where("id = ? AND is_deleted = 0", roster.StudentID).
		Update("political_status", "member").Error; err != nil {
		// 非致命错误
		fmt.Printf("[TY-Probationary] 更新学生政治面貌失败: studentID=%d err=%v\n", roster.StudentID, err)
	}

	return nil
}

// GetProbationaryMeetingByID 获取转正大会详情。
func (s *ProbationaryService) GetProbationaryMeetingByID(id int64) (*ProbationaryMeetingView, error) {
	meeting, err := s.repo.GetProbationaryMeetingByID(id)
	if err != nil {
		return nil, err
	}
	return s.meetingToView(*meeting), nil
}

// ListProbationaryMeetingsByApplicationID 按申请ID查询转正大会列表。
func (s *ProbationaryService) ListProbationaryMeetingsByApplicationID(applicationID int64) ([]ProbationaryMeetingView, error) {
	meetings, err := s.repo.ListProbationaryMeetingsByApplicationID(applicationID)
	if err != nil {
		return nil, err
	}

	views := make([]ProbationaryMeetingView, 0, len(meetings))
	for _, m := range meetings {
		views = append(views, *s.meetingToView(m))
	}
	return views, nil
}

// ListProbationaryMeetings 分页查询转正大会，applicationID 可选。
// applicationID 为 nil 时返回全部；page/pageSize 由调用方约束（≥1）。
func (s *ProbationaryService) ListProbationaryMeetings(applicationID *int64, page, pageSize int) (*ProbationaryMeetingListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	meetings, total, err := s.repo.ListProbationaryMeetings(applicationID, page, pageSize)
	if err != nil {
		return nil, err
	}

	views := make([]ProbationaryMeetingView, 0, len(meetings))
	for _, m := range meetings {
		views = append(views, *s.meetingToView(m))
	}
	return &ProbationaryMeetingListResult{
		Items:    views,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ---- 内部方法 ----

// recordToView 将考察记录模型转为视图。
func (s *ProbationaryService) recordToView(record models.TyProbationaryRecord) *ProbationaryRecordView {
	v := &ProbationaryRecordView{
		ID:            record.ID,
		ApplicationID: record.ApplicationID,
		RecordYear:    record.RecordYear,
		RecordQuarter: record.RecordQuarter,
		Summary:       record.Summary,
		CreatedAt:     record.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:     record.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	s.fillStudentInfo(v.ApplicationID, func(studentID int64, name, no string) {
		v.StudentID = studentID
		v.StudentName = name
		v.StudentNo = no
	})
	return v
}

// meetingToView 将转正大会模型转为视图。
func (s *ProbationaryService) meetingToView(meeting models.TyProbationaryMeeting) *ProbationaryMeetingView {
	v := &ProbationaryMeetingView{
		ID:                  meeting.ID,
		BizNo:               meeting.BizNo,
		ApplicationID:       meeting.ApplicationID,
		SelfApplicationPath: meeting.SelfApplicationPath,
		MeetingAt:           meeting.MeetingAt.Format("2006-01-02T15:04:05+08:00"),
		ExpectedCount:       meeting.ExpectedCount,
		ActualCount:         meeting.ActualCount,
		ApproveCount:        meeting.ApproveCount,
		Decision:            meeting.Decision,
		DecisionText:        probationaryDecisionTextMap[meeting.Decision],
		CreatedAt:           meeting.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:           meeting.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	if meeting.FormalJoinAt != nil {
		t := meeting.FormalJoinAt.Format("2006-01-02")
		v.FormalJoinAt = &t
	}

	s.fillStudentInfo(v.ApplicationID, func(studentID int64, name, no string) {
		v.StudentID = studentID
		v.StudentName = name
		v.StudentNo = no
	})
	return v
}

// fillStudentInfo 通过入团申请ID补齐学生信息（姓名/学号）。
// 应用层查询失败不阻塞记录返回，仅留空。
func (s *ProbationaryService) fillStudentInfo(applicationID int64, apply func(studentID int64, name, no string)) {
	if applicationID <= 0 || s.appRepo == nil {
		return
	}
	app, err := s.appRepo.GetByID(applicationID)
	if err != nil || app == nil {
		return
	}
	student, err := s.appRepo.GetStudentByID(app.StudentID)
	if err != nil || student == nil {
		return
	}
	apply(student.ID, student.Name, student.StudentNo)
}

// publishProbEvent 发布预备期相关事件。
func (s *ProbationaryService) publishProbEvent(evtType string, actorID int64, actorRole, ip, ua string, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "ty.probationary",
		AggregateID: "",
		EventType:   evtType,
		Module:      "TY",
		ActorID:     actorID,
		ActorRole:   actorRole,
		Payload:     payload,
		BizNo:       "",
		IP:          ip,
		UA:          ua,
	})
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/st/repository"
	stsm "student-system/internal/modules/st/statemachine"
	"student-system/internal/statem"
)

// ActivityService 活动业务服务层。
type ActivityService struct {
	repo *repository.ActivityRepository
	db   *gorm.DB
	bus  *eventx.Bus
}

// NewActivityService 创建活动服务。
func NewActivityService(repo *repository.ActivityRepository, db *gorm.DB, bus *eventx.Bus) *ActivityService {
	return &ActivityService{repo: repo, db: db, bus: bus}
}

// ---- DTO ----

// ActivityListResult 活动列表结果。
type ActivityListResult struct {
	Items    []ActivityView `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ActivityView 活动视图。
type ActivityView struct {
	ID                   int64  `json:"id"`
	BizNo                string `json:"biz_no"`
	AssociationID        int64  `json:"association_id"`
	AssociationName      string `json:"association_name"`
	Title                string `json:"title"`
	Level                string `json:"level"`
	ExpectedParticipants int    `json:"expected_participants"`
	BudgetCents          int64  `json:"budget_cents"`
	Location             string `json:"location"`
	StartedAt            string `json:"started_at"`
	EndedAt              string `json:"ended_at"`
	Status               string `json:"status"`
	StatusText           string `json:"status_text"`
	RejectCount          int    `json:"reject_count"`
	ApprovalChainCompleted bool `json:"approval_chain_completed"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

// CreateActivityRequest 创建活动请求。
type CreateActivityRequest struct {
	AssociationID        int64  `json:"association_id" binding:"required"`
	Title                string `json:"title" binding:"required"`
	Level                string `json:"level" binding:"required"`
	ExpectedParticipants int    `json:"expected_participants" binding:"required"`
	BudgetCents          int64  `json:"budget_cents"`
	PlanFileID           *int64 `json:"plan_file_id"`
	EmergencyPlanFileID  *int64 `json:"emergency_plan_file_id"`
	SafetyCommitFileID   *int64 `json:"safety_commit_file_id"`
	Location             string `json:"location" binding:"required"`
	StartedAt            string `json:"started_at" binding:"required"`
	EndedAt              string `json:"ended_at" binding:"required"`
	ExpectedCount        *int   `json:"expected_count"`
}

// UpdateActivityRequest 更新活动请求。
type UpdateActivityRequest struct {
	Title                *string `json:"title"`
	Level                *string `json:"level"`
	ExpectedParticipants *int    `json:"expected_participants"`
	BudgetCents          *int64  `json:"budget_cents"`
	Location             *string `json:"location"`
	StartedAt            *string `json:"started_at"`
	EndedAt              *string `json:"ended_at"`
}

// ---- 审批/签到/总结 DTO ----

// ApproveActivityRequest 审批活动请求。
type ApproveActivityRequest struct {
	StepNo  int    `json:"step_no" binding:"required"`
	Result  string `json:"result" binding:"required"`  // approve / reject
	Opinion string `json:"opinion"` // 驳回时≥30字
}

// CheckinRequest 签到请求。
type CheckinRequest struct {
	StudentID int64    `json:"student_id"` // 可选：社长代签时指定学生
	Method    string   `json:"method" binding:"required"` // qrcode / gps / manual
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
}

// SubmitSummaryRequest 提交活动总结请求。
type SubmitSummaryRequest struct {
	Participants     int    `json:"participants" binding:"required"`
	Photos           []int64 `json:"photos"`
	GoalScore        *int   `json:"goal_score"` // 0-5 目标达成度
	Improvements     string `json:"improvements"`
}

// ApprovalView 审批记录视图。
type ApprovalView struct {
	ID           int64  `json:"id"`
	ActivityID   int64  `json:"activity_id"`
	StepNo       int    `json:"step_no"`
	StepText     string `json:"step_text"`
	ApproverRole string `json:"approver_role"`
	ApproverID   int64  `json:"approver_id,omitempty"`
	ApproverName string `json:"approver_name,omitempty"`
	Decision     string `json:"decision"`
	DecisionText string `json:"decision_text"`
	Opinion      string `json:"opinion"`
	DecidedAt    string `json:"decided_at,omitempty"`
}

// CheckinView 签到视图。
type CheckinView struct {
	ID                int64  `json:"id"`
	ActivityID        int64  `json:"activity_id"`
	StudentID         int64  `json:"student_id"`
	StudentName       string `json:"student_name"`
	CheckinAt         string `json:"checkin_at"`
	Method            string `json:"method"`
	MethodText        string `json:"method_text"`
	IsLate            bool   `json:"is_late"`
	LateMinutes       int    `json:"late_minutes"`
	CountedAsAbsent   bool   `json:"counted_as_absent"`
	IsPresent         int    `json:"is_present"`
}

// CheckinListResult 签到列表结果。
type CheckinListResult struct {
	Items    []CheckinView `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// SummaryView 活动总结视图。
type SummaryView struct {
	ID           int64    `json:"id"`
	ActivityID   int64    `json:"activity_id"`
	Participants int      `json:"participants"`
	GoalScore    *int     `json:"goal_score"`
	Improvements string   `json:"improvements"`
	Photos       []int64  `json:"photos"`
	SubmittedAt  string   `json:"submitted_at"`
	IsOverdue    int      `json:"is_overdue"`
}

// TimelineEntry 事件时间线条目。
type TimelineEntry struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type"`
	ActorID    int64                  `json:"actor_id"`
	ActorRole  string                 `json:"actor_role"`
	OccurredAt string                 `json:"occurred_at"`
	Payload    map[string]interface{} `json:"payload"`
}

// ---- 状态映射 ----

var activityStatusTextMap = map[string]string{
	"S0":        "草稿",
	"S1":        "待审",
	"S2":        "审批中",
	"S3":        "通过",
	"S4":        "驳回",
	"cancelled": "已取消",
}

// 签到方式中文映射。
var checkinMethodTextMap = map[string]string{
	"qrcode": "二维码签到",
	"gps":    "GPS签到",
	"manual": "手动签到",
}

// 审批结果中文映射。
var decisionTextMap = map[string]string{
	"pass":   "通过",
	"reject": "驳回",
}

// ---- 业务方法 ----

// List 分页查询活动列表。
func (s *ActivityService) List(associationID int64, status string, page, pageSize int) (*ActivityListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	acts, total, err := s.repo.List(associationID, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]ActivityView, 0, len(acts))
	for _, a := range acts {
		items = append(items, s.toView(a))
	}

	return &ActivityListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取活动详情。
func (s *ActivityService) Get(id int64) (*ActivityView, error) {
	act, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	v := s.toView(*act)
	return &v, nil
}

// Create 创建活动（保存为 S0 草稿）。
func (s *ActivityService) Create(userID int64, req *CreateActivityRequest) (*ActivityView, error) {
	// 校验活动等级
	if req.Level != "A" && req.Level != "B" && req.Level != "C" && req.Level != "D" {
		return nil, fmt.Errorf("活动等级必须为 A/B/C/D")
	}

	// 校验社团存在
	if _, err := s.repo.GetAssociationByID(req.AssociationID); err != nil {
		return nil, fmt.Errorf("社团不存在")
	}

	// 校验 A/B 级必须上传应急预案
	if (req.Level == "A" || req.Level == "B") && req.EmergencyPlanFileID == nil {
		return nil, fmt.Errorf("A/B 级活动必须上传应急预案")
	}

	// 校验预算非负
	if req.BudgetCents < 0 {
		return nil, fmt.Errorf("预算不能为负数")
	}

	// 解析时间
	startedAt, err := time.Parse(time.RFC3339, req.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误，请使用 RFC3339 格式")
	}
	endedAt, err := time.Parse(time.RFC3339, req.EndedAt)
	if err != nil {
		return nil, fmt.Errorf("结束时间格式错误，请使用 RFC3339 格式")
	}
	if !endedAt.After(startedAt) {
		return nil, fmt.Errorf("结束时间必须晚于开始时间")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "ST")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	act := &models.StActivity{
		BizNo:                bizNo,
		AssociationID:        req.AssociationID,
		Title:                req.Title,
		ActivityLevel:        req.Level,
		ExpectedParticipants: req.ExpectedParticipants,
		BudgetCents:          req.BudgetCents,
		PlanFileID:           req.PlanFileID,
		EmergencyPlanFileID:  req.EmergencyPlanFileID,
		SafetyCommitFileID:   req.SafetyCommitFileID,
		Location:             req.Location,
		StartedAt:            startedAt,
		EndedAt:              endedAt,
		ExpectedCount:        req.ExpectedCount,
		Status:               "S0",
	}

	if err := s.repo.Create(act); err != nil {
		return nil, err
	}

	// 发布事件
	s.publishEvent(act, "StActivityCreated", userID, "", "", "", map[string]interface{}{
		"activity_id":    act.ID,
		"biz_no":         act.BizNo,
		"title":          act.Title,
		"level":          act.ActivityLevel,
		"association_id": act.AssociationID,
	})

	return s.Get(act.ID)
}

// Update 更新活动（仅 S0 草稿状态可改）。
func (s *ActivityService) Update(id, userID int64, req *UpdateActivityRequest) (*ActivityView, error) {
	act, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("活动不存在")
	}

	if act.Status != "S0" {
		return nil, fmt.Errorf("仅草稿状态可修改")
	}

	if req.Title != nil {
		if utf8.RuneCountInString(*req.Title) == 0 {
			return nil, fmt.Errorf("活动标题不能为空")
		}
		act.Title = *req.Title
	}
	if req.Level != nil {
		act.ActivityLevel = *req.Level
	}
	if req.ExpectedParticipants != nil {
		act.ExpectedParticipants = *req.ExpectedParticipants
	}
	if req.BudgetCents != nil {
		act.BudgetCents = *req.BudgetCents
	}
	if req.Location != nil {
		act.Location = *req.Location
	}
	if req.StartedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartedAt)
		if err != nil {
			return nil, fmt.Errorf("开始时间格式错误")
		}
		act.StartedAt = t
	}
	if req.EndedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndedAt)
		if err != nil {
			return nil, fmt.Errorf("结束时间格式错误")
		}
		act.EndedAt = t
	}

	if err := s.repo.Update(act); err != nil {
		return nil, err
	}

	return s.Get(act.ID)
}

// Submit 提交活动（S0 → S1）。
func (s *ActivityService) Submit(id, userID int64, actorName, actorRole, ip, ua string) (*ActivityView, error) {
	act, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("活动不存在")
	}

	if act.Status != stsm.StateDraft {
		return nil, fmt.Errorf("仅草稿状态可提交")
	}

	// 使用状态机引擎
	sm, err := stsm.NewActivitySM(act.ActivityLevel)
	if err != nil {
		return nil, err
	}

	to, err := sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, act.Status, stsm.ActionSubmit)
	if err != nil {
		return nil, err
	}

	act.Status = to
	act.LastAction = stsm.ActionSubmit
	if err := s.repo.Update(act); err != nil {
		return nil, err
	}

	s.publishEvent(act, "StActivitySubmitted", userID, actorRole, ip, ua, map[string]interface{}{
		"from": stsm.StateDraft,
		"to":   to,
	})

	return s.Get(act.ID)
}

// SoftDelete 软删除活动（仅 S0/S4）。
func (s *ActivityService) SoftDelete(id, userID int64) error {
	act, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("活动不存在")
	}
	if act.Status != "S0" && act.Status != "S4" {
		return fmt.Errorf("仅草稿或驳回状态可删除")
	}
	return s.repo.SoftDelete(id)
}

// ---- 审批流 ----

// Withdraw 撤回活动（S1 → S0）。
func (s *ActivityService) Withdraw(id, userID int64, actorName, actorRole, ip, ua string) (*ActivityView, error) {
	act, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("活动不存在")
	}

	if act.Status != stsm.StatePending {
		return nil, fmt.Errorf("仅待审状态可撤回")
	}

	sm, err := stsm.NewActivitySM(act.ActivityLevel)
	if err != nil {
		return nil, err
	}

	to, err := sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, act.Status, stsm.ActionWithdraw)
	if err != nil {
		return nil, err
	}

	act.Status = to
	act.LastAction = stsm.ActionWithdraw
	if err := s.repo.Update(act); err != nil {
		return nil, err
	}

	s.publishEvent(act, "StActivityWithdrawn", userID, actorRole, ip, ua, map[string]interface{}{
		"from": stsm.StatePending,
		"to":   to,
	})

	return s.Get(act.ID)
}

// Approve 审批活动（分级动态审批链）。
func (s *ActivityService) Approve(id, userID int64, req *ApproveActivityRequest, actorName, actorRole, ip, ua string) (*ActivityView, error) {
	if req == nil {
		return nil, fmt.Errorf("参数不能为空")
	}
	if req.Result != "approve" && req.Result != "reject" {
		return nil, fmt.Errorf("无效的审批结果")
	}

	// 1. 校验活动存在且状态为 S1 或 S2
	act, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("活动不存在")
	}
	if act.Status != stsm.StatePending && act.Status != stsm.StateInReview {
		return nil, fmt.Errorf("活动状态不允许审批")
	}

	// 2. 校验 stepNo 在当前等级的审批链范围内
	maxStep := stsm.MaxStepNo(act.ActivityLevel)
	if req.StepNo < 1 || req.StepNo > maxStep {
		return nil, fmt.Errorf("审批步骤编号超出范围（1~%d）", maxStep)
	}

	// 3. 校验前置步骤已通过
	if req.StepNo > 1 {
		ok, err := s.repo.HasApprovedStep(id, req.StepNo-1)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("前置审批步骤尚未通过")
		}
	}

	// 4. 校验当前步骤未审批过
	existing, err := s.repo.GetApprovalByStep(id, req.StepNo)
	if err == nil && existing != nil && existing.Decision == "pass" {
		return nil, fmt.Errorf("该步骤已通过，请勿重复审批")
	}
	// 已有审批记录（不管什么结果），不允许重复审批
	if err == nil && existing != nil {
		return nil, fmt.Errorf("该步骤已审批，请勿重复操作")
	}

	// 5. 校验角色权限
	roles, err := s.repo.FindUserRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败")
	}
	if !hasAnyRole(roles, stsm.StepRolesOfNo(req.StepNo)...) {
		return nil, fmt.Errorf("无该步骤审批权限")
	}

	// 6. 驳回时意见必须 ≥30字
	if req.Result == "reject" && utf8.RuneCountInString(req.Opinion) < 30 {
		return nil, fmt.Errorf("驳回意见至少 30 字")
	}

	// 7. 使用状态机引擎计算 to 状态
	sm, err := stsm.NewActivitySM(act.ActivityLevel)
	if err != nil {
		return nil, err
	}

	action := stsm.ResolveAction(req.StepNo, req.Result)
	from := act.Status
	to, err := sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
		Payload: map[string]interface{}{
			"step_no": req.StepNo,
			"result":  req.Result,
			"opinion": req.Opinion,
		},
	}, from, action)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// 8. 事务：写入审批记录 + 更新活动状态
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		decision := "pass"
		if req.Result == "reject" {
			decision = "reject"
		}

		rec := &models.StActivityApproval{
			ActivityID:     act.ID,
			StepNo:         req.StepNo,
			ApproverRole:   primaryRole(roles),
			ApproverUserID: &userID,
			Decision:       decision,
			Opinion:        req.Opinion,
			DecidedAt:      &now,
		}
		if err := tx.Create(rec).Error; err != nil {
			return fmt.Errorf("写入审批记录失败: %w", err)
		}

		act.Status = to
		act.LastAction = action
		// 10. 驳回时 reject_count +1
		if req.Result == "reject" {
			act.RejectCount++
		}
		return tx.Save(act).Error
	}); err != nil {
		return nil, err
	}

	// 10. 累计驳回≥3 记录事件
	if req.Result == "reject" && act.RejectCount >= 3 {
		s.publishEvent(act, "StActivityRejectThreshold", userID, actorRole, ip, ua, map[string]interface{}{
			"reject_count": act.RejectCount,
			"step_no":      req.StepNo,
		})
	}

	// 发布审批事件
	eventType := "StActivityApproved"
	if req.Result == "reject" {
		eventType = "StActivityRejected"
	}
	s.publishEvent(act, eventType, userID, actorRole, ip, ua, map[string]interface{}{
		"step_no":  req.StepNo,
		"result":   req.Result,
		"opinion":  req.Opinion,
		"from":     from,
		"to":       to,
		"approver": actorName,
	})

	return s.Get(act.ID)
}

// ListApprovals 列出审批记录。
func (s *ActivityService) ListApprovals(activityID int64) ([]ApprovalView, error) {
	records, err := s.repo.ListApprovals(activityID)
	if err != nil {
		return nil, err
	}

	views := make([]ApprovalView, 0, len(records))
	for _, r := range records {
		v := ApprovalView{
			ID:           r.ID,
			ActivityID:   r.ActivityID,
			StepNo:       r.StepNo,
			StepText:     stsm.StepTextOfNo(r.StepNo),
			ApproverRole: r.ApproverRole,
			Decision:     r.Decision,
			DecisionText: decisionTextMap[r.Decision],
			Opinion:      r.Opinion,
		}
		if r.ApproverUserID != nil {
			v.ApproverID = *r.ApproverUserID
		}
		if r.DecidedAt != nil {
			v.DecidedAt = r.DecidedAt.Format("2006-01-02T15:04:05+08:00")
		}
		// 加载审批人姓名
		if r.ApproverUserID != nil {
			if user, err := s.repo.GetUserByID(*r.ApproverUserID); err == nil {
				v.ApproverName = user.DisplayName
			}
		}
		views = append(views, v)
	}
	return views, nil
}

// Timeline 事件时间线。
func (s *ActivityService) Timeline(activityID int64) ([]TimelineEntry, error) {
	act, err := s.repo.GetByID(activityID)
	if err != nil {
		return nil, fmt.Errorf("活动不存在")
	}

	logs, err := s.bus.QueryByAggregate("st.activity", act.BizNo)
	if err != nil {
		return nil, err
	}

	entries := make([]TimelineEntry, 0, len(logs))
	for _, l := range logs {
		var payload map[string]interface{}
		if l.PayloadJSON != "" {
			_ = json.Unmarshal([]byte(l.PayloadJSON), &payload)
		}
		entries = append(entries, TimelineEntry{
			EventID:    l.EventID,
			EventType:  l.EventType,
			ActorID:    l.ActorID,
			ActorRole:  l.ActorRole,
			OccurredAt: l.OccurredAt.Format("2006-01-02T15:04:05+08:00"),
			Payload:    payload,
		})
	}
	return entries, nil
}

// ---- 签到 ----

// Checkin 签到。
func (s *ActivityService) Checkin(activityID, studentID int64, req *CheckinRequest) (*CheckinView, error) {
	// 1. 校验活动存在且状态为 S3
	act, err := s.repo.GetByID(activityID)
	if err != nil {
		return nil, fmt.Errorf("活动不存在")
	}
	if act.Status != stsm.StatePassed {
		return nil, fmt.Errorf("活动未通过审批，不可签到")
	}

	// 2. 校验当前时间在签到窗口内：活动开始前30分钟 ~ 活动开始后15分钟
	now := time.Now()
	checkinOpen := act.StartedAt.Add(-30 * time.Minute)
	checkinClose := act.StartedAt.Add(15 * time.Minute)
	if now.Before(checkinOpen) {
		return nil, fmt.Errorf("签到尚未开始（活动开始前30分钟开放）")
	}

	// 5. 同一学生同一活动不可重复签到
	exists, err := s.repo.HasCheckin(activityID, studentID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("已签到，不可重复签到")
	}

	// 6. GPS 签到时校验经纬度不为零
	if req.Method == "gps" {
		if req.Lat == nil || req.Lng == nil || *req.Lat == 0 || *req.Lng == 0 {
			return nil, fmt.Errorf("GPS签到需提供有效经纬度")
		}
	}

	// 计算迟到信息
	isLate := 0
	lateMinutes := 0
	isPresent := 1

	if now.After(act.StartedAt) {
		// 迟到
		lateDuration := now.Sub(act.StartedAt)
		lateMinutes = int(lateDuration.Minutes())
		isLate = 1

		// 迟到超过15分钟视为缺勤
		if lateMinutes > 15 {
			isPresent = 0
		}
	}

	// 超过签到关闭时间（开始后15分钟）仍允许签到但标记缺勤
	if now.After(checkinClose) && isPresent == 1 {
		// 超过15分钟窗口，强制标记
		isLate = 1
		isPresent = 0
		lateDuration := now.Sub(act.StartedAt)
		lateMinutes = int(lateDuration.Minutes())
	}

	rec := &models.StActivityCheckin{
		ActivityID:  activityID,
		StudentID:   studentID,
		CheckinAt:   now,
		Method:      req.Method,
		IsLate:      isLate,
		LateMinutes: lateMinutes,
		IsPresent:   isPresent,
	}
	if err := s.repo.CreateCheckin(rec); err != nil {
		return nil, err
	}

	// 构建视图
	studentName := ""
	if student, err := s.repo.GetStudentByID(studentID); err == nil {
		studentName = student.Name
	}

	view := &CheckinView{
		ID:              rec.ID,
		ActivityID:      rec.ActivityID,
		StudentID:       rec.StudentID,
		StudentName:     studentName,
		CheckinAt:       rec.CheckinAt.Format("2006-01-02T15:04:05+08:00"),
		Method:          rec.Method,
		MethodText:      checkinMethodTextMap[rec.Method],
		IsLate:          rec.IsLate != 0,
		LateMinutes:     rec.LateMinutes,
		CountedAsAbsent: rec.IsPresent == 0,
		IsPresent:       rec.IsPresent,
	}

	return view, nil
}

// ListCheckins 签到列表。
func (s *ActivityService) ListCheckins(activityID int64, page, pageSize int) (*CheckinListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := s.repo.ListCheckins(activityID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]CheckinView, 0, len(records))
	for _, r := range records {
		studentName := ""
		if student, err := s.repo.GetStudentByID(r.StudentID); err == nil {
			studentName = student.Name
		}

		items = append(items, CheckinView{
			ID:              r.ID,
			ActivityID:      r.ActivityID,
			StudentID:       r.StudentID,
			StudentName:     studentName,
			CheckinAt:       r.CheckinAt.Format("2006-01-02T15:04:05+08:00"),
			Method:          r.Method,
			MethodText:      checkinMethodTextMap[r.Method],
			IsLate:          r.IsLate != 0,
			LateMinutes:     r.LateMinutes,
			CountedAsAbsent: r.IsPresent == 0,
			IsPresent:       r.IsPresent,
		})
	}

	return &CheckinListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ---- 总结 ----

// SubmitSummary 提交活动总结。
func (s *ActivityService) SubmitSummary(activityID, userID int64, req *SubmitSummaryRequest) (*SummaryView, error) {
	// 1. 校验活动存在且状态为 S3
	act, err := s.repo.GetByID(activityID)
	if err != nil {
		return nil, fmt.Errorf("活动不存在")
	}
	if act.Status != stsm.StatePassed {
		return nil, fmt.Errorf("活动未通过审批，不可提交总结")
	}

	// 2. 校验活动已结束
	now := time.Now()
	if now.Before(act.EndedAt) {
		return nil, fmt.Errorf("活动尚未结束，不可提交总结")
	}

	// 3. 校验同一活动不可重复提交总结
	if _, err := s.repo.GetSummaryByActivity(activityID); err == nil {
		return nil, fmt.Errorf("该活动已提交总结，不可重复提交")
	}

	// 4. 判断是否超期（活动结束3天后提交算超期）
	isOverdue := 0
	deadline := act.EndedAt.Add(3 * 24 * time.Hour)
	if now.After(deadline) {
		isOverdue = 1
	}

	rec := &models.StActivitySummary{
		ActivityID:         activityID,
		ActualParticipants: req.Participants,
		AchievementScore:   req.GoalScore,
		Suggestions:        req.Improvements,
		SubmittedAt:        now,
		IsOverdue:          isOverdue,
	}
	if err := s.repo.CreateSummary(rec); err != nil {
		return nil, err
	}

	// 保存活动现场照片关联
	photos := make([]int64, 0)
	if len(req.Photos) > 0 {
		now2 := time.Now()
		for _, fid := range req.Photos {
			if err := s.repo.CreatePhoto(&models.StActivityPhoto{
				ActivityID: activityID,
				FileID:     fid,
				TakenAt:    &now2,
			}); err == nil {
				photos = append(photos, fid)
			}
		}
	}

	s.publishEvent(act, "StActivitySummarySubmitted", userID, "", "", "", map[string]interface{}{
		"activity_id": activityID,
		"participants": req.Participants,
		"is_overdue":  isOverdue,
	})

	return &SummaryView{
		ID:           rec.ID,
		ActivityID:   rec.ActivityID,
		Participants: rec.ActualParticipants,
		GoalScore:    rec.AchievementScore,
		Improvements: rec.Suggestions,
		Photos:       photos,
		SubmittedAt:  rec.SubmittedAt.Format("2006-01-02T15:04:05+08:00"),
		IsOverdue:    rec.IsOverdue,
	}, nil
}

// GetSummary 获取活动总结。
func (s *ActivityService) GetSummary(activityID int64) (*SummaryView, error) {
	rec, err := s.repo.GetSummaryByActivity(activityID)
	if err != nil {
		return nil, fmt.Errorf("活动总结不存在")
	}

	photos, _ := s.repo.ListPhotosByActivity(activityID)
	photoIDs := make([]int64, 0, len(photos))
	for _, p := range photos {
		photoIDs = append(photoIDs, p.FileID)
	}

	return &SummaryView{
		ID:           rec.ID,
		ActivityID:   rec.ActivityID,
		Participants: rec.ActualParticipants,
		GoalScore:    rec.AchievementScore,
		Improvements: rec.Suggestions,
		Photos:       photoIDs,
		SubmittedAt:  rec.SubmittedAt.Format("2006-01-02T15:04:05+08:00"),
		IsOverdue:    rec.IsOverdue,
	}, nil
}

// ---- 内部方法 ----

func (s *ActivityService) toView(act models.StActivity) ActivityView {
	v := ActivityView{
		ID:                   act.ID,
		BizNo:                act.BizNo,
		AssociationID:        act.AssociationID,
		Title:                act.Title,
		Level:                act.ActivityLevel,
		ExpectedParticipants: act.ExpectedParticipants,
		BudgetCents:          act.BudgetCents,
		Location:             act.Location,
		StartedAt:            act.StartedAt.Format("2006-01-02T15:04:05+08:00"),
		EndedAt:              act.EndedAt.Format("2006-01-02T15:04:05+08:00"),
		Status:               act.Status,
		StatusText:           activityStatusTextMap[act.Status],
		RejectCount:          act.RejectCount,
		ApprovalChainCompleted: act.Status == stsm.StatePassed,
		CreatedAt:            act.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:            act.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 加载社团名称
	if assoc, err := s.repo.GetAssociationByID(act.AssociationID); err == nil {
		v.AssociationName = assoc.Name
	}

	return v
}

// publishEvent 发布业务事件到 event_log。
func (s *ActivityService) publishEvent(act *models.StActivity, evtType string, actorID int64, actorRole, ip, ua string, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["activity_id"] = act.ID
	payload["biz_no"] = act.BizNo
	payload["status"] = act.Status

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "st.activity",
		AggregateID: act.BizNo,
		EventType:   evtType,
		Module:      "ST",
		ActorID:     actorID,
		ActorRole:   actorRole,
		Payload:     payload,
		BizNo:       act.BizNo,
		IP:          ip,
		UA:          ua,
	})
}

// hasAnyRole 判断角色列表中是否包含任一目标角色。
func hasAnyRole(roles []string, targets ...string) bool {
	set := make(map[string]struct{}, len(targets))
	for _, t := range targets {
		set[t] = struct{}{}
	}
	for _, r := range roles {
		if _, ok := set[r]; ok {
			return true
		}
	}
	return false
}

// primaryRole 取一个最具代表性的角色（用于审批记录展示）。
func primaryRole(roles []string) string {
	priority := []string{"R-SY-ADMIN", "R-SY-LEAGUE", "R-COL-LEAGUE", "R-COL-COUN", "R-COL-TUTOR", "R-STU-LEAGUE", "R-STU-NORM"}
	for _, p := range priority {
		for _, r := range roles {
			if r == p {
				return p
			}
		}
	}
	if len(roles) > 0 {
		return roles[0]
	}
	return ""
}

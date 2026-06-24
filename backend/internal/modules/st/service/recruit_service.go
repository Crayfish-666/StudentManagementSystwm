// Package service 提供 ST 模块招新业务逻辑。
//
// 设计依据：docs/01 §5.3.4 招新子流程；docs/03 §6.2.5/6.2.6；docs/04 §7.3。
// 硬规则：
//   1. 招新计划 status 四态：S0 草稿 / S1 待审 / S3 通过（可投递）/ S4 驳回
//   2. 同一学生同一学年最多加入 3 个社团（已 accepted 计）
//   3. 招新结果录入期限：5 个工作日
//   4. 状态变更必须走 statem.Apply()，禁止直接 UPDATE
package service

import (
	"context"
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

// RecruitService 招新业务服务层。
type RecruitService struct {
	repo *repository.RecruitRepository
	db   *gorm.DB
	bus  *eventx.Bus
	sm   *statem.Engine
}

// NewRecruitService 创建招新服务。
func NewRecruitService(repo *repository.RecruitRepository, db *gorm.DB, bus *eventx.Bus) *RecruitService {
	return &RecruitService{
		repo: repo,
		db:   db,
		bus:  bus,
		sm:   stsm.NewRecruitPlanSM(),
	}
}

// ---- 招新计划 DTO ----

// RecruitPlanListResult 招新计划列表结果。
type RecruitPlanListResult struct {
	Items    []RecruitPlanView `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// RecruitPlanView 招新计划视图。
type RecruitPlanView struct {
	ID               int64   `json:"id"`
	BizNo            string  `json:"biz_no"`
	AssociationID    int64   `json:"association_id"`
	AssociationName  string  `json:"association_name"`
	Season           string  `json:"season"`
	SeasonText       string  `json:"season_text"`
	AcademicYear     string  `json:"academic_year"`
	TargetCount      int     `json:"target_count"`
	PlanFileID       *int64  `json:"plan_file_id,omitempty"`
	AssessmentMethod string  `json:"assessment_method"`
	InterviewAt      string  `json:"interview_at,omitempty"`
	Status           string  `json:"status"`
	StatusText       string  `json:"status_text"`
	ResultDeadline   string  `json:"result_deadline,omitempty"`
	IsFinished       int     `json:"is_finished"`
	FinishedAt       string  `json:"finished_at,omitempty"`
	FinishedBy       *int64  `json:"finished_by,omitempty"`
	FinishedReason   string  `json:"finished_reason,omitempty"`
	// RecruitPhase 招新阶段（计算字段，由 status + is_finished 推导）：
	//   - not_open  ：未发布（S0/S1/S4）
	//   - ongoing  ：招新中（S3 且 is_finished=0）
	//   - finished ：已结束（S3 且 is_finished=1）
	RecruitPhase     string `json:"recruit_phase"`
	RecruitPhaseText string `json:"recruit_phase_text"`
	ApplyCount       int64  `json:"apply_count"`
	AcceptedCount    int64  `json:"accepted_count"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// CreateRecruitPlanRequest 创建招新计划请求。
type CreateRecruitPlanRequest struct {
	AssociationID    int64   `json:"association_id" binding:"required"`
	Season           string  `json:"season" binding:"required"` // autumn / spring
	AcademicYear     string  `json:"academic_year" binding:"required"`
	TargetCount      int     `json:"target_count" binding:"required"`
	PlanFileID       *int64  `json:"plan_file_id"`
	AssessmentMethod string  `json:"assessment_method"`
	InterviewAt      *string `json:"interview_at"`
}

// UpdateRecruitPlanRequest 更新招新计划请求（仅 S0 可改）。
type UpdateRecruitPlanRequest struct {
	Season           *string `json:"season"`
	AcademicYear     *string `json:"academic_year"`
	TargetCount      *int    `json:"target_count"`
	PlanFileID       *int64  `json:"plan_file_id"`
	AssessmentMethod *string `json:"assessment_method"`
	InterviewAt      *string `json:"interview_at"`
}

// FinishRecruitPlanRequest 提前结束招新请求（不可逆）。
type FinishRecruitPlanRequest struct {
	Reason string `json:"reason"` // 可选：结束原因
}

// ---- 招新申请 DTO ----

// RecruitApplyListResult 招新申请列表结果。
type RecruitApplyListResult struct {
	Items    []RecruitApplyView `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// RecruitApplyView 招新申请视图。
type RecruitApplyView struct {
	ID                  int64  `json:"id"`
	PlanID              int64  `json:"plan_id"`
	PlanBizNo           string `json:"plan_biz_no"`
	PlanSeason          string `json:"plan_season"`
	PlanAcademicYear    string `json:"plan_academic_year"`
	AssociationID       int64  `json:"association_id"`
	AssociationName     string `json:"association_name"`
	StudentID           int64  `json:"student_id"`
	StudentNo           string `json:"student_no"`
	StudentName         string `json:"student_name"`
	ResumeFileID        *int64 `json:"resume_file_id,omitempty"`
	Result              string `json:"result"`
	ResultText          string `json:"result_text"`
	ResultAt            string `json:"result_at,omitempty"`
	CreatedAt           string `json:"created_at"`
}

// CreateRecruitApplyRequest 学生投递招新申请。
type CreateRecruitApplyRequest struct {
	PlanID       int64  `json:"plan_id" binding:"required"`
	ResumeFileID *int64 `json:"resume_file_id"`
}

// SubmitApplyResultRequest 录入面试结果。
type SubmitApplyResultRequest struct {
	Result string `json:"result" binding:"required"` // accepted / rejected
}

// ---- 状态映射 ----

var planStatusTextMap = map[string]string{
	"S0": "草稿",
	"S1": "待审",
	"S3": "已通过",
	"S4": "已驳回",
}

var planSeasonTextMap = map[string]string{
	"autumn": "秋季招新",
	"spring": "春季补招",
}

// 招新阶段枚举与中文文本。
const (
	planPhaseNotOpen  = "not_open"
	planPhaseOngoing  = "ongoing"
	planPhaseFinished = "finished"
)

var planPhaseTextMap = map[string]string{
	planPhaseNotOpen:  "未发布",
	planPhaseOngoing:  "招新中",
	planPhaseFinished: "已结束",
}

// calcRecruitPhase 根据 status + is_finished 推导招新阶段。
func calcRecruitPhase(status string, isFinished int) string {
	if status == stsm.PlanStatePassed {
		if isFinished == 1 {
			return planPhaseFinished
		}
		return planPhaseOngoing
	}
	return planPhaseNotOpen
}

var applyResultTextMap = map[string]string{
	"pending":  "待面试",
	"accepted": "已录用",
	"rejected": "未通过",
}

// 招新计划审批允许角色。
var planApproveRoles = []string{
	"R-SY-ADMIN", "R-SY-LEAGUE", "R-COL-LEAGUE", "R-COL-COUN", "R-COL-TUTOR",
}

// ---- 招新计划业务方法 ----

// ListPlans 分页查询招新计划。
func (s *RecruitService) ListPlans(associationID int64, status, academicYear string, page, pageSize int) (*RecruitPlanListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	plans, total, err := s.repo.ListPlans(associationID, status, academicYear, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]RecruitPlanView, 0, len(plans))
	for _, p := range plans {
		items = append(items, s.toPlanView(p))
	}
	return &RecruitPlanListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetPlan 查询招新计划详情。
func (s *RecruitService) GetPlan(id int64) (*RecruitPlanView, error) {
	plan, err := s.repo.GetPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	v := s.toPlanView(*plan)
	return &v, nil
}

// CreatePlan 创建招新计划（S0 草稿）。
func (s *RecruitService) CreatePlan(userID int64, req *CreateRecruitPlanRequest, actorName, actorRole, ip, ua string) (*RecruitPlanView, error) {
	// 1. 校验社团存在
	if _, err := s.repo.GetAssociationByID(req.AssociationID); err != nil {
		return nil, fmt.Errorf("社团不存在")
	}

	// 2. 校验季节
	if req.Season != "autumn" && req.Season != "spring" {
		return nil, fmt.Errorf("招新季节必须为 autumn(秋) 或 spring(春)")
	}

	// 3. 校验目标人数
	if req.TargetCount <= 0 {
		return nil, fmt.Errorf("目标人数必须大于 0")
	}

	// 4. 解析面试时间（可选）
	var interviewAt *time.Time
	if req.InterviewAt != nil && *req.InterviewAt != "" {
		t, err := time.Parse(time.RFC3339, *req.InterviewAt)
		if err != nil {
			return nil, fmt.Errorf("面试时间格式错误，请使用 RFC3339 格式")
		}
		interviewAt = &t
	}

	// 5. 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "ST")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	plan := &models.StRecruitPlan{
		BizNo:            bizNo,
		AssociationID:    req.AssociationID,
		Season:           req.Season,
		AcademicYear:     req.AcademicYear,
		TargetCount:      req.TargetCount,
		PlanFileID:       req.PlanFileID,
		AssessmentMethod: req.AssessmentMethod,
		InterviewAt:      interviewAt,
		Status:           stsm.PlanStateDraft,
	}
	if err := s.repo.CreatePlan(plan); err != nil {
		return nil, err
	}

	// 6. 发布事件
	s.publishPlanEvent(plan, "StRecruitPlanCreated", userID, actorRole, ip, ua, map[string]interface{}{
		"plan_id":       plan.ID,
		"biz_no":        plan.BizNo,
		"association":   plan.AssociationID,
		"season":        plan.Season,
		"academic_year": plan.AcademicYear,
	})

	return s.GetPlan(plan.ID)
}

// UpdatePlan 更新招新计划（仅 S0 草稿可改）。
func (s *RecruitService) UpdatePlan(id, userID int64, req *UpdateRecruitPlanRequest) (*RecruitPlanView, error) {
	plan, err := s.repo.GetPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	if plan.Status != stsm.PlanStateDraft {
		return nil, fmt.Errorf("仅草稿状态可修改")
	}

	if req.Season != nil {
		if *req.Season != "autumn" && *req.Season != "spring" {
			return nil, fmt.Errorf("招新季节必须为 autumn(秋) 或 spring(春)")
		}
		plan.Season = *req.Season
	}
	if req.AcademicYear != nil {
		if utf8.RuneCountInString(*req.AcademicYear) == 0 {
			return nil, fmt.Errorf("学年不能为空")
		}
		plan.AcademicYear = *req.AcademicYear
	}
	if req.TargetCount != nil {
		if *req.TargetCount <= 0 {
			return nil, fmt.Errorf("目标人数必须大于 0")
		}
		plan.TargetCount = *req.TargetCount
	}
	if req.PlanFileID != nil {
		plan.PlanFileID = req.PlanFileID
	}
	if req.AssessmentMethod != nil {
		plan.AssessmentMethod = *req.AssessmentMethod
	}
	if req.InterviewAt != nil {
		if *req.InterviewAt == "" {
			plan.InterviewAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.InterviewAt)
			if err != nil {
				return nil, fmt.Errorf("面试时间格式错误")
			}
			plan.InterviewAt = &t
		}
	}

	if err := s.repo.UpdatePlan(plan); err != nil {
		return nil, err
	}
	return s.GetPlan(plan.ID)
}

// SubmitPlan 提交招新计划（S0 → S1）。
func (s *RecruitService) SubmitPlan(id, userID int64, actorName, actorRole, ip, ua string) (*RecruitPlanView, error) {
	plan, err := s.repo.GetPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	if plan.Status != stsm.PlanStateDraft {
		return nil, fmt.Errorf("仅草稿状态可提交")
	}

	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, plan.Status, stsm.PlanActionSubmit)
	if err != nil {
		return nil, err
	}

	plan.Status = to
	if err := s.repo.UpdatePlan(plan); err != nil {
		return nil, err
	}

	s.publishPlanEvent(plan, "StRecruitPlanSubmitted", userID, actorRole, ip, ua, map[string]interface{}{
		"from": stsm.PlanStateDraft,
		"to":   to,
	})
	return s.GetPlan(plan.ID)
}

// WithdrawPlan 撤回招新计划（S1 → S0）。
func (s *RecruitService) WithdrawPlan(id, userID int64, actorName, actorRole, ip, ua string) (*RecruitPlanView, error) {
	plan, err := s.repo.GetPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	if plan.Status != stsm.PlanStatePending {
		return nil, fmt.Errorf("仅待审状态可撤回")
	}

	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, plan.Status, stsm.PlanActionWithdraw)
	if err != nil {
		return nil, err
	}

	plan.Status = to
	if err := s.repo.UpdatePlan(plan); err != nil {
		return nil, err
	}

	s.publishPlanEvent(plan, "StRecruitPlanWithdrawn", userID, actorRole, ip, ua, map[string]interface{}{
		"from": stsm.PlanStatePending,
		"to":   to,
	})
	return s.GetPlan(plan.ID)
}

// ApprovePlan 审批通过招新计划（S1 → S3）。
func (s *RecruitService) ApprovePlan(id, userID int64, actorName, actorRole, ip, ua string) (*RecruitPlanView, error) {
	// 1. 校验计划存在且状态
	plan, err := s.repo.GetPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	if plan.Status != stsm.PlanStatePending {
		return nil, fmt.Errorf("仅待审状态可审批")
	}

	// 2. 校验角色权限
	roles, err := s.repo.FindUserRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败")
	}
	if !hasAnyRole(roles, planApproveRoles...) {
		return nil, fmt.Errorf("无招新计划审批权限")
	}

	// 3. 状态机推进
	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, plan.Status, stsm.PlanActionApprove)
	if err != nil {
		return nil, err
	}

	plan.Status = to
	if err := s.repo.UpdatePlan(plan); err != nil {
		return nil, err
	}

	s.publishPlanEvent(plan, "StRecruitPlanApproved", userID, actorRole, ip, ua, map[string]interface{}{
		"from": stsm.PlanStatePending,
		"to":   to,
	})
	return s.GetPlan(plan.ID)
}

// RejectPlan 驳回招新计划（S1 → S4）。
func (s *RecruitService) RejectPlan(id, userID int64, opinion, actorName, actorRole, ip, ua string) (*RecruitPlanView, error) {
	// 1. 校验计划存在且状态
	plan, err := s.repo.GetPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	if plan.Status != stsm.PlanStatePending {
		return nil, fmt.Errorf("仅待审状态可驳回")
	}

	// 2. 校验角色权限
	roles, err := s.repo.FindUserRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败")
	}
	if !hasAnyRole(roles, planApproveRoles...) {
		return nil, fmt.Errorf("无招新计划审批权限")
	}

	// 3. 驳回意见必须 ≥10 字
	if utf8.RuneCountInString(opinion) < 10 {
		return nil, fmt.Errorf("驳回意见至少 10 字")
	}

	// 4. 状态机推进
	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
		Payload: map[string]interface{}{
			"opinion": opinion,
		},
	}, plan.Status, stsm.PlanActionReject)
	if err != nil {
		return nil, err
	}

	plan.Status = to
	if err := s.repo.UpdatePlan(plan); err != nil {
		return nil, err
	}

	s.publishPlanEvent(plan, "StRecruitPlanRejected", userID, actorRole, ip, ua, map[string]interface{}{
		"from":    stsm.PlanStatePending,
		"to":      to,
		"opinion": opinion,
	})
	return s.GetPlan(plan.ID)
}

// PublishPlan 发布招新计划（生成 5 工作日结果录入期限）。
func (s *RecruitService) PublishPlan(id, userID int64, actorName, actorRole, ip, ua string) (*RecruitPlanView, error) {
	plan, err := s.repo.GetPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	if plan.Status != stsm.PlanStatePassed {
		return nil, fmt.Errorf("仅已通过审批的计划可发布")
	}

	// 1. 状态机推进（S3 保持 S3）
	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, plan.Status, stsm.PlanActionPublish)
	if err != nil {
		return nil, err
	}

	// 2. 设置结果录入期限：发布后 5 个工作日（按 7 自然日兜底）
	deadline := time.Now().AddDate(0, 0, 7)
	plan.ResultDeadline = &deadline
	plan.Status = to
	if err := s.repo.UpdatePlan(plan); err != nil {
		return nil, err
	}

	s.publishPlanEvent(plan, "StRecruitPlanPublished", userID, actorRole, ip, ua, map[string]interface{}{
		"result_deadline": deadline.Format("2006-01-02"),
	})
	return s.GetPlan(plan.ID)
}

// FinishPlan 提前结束招新（仅 S3 + 未结束 状态可用，操作不可逆）。
//
// 业务规则：
//   - 必须 status=S3（已通过/可投递）
//   - 必须 is_finished=0（招新中）
//   - 不动 status 字段，仅更新 is_finished / finished_at / finished_by / finished_reason
//   - 不走 statem 状态机（status 字段无变化），但写业务事件留痕
func (s *RecruitService) FinishPlan(id, userID int64, req *FinishRecruitPlanRequest, actorName, actorRole, ip, ua string) (*RecruitPlanView, error) {
	// 1. 校验计划存在
	plan, err := s.repo.GetPlanByID(id)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	// 2. 仅 S3 可结束
	if plan.Status != stsm.PlanStatePassed {
		return nil, fmt.Errorf("仅已通过审批的招新计划可结束")
	}
	// 3. 未结束的才可结束（不可逆）
	if plan.IsFinished != 0 {
		return nil, fmt.Errorf("该招新计划已结束，不可重复结束")
	}
	// 4. 权限：与招新计划审批同集
	roles, err := s.repo.FindUserRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败")
	}
	if !hasAnyRole(roles, planApproveRoles...) {
		return nil, fmt.Errorf("无招新计划结束权限")
	}

	// 5. 写入结束标志
	now := time.Now()
	plan.IsFinished = 1
	plan.FinishedAt = &now
	plan.FinishedBy = &userID
	if req != nil {
		plan.FinishedReason = req.Reason
	}
	if err := s.repo.UpdatePlan(plan); err != nil {
		return nil, err
	}

	// 6. 业务事件留痕
	s.publishPlanEvent(plan, "StRecruitPlanFinished", userID, actorRole, ip, ua, map[string]interface{}{
		"reason":      plan.FinishedReason,
		"finished_at": now.Format("2006-01-02 15:04:05"),
	})
	return s.GetPlan(plan.ID)
}

// ---- 招新申请业务方法 ----

// ListApplies 分页查询招新申请。
func (s *RecruitService) ListApplies(planID, studentID int64, result string, page, pageSize int) (*RecruitApplyListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	apps, total, err := s.repo.ListApplies(planID, studentID, result, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 批量预加载 plan / association，避免 N+1
	planMap, assocMap := s.loadPlanAndAssocMaps(apps)

	items := make([]RecruitApplyView, 0, len(apps))
	for _, a := range apps {
		items = append(items, s.toApplyView(a, planMap, assocMap))
	}
	return &RecruitApplyListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// loadPlanAndAssocMaps 一次性预加载本页申请关联的招新计划和社团。
func (s *RecruitService) loadPlanAndAssocMaps(apps []models.StRecruitApply) (map[int64]*models.StRecruitPlan, map[int64]*models.StAssociation) {
	planMap := make(map[int64]*models.StRecruitPlan, len(apps))
	if len(apps) == 0 {
		return planMap, map[int64]*models.StAssociation{}
	}
	planIDs := make([]int64, 0, len(apps))
	seen := make(map[int64]bool, len(apps))
	for _, a := range apps {
		if !seen[a.PlanID] {
			planIDs = append(planIDs, a.PlanID)
			seen[a.PlanID] = true
		}
	}
	plans, _ := s.repo.ListPlansByIDs(planIDs)
	for i := range plans {
		planMap[plans[i].ID] = &plans[i]
	}
	// 聚合 plan 中的 association_id 集合
	assocIDs := make([]int64, 0, len(plans))
	seenA := make(map[int64]bool, len(plans))
	for _, p := range plans {
		if !seenA[p.AssociationID] {
			assocIDs = append(assocIDs, p.AssociationID)
			seenA[p.AssociationID] = true
		}
	}
	assocMap, _ := s.repo.ListAssociationsByIDs(assocIDs)
	return planMap, assocMap
}

// CreateApply 学生投递招新申请。
func (s *RecruitService) CreateApply(userID, studentID int64, req *CreateRecruitApplyRequest, actorName, actorRole, ip, ua string) (*RecruitApplyView, error) {
	// 1. 校验计划存在且已发布
	plan, err := s.repo.GetPlanByID(req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}
	if plan.Status != stsm.PlanStatePassed {
		return nil, fmt.Errorf("该招新计划未发布，不可投递")
	}
	if plan.IsFinished != 0 {
		return nil, fmt.Errorf("该招新计划已结束，不可投递")
	}

	// 2. 校验学生存在
	if _, err := s.repo.GetStudentByID(studentID); err != nil {
		return nil, fmt.Errorf("学生不存在")
	}

	// 3. 同一学生同一计划不可重复投递
	exists, err := s.repo.HasApplyInPlan(req.PlanID, studentID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("您已投递过该招新计划，不可重复投递")
	}

	// 4. 硬卡控：同一学生同一学年最多加入 3 个社团
	count, err := s.repo.CountAcceptedAssociationsInYear(studentID, plan.AcademicYear)
	if err != nil {
		return nil, err
	}
	if count >= 3 {
		return nil, fmt.Errorf("您本学年已加入 %d 个社团，最多加入 3 个", count)
	}

	// 5. 创建申请
	app := &models.StRecruitApply{
		PlanID:       req.PlanID,
		StudentID:    studentID,
		ResumeFileID: req.ResumeFileID,
		Result:       "pending",
	}
	if err := s.repo.CreateApply(app); err != nil {
		return nil, err
	}

	// 6. 发布事件
	if s.bus != nil {
		_ = s.bus.Publish(&eventx.Event{
			Aggregate:   "st.recruit_apply",
			AggregateID: applyBizNo(app.ID),
			EventType:   "StRecruitApplySubmitted",
			Module:      "ST",
			ActorID:     userID,
			ActorRole:   actorRole,
			Payload: map[string]interface{}{
				"apply_id":   app.ID,
				"plan_id":    app.PlanID,
				"student_id": app.StudentID,
			},
			BizNo: applyBizNo(app.ID),
			IP:    ip,
			UA:    ua,
		})
	}

	// 单条预加载 plan + association
	planMap, assocMap := s.loadPlanAndAssocMaps([]models.StRecruitApply{*app})
	v := s.toApplyView(*app, planMap, assocMap)
	return &v, nil
}

// SubmitApplyResult 录入面试结果（accepted/rejected）。
// accepted 时若学生不在社团成员名单中，则自动加入 st_assoc_member。
func (s *RecruitService) SubmitApplyResult(applyID, userID int64, req *SubmitApplyResultRequest, actorName, actorRole, ip, ua string) (*RecruitApplyView, error) {
	if req.Result != "accepted" && req.Result != "rejected" {
		return nil, fmt.Errorf("结果必须为 accepted 或 rejected")
	}

	app, err := s.repo.GetApplyByID(applyID)
	if err != nil {
		return nil, fmt.Errorf("招新申请不存在")
	}
	if app.Result != "pending" {
		return nil, fmt.Errorf("该申请已录入结果，不可重复操作")
	}

	// 校验计划存在（用于回填社团成员）
	plan, err := s.repo.GetPlanByID(app.PlanID)
	if err != nil {
		return nil, fmt.Errorf("招新计划不存在")
	}

	now := time.Now()
	app.Result = req.Result
	app.ResultAt = &now

	// 事务：更新申请结果 + 录用时加入社团成员
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(app).Error; err != nil {
			return fmt.Errorf("更新申请结果失败: %w", err)
		}

		// 录用：自动加入社团成员表
		if req.Result == "accepted" {
			// 检查是否已是成员
			var existCount int64
			if err := tx.Model(&models.StAssocMember{}).
				Where("association_id = ? AND student_id = ? AND is_deleted = 0 AND left_at IS NULL",
					plan.AssociationID, app.StudentID).
				Count(&existCount).Error; err != nil {
				return err
			}
			if existCount == 0 {
				member := &models.StAssocMember{
					AssociationID: plan.AssociationID,
					StudentID:     app.StudentID,
					Role:          "member",
					JoinedAt:      now,
					IsCoreOfficer: 0,
				}
				if err := tx.Create(member).Error; err != nil {
					return fmt.Errorf("加入社团成员失败: %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 发布事件
	if s.bus != nil {
		_ = s.bus.Publish(&eventx.Event{
			Aggregate:   "st.recruit_apply",
			AggregateID: applyBizNo(app.ID),
			EventType:   "StRecruitApplyResulted",
			Module:      "ST",
			ActorID:     userID,
			ActorRole:   actorRole,
			Payload: map[string]interface{}{
				"apply_id":   app.ID,
				"plan_id":    app.PlanID,
				"student_id": app.StudentID,
				"result":     req.Result,
			},
			BizNo: applyBizNo(app.ID),
			IP:    ip,
			UA:    ua,
		})
	}

	// 单条预加载 plan + association
	planMap, assocMap := s.loadPlanAndAssocMaps([]models.StRecruitApply{*app})
	v := s.toApplyView(*app, planMap, assocMap)
	return &v, nil
}

// ---- 内部方法 ----

// GetStudentIDByUserID 根据登录用户 ID 获取其关联的 student_id（学生身份）。
// 学生在投递招新申请 / 查看自己的申请时使用。
func (s *RecruitService) GetStudentIDByUserID(userID int64) (int64, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return 0, err
	}
	if user.StudentID == nil {
		return 0, nil
	}
	return *user.StudentID, nil
}

func (s *RecruitService) toPlanView(p models.StRecruitPlan) RecruitPlanView {
	v := RecruitPlanView{
		ID:               p.ID,
		BizNo:            p.BizNo,
		AssociationID:    p.AssociationID,
		Season:           p.Season,
		SeasonText:       planSeasonTextMap[p.Season],
		AcademicYear:     p.AcademicYear,
		TargetCount:      p.TargetCount,
		PlanFileID:       p.PlanFileID,
		AssessmentMethod: p.AssessmentMethod,
		Status:           p.Status,
		StatusText:       planStatusTextMap[p.Status],
		IsFinished:       p.IsFinished,
		FinishedBy:       p.FinishedBy,
		FinishedReason:   p.FinishedReason,
		RecruitPhase:     calcRecruitPhase(p.Status, p.IsFinished),
		RecruitPhaseText: planPhaseTextMap[calcRecruitPhase(p.Status, p.IsFinished)],
		CreatedAt:        p.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:        p.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if p.InterviewAt != nil {
		v.InterviewAt = p.InterviewAt.Format("2006-01-02T15:04:05+08:00")
	}
	if p.ResultDeadline != nil {
		v.ResultDeadline = p.ResultDeadline.Format("2006-01-02")
	}
	if p.FinishedAt != nil {
		v.FinishedAt = p.FinishedAt.Format("2006-01-02T15:04:05+08:00")
	}
	// 社团名
	if assoc, err := s.repo.GetAssociationByID(p.AssociationID); err == nil {
		v.AssociationName = assoc.Name
	}
	// 申请统计（仅已通过/已发布后才有意义）
	if p.Status == stsm.PlanStatePassed {
		apps, total, _ := s.repo.ListApplies(p.ID, 0, "", 1, 1000)
		v.ApplyCount = total
		for _, a := range apps {
			if a.Result == "accepted" {
				v.AcceptedCount++
			}
		}
	}
	return v
}

func (s *RecruitService) toApplyView(a models.StRecruitApply, planMap map[int64]*models.StRecruitPlan, assocMap map[int64]*models.StAssociation) RecruitApplyView {
	v := RecruitApplyView{
		ID:           a.ID,
		PlanID:       a.PlanID,
		StudentID:    a.StudentID,
		ResumeFileID: a.ResumeFileID,
		Result:       a.Result,
		ResultText:   applyResultTextMap[a.Result],
		CreatedAt:    a.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if a.ResultAt != nil {
		v.ResultAt = a.ResultAt.Format("2006-01-02T15:04:05+08:00")
	}
	if student, err := s.repo.GetStudentByID(a.StudentID); err == nil {
		v.StudentNo = student.StudentNo
		v.StudentName = student.Name
	}
	// 关联招新计划 + 社团信息
	if plan, ok := planMap[a.PlanID]; ok && plan != nil {
		v.PlanBizNo = plan.BizNo
		v.PlanSeason = plan.Season
		v.PlanAcademicYear = plan.AcademicYear
		v.AssociationID = plan.AssociationID
		if assoc, ok := assocMap[plan.AssociationID]; ok && assoc != nil {
			v.AssociationName = assoc.Name
		}
	}
	return v
}

// publishPlanEvent 发布招新计划相关事件。
func (s *RecruitService) publishPlanEvent(plan *models.StRecruitPlan, evtType string, actorID int64, actorRole, ip, ua string, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["plan_id"] = plan.ID
	payload["biz_no"] = plan.BizNo
	payload["status"] = plan.Status

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "st.recruit_plan",
		AggregateID: plan.BizNo,
		EventType:   evtType,
		Module:      "ST",
		ActorID:     actorID,
		ActorRole:   actorRole,
		Payload:     payload,
		BizNo:       plan.BizNo,
		IP:          ip,
		UA:          ua,
	})
}

// BizNoFallback 给招新申请合成一个稳定事件聚合 ID（无 biz_no 字段）。
// 实际定义在文件末尾，避免循环引用。

// applyBizNo 给招新申请合成一个稳定事件聚合 ID。
func applyBizNo(id int64) string {
	return fmt.Sprintf("APPLY-%d", id)
}

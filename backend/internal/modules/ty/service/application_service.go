package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/ty/repository"
	tysm "student-system/internal/modules/ty/statemachine"
	"student-system/internal/statem"
	"student-system/pkg/cryptox"
)

// ApplicationService 入团申请业务服务层。
type ApplicationService struct {
	repo *repository.ApplicationRepository
	db   *gorm.DB
	sm   *statem.Engine
	bus  *eventx.Bus
}

// NewApplicationService 创建入团申请服务。
func NewApplicationService(repo *repository.ApplicationRepository, db *gorm.DB, bus *eventx.Bus) *ApplicationService {
	return &ApplicationService{
		repo: repo,
		db:   db,
		sm:   tysm.NewApplicationSM(),
		bus:  bus,
	}
}

// ---- DTO 定义 ----

// ApplicationListResult 入团申请列表结果。
type ApplicationListResult struct {
	Items    []ApplicationView `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// ApplicationView 入团申请视图。
type ApplicationView struct {
	ID               int64      `json:"id"`
	BizNo            string     `json:"biz_no"`
	StudentID        int64      `json:"student_id"`
	StudentName      string     `json:"student_name"`
	StudentNo        string     `json:"student_no"`
	BranchID         int64      `json:"branch_id"`
	BranchName       string     `json:"branch_name"`
	CollegeID        int64      `json:"college_id"`
	CollegeName      string     `json:"college_name"`
	ApplyDate        string     `json:"apply_date"`
	SelfStatement    string     `json:"self_statement"`
	FamilyMembers    string     `json:"family_members_json"`
	RewardsPunish    string     `json:"rewards_punishments"`
	Status           string     `json:"status"`
	StatusText       string     `json:"status_text"`
	CounselorOpinion string     `json:"counselor_opinion"`
	CounselorUserID  *int64     `json:"counselor_user_id,omitempty"`
	CounselorAt      *string    `json:"counselor_at,omitempty"`
	CollegeOpinion   string     `json:"college_opinion"`
	CollegeUserID    *int64     `json:"college_user_id,omitempty"`
	CollegeAt        *string    `json:"college_at,omitempty"`
	SchoolOpinion    string     `json:"school_opinion"`
	SchoolUserID     *int64     `json:"school_user_id,omitempty"`
	SchoolAt         *string    `json:"school_at,omitempty"`
	RejectReason     string     `json:"reject_reason"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
}

// CreateApplicationRequest 创建入团申请请求。
type CreateApplicationRequest struct {
	BranchID      int64  `json:"branch_id" binding:"required"`
	ApplyDate     string `json:"apply_date" binding:"required"`
	SelfStatement string `json:"self_statement" binding:"required"`
	FamilyMembers string `json:"family_members_json"`
	RewardsPunish string `json:"rewards_punishments"`
}

// UpdateApplicationRequest 更新入团申请请求。
type UpdateApplicationRequest struct {
	BranchID      *int64  `json:"branch_id"`
	ApplyDate     *string `json:"apply_date"`
	SelfStatement *string `json:"self_statement"`
	FamilyMembers *string `json:"family_members_json"`
	RewardsPunish *string `json:"rewards_punishments"`
}

// WithdrawRequest 撤回请求。
type WithdrawRequest struct {
	Reason string `json:"reason"`
}

// ---- 状态映射 ----

var statusTextMap = map[string]string{
	"S0": "草稿",
	"S1": "待审",
	"S2": "审批中",
	"S3": "通过",
	"S4": "驳回",
}

// ---- 业务方法 ----

// List 分页查询入团申请列表。
// 辅导员(R-COL-COUN)只能看到自己负责专业下的学生申请。
func (s *ApplicationService) List(userID int64, status string, studentID, collegeID int64, page, pageSize int) (*ApplicationListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 辅导员专业级过滤：仅展示本专业学生的申请
	var majorIDs []int64
	roles, _ := s.repo.FindUserRoles(userID)
	if hasAny(roles, "R-COL-COUN") && !hasAny(roles, "R-SY-ADMIN", "R-SY-LEAGUE", "R-COL-LEAGUE") {
		majorIDs, _ = s.repo.FindCounselorMajorIDs(userID)
	}

	apps, total, err := s.repo.List(status, studentID, collegeID, majorIDs, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 预加载关联名称
	colleges, _ := s.repo.ListColleges()
	collegeMap := make(map[int64]string)
	for _, c := range colleges {
		collegeMap[c.ID] = c.Name
	}

	items := make([]ApplicationView, 0, len(apps))
	for _, app := range apps {
		v := s.toView(app, collegeMap)
		items = append(items, v)
	}

	return &ApplicationListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取入团申请详情。
func (s *ApplicationService) Get(id int64) (*ApplicationView, error) {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	colleges, _ := s.repo.ListColleges()
	collegeMap := make(map[int64]string)
	for _, c := range colleges {
		collegeMap[c.ID] = c.Name
	}

	v := s.toView(*app, collegeMap)
	return &v, nil
}

// Create 创建入团申请（保存为 S0 草稿）。
func (s *ApplicationService) Create(userID int64, req *CreateApplicationRequest) (*ApplicationView, error) {
	// 获取关联学生
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if user.StudentID == nil {
		return nil, fmt.Errorf("当前用户未关联学生身份")
	}
	studentID := *user.StudentID

	// 校验年龄 14-28 周岁
	student, err := s.repo.GetStudentByID(studentID)
	if err != nil {
		return nil, fmt.Errorf("学生信息不存在")
	}
	if err := s.validateAge(student); err != nil {
		return nil, err
	}

	// 校验自述字数 ≥ 500（与 SQLite length() 保持一致：按字符数而非字节数）
	if utf8.RuneCountInString(req.SelfStatement) < 500 {
		return nil, fmt.Errorf("思想政治表现自述字数须 ≥ 500")
	}

	// 校验所选团支部必须属于学生所在学院（不允许跨院提交）
	branch, err := s.repo.GetBranchByID(req.BranchID)
	if err != nil {
		return nil, fmt.Errorf("所选团支部不存在")
	}
	if student.CollegeID == nil {
		return nil, fmt.Errorf("学生未关联院系，无法提交申请")
	}
	if branch.CollegeID != *student.CollegeID {
		return nil, fmt.Errorf("所选团支部不属于学生所在院系，请重新选择本班团支部")
	}

	// 校验政治面貌：已是团员 / 预备党员 / 党员的不可发起入团申请
	// 字典 sys_dict.category=political_status 的 code: masses(群众) activist(入团积极分子) probationary(预备团员) member(团员) party_*(党员)
	if rejectMsg := s.checkPoliticalStatus(student.PoliticalStatus); rejectMsg != "" {
		return nil, fmt.Errorf("%s", rejectMsg)
	}

	// 全周期 1 单限制：每名学生终身只能有一条入团申请
	if has, existing, _ := s.repo.HasActiveApplication(studentID, 0); has {
		stateText := statusTextMap[existing.Status]
		return nil, fmt.Errorf("每名学生只能提交一次入团申请，您已存在申请记录（%s，编号 %s）", stateText, existing.BizNo)
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "TY")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// 解析申请日期
	applyDate, err := time.Parse("2006-01-02", req.ApplyDate)
	if err != nil {
		return nil, fmt.Errorf("申请日期格式错误")
	}

	app := models.TyApplication{
		BizNo:         bizNo,
		StudentID:     studentID,
		BranchID:      req.BranchID,
		ApplyDate:     applyDate,
		SelfStatement: req.SelfStatement,
		FamilyMembers: req.FamilyMembers,
		RewardsPunish: req.RewardsPunish,
		Status:        "S0",
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
	}

	if err := s.repo.Create(&app); err != nil {
		return nil, err
	}

	return s.Get(app.ID)
}

// Update 更新入团申请（仅 S0 状态可改）。
func (s *ApplicationService) Update(id, userID int64, req *UpdateApplicationRequest) (*ApplicationView, error) {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("申请不存在")
	}

	if app.Status != "S0" {
		return nil, fmt.Errorf("仅草稿状态可修改")
	}

	// 权限校验：只有创建者本人可修改
	if app.CreatedBy != nil && *app.CreatedBy != userID {
		return nil, fmt.Errorf("无权修改他人申请")
	}

	if req.BranchID != nil {
		app.BranchID = *req.BranchID
	}
	if req.ApplyDate != nil {
		applyDate, err := time.Parse("2006-01-02", *req.ApplyDate)
		if err != nil {
			return nil, fmt.Errorf("申请日期格式错误")
		}
		app.ApplyDate = applyDate
	}
	if req.SelfStatement != nil {
		if utf8.RuneCountInString(*req.SelfStatement) < 500 {
			return nil, fmt.Errorf("思想政治表现自述字数须 ≥ 500")
		}
		app.SelfStatement = *req.SelfStatement
	}
	if req.FamilyMembers != nil {
		app.FamilyMembers = *req.FamilyMembers
	}
	if req.RewardsPunish != nil {
		app.RewardsPunish = *req.RewardsPunish
	}
	app.UpdatedBy = &userID

	if err := s.repo.Update(app); err != nil {
		return nil, err
	}

	return s.Get(app.ID)
}

// Submit 提交入团申请（S0 → S1）。
func (s *ApplicationService) Submit(id, userID int64, actorName, actorRole, ip, ua string) (*ApplicationView, error) {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("申请不存在")
	}

	if app.Status != tysm.StateDraft {
		return nil, fmt.Errorf("仅草稿状态可提交")
	}

	// 权限校验
	if app.CreatedBy != nil && *app.CreatedBy != userID {
		return nil, fmt.Errorf("无权提交他人申请")
	}

	// 校验自述字数
	if utf8.RuneCountInString(app.SelfStatement) < 500 {
		return nil, fmt.Errorf("思想政治表现自述字数须 ≥ 500")
	}

	// 校验年龄
	student, err := s.repo.GetStudentByID(app.StudentID)
	if err != nil {
		return nil, fmt.Errorf("学生信息不存在")
	}
	if err := s.validateAge(student); err != nil {
		return nil, err
	}

	// 校验政治面貌：已是团员/党员的不可提交入团申请
	if rejectMsg := s.checkPoliticalStatus(student.PoliticalStatus); rejectMsg != "" {
		return nil, fmt.Errorf("%s", rejectMsg)
	}

	// 校验同期无 S1/S2 申请
	hasPending, err := s.repo.HasPending(app.StudentID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, fmt.Errorf("已存在审批中申请，请勿重复提交")
	}

	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, app.Status, tysm.ActionSubmit)
	if err != nil {
		return nil, err
	}

	app.Status = to
	app.UpdatedBy = &userID
	if err := s.repo.Update(app); err != nil {
		return nil, err
	}

	s.publishEvent(app, "TyApplicationSubmitted", userID, actorRole, ip, ua, map[string]interface{}{
		"from": tysm.StateDraft,
		"to":   to,
	})

	return s.Get(app.ID)
}

// Withdraw 撤回入团申请（S1 → S0）。
func (s *ApplicationService) Withdraw(id, userID int64, reason, actorName, actorRole, ip, ua string) (*ApplicationView, error) {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("申请不存在")
	}

	if app.Status != tysm.StatePending {
		return nil, fmt.Errorf("仅待审状态可撤回")
	}

	// 权限校验
	if app.CreatedBy != nil && *app.CreatedBy != userID {
		return nil, fmt.Errorf("无权撤回他人申请")
	}

	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, app.Status, tysm.ActionWithdraw)
	if err != nil {
		return nil, err
	}

	app.Status = to
	app.UpdatedBy = &userID
	if err := s.repo.Update(app); err != nil {
		return nil, err
	}

	s.publishEvent(app, "TyApplicationWithdrawn", userID, actorRole, ip, ua, map[string]interface{}{
		"from":   tysm.StatePending,
		"to":     to,
		"reason": reason,
	})

	return s.Get(app.ID)
}

// SoftDelete 软删除入团申请（仅 S0/S4）。
func (s *ApplicationService) SoftDelete(id, userID int64) error {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("申请不存在")
	}

	if app.Status != "S0" && app.Status != "S4" {
		return fmt.Errorf("仅草稿或驳回状态可删除")
	}

	// 权限校验
	if app.CreatedBy != nil && *app.CreatedBy != userID {
		return fmt.Errorf("无权删除他人申请")
	}

	return s.repo.SoftDelete(id)
}

// ---- 三级审批流（S06）----

// ApproveRequest 审批/驳回请求 DTO。
type ApproveRequest struct {
	Step    string `json:"step" binding:"required"`    // counselor / college / school
	Result  string `json:"result" binding:"required"`  // approve / reject
	Opinion string `json:"opinion" binding:"required"` // 审批意见 ≥5 字
}

// ApprovalRecordView 审批记录视图。
type ApprovalRecordView struct {
	ID            int64  `json:"id"`
	ApplicationID int64  `json:"application_id"`
	Step          string `json:"step"`
	StepText      string `json:"step_text"`
	ApproverID    int64  `json:"approver_id"`
	ApproverName  string `json:"approver_name"`
	ApproverRole  string `json:"approver_role"`
	Result        string `json:"result"`
	ResultText    string `json:"result_text"`
	Opinion       string `json:"opinion"`
	FromStatus    string `json:"from_status"`
	ToStatus      string `json:"to_status"`
	OccurredAt    string `json:"occurred_at"`
}

// TimelineEntry 时间线条目（事件流投影）。
type TimelineEntry struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type"`
	ActorID    int64                  `json:"actor_id"`
	ActorRole  string                 `json:"actor_role"`
	OccurredAt string                 `json:"occurred_at"`
	Payload    map[string]interface{} `json:"payload"`
}

var stepTextMap = map[string]string{
	tysm.StepCounselor: "辅导员/团支部初审",
	tysm.StepCollege:   "院系团委复核",
	tysm.StepSchool:    "校团委终审",
}

var resultTextMap = map[string]string{
	"approve": "通过",
	"reject":  "驳回",
}

// Approve 审批入团申请（含通过/驳回）。
//
// 三级审批链：
//   - counselor：S1 → S2（辅导员初审通过 / 驳回 → S4）
//   - college：S2 (counselor 已通过) → S2（院系复核通过 / 驳回 → S4）
//   - school：S2 (college 已通过) → S3（校级终审通过 / 驳回 → S4）
//
// 校验：
//   - 角色权限（counselor: R-COL-COUN/R-COL-LEAGUE；college: R-COL-LEAGUE；school: R-SY-LEAGUE/R-SY-ADMIN）。
//   - 院系隔离：辅导员/院系角色仅能审批本 college 的申请。
//   - 步骤前置：college 步骤要求 counselor 已通过；school 步骤要求 college 已通过。
//   - 同一 step 仅允许 1 条 approve 记录。
func (s *ApplicationService) Approve(
	id, userID int64,
	req *ApproveRequest,
	actorName, actorRole, ip, ua string,
) (*ApplicationView, error) {
	if req == nil {
		return nil, fmt.Errorf("参数不能为空")
	}
	if req.Step != tysm.StepCounselor && req.Step != tysm.StepCollege && req.Step != tysm.StepSchool {
		return nil, fmt.Errorf("无效的审批步骤")
	}
	if req.Result != "approve" && req.Result != "reject" {
		return nil, fmt.Errorf("无效的审批结果")
	}
	if len([]rune(req.Opinion)) < 5 {
		return nil, fmt.Errorf("审批意见至少 5 字")
	}

	app, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("申请不存在")
	}

	// 角色 + 院系权限校验
	roles, err := s.repo.FindUserRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败")
	}
	if err := s.checkApprovalAuth(req.Step, roles); err != nil {
		return nil, err
	}
	if err := s.checkCollegeScope(userID, app, req.Step, roles); err != nil {
		return nil, err
	}

	// 步骤前置：审批顺序不可跳跃
	if req.Step == tysm.StepCollege {
		ok, err := s.repo.HasApprovedStep(id, tysm.StepCounselor)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("辅导员初审尚未通过，院系不可复核")
		}
	}
	if req.Step == tysm.StepSchool {
		ok, err := s.repo.HasApprovedStep(id, tysm.StepCollege)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("院系复核尚未通过，校级不可终审")
		}
	}

	// 同一 step 不允许重复通过
	if req.Result == "approve" {
		ok, err := s.repo.HasApprovedStep(id, req.Step)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, fmt.Errorf("该步骤已通过，请勿重复审批")
		}
	}

	action := tysm.ResolveAction(req.Step, req.Result)
	if action == "" {
		return nil, fmt.Errorf("无法解析审批动作")
	}
	from := app.Status
	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
		Payload: map[string]interface{}{
			"step":    req.Step,
			"result":  req.Result,
			"opinion": req.Opinion,
		},
	}, from, action)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// 事务：写入审批记录 + 更新申请单
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		rec := &models.TyApprovalRecord{
			ApplicationID: app.ID,
			Step:          req.Step,
			ApproverID:    userID,
			ApproverName:  actorName,
			ApproverRole:  primaryRole(roles),
			Result:        req.Result,
			Opinion:       req.Opinion,
			FromStatus:    from,
			ToStatus:      to,
			OccurredAt:    now,
			IP:            ip,
		}
		if err := tx.Create(rec).Error; err != nil {
			return fmt.Errorf("写入审批记录失败: %w", err)
		}

		// 同步更新申请单冗余字段（与 docs/03 §5.2.2 字段对齐）
		switch req.Step {
		case tysm.StepCounselor:
			app.CounselorOpinion = req.Opinion
			app.CounselorUserID = &userID
			app.CounselorAt = &now
		case tysm.StepCollege:
			app.CollegeOpinion = req.Opinion
			app.CollegeUserID = &userID
			app.CollegeAt = &now
		case tysm.StepSchool:
			app.SchoolOpinion = req.Opinion
			app.SchoolUserID = &userID
			app.SchoolAt = &now
		}
		if req.Result == "reject" {
			app.RejectReason = req.Opinion
		}
		app.Status = to
		app.UpdatedBy = &userID
		return tx.Save(app).Error
	}); err != nil {
		return nil, err
	}

	// 终审通过（S3）→ 自动更新学生政治面貌为"member（共青团员）"
	if to == tysm.StatePassed {
		if upErr := s.repo.UpdateStudentPoliticalStatus(app.StudentID, "member"); upErr != nil {
			// 非致命错误：不影响审批结果，仅记录
			fmt.Printf("[TY-Approve] 更新学生政治面貌失败: studentID=%d err=%v", app.StudentID, upErr)
		}
	}

	// 发布事件（事务外）
	eventType := "TyApplicationApproved"
	if req.Result == "reject" {
		eventType = "TyApplicationRejected"
	}
	s.publishEvent(app, eventType, userID, actorRole, ip, ua, map[string]interface{}{
		"step":     req.Step,
		"result":   req.Result,
		"opinion":  req.Opinion,
		"from":     from,
		"to":       to,
		"approver": actorName,
	})

	return s.Get(app.ID)
}

// ListApprovals 列出某申请单的审批记录。
func (s *ApplicationService) ListApprovals(applicationID int64) ([]ApprovalRecordView, error) {
	records, err := s.repo.ListApprovalRecords(applicationID)
	if err != nil {
		return nil, err
	}
	views := make([]ApprovalRecordView, 0, len(records))
	for _, r := range records {
		views = append(views, ApprovalRecordView{
			ID:            r.ID,
			ApplicationID: r.ApplicationID,
			Step:          r.Step,
			StepText:      stepTextMap[r.Step],
			ApproverID:    r.ApproverID,
			ApproverName:  r.ApproverName,
			ApproverRole:  r.ApproverRole,
			Result:        r.Result,
			ResultText:    resultTextMap[r.Result],
			Opinion:       r.Opinion,
			FromStatus:    r.FromStatus,
			ToStatus:      r.ToStatus,
			OccurredAt:    r.OccurredAt.Format("2006-01-02T15:04:05+08:00"),
		})
	}
	return views, nil
}

// Timeline 返回申请单的事件流时间线（来自 event_log 投影）。
func (s *ApplicationService) Timeline(applicationID int64) ([]TimelineEntry, error) {
	app, err := s.repo.GetByID(applicationID)
	if err != nil {
		return nil, fmt.Errorf("申请不存在")
	}
	logs, err := s.bus.QueryByAggregate("ty.application", app.BizNo)
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

// ListPending 列出当前用户应处理的待办审批。
//
// 规则：
//   - R-COL-COUN/R-COL-LEAGUE：S1 状态且院系匹配
//   - R-COL-LEAGUE：S2 状态且 counselor 已通过、本院系
//   - R-SY-LEAGUE/R-SY-ADMIN：S2 状态且 college 已通过
func (s *ApplicationService) ListPending(userID int64, page, pageSize int) (*ApplicationListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	roles, err := s.repo.FindUserRoles(userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return &ApplicationListResult{Items: []ApplicationView{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}

	colleges, _ := s.repo.ListColleges()
	collegeMap := make(map[int64]string)
	for _, c := range colleges {
		collegeMap[c.ID] = c.Name
	}

	// 取该用户的全部相关申请，再过滤"是否轮到我审"
	statusList := []string{}
	if hasAny(roles, "R-COL-COUN", "R-COL-LEAGUE") {
		statusList = append(statusList, tysm.StatePending, tysm.StateInReview)
	}
	if hasAny(roles, "R-SY-LEAGUE", "R-SY-ADMIN") {
		// admin 视角看全部 S2
		statusList = appendUnique(statusList, tysm.StateInReview)
	}
	if len(statusList) == 0 {
		return &ApplicationListResult{Items: []ApplicationView{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}

	var collegeIDs []int64
	if !hasAny(roles, "R-SY-LEAGUE", "R-SY-ADMIN") {
		// 院系级用户：限定本人 scope_college_id
		collegeIDs, _ = s.repo.FindUserScopeCollegeIDs(userID)
	}

	apps, err := s.fetchPendingApps(statusList, collegeIDs)
	if err != nil {
		return nil, err
	}

	// 过滤：是否轮到该用户审批 + 辅导员专业级隔离
	var counselorMajorIDs []int64
	isCounselor := hasAny(roles, "R-COL-COUN")
	if isCounselor {
		counselorMajorIDs, _ = s.repo.FindCounselorMajorIDs(userID)
	}

	filtered := make([]models.TyApplication, 0, len(apps))
	for _, app := range apps {
		nextStep := s.nextStepFor(app.ID, app.Status)
		if nextStep == "" {
			continue
		}
		if !s.canApproveStep(nextStep, roles) {
			continue
		}
		// 辅导员步骤：仅展示本专业学生的申请
		if nextStep == tysm.StepCounselor && isCounselor && len(counselorMajorIDs) > 0 {
			student, err := s.repo.GetStudentByID(app.StudentID)
			if err != nil || student.MajorID == nil {
				continue
			}
			if !containsInt64(counselorMajorIDs, *student.MajorID) {
				continue
			}
		}
		filtered = append(filtered, app)
	}

	total := int64(len(filtered))
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	pageItems := filtered[start:end]

	views := make([]ApplicationView, 0, len(pageItems))
	for _, app := range pageItems {
		v := s.toView(app, collegeMap)
		// 附加：下一步骤（前端展示用）
		next := s.nextStepFor(app.ID, app.Status)
		if next != "" {
			v.StatusText = stepTextMap[next] + "·待审"
		}
		views = append(views, v)
	}

	return &ApplicationListResult{
		Items:    views,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ---- 内部审批辅助 ----

// checkApprovalAuth 校验角色是否允许执行某 step 的审批。
func (s *ApplicationService) checkApprovalAuth(step string, roles []string) error {
	switch step {
	case tysm.StepCounselor:
		if !hasAny(roles, "R-COL-COUN", "R-COL-LEAGUE", "R-SY-ADMIN") {
			return fmt.Errorf("无该步骤审批权限")
		}
	case tysm.StepCollege:
		if !hasAny(roles, "R-COL-LEAGUE", "R-SY-ADMIN") {
			return fmt.Errorf("无该步骤审批权限")
		}
	case tysm.StepSchool:
		if !hasAny(roles, "R-SY-LEAGUE", "R-SY-ADMIN") {
			return fmt.Errorf("无该步骤审批权限")
		}
	}
	return nil
}

// checkCollegeScope 校验院系级用户仅能审批本院系申请；
// 辅导员(counselor)步骤额外校验专业级隔离：只能审批自己负责专业下的学生。
func (s *ApplicationService) checkCollegeScope(userID int64, app *models.TyApplication, step string, roles []string) error {
	// admin / 校级用户 不受院系隔离限制
	if hasAny(roles, "R-SY-ADMIN", "R-SY-LEAGUE") {
		return nil
	}
	if step != tysm.StepCounselor && step != tysm.StepCollege {
		return nil
	}

	branch, err := s.repo.GetBranchByID(app.BranchID)
	if err != nil {
		return fmt.Errorf("申请所属团支部异常")
	}

	scopeIDs, err := s.repo.FindUserScopeCollegeIDs(userID)
	if err != nil {
		return err
	}
	if len(scopeIDs) == 0 {
		// 未配置 scope_college_id 的院系级用户：默认放行（兼容种子数据未设置情况）
		return nil
	}
	for _, cid := range scopeIDs {
		if cid == branch.CollegeID {
			// 院系匹配，继续校验专业级隔离
			goto collegeMatched
		}
	}
	return fmt.Errorf("仅可审批本院系申请")

collegeMatched:
	// 辅导员步骤：校验专业级隔离
	if step == tysm.StepCounselor && hasAny(roles, "R-COL-COUN") {
		// 查询申请学生的专业
		student, err := s.repo.GetStudentByID(app.StudentID)
		if err != nil {
			return fmt.Errorf("申请人学生信息异常")
		}
		if student.MajorID == nil {
			// 学生未关联专业，放行（兼容数据不完整场景）
			return nil
		}

		// 查询辅导员负责的专业列表
		majorIDs, err := s.repo.FindCounselorMajorIDs(userID)
		if err != nil {
			return fmt.Errorf("查询辅导员负责专业失败")
		}
		if len(majorIDs) == 0 {
			// 辅导员未分配班级，放行（兼容未配置场景）
			return nil
		}
		for _, mid := range majorIDs {
			if mid == *student.MajorID {
				return nil
			}
		}
		return fmt.Errorf("仅可审批本专业学生的申请")
	}

	return nil
}

// nextStepFor 计算下一应处理的 step；若无返回空串。
func (s *ApplicationService) nextStepFor(applicationID int64, status string) string {
	switch status {
	case tysm.StatePending:
		return tysm.StepCounselor
	case tysm.StateInReview:
		// 检查 college 是否已通过
		ok, _ := s.repo.HasApprovedStep(applicationID, tysm.StepCollege)
		if !ok {
			return tysm.StepCollege
		}
		return tysm.StepSchool
	}
	return ""
}

// canApproveStep 判断角色是否可审批指定 step。
func (s *ApplicationService) canApproveStep(step string, roles []string) bool {
	switch step {
	case tysm.StepCounselor:
		return hasAny(roles, "R-COL-COUN", "R-COL-LEAGUE", "R-SY-ADMIN")
	case tysm.StepCollege:
		return hasAny(roles, "R-COL-LEAGUE", "R-SY-ADMIN")
	case tysm.StepSchool:
		return hasAny(roles, "R-SY-LEAGUE", "R-SY-ADMIN")
	}
	return false
}

// fetchPendingApps 拉取候选申请。
func (s *ApplicationService) fetchPendingApps(statusList []string, collegeIDs []int64) ([]models.TyApplication, error) {
	q := s.db.Model(&models.TyApplication{}).Where("is_deleted = 0").Where("status IN ?", statusList)
	if len(collegeIDs) > 0 {
		q = q.Where("branch_id IN (SELECT id FROM ty_branch WHERE college_id IN ? AND is_deleted = 0)", collegeIDs)
	}
	var apps []models.TyApplication
	if err := q.Order("id DESC").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

// publishEvent 发布业务事件到 event_log。
func (s *ApplicationService) publishEvent(app *models.TyApplication, evtType string, actorID int64, actorRole, ip, ua string, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["application_id"] = app.ID
	payload["biz_no"] = app.BizNo
	payload["status"] = app.Status

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "ty.application",
		AggregateID: app.BizNo,
		EventType:   evtType,
		Module:      "TY",
		ActorID:     actorID,
		ActorRole:   actorRole,
		Payload:     payload,
		BizNo:       app.BizNo,
		IP:          ip,
		UA:          ua,
	})
}

// hasAny 判断角色列表中是否包含任一目标角色。
func hasAny(roles []string, targets ...string) bool {
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

// containsInt64 判断 int64 切片中是否包含目标值。
func containsInt64(list []int64, target int64) bool {
	for _, v := range list {
		if v == target {
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

func appendUnique(list []string, v string) []string {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

// ---- 内部方法 ----

// BranchView 团支部视图。
type BranchView struct {
	ID        int64  `json:"id"`
	BizNo     string `json:"biz_no"`
	Name      string `json:"name"`
	CollegeID int64  `json:"college_id"`
}

// ListBranches 查询团支部列表（下拉选择用）。
func (s *ApplicationService) ListBranches(collegeID int64) ([]BranchView, error) {
	branches, err := s.repo.ListBranchesByCollege(collegeID)
	if err != nil {
		return nil, err
	}

	views := make([]BranchView, 0, len(branches))
	for _, b := range branches {
		views = append(views, BranchView{
			ID:        b.ID,
			BizNo:     b.BizNo,
			Name:      b.Name,
			CollegeID: b.CollegeID,
		})
	}
	return views, nil
}

// validateAge 校验申请人年龄 14-28 周岁。
func (s *ApplicationService) validateAge(student *models.IdxStudent) error {
	if student.IDCardEnc == "" {
		// 无身份证信息，跳过年龄校验（管理员创建场景）
		return nil
	}

	idCard, err := cryptox.Decrypt(student.IDCardEnc)
	if err != nil {
		return nil // 解密失败，跳过
	}

	if len(idCard) < 14 {
		return nil
	}

	// 从身份证提取出生日期
	birthStr := idCard[6:10] + "-" + idCard[10:12] + "-" + idCard[12:14]
	birthDate, err := time.Parse("2006-01-02", birthStr)
	if err != nil {
		return nil
	}

	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	if age < 14 || age > 28 {
		return fmt.Errorf("申请人年龄超出 14-28 周岁范围")
	}

	return nil
}

// toView 将模型转为视图。
func (s *ApplicationService) toView(app models.TyApplication, collegeMap map[int64]string) ApplicationView {
	v := ApplicationView{
		ID:               app.ID,
		BizNo:            app.BizNo,
		StudentID:        app.StudentID,
		BranchID:         app.BranchID,
		ApplyDate:        app.ApplyDate.Format("2006-01-02"),
		SelfStatement:    app.SelfStatement,
		FamilyMembers:    app.FamilyMembers,
		RewardsPunish:    app.RewardsPunish,
		Status:           app.Status,
		StatusText:       statusTextMap[app.Status],
		CounselorOpinion: app.CounselorOpinion,
		CounselorUserID:  app.CounselorUserID,
		CollegeOpinion:   app.CollegeOpinion,
		CollegeUserID:    app.CollegeUserID,
		SchoolOpinion:    app.SchoolOpinion,
		SchoolUserID:     app.SchoolUserID,
		RejectReason:     app.RejectReason,
		CreatedAt:        app.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:        app.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	if app.CounselorAt != nil {
		t := app.CounselorAt.Format("2006-01-02T15:04:05+08:00")
		v.CounselorAt = &t
	}
	if app.CollegeAt != nil {
		t := app.CollegeAt.Format("2006-01-02T15:04:05+08:00")
		v.CollegeAt = &t
	}
	if app.SchoolAt != nil {
		t := app.SchoolAt.Format("2006-01-02T15:04:05+08:00")
		v.SchoolAt = &t
	}

	// 加载学生姓名
	if student, err := s.repo.GetStudentByID(app.StudentID); err == nil {
		v.StudentName = student.Name
		v.StudentNo = student.StudentNo
	}

	// 加载团支部名称 + 院系：以 branch_id 为唯一事实来源，避免跨院错配
	if branch, err := s.repo.GetBranchByID(app.BranchID); err == nil {
		v.BranchName = branch.Name
		v.CollegeID = branch.CollegeID
		v.CollegeName = collegeMap[branch.CollegeID]
	}

	return v
}

// checkPoliticalStatus 校验政治面貌是否允许发起入团申请。
// 政治面貌字典 sys_dict.category=political_status 的 code:
//   - masses / 群众             （允许入团）
//   - activist / 入团积极分子   （允许入团）
//   - probationary / 预备团员   （不允许）
//   - member / 共青团员         （不允许）
//   - party_probationary / 预备党员（不允许）
//   - party_member / 中共党员   （不允许）
//
// 返回空字符串表示允许；非空字符串为拒绝原因。
func (s *ApplicationService) checkPoliticalStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "masses", "群众":
		return ""
	case "activist", "入团积极分子":
		return ""
	case "probationary", "预备团员":
		return "您已是预备团员，无需再次发起入团申请"
	case "member", "共青团员":
		return "您已是共青团员，无需发起入团申请"
	case "party_probationary", "预备党员":
		return "您已是预备党员，无需发起入团申请"
	case "party_member", "中共党员":
		return "您已是中共党员，无需发起入团申请"
	default:
		// 未知值、空值、字典码不匹配时，按"群众"放行（保守策略：不影响新生）
		return ""
	}
}

// ---- 发展轨迹 ----

// DevelopmentTrackEntry 发展轨迹条目。
type DevelopmentTrackEntry struct {
	Module      string `json:"module"`       // application / recommendation / cultivation / development_object / political_review / development_meeting / probationary
	ModuleText  string `json:"module_text"`  // 模块中文名
	TargetID    int64  `json:"target_id"`     // 对应模块主记录 ID
	BizNo       string `json:"biz_no"`       // 业务编号
	Status      string `json:"status"`       // 当前状态
	StatusText  string `json:"status_text"`  // 状态中文名
	OccurredAt  string `json:"occurred_at"`  // 发生时间
	Step        string `json:"step"`         // 审批步骤（仅审批记录有值）
	StepText    string `json:"step_text"`    // 步骤中文名
	Result      string `json:"result"`       // 审批结果（approve/reject/空）
	ResultText  string `json:"result_text"`  // 结果中文名
	ApproverName string `json:"approver_name"` // 审批人
	Opinion     string `json:"opinion"`       // 审批意见
	FromStatus  string `json:"from_status"`   // 状态流转-源
	ToStatus    string `json:"to_status"`     // 状态流转-目标
}

// DevelopmentTrackResult 发展轨迹结果。
type DevelopmentTrackResult struct {
	StudentID   int64                  `json:"student_id"`
	StudentName string                 `json:"student_name"`
	PoliticalStatus string             `json:"political_status"`
	PoliticalStatusText string         `json:"political_status_text"`
	Entries      []DevelopmentTrackEntry `json:"entries"`
}

var moduleTextMap = map[string]string{
	"application":        "入团申请",
	"recommendation":     "推优大会",
	"cultivation":        "培养考察",
	"development_object": "列为发展对象",
	"political_review":   "政审",
	"development_meeting":"发展大会",
	"probationary":       "转正",
}

var stepTextMapFull = map[string]string{
	"counselor": "辅导员/团支部初审",
	"college":   "院系团委复核",
	"school":    "校团委终审",
	"branch":    "团支部大会讨论",
	"meeting":   "大会表决",
}

var politicalStatusTextMap = map[string]string{
	"masses":            "群众",
	"activist":         "入团积极分子",
	"probationary":      "预备团员",
	"member":            "共青团员",
	"party_probationary":"预备党员",
	"party_member":      "中共党员",
}

// DevelopmentTrack 查询某学生的团员发展全流程轨迹。
// 聚合 ty_approval_record 全模块审批记录 + 各阶段关键事件。
func (s *ApplicationService) DevelopmentTrack(studentID int64) (*DevelopmentTrackResult, error) {
	// 查询学生信息
	student, err := s.repo.GetStudentByID(studentID)
	if err != nil {
		return nil, fmt.Errorf("学生不存在")
	}

	// 查询该学生的入团申请
	var apps []models.TyApplication
	if err := s.db.Where("student_id = ? AND is_deleted = 0", studentID).Order("id ASC").Find(&apps).Error; err != nil {
		return nil, err
	}

	entries := make([]DevelopmentTrackEntry, 0)

	for _, app := range apps {
		// 1. 入团申请阶段（使用申请日期作为时间锚点，排在审批记录之前）
		entries = append(entries, DevelopmentTrackEntry{
			Module:      "application",
			ModuleText:  moduleTextMap["application"],
			TargetID:    app.ID,
			BizNo:       app.BizNo,
			Status:      app.Status,
			StatusText:  statusTextMap[app.Status],
			OccurredAt:  app.ApplyDate.Format("2006-01-02T15:04:05+08:00"),
		})

		// 2. 查询该申请关联的全流程审批记录
		records, err := s.repo.ListAllApprovalRecordsByApplication(app.ID)
		if err != nil {
			continue
		}
		for _, rec := range records {
			entry := DevelopmentTrackEntry{
				Module:       rec.Module,
				ModuleText:   moduleTextMap[rec.Module],
				TargetID:     rec.TargetID,
				BizNo:        app.BizNo,
				OccurredAt:   rec.OccurredAt.Format("2006-01-02T15:04:05+08:00"),
				Step:         rec.Step,
				StepText:     stepTextMapFull[rec.Step],
				Result:       rec.Result,
				ResultText:   resultTextMap[rec.Result],
				ApproverName: rec.ApproverName,
				Opinion:      rec.Opinion,
				FromStatus:   rec.FromStatus,
				ToStatus:     rec.ToStatus,
			}
			entries = append(entries, entry)
		}

		// 3. 查询推优大会记录
		var recMeetings []models.TyRecommendationMeeting
		s.db.Where("application_id = ? AND is_deleted = 0", app.ID).Order("meeting_at ASC").Find(&recMeetings)
		for _, m := range recMeetings {
			entries = append(entries, DevelopmentTrackEntry{
				Module:     "recommendation",
				ModuleText: moduleTextMap["recommendation"],
				TargetID:   m.ID,
				BizNo:      m.BizNo,
				Status:     m.Decision,
				StatusText: map[string]string{"pass": "通过", "reject": "未通过", "": "待决议"}[m.Decision],
				OccurredAt: m.MeetingAt.Format("2006-01-02T15:04:05+08:00"),
			})
		}

		// 4. 查询培养考察记录
		var cultRecords []models.TyCultivationRecord
		s.db.Where("application_id = ? AND is_deleted = 0", app.ID).Order("created_at ASC").Find(&cultRecords)
		for _, cr := range cultRecords {
			entries = append(entries, DevelopmentTrackEntry{
				Module:     "cultivation",
				ModuleText: moduleTextMap["cultivation"],
				TargetID:   cr.ID,
				BizNo:      app.BizNo,
				Status:     "recorded",
				StatusText: fmt.Sprintf("%d年%d月记录", cr.RecordYear, cr.RecordMonth),
				OccurredAt: cr.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			})
		}

		// 5. 查询发展对象记录
		var devObjs []models.TyDevelopmentObject
		s.db.Where("application_id = ? AND is_deleted = 0", app.ID).Order("id ASC").Find(&devObjs)
		for _, d := range devObjs {
			entries = append(entries, DevelopmentTrackEntry{
				Module:     "development_object",
				ModuleText: moduleTextMap["development_object"],
				TargetID:   d.ID,
				BizNo:      d.BizNo,
				Status:     d.Status,
				StatusText: devObjStatusTextMap[d.Status],
				OccurredAt: d.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			})
		}

		// 6. 查询政审记录
		var polReviews []models.TyPoliticalReview
		s.db.Where("application_id = ? AND is_deleted = 0", app.ID).Order("id ASC").Find(&polReviews)
		for _, p := range polReviews {
			conclusionText := map[string]string{"pass": "合格", "basic_pass": "基本合格", "fail": "不合格"}[p.Conclusion]
			entries = append(entries, DevelopmentTrackEntry{
				Module:     "political_review",
				ModuleText: moduleTextMap["political_review"],
				TargetID:   p.ID,
				BizNo:      app.BizNo,
				Status:     p.Conclusion,
				StatusText: conclusionText,
				OccurredAt: p.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			})
		}

		// 7. 查询发展大会记录
		var devMeetings []models.TyDevelopmentMeeting
		s.db.Where("application_id = ? AND is_deleted = 0", app.ID).Order("id ASC").Find(&devMeetings)
		for _, dm := range devMeetings {
			entries = append(entries, DevelopmentTrackEntry{
				Module:     "development_meeting",
				ModuleText: moduleTextMap["development_meeting"],
				TargetID:   dm.ID,
				BizNo:      dm.BizNo,
				Status:     dm.Decision,
				StatusText: map[string]string{"pass": "通过", "reject": "未通过", "": "待决议"}[dm.Decision],
				OccurredAt: dm.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			})
		}

		// 8. 查询转正大会记录
		var probMeetings []models.TyProbationaryMeeting
		s.db.Where("application_id = ? AND is_deleted = 0", app.ID).Order("id ASC").Find(&probMeetings)
		for _, pm := range probMeetings {
			entries = append(entries, DevelopmentTrackEntry{
				Module:     "probationary",
				ModuleText: moduleTextMap["probationary"],
				TargetID:   pm.ID,
				BizNo:      pm.BizNo,
				Status:     pm.Decision,
				StatusText: map[string]string{"pass": "通过", "reject": "未通过"}[pm.Decision],
				OccurredAt: pm.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			})
		}
	}

	// 按时间排序
	sortEntriesByTime(entries)

	return &DevelopmentTrackResult{
		StudentID:          studentID,
		StudentName:        student.Name,
		PoliticalStatus:    student.PoliticalStatus,
		PoliticalStatusText: politicalStatusTextMap[student.PoliticalStatus],
		Entries:            entries,
	}, nil
}

// sortEntriesByTime 按时间正序排列轨迹条目。
func sortEntriesByTime(entries []DevelopmentTrackEntry) {
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].OccurredAt > entries[j].OccurredAt {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

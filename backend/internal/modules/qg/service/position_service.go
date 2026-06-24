package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/qg/repository"
)

// PositionService 岗位+申请业务服务层。
type PositionService struct {
	repo *repository.PositionRepository
	db   *gorm.DB
	bus  *eventx.Bus
}

// NewPositionService 创建岗位服务。
func NewPositionService(repo *repository.PositionRepository, db *gorm.DB, bus *eventx.Bus) *PositionService {
	return &PositionService{repo: repo, db: db, bus: bus}
}

// ---- DTO ----

// PositionListResult 岗位列表结果。
type PositionListResult struct {
	Items    []PositionView `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// PositionView 岗位视图。
type PositionView struct {
	ID               int64   `json:"id"`
	BizNo            string  `json:"biz_no"`
	DeptType         string  `json:"dept_type"`
	DeptName         string  `json:"dept_name"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Headcount        int     `json:"headcount"`
	WeeklyHoursLimit int     `json:"weekly_hours_limit"`
	HourlyRateCents  int64   `json:"hourly_rate_cents"`
	StartAt          string  `json:"start_at"`
	EndAt            string  `json:"end_at"`
	RiskNotes        string  `json:"risk_notes"`
	KpiJSON          string  `json:"kpi_json"`
	Status           string  `json:"status"`
	StatusText       string  `json:"status_text"`
	SupervisorUserID *int64  `json:"supervisor_user_id,omitempty"`
	SupervisorName   string  `json:"supervisor_name,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// CreatePositionRequest 创建岗位请求。
type CreatePositionRequest struct {
	DeptType         string `json:"dept_type" binding:"required"`
	DeptName         string `json:"dept_name" binding:"required"`
	Title            string `json:"title" binding:"required"`
	Description      string `json:"description" binding:"required"`
	Headcount        int    `json:"headcount" binding:"required"`
	WeeklyHoursLimit int    `json:"weekly_hours_limit" binding:"required"`
	HourlyRateCents  int64  `json:"hourly_rate_cents" binding:"required"`
	StartAt          string `json:"start_at" binding:"required"`
	EndAt            string `json:"end_at" binding:"required"`
	RiskNotes        string `json:"risk_notes"`
	KpiJSON          string `json:"kpi_json"`
	SupervisorUserID *int64 `json:"supervisor_user_id"`
}

// ApplyPositionRequest 申请岗位请求。
type ApplyPositionRequest struct {
	PositionID   int64  `json:"position_id" binding:"required"`
	ResumeFileID *int64 `json:"resume_file_id"`
}

// ApplyView 岗位申请视图。
type ApplyView struct {
	ID              int64   `json:"id"`
	BizNo           string  `json:"biz_no"`
	PositionID      int64   `json:"position_id"`
	PositionTitle   string  `json:"position_title"`
	StudentID       int64   `json:"student_id"`
	StudentName     string  `json:"student_name"`
	ResumeFileID    *int64  `json:"resume_file_id,omitempty"`
	ApplyStatus     string  `json:"apply_status"`
	ApplyStatusText string  `json:"apply_status_text"`
	InterviewAt     *string `json:"interview_at,omitempty"`
	InterviewNote   string  `json:"interview_note"`
	ConfirmDeadline *string `json:"confirm_deadline,omitempty"`
	ConfirmedAt     *string `json:"confirmed_at,omitempty"`
	OnBoardAt       *string `json:"on_board_at,omitempty"`
	OffBoardAt      *string `json:"off_board_at,omitempty"`
	Status          string  `json:"status"`
	StatusText      string  `json:"status_text"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// ---- 状态映射 ----

var positionStatusTextMap = map[string]string{
	"S0":     "草稿",
	"S1":     "待审",
	"S2":     "院系通过",
	"S3":     "终审通过",
	"S4":     "已驳回",
	"closed": "已关闭",
}

var applyStatusTextMap = map[string]string{
	"pending":   "待处理",
	"interview": "面试中",
	"accepted":  "已录用",
	"rejected":  "已拒绝",
	"abandoned": "已放弃",
	"expired":   "已过期",
}

var applyOnboardStatusTextMap = map[string]string{
	"onboarding": "入职中",
	"on_job":     "在岗",
	"renewal":    "续聘",
	"terminated": "已解聘",
	"closed":     "已关闭",
}

// ---- 岗位业务方法 ----

// List 分页查询岗位列表。
// keyword：按部门名称 dept_name 模糊搜索。
func (s *PositionService) List(keyword, deptType, status string, page, pageSize int) (*PositionListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	positions, total, err := s.repo.ListPositions(keyword, deptType, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]PositionView, 0, len(positions))
	for _, pos := range positions {
		v := s.toPositionView(pos)
		items = append(items, v)
	}

	return &PositionListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取岗位详情。
func (s *PositionService) Get(id int64) (*PositionView, error) {
	pos, err := s.repo.GetPositionByID(id)
	if err != nil {
		return nil, fmt.Errorf("岗位不存在")
	}

	v := s.toPositionView(*pos)
	return &v, nil
}

// Create 创建岗位。
func (s *PositionService) Create(userID int64, req *CreatePositionRequest) (*PositionView, error) {
	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "QG-POS")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// 解析时间
	startAt, err := parseTime(req.StartAt)
	if err != nil {
		return nil, fmt.Errorf("岗位开始时间格式错误")
	}
	endAt, err := parseTime(req.EndAt)
	if err != nil {
		return nil, fmt.Errorf("岗位结束时间格式错误")
	}

	pos := &models.QgPosition{
		BizNo:            bizNo,
		DeptType:         req.DeptType,
		DeptName:         req.DeptName,
		Title:            req.Title,
		Description:      req.Description,
		Headcount:        req.Headcount,
		WeeklyHoursLimit: req.WeeklyHoursLimit,
		HourlyRateCents:  req.HourlyRateCents,
		StartAt:          startAt,
		EndAt:            endAt,
		RiskNotes:        req.RiskNotes,
		KpiJSON:          req.KpiJSON,
		Status:           "S0",
		SupervisorUserID: req.SupervisorUserID,
		CreatedBy:        &userID,
		UpdatedBy:        &userID,
	}

	if err := s.repo.CreatePosition(pos); err != nil {
		return nil, err
	}

	return s.Get(pos.ID)
}

// Submit 提交岗位（S0→S1）。
func (s *PositionService) Submit(id int64) (*PositionView, error) {
	pos, err := s.repo.GetPositionByID(id)
	if err != nil {
		return nil, fmt.Errorf("岗位不存在")
	}

	if pos.Status != "S0" {
		return nil, fmt.Errorf("当前状态不允许提交")
	}

	pos.Status = "S1"
	if err := s.repo.UpdatePosition(pos); err != nil {
		return nil, err
	}

	return s.Get(id)
}

// Approve 审批岗位（college: S1→S2, school: S2→S3）。
func (s *PositionService) Approve(id, userID int64, level string) (*PositionView, error) {
	pos, err := s.repo.GetPositionByID(id)
	if err != nil {
		return nil, fmt.Errorf("岗位不存在")
	}

	switch {
	case pos.Status == "S1" && level == "college":
		pos.Status = "S2"
	case pos.Status == "S2" && level == "school":
		pos.Status = "S3"
	default:
		return nil, fmt.Errorf("当前状态不允许审批")
	}

	updatedBy := userID
	pos.UpdatedBy = &updatedBy

	if err := s.repo.UpdatePosition(pos); err != nil {
		return nil, err
	}

	// 岗位审批通过触发事件
	if pos.Status == "S3" && s.bus != nil {
		_ = s.bus.Publish(&eventx.Event{
			Aggregate:   "qg.position",
			AggregateID: pos.BizNo,
			EventType:   "QgPositionApproved",
			Module:      "QG",
			ActorID:     userID,
			Payload: map[string]interface{}{
				"position_id": pos.ID,
				"biz_no":      pos.BizNo,
				"title":       pos.Title,
				"dept_type":   pos.DeptType,
			},
			BizNo: pos.BizNo,
		})
	}

	return s.Get(id)
}

// Reject 驳回岗位。
func (s *PositionService) Reject(id, userID int64, opinion string) (*PositionView, error) {
	pos, err := s.repo.GetPositionByID(id)
	if err != nil {
		return nil, fmt.Errorf("岗位不存在")
	}

	if pos.Status != "S1" && pos.Status != "S2" {
		return nil, fmt.Errorf("当前状态不允许驳回")
	}

	pos.Status = "S4"
	updatedBy := userID
	pos.UpdatedBy = &updatedBy

	if err := s.repo.UpdatePosition(pos); err != nil {
		return nil, err
	}

	return s.Get(id)
}

// Delete 软删除岗位。
func (s *PositionService) Delete(id int64) error {
	return s.repo.SoftDeletePosition(id)
}

// ---- 申请业务方法 ----

// Apply 学生申请岗位。
func (s *PositionService) Apply(userID, studentID int64, req *ApplyPositionRequest) (*ApplyView, error) {
	// BR-QG-01: 必须有困难认定（level != none 且 status=S3）
	cert, err := s.repo.GetActiveDifficultyCert(studentID)
	if err != nil || cert == nil {
		return nil, &BizError{Code: 40301, Msg: "必须先完成困难认定才能申请岗位"}
	}
	if cert.Level == "none" {
		return nil, &BizError{Code: 40301, Msg: "困难认定等级为'不困难'，无法申请岗位"}
	}

	// 同岗位不可重复投递
	exists, err := s.repo.ExistsApplyByPositionAndStudent(req.PositionID, studentID)
	if err != nil {
		return nil, fmt.Errorf("查询申请记录失败: %w", err)
	}
	if exists {
		return nil, &BizError{Code: 40901, Msg: "已申请该岗位，不可重复投递"}
	}

	// 同时在岗 ≤ 1
	onJobCount, err := s.repo.CountOnJobByStudent(studentID)
	if err != nil {
		return nil, fmt.Errorf("查询在岗记录失败: %w", err)
	}
	if onJobCount >= 1 {
		return nil, &BizError{Code: 40902, Msg: "已在岗1个岗位，不可再申请"}
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "QG")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	apply := &models.QgPositionApply{
		BizNo:        bizNo,
		PositionID:   req.PositionID,
		StudentID:    studentID,
		ResumeFileID: req.ResumeFileID,
		ApplyStatus:  "pending",
		Status:       "onboarding",
	}

	if err := s.repo.CreateApply(apply); err != nil {
		return nil, err
	}

	return s.GetApply(apply.ID)
}

// AcceptApply 录用（设置 confirm_deadline +3工作日）。
func (s *PositionService) AcceptApply(id, userID int64) (*ApplyView, error) {
	apply, err := s.repo.GetApplyByID(id)
	if err != nil {
		return nil, fmt.Errorf("申请记录不存在")
	}

	if apply.ApplyStatus != "pending" && apply.ApplyStatus != "interview" {
		return nil, fmt.Errorf("当前状态不允许录用")
	}

	apply.ApplyStatus = "accepted"
	// 设置确认截止时间（+3工作日，简化处理按+3自然日）
	deadline := addWorkDays(time.Now(), 3)
	apply.ConfirmDeadline = &deadline

	if err := s.repo.UpdateApply(apply); err != nil {
		return nil, err
	}

	return s.GetApply(id)
}

// ConfirmApply 学生确认录用。
func (s *PositionService) ConfirmApply(id int64) (*ApplyView, error) {
	apply, err := s.repo.GetApplyByID(id)
	if err != nil {
		return nil, fmt.Errorf("申请记录不存在")
	}

	if apply.ApplyStatus != "accepted" {
		return nil, fmt.Errorf("当前状态不允许确认")
	}

	// 检查确认截止时间
	if apply.ConfirmDeadline != nil && time.Now().After(*apply.ConfirmDeadline) {
		apply.ApplyStatus = "expired"
		_ = s.repo.UpdateApply(apply)
		return nil, &BizError{Code: 41001, Msg: "确认截止时间已过，申请已过期"}
	}

	now := time.Now()
	apply.ConfirmedAt = &now
	apply.ApplyStatus = "accepted"

	if err := s.repo.UpdateApply(apply); err != nil {
		return nil, err
	}

	return s.GetApply(id)
}

// Onboard 上岗。
func (s *PositionService) Onboard(id int64) (*ApplyView, error) {
	apply, err := s.repo.GetApplyByID(id)
	if err != nil {
		return nil, fmt.Errorf("申请记录不存在")
	}

	if apply.ApplyStatus != "accepted" {
		return nil, fmt.Errorf("当前状态不允许上岗")
	}

	now := time.Now()
	apply.OnBoardAt = &now
	apply.Status = "on_job"

	if err := s.repo.UpdateApply(apply); err != nil {
		return nil, err
	}

	return s.GetApply(id)
}

// GetApply 获取申请详情。
func (s *PositionService) GetApply(id int64) (*ApplyView, error) {
	apply, err := s.repo.GetApplyByID(id)
	if err != nil {
		return nil, fmt.Errorf("申请记录不存在")
	}

	v := s.toApplyView(*apply)
	return &v, nil
}

// ApplyListResult 岗位申请列表结果。
type ApplyListResult struct {
	Items    []ApplyView `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ListApplies 分页查询岗位申请列表（前端下拉选择器取数）。
// status 可选,如 "on_job";keyword 可对岗位标题 / 学生姓名 / 学号做模糊匹配。
func (s *PositionService) ListApplies(status, keyword string, page, pageSize int) (*ApplyListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	applies, total, err := s.repo.ListApplies(status, keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]ApplyView, 0, len(applies))
	for _, a := range applies {
		items = append(items, s.toApplyView(a))
	}

	return &ApplyListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ---- 内部方法 ----

func (s *PositionService) toPositionView(pos models.QgPosition) PositionView {
	v := PositionView{
		ID:               pos.ID,
		BizNo:            pos.BizNo,
		DeptType:         pos.DeptType,
		DeptName:         pos.DeptName,
		Title:            pos.Title,
		Description:      pos.Description,
		Headcount:        pos.Headcount,
		WeeklyHoursLimit: pos.WeeklyHoursLimit,
		HourlyRateCents:  pos.HourlyRateCents,
		StartAt:          pos.StartAt.Format("2006-01-02T15:04:05+08:00"),
		EndAt:            pos.EndAt.Format("2006-01-02T15:04:05+08:00"),
		RiskNotes:        pos.RiskNotes,
		KpiJSON:          pos.KpiJSON,
		Status:           pos.Status,
		StatusText:       positionStatusTextMap[pos.Status],
		SupervisorUserID: pos.SupervisorUserID,
		CreatedAt:        pos.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:        pos.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 加载主管姓名
	if pos.SupervisorUserID != nil {
		if u, err := s.repo.GetUserByID(*pos.SupervisorUserID); err == nil {
			v.SupervisorName = u.DisplayName
		}
	}

	return v
}

func (s *PositionService) toApplyView(apply models.QgPositionApply) ApplyView {
	v := ApplyView{
		ID:            apply.ID,
		BizNo:         apply.BizNo,
		PositionID:    apply.PositionID,
		StudentID:     apply.StudentID,
		ResumeFileID:  apply.ResumeFileID,
		ApplyStatus:   apply.ApplyStatus,
		ApplyStatusText: applyStatusTextMap[apply.ApplyStatus],
		InterviewNote: apply.InterviewNote,
		Status:        apply.Status,
		StatusText:    applyOnboardStatusTextMap[apply.Status],
		CreatedAt:     apply.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:     apply.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	if apply.InterviewAt != nil {
		t := apply.InterviewAt.Format("2006-01-02T15:04:05+08:00")
		v.InterviewAt = &t
	}
	if apply.ConfirmDeadline != nil {
		t := apply.ConfirmDeadline.Format("2006-01-02T15:04:05+08:00")
		v.ConfirmDeadline = &t
	}
	if apply.ConfirmedAt != nil {
		t := apply.ConfirmedAt.Format("2006-01-02T15:04:05+08:00")
		v.ConfirmedAt = &t
	}
	if apply.OnBoardAt != nil {
		t := apply.OnBoardAt.Format("2006-01-02T15:04:05+08:00")
		v.OnBoardAt = &t
	}
	if apply.OffBoardAt != nil {
		t := apply.OffBoardAt.Format("2006-01-02T15:04:05+08:00")
		v.OffBoardAt = &t
	}

	// 加载岗位标题
	if pos, err := s.repo.GetPositionByID(apply.PositionID); err == nil {
		v.PositionTitle = pos.Title
	}

	// 加载学生姓名
	if student, err := s.repo.GetStudentByID(apply.StudentID); err == nil {
		v.StudentName = student.Name
	}

	return v
}

// addWorkDays 添加工作日（简化处理：跳过周末）。
func addWorkDays(start time.Time, days int) time.Time {
	result := start
	for i := 0; i < days; {
		result = result.AddDate(0, 0, 1)
		wd := result.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			i++
		}
	}
	return result
}

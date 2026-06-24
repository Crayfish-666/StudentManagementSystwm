package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/qg/repository"
	qgsm "student-system/internal/modules/qg/statemachine"
	"student-system/internal/statem"
)

// AssessmentService 月度考核+薪酬业务服务层。
type AssessmentService struct {
	repo *repository.AssessmentRepository
	db   *gorm.DB
	sm   *statem.Engine
	bus  *eventx.Bus
}

// NewAssessmentService 创建考核薪酬服务。
func NewAssessmentService(repo *repository.AssessmentRepository, db *gorm.DB, bus *eventx.Bus) *AssessmentService {
	return &AssessmentService{repo: repo, db: db, sm: qgsm.NewAssessSM(), bus: bus}
}

// ---- DTO ----

// AssessmentListResult 考核列表结果。
type AssessmentListResult struct {
	Items    []AssessmentView `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// AssessmentView 考核视图。
type AssessmentView struct {
	ID                 int64   `json:"id"`
	BizNo              string  `json:"biz_no"`
	ApplyID            int64   `json:"apply_id"`
	PositionTitle      string  `json:"position_title"`
	StudentID          int64   `json:"student_id"`
	StudentName        string  `json:"student_name"`
	AssessYear         int     `json:"assess_year"`
	AssessMonth        int     `json:"assess_month"`
	ScoreAttendance    int     `json:"score_attendance"`
	ScoreWorkComplete  int     `json:"score_work_complete"`
	ScoreComprehensive int     `json:"score_comprehensive"`
	WeightedScore      float64 `json:"weighted_score"`
	Coefficient        float64 `json:"coefficient"`
	CoefficientText    string  `json:"coefficient_text"`
	IsObservation      int     `json:"is_observation"`
	Note               string  `json:"note"`
	Status             string  `json:"status"`
	StatusText         string  `json:"status_text"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

// CreateAssessmentRequest 创建考核请求。
type CreateAssessmentRequest struct {
	ApplyID            int64  `json:"apply_id" binding:"required"`
	AssessYear         int    `json:"assess_year" binding:"required"`
	AssessMonth        int    `json:"assess_month" binding:"required"`
	ScoreAttendance    int    `json:"score_attendance" binding:"required"`
	ScoreWorkComplete  int    `json:"score_work_complete" binding:"required"`
	ScoreComprehensive int    `json:"score_comprehensive" binding:"required"`
	IsObservation      int    `json:"is_observation"`
	Note               string `json:"note"`
}

// PayrollListResult 薪酬列表结果。
type PayrollListResult struct {
	Items    []PayrollView `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// PayrollView 薪酬视图。
type PayrollView struct {
	ID                int64   `json:"id"`
	BizNo             string  `json:"biz_no"`
	StudentID         int64   `json:"student_id"`
	StudentName       string  `json:"student_name"`
	ApplyID           int64   `json:"apply_id"`
	PositionTitle     string  `json:"position_title"`
	PayYear           int     `json:"pay_year"`
	PayMonth          int     `json:"pay_month"`
	TotalHours        float64 `json:"total_hours"`
	GrossCents        int64   `json:"gross_cents"`
	TaxCents          int64   `json:"tax_cents"`
	DeductionCents    int64   `json:"deduction_cents"`
	NetCents          int64   `json:"net_cents"`
	Coefficient       float64 `json:"coefficient"`
	Status            string  `json:"status"`
	StatusText        string  `json:"status_text"`
	ReviewedBy        *int64  `json:"reviewed_by,omitempty"`
	PaidAt            *string `json:"paid_at,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// ---- 状态映射 ----

var assessStatusTextMap = map[string]string{
	"S1": "待确认",
	"S3": "已确认",
}

var payrollStatusTextMap = map[string]string{
	"draft":   "草稿",
	"reviewed": "已复核",
	"paid":    "已发放",
	"failed":  "发放失败",
}

var coefficientTextMap = map[float64]string{
	1.0: "全额",
	0.8: "八折",
	0.5: "五折",
	0.0: "无",
}

// ---- 考核业务方法 ----

// CreateAssessment 创建月度考核。
// 唯一性约束：SSOT §8.2.7 规定 UNIQUE(apply_id, assess_year, assess_month)。
// 这里先做存在性预检,并在 INSERT 后兜底捕获 UNIQUE 索引冲突,统一返回业务错误 "考核已存在"，
// 由 Handler 映射为 40905 错误码。
func (s *AssessmentService) CreateAssessment(userID int64, req *CreateAssessmentRequest) (*AssessmentView, error) {
	// 预检:同 apply + 年月 已有记录则拒绝
	if existing, err := s.repo.GetAssessmentByApplyAndMonth(req.ApplyID, req.AssessYear, req.AssessMonth); err == nil && existing != nil {
		return nil, fmt.Errorf("考核已存在")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询考核是否已存在失败: %w", err)
	}

	// 计算加权得分
	weightedScore := float64(req.ScoreAttendance)*0.4 + float64(req.ScoreWorkComplete)*0.4 + float64(req.ScoreComprehensive)*0.2
	weightedScore = math.Round(weightedScore*100) / 100

	// 计算考核系数
	var coefficient float64
	switch {
	case weightedScore >= 85:
		coefficient = 1.0
	case weightedScore >= 60:
		coefficient = 0.8
	default:
		coefficient = 0.5
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "QG")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// 获取申请记录以取得学生ID
	apply, err := s.repo.GetApplyByID(req.ApplyID)
	if err != nil {
		return nil, fmt.Errorf("岗位申请记录不存在")
	}

	assess := &models.QgMonthlyAssess{
		BizNo:              bizNo,
		ApplyID:            req.ApplyID,
		StudentID:          apply.StudentID,
		AssessYear:         req.AssessYear,
		AssessMonth:        req.AssessMonth,
		ScoreAttendance:    req.ScoreAttendance,
		ScoreWorkComplete:  req.ScoreWorkComplete,
		ScoreComprehensive: req.ScoreComprehensive,
		WeightedScore:      weightedScore,
		Coefficient:        coefficient,
		IsObservation:      req.IsObservation,
		Note:               req.Note,
		Status:             "S1",
	}

	if err := s.repo.CreateAssessment(assess); err != nil {
		// 兜底:并发场景下预检可能漏过,捕获 UNIQUE 索引冲突并转为业务错误。
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("考核已存在")
		}
		return nil, fmt.Errorf("创建考核记录失败: %w", err)
	}

	return s.GetAssess(assess.ID)
}

// isUniqueViolation 判断 GORM/SQLite 错误是否为 UNIQUE 约束冲突。
// SQLite 错误格式: "UNIQUE constraint failed: <table>.<col>, ..."
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "Duplicate entry")
}

// ListAssess 分页查询考核列表。
func (s *AssessmentService) ListAssess(year, month int, applyID int64, positionTitle string, page, pageSize int) (*AssessmentListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	assessments, total, err := s.repo.ListAssessments(year, month, applyID, positionTitle, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]AssessmentView, 0, len(assessments))
	for _, a := range assessments {
		v := s.toAssessView(a)
		items = append(items, v)
	}

	return &AssessmentListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetAssess 获取考核详情。
func (s *AssessmentService) GetAssess(id int64) (*AssessmentView, error) {
	assess, err := s.repo.GetAssessmentByID(id)
	if err != nil {
		return nil, fmt.Errorf("考核记录不存在")
	}

	v := s.toAssessView(*assess)
	return &v, nil
}

// ConfirmAssessment 确认月度考核（S1 → S3）。
// 仅 S1(待确认) 允许被确认；S3(已确认) 重复确认会被状态机阻断并返回业务错误。
// 状态变更走 statem.Engine.Apply，持久化与 event_log 写入在同一事务内。
func (s *AssessmentService) ConfirmAssessment(id, userID int64, actorName, actorRole, ip, ua string) (*AssessmentView, error) {
	assess, err := s.repo.GetAssessmentByID(id)
	if err != nil {
		return nil, fmt.Errorf("考核记录不存在")
	}

	from := assess.Status
	to, err := s.sm.Apply(&statem.BizCtx{
		Ctx:       context.Background(),
		ActorID:   userID,
		ActorName: actorName,
		ActorRole: actorRole,
		IP:        ip,
		UA:        ua,
	}, from, qgsm.ActionConfirm)
	if err != nil {
		return nil, err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		assess.Status = to
		return tx.Save(assess).Error
	}); err != nil {
		return nil, fmt.Errorf("更新考核状态失败: %w", err)
	}

	s.publishAssessEvent(assess, "QgMonthlyAssessConfirmed", userID, actorRole, ip, ua, map[string]interface{}{
		"from": from,
		"to":   to,
	})

	return s.GetAssess(id)
}

// publishAssessEvent 发布月度考核领域事件到 event_log。
func (s *AssessmentService) publishAssessEvent(assess *models.QgMonthlyAssess, evtType string, actorID int64, actorRole, ip, ua string, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["assess_id"] = assess.ID
	payload["biz_no"] = assess.BizNo
	payload["status"] = assess.Status

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "qg.monthly_assess",
		AggregateID: assess.BizNo,
		EventType:   evtType,
		Module:      "QG",
		ActorID:     actorID,
		ActorRole:   actorRole,
		Payload:     payload,
		BizNo:       assess.BizNo,
		IP:          ip,
		UA:          ua,
	})
}

// ---- 薪酬业务方法 ----

// ComputePayroll 计算薪酬。
// 公式: Σ(每日有效工时) × 小时薪酬标准 × 考核系数
func (s *AssessmentService) ComputePayroll(applyID int64, year, month int) (*PayrollView, error) {
	// 获取申请记录
	apply, err := s.repo.GetApplyByID(applyID)
	if err != nil {
		return nil, fmt.Errorf("岗位申请记录不存在")
	}

	// 获取岗位信息（含时薪）
	position, err := s.repo.GetPositionByID(apply.PositionID)
	if err != nil {
		return nil, fmt.Errorf("岗位信息不存在")
	}

	// 获取当月考核
	assess, err := s.repo.GetAssessmentByApplyAndMonth(applyID, year, month)
	if err != nil {
		return nil, fmt.Errorf("当月考核记录不存在，请先创建考核")
	}

	// 获取当月有效工时记录
	attendances, err := s.repo.ListAttendancesByApplyAndMonth(applyID, year, month)
	if err != nil {
		return nil, fmt.Errorf("查询工时记录失败: %w", err)
	}

	// 计算总工时
	var totalHours float64
	for _, a := range attendances {
		totalHours += a.EffectiveHours
	}
	totalHours = math.Round(totalHours*100) / 100

	// 计算薪酬
	grossCents := int64(math.Round(totalHours * float64(position.HourlyRateCents) * assess.Coefficient))
	taxCents := int64(0)     // 简化处理，暂不扣税
	deductionCents := int64(0) // 简化处理，暂不扣减
	netCents := grossCents - taxCents - deductionCents

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "QG")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// 事务：创建薪酬 + 薪酬明细
	payroll := &models.QgPayroll{
		BizNo:      bizNo,
		StudentID:  apply.StudentID,
		ApplyID:    applyID,
		PayYear:    year,
		PayMonth:   month,
		TotalHours: totalHours,
		GrossCents: grossCents,
		TaxCents:   taxCents,
		DeductionCents: deductionCents,
		NetCents:   netCents,
		Coefficient: assess.Coefficient,
		Status:     "draft",
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(payroll).Error; err != nil {
			return err
		}

		// 创建薪酬明细
		for _, a := range attendances {
			amountCents := int64(math.Round(a.EffectiveHours * float64(position.HourlyRateCents) * assess.Coefficient))
			detail := &models.QgPayrollDetail{
				PayrollID:    payroll.ID,
				AttendanceID: a.ID,
				WorkDate:     a.WorkDate,
				Hours:        a.EffectiveHours,
				RateCents:    position.HourlyRateCents,
				AmountCents:  amountCents,
			}
			if err := tx.Create(detail).Error; err != nil {
				return fmt.Errorf("创建薪酬明细失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.GetPayroll(payroll.ID)
}

// ListPayroll 分页查询薪酬列表。
func (s *AssessmentService) ListPayroll(year, month int, status, positionTitle string, page, pageSize int) (*PayrollListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	payrolls, total, err := s.repo.ListPayrolls(year, month, status, positionTitle, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]PayrollView, 0, len(payrolls))
	for _, p := range payrolls {
		v := s.toPayrollView(p)
		items = append(items, v)
	}

	return &PayrollListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetPayroll 获取薪酬详情。
func (s *AssessmentService) GetPayroll(id int64) (*PayrollView, error) {
	payroll, err := s.repo.GetPayrollByID(id)
	if err != nil {
		return nil, fmt.Errorf("薪酬记录不存在")
	}

	v := s.toPayrollView(*payroll)
	return &v, nil
}

// ReviewPayroll 复核薪酬（draft→reviewed）。
func (s *AssessmentService) ReviewPayroll(id, userID int64) (*PayrollView, error) {
	payroll, err := s.repo.GetPayrollByID(id)
	if err != nil {
		return nil, fmt.Errorf("薪酬记录不存在")
	}

	if payroll.Status != "draft" {
		return nil, fmt.Errorf("当前状态不允许复核")
	}

	payroll.Status = "reviewed"
	payroll.ReviewedBy = &userID

	if err := s.repo.UpdatePayroll(payroll); err != nil {
		return nil, err
	}

	return s.GetPayroll(id)
}

// PayPayroll 发放薪酬（reviewed→paid）。
func (s *AssessmentService) PayPayroll(id, userID int64) (*PayrollView, error) {
	payroll, err := s.repo.GetPayrollByID(id)
	if err != nil {
		return nil, fmt.Errorf("薪酬记录不存在")
	}

	if payroll.Status != "reviewed" {
		return nil, fmt.Errorf("当前状态不允许发放")
	}

	now := time.Now()
	payroll.Status = "paid"
	payroll.PaidAt = &now

	if err := s.repo.UpdatePayroll(payroll); err != nil {
		return nil, err
	}

	return s.GetPayroll(id)
}

// ---- 内部方法 ----

// AttendancePreview 出勤分自动计算的预览结果（不写库）。
// 用于"创建月度考核"对话框实时回填"出勤分"输入框。
type AttendancePreview struct {
	ShouldHours     float64 `json:"should_hours"`    // 标准工时（小时）= weekly_hours_limit × 4
	ActualHours     float64 `json:"actual_hours"`    // 实出勤工时（小时）= sum(qg_attendance.effective_hours)
	OnBoardAt       string  `json:"on_board_at"`     // 学生实际上岗日期（YYYY-MM-DD）
	ScoreAttendance int     `json:"score_attendance"` // 计算出的出勤分（0-100）
	Formula         string  `json:"formula"`         // 公式说明
}

// PreviewAttendance 计算某 apply 在某年某月的出勤分预览。
// 算法：出勤分 = round(实出勤工时 / 标准工时 × 100), 封顶 100。
//   - 标准工时 = qg_position.weekly_hours_limit × 4（约 4 周一月）
//   - 实出勤工时 = sum(qg_attendance.effective_hours), 当月, is_deleted=0
//   - 兼职/寒暑假标准工时由岗位表的 weekly_hours_limit 决定（V1 简化）
//   - 综合分保持人工输入，本接口不计算
func (s *AssessmentService) PreviewAttendance(applyID int64, year, month int) (*AttendancePreview, error) {
	apply, err := s.repo.GetApplyByID(applyID)
	if err != nil {
		return nil, fmt.Errorf("岗位申请记录不存在")
	}
	if month < 1 || month > 12 || year < 1900 {
		return nil, fmt.Errorf("年月参数非法")
	}

	// 标准工时（来自岗位周工时上限 × 4）
	pos, err := s.repo.GetPositionByApplyID(applyID)
	if err != nil || pos == nil {
		return nil, fmt.Errorf("岗位信息不存在")
	}
	weeklyHours := pos.WeeklyHoursLimit
	if weeklyHours <= 0 {
		weeklyHours = 10 // 兜底: 岗位未配周工时上限时按 10h/周(40h/月)
	}
	shouldHours := float64(weeklyHours) * 4

	// 实出勤工时
	actualHours, err := s.repo.SumEffectiveHoursByApplyAndMonth(applyID, year, month)
	if err != nil {
		return nil, fmt.Errorf("统计实出勤工时失败: %w", err)
	}

	score := 0
	if shouldHours > 0 {
		score = int(math.Round(actualHours / shouldHours * 100))
		if score > 100 {
			score = 100
		}
		if score < 0 {
			score = 0
		}
	}

	onBoardStr := ""
	if apply.OnBoardAt != nil {
		onBoardStr = apply.OnBoardAt.Format("2006-01-02")
	}

	return &AttendancePreview{
		ShouldHours:     shouldHours,
		ActualHours:     actualHours,
		OnBoardAt:       onBoardStr,
		ScoreAttendance: score,
		Formula:         fmt.Sprintf("实出勤工时 / 标准工时(每周上限 %dh × 4 周 ≈ %dh) × 100(封顶 100)", weeklyHours, int(shouldHours)),
	}, nil
}

// countWorkdays 统计 [start, end) 区间内周一~五的天数(半开区间)。
// 保留以备其他业务复用;当前 PreviewAttendance 已改用"工时制"算法。
func countWorkdays(start, end time.Time) int {
	if !end.After(start) {
		return 0
	}
	count := 0
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		wd := d.Weekday()
		if wd >= time.Monday && wd <= time.Friday {
			count++
		}
	}
	return count
}

// monthRange 返回 [月初, 下月初) 的半开区间。
// year/month 非法时返回两个零时间。
func monthRange(year, month int) (time.Time, time.Time) {
	if year < 1900 || month < 1 || month > 12 {
		return time.Time{}, time.Time{}
	}
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, 0)
	return first, last
}

func (s *AssessmentService) toAssessView(a models.QgMonthlyAssess) AssessmentView {
	v := AssessmentView{
		ID:                 a.ID,
		BizNo:              a.BizNo,
		ApplyID:            a.ApplyID,
		StudentID:          a.StudentID,
		AssessYear:         a.AssessYear,
		AssessMonth:        a.AssessMonth,
		ScoreAttendance:    a.ScoreAttendance,
		ScoreWorkComplete:  a.ScoreWorkComplete,
		ScoreComprehensive: a.ScoreComprehensive,
		WeightedScore:      a.WeightedScore,
		Coefficient:        a.Coefficient,
		CoefficientText:    coefficientTextMap[a.Coefficient],
		IsObservation:      a.IsObservation,
		Note:               a.Note,
		Status:             a.Status,
		StatusText:         assessStatusTextMap[a.Status],
		CreatedAt:          a.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:          a.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 加载岗位标题
	if apply, err := s.repo.GetApplyByID(a.ApplyID); err == nil {
		if pos, err := s.repo.GetPositionByID(apply.PositionID); err == nil {
			v.PositionTitle = pos.Title
		}
	}

	// 加载学生姓名
	if student, err := s.repo.GetStudentByID(a.StudentID); err == nil {
		v.StudentName = student.Name
	}

	return v
}

func (s *AssessmentService) toPayrollView(p models.QgPayroll) PayrollView {
	v := PayrollView{
		ID:             p.ID,
		BizNo:          p.BizNo,
		StudentID:      p.StudentID,
		ApplyID:        p.ApplyID,
		PayYear:        p.PayYear,
		PayMonth:       p.PayMonth,
		TotalHours:     p.TotalHours,
		GrossCents:     p.GrossCents,
		TaxCents:       p.TaxCents,
		DeductionCents: p.DeductionCents,
		NetCents:       p.NetCents,
		Coefficient:    p.Coefficient,
		Status:         p.Status,
		StatusText:     payrollStatusTextMap[p.Status],
		ReviewedBy:     p.ReviewedBy,
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	if p.PaidAt != nil {
		t := p.PaidAt.Format("2006-01-02T15:04:05+08:00")
		v.PaidAt = &t
	}

	// 加载学生姓名
	if student, err := s.repo.GetStudentByID(p.StudentID); err == nil {
		v.StudentName = student.Name
	}

	// 加载岗位标题
	if apply, err := s.repo.GetApplyByID(p.ApplyID); err == nil {
		if pos, err := s.repo.GetPositionByID(apply.PositionID); err == nil {
			v.PositionTitle = pos.Title
		}
	}

	return v
}

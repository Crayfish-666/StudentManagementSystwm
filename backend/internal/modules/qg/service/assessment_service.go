package service

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/qg/repository"
)

// AssessmentService 月度考核+薪酬业务服务层。
type AssessmentService struct {
	repo *repository.AssessmentRepository
	db   *gorm.DB
}

// NewAssessmentService 创建考核薪酬服务。
func NewAssessmentService(repo *repository.AssessmentRepository, db *gorm.DB) *AssessmentService {
	return &AssessmentService{repo: repo, db: db}
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
func (s *AssessmentService) CreateAssessment(userID int64, req *CreateAssessmentRequest) (*AssessmentView, error) {
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
		return nil, err
	}

	return s.GetAssess(assess.ID)
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

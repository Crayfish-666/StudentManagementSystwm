package repository

import (
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// AssessmentRepository 月度考核+薪酬数据访问层。
type AssessmentRepository struct {
	db *gorm.DB
}

// NewAssessmentRepository 创建考核薪酬仓储。
func NewAssessmentRepository(db *gorm.DB) *AssessmentRepository {
	return &AssessmentRepository{db: db}
}

// ---- 考核 ----

// CreateAssessment 创建考核记录。
func (r *AssessmentRepository) CreateAssessment(assess *models.QgMonthlyAssess) error {
	return r.db.Create(assess).Error
}

// GetAssessmentByID 按 ID 查询考核记录。
func (r *AssessmentRepository) GetAssessmentByID(id int64) (*models.QgMonthlyAssess, error) {
	var assess models.QgMonthlyAssess
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&assess).Error; err != nil {
		return nil, err
	}
	return &assess, nil
}

// ListAssessments 分页查询考核列表。
// positionTitle 非空时按岗位标题模糊匹配（join qg_position_apply + qg_position）。
func (r *AssessmentRepository) ListAssessments(year, month int, applyID int64, positionTitle string, page, pageSize int) ([]models.QgMonthlyAssess, int64, error) {
	query := r.db.Where("qg_monthly_assess.is_deleted = 0")
	if year > 0 {
		query = query.Where("qg_monthly_assess.assess_year = ?", year)
	}
	if month > 0 {
		query = query.Where("qg_monthly_assess.assess_month = ?", month)
	}
	if applyID > 0 {
		query = query.Where("qg_monthly_assess.apply_id = ?", applyID)
	}
	if positionTitle != "" {
		query = query.
			Joins("JOIN qg_position_apply pa ON pa.id = qg_monthly_assess.apply_id").
			Joins("JOIN qg_position p ON p.id = pa.position_id").
			Where("p.title LIKE ?", "%"+positionTitle+"%")
	}

	var total int64
	if err := query.Model(&models.QgMonthlyAssess{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var assessments []models.QgMonthlyAssess
	offset := (page - 1) * pageSize
	if err := query.Order("qg_monthly_assess.created_at DESC").Offset(offset).Limit(pageSize).Find(&assessments).Error; err != nil {
		return nil, 0, err
	}

	return assessments, total, nil
}

// GetAssessmentByApplyAndMonth 按申请和月份查询考核记录。
func (r *AssessmentRepository) GetAssessmentByApplyAndMonth(applyID int64, year, month int) (*models.QgMonthlyAssess, error) {
	var assess models.QgMonthlyAssess
	if err := r.db.Where("apply_id = ? AND assess_year = ? AND assess_month = ? AND is_deleted = 0", applyID, year, month).
		First(&assess).Error; err != nil {
		return nil, err
	}
	return &assess, nil
}

// ---- 薪酬 ----

// CreatePayroll 创建薪酬记录。
func (r *AssessmentRepository) CreatePayroll(payroll *models.QgPayroll) error {
	return r.db.Create(payroll).Error
}

// GetPayrollByID 按 ID 查询薪酬记录。
func (r *AssessmentRepository) GetPayrollByID(id int64) (*models.QgPayroll, error) {
	var payroll models.QgPayroll
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&payroll).Error; err != nil {
		return nil, err
	}
	return &payroll, nil
}

// ListPayrolls 分页查询薪酬列表。
// positionTitle 非空时按岗位标题模糊匹配（join qg_position_apply + qg_position）。
func (r *AssessmentRepository) ListPayrolls(year, month int, status, positionTitle string, page, pageSize int) ([]models.QgPayroll, int64, error) {
	query := r.db.Where("qg_payroll.is_deleted = 0")
	if year > 0 {
		query = query.Where("qg_payroll.pay_year = ?", year)
	}
	if month > 0 {
		query = query.Where("qg_payroll.pay_month = ?", month)
	}
	if status != "" {
		query = query.Where("qg_payroll.status = ?", status)
	}
	if positionTitle != "" {
		query = query.
			Joins("JOIN qg_position_apply pa ON pa.id = qg_payroll.apply_id").
			Joins("JOIN qg_position p ON p.id = pa.position_id").
			Where("p.title LIKE ?", "%"+positionTitle+"%")
	}

	var total int64
	if err := query.Model(&models.QgPayroll{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var payrolls []models.QgPayroll
	offset := (page - 1) * pageSize
	if err := query.Order("qg_payroll.created_at DESC").Offset(offset).Limit(pageSize).Find(&payrolls).Error; err != nil {
		return nil, 0, err
	}

	return payrolls, total, nil
}

// UpdatePayroll 更新薪酬记录。
func (r *AssessmentRepository) UpdatePayroll(payroll *models.QgPayroll) error {
	return r.db.Save(payroll).Error
}

// ---- 辅助 ----

// GetApplyByID 查询岗位申请记录。
func (r *AssessmentRepository) GetApplyByID(id int64) (*models.QgPositionApply, error) {
	var apply models.QgPositionApply
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&apply).Error; err != nil {
		return nil, err
	}
	return &apply, nil
}

// GetPositionByID 查询岗位信息。
func (r *AssessmentRepository) GetPositionByID(id int64) (*models.QgPosition, error) {
	var pos models.QgPosition
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&pos).Error; err != nil {
		return nil, err
	}
	return &pos, nil
}

// GetStudentByID 查询学生信息。
func (r *AssessmentRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// ListAttendancesByApplyAndMonth 查询指定申请指定月份的工时记录。
func (r *AssessmentRepository) ListAttendancesByApplyAndMonth(applyID int64, year, month int) ([]models.QgAttendance, error) {
	var records []models.QgAttendance
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)
	if err := r.db.Where("apply_id = ? AND work_date >= ? AND work_date < ? AND is_deleted = 0", applyID, startDate, endDate).
		Order("work_date ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

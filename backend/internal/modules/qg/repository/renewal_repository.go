package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// RenewalRepository 续聘/解聘+申诉数据访问层。
type RenewalRepository struct {
	db *gorm.DB
}

// NewRenewalRepository 创建续聘/解聘+申诉仓储。
func NewRenewalRepository(db *gorm.DB) *RenewalRepository {
	return &RenewalRepository{db: db}
}

// ---- 续聘/解聘 ----

// CreateRenewal 创建续聘/解聘记录。
func (r *RenewalRepository) CreateRenewal(renewal *models.QgRenewalTerm) error {
	return r.db.Create(renewal).Error
}

// GetRenewalByID 按 ID 查询续聘/解聘记录。
func (r *RenewalRepository) GetRenewalByID(id int64) (*models.QgRenewalTerm, error) {
	var renewal models.QgRenewalTerm
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&renewal).Error; err != nil {
		return nil, err
	}
	return &renewal, nil
}

// UpdateRenewal 更新续聘/解聘记录。
func (r *RenewalRepository) UpdateRenewal(renewal *models.QgRenewalTerm) error {
	return r.db.Save(renewal).Error
}

// ListRenewals 分页查询续聘/解聘列表。
func (r *RenewalRepository) ListRenewals(applyID int64, renewalType string, page, pageSize int) ([]models.QgRenewalTerm, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if applyID > 0 {
		query = query.Where("apply_id = ?", applyID)
	}
	if renewalType != "" {
		query = query.Where("type = ?", renewalType)
	}

	var total int64
	if err := query.Model(&models.QgRenewalTerm{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var renewals []models.QgRenewalTerm
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&renewals).Error; err != nil {
		return nil, 0, err
	}

	return renewals, total, nil
}

// ---- 申诉 ----

// CreateComplaint 创建申诉记录。
func (r *RenewalRepository) CreateComplaint(complaint *models.QgComplaint) error {
	return r.db.Create(complaint).Error
}

// GetComplaintByID 按 ID 查询申诉记录。
func (r *RenewalRepository) GetComplaintByID(id int64) (*models.QgComplaint, error) {
	var complaint models.QgComplaint
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&complaint).Error; err != nil {
		return nil, err
	}
	return &complaint, nil
}

// UpdateComplaint 更新申诉记录。
func (r *RenewalRepository) UpdateComplaint(complaint *models.QgComplaint) error {
	return r.db.Save(complaint).Error
}

// ListComplaints 分页查询申诉列表。
func (r *RenewalRepository) ListComplaints(studentID int64, targetType, status string, page, pageSize int) ([]models.QgComplaint, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if studentID > 0 {
		query = query.Where("student_id = ?", studentID)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Model(&models.QgComplaint{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var complaints []models.QgComplaint
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&complaints).Error; err != nil {
		return nil, 0, err
	}

	return complaints, total, nil
}

// ---- 辅助 ----

// GetApplyByID 查询岗位申请记录。
func (r *RenewalRepository) GetApplyByID(id int64) (*models.QgPositionApply, error) {
	var apply models.QgPositionApply
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&apply).Error; err != nil {
		return nil, err
	}
	return &apply, nil
}

// GetPositionByID 查询岗位信息。
func (r *RenewalRepository) GetPositionByID(id int64) (*models.QgPosition, error) {
	var pos models.QgPosition
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&pos).Error; err != nil {
		return nil, err
	}
	return &pos, nil
}

// GetStudentByID 查询学生信息。
func (r *RenewalRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetUserByID 查询用户信息。
func (r *RenewalRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

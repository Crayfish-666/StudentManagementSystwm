package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// AssociationRepository 社团数据访问层。
type AssociationRepository struct {
	db *gorm.DB
}

// NewAssociationRepository 创建社团仓储。
func NewAssociationRepository(db *gorm.DB) *AssociationRepository {
	return &AssociationRepository{db: db}
}

// List 分页查询社团列表。
func (r *AssociationRepository) List(status string, collegeID int64, keyword string, page, pageSize int) ([]models.StAssociation, int64, error) {
	query := r.db.Where("is_deleted = 0")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if collegeID > 0 {
		query = query.Where("college_id = ?", collegeID)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Model(&models.StAssociation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var assocs []models.StAssociation
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&assocs).Error; err != nil {
		return nil, 0, err
	}

	return assocs, total, nil
}

// GetByID 按 ID 查询社团。
func (r *AssociationRepository) GetByID(id int64) (*models.StAssociation, error) {
	var assoc models.StAssociation
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&assoc).Error; err != nil {
		return nil, err
	}
	return &assoc, nil
}

// Create 创建社团。
func (r *AssociationRepository) Create(assoc *models.StAssociation) error {
	return r.db.Create(assoc).Error
}

// Update 更新社团。
func (r *AssociationRepository) Update(assoc *models.StAssociation) error {
	return r.db.Save(assoc).Error
}

// SoftDelete 软删除社团。
func (r *AssociationRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.StAssociation{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// CountByName 按名称统计社团数（同名检测）。
func (r *AssociationRepository) CountByName(name string, excludeID int64) (int64, error) {
	var count int64
	query := r.db.Model(&models.StAssociation{}).Where("name = ? AND is_deleted = 0", name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByTutor 统计指导教师当前指导的活跃社团数。
func (r *AssociationRepository) CountByTutor(tutorUserID int64) (int64, error) {
	var count int64
	if err := r.db.Model(&models.StAssociation{}).
		Where("tutor_user_id = ? AND is_deleted = 0 AND status NOT IN ('cancelled')", tutorUserID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ---- 发起人 ----

// CreateFounder 创建发起人记录。
func (r *AssociationRepository) CreateFounder(founder *models.StFounder) error {
	return r.db.Create(founder).Error
}

// ListFoundersByAssoc 查询社团所有发起人。
func (r *AssociationRepository) ListFoundersByAssoc(associationID int64) ([]models.StFounder, error) {
	var founders []models.StFounder
	if err := r.db.Where("association_id = ? AND is_deleted = 0", associationID).Find(&founders).Error; err != nil {
		return nil, err
	}
	return founders, nil
}

// ---- 成员 ----

// ListMembersByAssoc 查询社团成员列表。
func (r *AssociationRepository) ListMembersByAssoc(associationID int64) ([]models.StAssocMember, error) {
	var members []models.StAssocMember
	if err := r.db.Where("association_id = ? AND is_deleted = 0", associationID).
		Order("role ASC, joined_at ASC").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// CreateMember 创建成员记录。
func (r *AssociationRepository) CreateMember(member *models.StAssocMember) error {
	return r.db.Create(member).Error
}

// ---- 辅助 ----

// GetUserByID 查询用户信息。
func (r *AssociationRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetStudentByID 查询学生信息。
func (r *AssociationRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// ListColleges 查询所有院系。
func (r *AssociationRepository) ListColleges() ([]models.SysCollege, error) {
	var colleges []models.SysCollege
	if err := r.db.Where("is_deleted = 0").Order("id ASC").Find(&colleges).Error; err != nil {
		return nil, err
	}
	return colleges, nil
}

// ListUsers 查询用户列表（用于指导教师下拉）。
func (r *AssociationRepository) ListUsers() ([]models.SysUser, error) {
	var users []models.SysUser
	if err := r.db.Where("is_deleted = 0").Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

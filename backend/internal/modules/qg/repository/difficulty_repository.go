package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// DifficultyRepository 困难认定数据访问层。
type DifficultyRepository struct {
	db *gorm.DB
}

// NewDifficultyRepository 创建困难认定仓储。
func NewDifficultyRepository(db *gorm.DB) *DifficultyRepository {
	return &DifficultyRepository{db: db}
}

// List 分页查询困难认定列表。
func (r *DifficultyRepository) List(level, status string, studentID int64, page, pageSize int) ([]models.QgDifficultyCert, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if studentID > 0 {
		query = query.Where("student_id = ?", studentID)
	}

	var total int64
	if err := query.Model(&models.QgDifficultyCert{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var certs []models.QgDifficultyCert
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&certs).Error; err != nil {
		return nil, 0, err
	}

	return certs, total, nil
}

// GetByID 按 ID 查询困难认定。
func (r *DifficultyRepository) GetByID(id int64) (*models.QgDifficultyCert, error) {
	var cert models.QgDifficultyCert
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// Create 创建困难认定。
func (r *DifficultyRepository) Create(cert *models.QgDifficultyCert) error {
	return r.db.Create(cert).Error
}

// Update 更新困难认定。
func (r *DifficultyRepository) Update(cert *models.QgDifficultyCert) error {
	return r.db.Save(cert).Error
}

// SoftDelete 软删除困难认定。
func (r *DifficultyRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.QgDifficultyCert{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// CountByStudentAndYear 统计指定学生指定学年的认定次数。
func (r *DifficultyRepository) CountByStudentAndYear(studentID int64, academicYear string) (int64, error) {
	var count int64
	if err := r.db.Model(&models.QgDifficultyCert{}).
		Where("student_id = ? AND academic_year = ? AND is_deleted = 0 AND status != 'S4'", studentID, academicYear).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetStudentByID 查询学生信息。
func (r *DifficultyRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetUserByID 查询用户信息。
func (r *DifficultyRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

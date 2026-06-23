package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// PositionRepository 岗位+申请数据访问层。
type PositionRepository struct {
	db *gorm.DB
}

// NewPositionRepository 创建岗位仓储。
func NewPositionRepository(db *gorm.DB) *PositionRepository {
	return &PositionRepository{db: db}
}

// ---- 岗位 ----

// ListPositions 分页查询岗位列表。
// keyword：对部门名称 dept_name 做 LIKE '%keyword%' 模糊匹配。
func (r *PositionRepository) ListPositions(keyword, deptType, status string, page, pageSize int) ([]models.QgPosition, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if keyword != "" {
		query = query.Where("dept_name LIKE ?", "%"+keyword+"%")
	}
	if deptType != "" {
		query = query.Where("dept_type = ?", deptType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Model(&models.QgPosition{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var positions []models.QgPosition
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&positions).Error; err != nil {
		return nil, 0, err
	}

	return positions, total, nil
}

// GetPositionByID 按 ID 查询岗位。
func (r *PositionRepository) GetPositionByID(id int64) (*models.QgPosition, error) {
	var pos models.QgPosition
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&pos).Error; err != nil {
		return nil, err
	}
	return &pos, nil
}

// CreatePosition 创建岗位。
func (r *PositionRepository) CreatePosition(pos *models.QgPosition) error {
	return r.db.Create(pos).Error
}

// UpdatePosition 更新岗位。
func (r *PositionRepository) UpdatePosition(pos *models.QgPosition) error {
	return r.db.Save(pos).Error
}

// SoftDeletePosition 软删除岗位。
func (r *PositionRepository) SoftDeletePosition(id int64) error {
	return r.db.Model(&models.QgPosition{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// ---- 申请 ----

// GetApplyByID 按 ID 查询岗位申请。
func (r *PositionRepository) GetApplyByID(id int64) (*models.QgPositionApply, error) {
	var apply models.QgPositionApply
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&apply).Error; err != nil {
		return nil, err
	}
	return &apply, nil
}

// CreateApply 创建岗位申请。
func (r *PositionRepository) CreateApply(apply *models.QgPositionApply) error {
	return r.db.Create(apply).Error
}

// UpdateApply 更新岗位申请。
func (r *PositionRepository) UpdateApply(apply *models.QgPositionApply) error {
	return r.db.Save(apply).Error
}

// ExistsApplyByPositionAndStudent 检查学生是否已申请某岗位。
func (r *PositionRepository) ExistsApplyByPositionAndStudent(positionID, studentID int64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.QgPositionApply{}).
		Where("position_id = ? AND student_id = ? AND is_deleted = 0 AND apply_status NOT IN ('rejected','abandoned','expired')", positionID, studentID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountOnJobByStudent 统计学生在岗数量。
func (r *PositionRepository) CountOnJobByStudent(studentID int64) (int64, error) {
	var count int64
	if err := r.db.Model(&models.QgPositionApply{}).
		Where("student_id = ? AND status = 'on_job' AND is_deleted = 0", studentID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetActiveDifficultyCert 获取学生有效的困难认定。
func (r *PositionRepository) GetActiveDifficultyCert(studentID int64) (*models.QgDifficultyCert, error) {
	var cert models.QgDifficultyCert
	if err := r.db.Where("student_id = ? AND status = 'S3' AND is_deleted = 0", studentID).
		Order("created_at DESC").First(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// GetStudentByID 查询学生信息。
func (r *PositionRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetUserByID 查询用户信息。
func (r *PositionRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

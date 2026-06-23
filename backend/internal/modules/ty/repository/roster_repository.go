package repository

import (
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// RosterRepository 团员花名册数据访问层。
type RosterRepository struct {
	db *gorm.DB
}

// NewRosterRepository 创建团员花名册仓储。
func NewRosterRepository(db *gorm.DB) *RosterRepository {
	return &RosterRepository{db: db}
}

// Create 创建团员花名册记录。
func (r *RosterRepository) Create(roster *models.TyMemberRoster) error {
	return r.db.Create(roster).Error
}

// GetByID 按ID查询团员花名册。
func (r *RosterRepository) GetByID(id int64) (*models.TyMemberRoster, error) {
	var roster models.TyMemberRoster
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&roster).Error; err != nil {
		return nil, err
	}
	return &roster, nil
}

// GetByStudentID 按学生ID查询团员花名册（取 active 状态）。
func (r *RosterRepository) GetByStudentID(studentID int64) (*models.TyMemberRoster, error) {
	var roster models.TyMemberRoster
	if err := r.db.Where("student_id = ? AND status = 'active' AND is_deleted = 0", studentID).
		First(&roster).Error; err != nil {
		return nil, err
	}
	return &roster, nil
}

// Update 更新团员花名册记录。
func (r *RosterRepository) Update(roster *models.TyMemberRoster) error {
	return r.db.Save(roster).Error
}

// List 列表查询团员花名册，支持按支部/状态/关键字筛选和分页。
//
// 参数：
//   - branchID: 支部ID（0=不限）
//   - status: 状态过滤（空=不限）
//   - keyword: 关键字搜索（匹配学生姓名/学号）
//   - page, pageSize: 分页参数
func (r *RosterRepository) List(branchID int64, status string, keyword string, page, pageSize int) ([]models.TyMemberRoster, int64, error) {
	query := r.db.Where("is_deleted = 0")

	if branchID > 0 {
		query = query.Where("branch_id = ?", branchID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("student_id IN (SELECT id FROM idx_student WHERE (name LIKE ? OR student_no LIKE ?) AND is_deleted = 0)",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := query.Model(&models.TyMemberRoster{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.TyMemberRoster
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CheckMemberNoExists 检查团员证号是否已存在（用于唯一性校验）。
func (r *RosterRepository) CheckMemberNoExists(memberNo string, excludeID int64) (bool, error) {
	var count int64
	query := r.db.Model(&models.TyMemberRoster{}).
		Where("member_no = ? AND is_deleted = 0", memberNo)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// TransferOut 团员转出操作。
func (r *RosterRepository) TransferOut(id int64, transferredAt *time.Time) error {
	now := transferredAt
	if now == nil {
		t := time.Now()
		now = &t
	}
	return r.db.Model(&models.TyMemberRoster{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(map[string]interface{}{
			"status":         "transferred",
			"transferred_at": now,
		}).Error
}

// Overtime 超龄离团操作（BR-TY-03）。
func (r *RosterRepository) Overtime(id int64) error {
	return r.db.Model(&models.TyMemberRoster{}).
		Where("id = ? AND is_deleted = 0", id).
		Update("is_overtime", 1).Error
}

// Archive 归档操作（BR-TY-04，保留5年）。
func (r *RosterRepository) Archive(id int64, keepUntil interface{}) error {
	return r.db.Model(&models.TyMemberRoster{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(map[string]interface{}{
			"status":              "archived",
			"archive_keep_until":  keepUntil,
		}).Error
}

// GetStudentByID 查询学生信息（用于关联显示）。
func (r *RosterRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetBranchByID 查询团支部信息（用于关联显示）。
func (r *RosterRepository) GetBranchByID(id int64) (*models.TyBranch, error) {
	var branch models.TyBranch
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

// CountByStatus 按状态统计团员数量。
func (r *RosterRepository) CountByStatus(status string) (int64, error) {
	var count int64
	query := r.db.Model(&models.TyMemberRoster{}).Where("is_deleted = 0")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

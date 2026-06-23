package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// ActivityRepository 活动数据访问层。
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository 创建活动仓储。
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// List 分页查询活动列表。
func (r *ActivityRepository) List(associationID int64, status string, page, pageSize int) ([]models.StActivity, int64, error) {
	query := r.db.Where("is_deleted = 0")

	if associationID > 0 {
		query = query.Where("association_id = ?", associationID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Model(&models.StActivity{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var acts []models.StActivity
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&acts).Error; err != nil {
		return nil, 0, err
	}

	return acts, total, nil
}

// GetByID 按 ID 查询活动。
func (r *ActivityRepository) GetByID(id int64) (*models.StActivity, error) {
	var act models.StActivity
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&act).Error; err != nil {
		return nil, err
	}
	return &act, nil
}

// Create 创建活动。
func (r *ActivityRepository) Create(act *models.StActivity) error {
	return r.db.Create(act).Error
}

// Update 更新活动。
func (r *ActivityRepository) Update(act *models.StActivity) error {
	return r.db.Save(act).Error
}

// SoftDelete 软删除活动。
func (r *ActivityRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.StActivity{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// GetAssociationByID 查询社团信息。
func (r *ActivityRepository) GetAssociationByID(id int64) (*models.StAssociation, error) {
	var assoc models.StAssociation
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&assoc).Error; err != nil {
		return nil, err
	}
	return &assoc, nil
}

// GetUserByID 查询用户信息。
func (r *ActivityRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetStudentByID 查询学生信息。
func (r *ActivityRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// ListColleges 查询所有院系。
func (r *ActivityRepository) ListColleges() ([]models.SysCollege, error) {
	var colleges []models.SysCollege
	if err := r.db.Where("is_deleted = 0").Order("id ASC").Find(&colleges).Error; err != nil {
		return nil, err
	}
	return colleges, nil
}

// ---- 审批记录相关 ----

// CreateApproval 创建审批记录。
func (r *ActivityRepository) CreateApproval(rec *models.StActivityApproval) error {
	return r.db.Create(rec).Error
}

// ListApprovals 查询活动的审批记录列表。
func (r *ActivityRepository) ListApprovals(activityID int64) ([]models.StActivityApproval, error) {
	var records []models.StActivityApproval
	if err := r.db.Where("activity_id = ?", activityID).Order("step_no ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// HasApprovedStep 检查指定步骤是否已通过。
func (r *ActivityRepository) HasApprovedStep(activityID int64, stepNo int) (bool, error) {
	var count int64
	if err := r.db.Model(&models.StActivityApproval{}).
		Where("activity_id = ? AND step_no = ? AND decision = ?", activityID, stepNo, "pass").
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetApprovalByStep 查询指定步骤的审批记录。
func (r *ActivityRepository) GetApprovalByStep(activityID int64, stepNo int) (*models.StActivityApproval, error) {
	var rec models.StActivityApproval
	if err := r.db.Where("activity_id = ? AND step_no = ?", activityID, stepNo).First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

// ---- 签到相关 ----

// CreateCheckin 创建签到记录。
func (r *ActivityRepository) CreateCheckin(rec *models.StActivityCheckin) error {
	return r.db.Create(rec).Error
}

// ListCheckins 分页查询签到列表。
func (r *ActivityRepository) ListCheckins(activityID int64, page, pageSize int) ([]models.StActivityCheckin, int64, error) {
	query := r.db.Where("activity_id = ?", activityID)

	var total int64
	if err := query.Model(&models.StActivityCheckin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []models.StActivityCheckin
	offset := (page - 1) * pageSize
	if err := query.Order("checkin_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// HasCheckin 检查学生是否已签到。
func (r *ActivityRepository) HasCheckin(activityID, studentID int64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.StActivityCheckin{}).
		Where("activity_id = ? AND student_id = ?", activityID, studentID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetCheckinCount 获取签到人数。
func (r *ActivityRepository) GetCheckinCount(activityID int64) (int64, error) {
	var count int64
	if err := r.db.Model(&models.StActivityCheckin{}).
		Where("activity_id = ? AND is_present = 1", activityID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ---- 总结相关 ----

// CreateSummary 创建活动总结。
func (r *ActivityRepository) CreateSummary(rec *models.StActivitySummary) error {
	return r.db.Create(rec).Error
}

// GetSummaryByActivity 根据活动 ID 查询总结。
func (r *ActivityRepository) GetSummaryByActivity(activityID int64) (*models.StActivitySummary, error) {
	var rec models.StActivitySummary
	if err := r.db.Where("activity_id = ? AND is_deleted = 0", activityID).First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

// CreatePhoto 创建活动照片记录。
func (r *ActivityRepository) CreatePhoto(rec *models.StActivityPhoto) error {
	return r.db.Create(rec).Error
}

// ListPhotosByActivity 查询活动照片列表。
func (r *ActivityRepository) ListPhotosByActivity(activityID int64) ([]models.StActivityPhoto, error) {
	var photos []models.StActivityPhoto
	if err := r.db.Where("activity_id = ? AND is_deleted = 0", activityID).Order("id ASC").Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}

// ---- 用户角色相关 ----

// FindUserRoles 查询用户角色码列表。
func (r *ActivityRepository) FindUserRoles(userID int64) ([]string, error) {
	type row struct {
		Code string
	}
	var rows []row
	if err := r.db.Table("sys_user_role AS ur").
		Select("r.code AS code").
		Joins("JOIN sys_role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND ur.is_deleted = 0 AND r.is_deleted = 0", userID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(rows))
	for _, x := range rows {
		codes = append(codes, x.Code)
	}
	return codes, nil
}

// FindUserScopeCollegeIDs 查询用户作用域内院系 ID 列表。
func (r *ActivityRepository) FindUserScopeCollegeIDs(userID int64) ([]int64, error) {
	type row struct {
		CollegeID int64
	}
	var rows []row
	if err := r.db.Table("sys_user_role AS ur").
		Select("ur.scope_college_id AS college_id").
		Where("ur.user_id = ? AND ur.is_deleted = 0 AND ur.scope_college_id IS NOT NULL", userID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, x := range rows {
		ids = append(ids, x.CollegeID)
	}
	return ids, nil
}

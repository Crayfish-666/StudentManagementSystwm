// Package repository 提供 ST 模块招新相关数据访问。
//
// 设计依据：docs/03 §6.2.5 st_recruit_plan / §6.2.6 st_recruit_apply。
package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// RecruitRepository 招新数据访问层。
type RecruitRepository struct {
	db *gorm.DB
}

// NewRecruitRepository 创建招新仓储。
func NewRecruitRepository(db *gorm.DB) *RecruitRepository {
	return &RecruitRepository{db: db}
}

// ---- 招新计划 ----

// ListPlans 分页查询招新计划，支持按社团、状态、学年过滤。
func (r *RecruitRepository) ListPlans(associationID int64, status, academicYear string, page, pageSize int) ([]models.StRecruitPlan, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if associationID > 0 {
		query = query.Where("association_id = ?", associationID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if academicYear != "" {
		query = query.Where("academic_year = ?", academicYear)
	}

	var total int64
	if err := query.Model(&models.StRecruitPlan{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var plans []models.StRecruitPlan
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&plans).Error; err != nil {
		return nil, 0, err
	}
	return plans, total, nil
}

// GetPlanByID 按 ID 查询招新计划。
func (r *RecruitRepository) GetPlanByID(id int64) (*models.StRecruitPlan, error) {
	var plan models.StRecruitPlan
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

// CreatePlan 创建招新计划。
func (r *RecruitRepository) CreatePlan(plan *models.StRecruitPlan) error {
	return r.db.Create(plan).Error
}

// UpdatePlan 更新招新计划。
func (r *RecruitRepository) UpdatePlan(plan *models.StRecruitPlan) error {
	return r.db.Save(plan).Error
}

// CountActivePlansByAssoc 统计社团当前有效计划数（用于业务校验）。
func (r *RecruitRepository) CountActivePlansByAssoc(associationID int64) (int64, error) {
	var count int64
	if err := r.db.Model(&models.StRecruitPlan{}).
		Where("association_id = ? AND is_deleted = 0 AND status IN ('S0','S1','S3')", associationID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ---- 招新申请 ----

// ListApplies 分页查询招新申请。
func (r *RecruitRepository) ListApplies(planID, studentID int64, result string, page, pageSize int) ([]models.StRecruitApply, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if planID > 0 {
		query = query.Where("plan_id = ?", planID)
	}
	if studentID > 0 {
		query = query.Where("student_id = ?", studentID)
	}
	if result != "" {
		query = query.Where("result = ?", result)
	}

	var total int64
	if err := query.Model(&models.StRecruitApply{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var apps []models.StRecruitApply
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

// GetApplyByID 按 ID 查询招新申请。
func (r *RecruitRepository) GetApplyByID(id int64) (*models.StRecruitApply, error) {
	var app models.StRecruitApply
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// ListPlansByIDs 按 ID 列表批量查询招新计划（用于申请列表预加载）。
func (r *RecruitRepository) ListPlansByIDs(ids []int64) ([]models.StRecruitPlan, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var plans []models.StRecruitPlan
	if err := r.db.Where("id IN ? AND is_deleted = 0", ids).Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

// ListAssociationsByIDs 按 ID 列表批量查询社团（用于申请列表预加载）。
func (r *RecruitRepository) ListAssociationsByIDs(ids []int64) (map[int64]*models.StAssociation, error) {
	if len(ids) == 0 {
		return map[int64]*models.StAssociation{}, nil
	}
	var assocs []models.StAssociation
	if err := r.db.Where("id IN ? AND is_deleted = 0", ids).Find(&assocs).Error; err != nil {
		return nil, err
	}
	m := make(map[int64]*models.StAssociation, len(assocs))
	for i := range assocs {
		m[assocs[i].ID] = &assocs[i]
	}
	return m, nil
}

// CreateApply 创建招新申请。
func (r *RecruitRepository) CreateApply(app *models.StRecruitApply) error {
	return r.db.Create(app).Error
}

// UpdateApply 更新招新申请。
func (r *RecruitRepository) UpdateApply(app *models.StRecruitApply) error {
	return r.db.Save(app).Error
}

// HasApplyInPlan 检查同一学生同一计划是否已投递（唯一约束兜底）。
func (r *RecruitRepository) HasApplyInPlan(planID, studentID int64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.StRecruitApply{}).
		Where("plan_id = ? AND student_id = ? AND is_deleted = 0", planID, studentID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountAcceptedAssociationsInYear 统计学生同一学年已接受的社团数（硬卡控 3）。
// 依据：docs/01 §5.3.4 "单一学生同一学年最多加入 3 个社团"。
func (r *RecruitRepository) CountAcceptedAssociationsInYear(studentID int64, academicYear string) (int64, error) {
	var count int64
	if err := r.db.Table("st_recruit_apply AS ra").
		Joins("JOIN st_recruit_plan rp ON rp.id = ra.plan_id AND rp.is_deleted = 0").
		Where("ra.student_id = ? AND ra.is_deleted = 0 AND ra.result = ? AND rp.academic_year = ?",
			studentID, "accepted", academicYear).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ---- 用户角色与成员 ----

// FindUserRoles 查询用户角色码列表。
func (r *RecruitRepository) FindUserRoles(userID int64) ([]string, error) {
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

// GetAssociationByID 查询社团信息。
func (r *RecruitRepository) GetAssociationByID(id int64) (*models.StAssociation, error) {
	var assoc models.StAssociation
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&assoc).Error; err != nil {
		return nil, err
	}
	return &assoc, nil
}

// GetStudentByID 查询学生信息。
func (r *RecruitRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetUserByID 查询用户信息。
func (r *RecruitRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

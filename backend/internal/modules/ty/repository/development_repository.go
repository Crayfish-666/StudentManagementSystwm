package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// DevelopmentObjectRepository 发展对象数据访问层。
type DevelopmentObjectRepository struct {
	db *gorm.DB
}

// NewDevelopmentObjectRepository 创建发展对象仓储。
func NewDevelopmentObjectRepository(db *gorm.DB) *DevelopmentObjectRepository {
	return &DevelopmentObjectRepository{db: db}
}

// Create 创建发展对象记录。
func (r *DevelopmentObjectRepository) Create(obj *models.TyDevelopmentObject) error {
	return r.db.Create(obj).Error
}

// GetByApplicationID 按申请ID查询发展对象（1:1关系）。
func (r *DevelopmentObjectRepository) GetByApplicationID(applicationID int64) (*models.TyDevelopmentObject, error) {
	var obj models.TyDevelopmentObject
	if err := r.db.Where("application_id = ? AND is_deleted = 0", applicationID).First(&obj).Error; err != nil {
		return nil, err
	}
	return &obj, nil
}

// GetByID 按ID查询发展对象。
func (r *DevelopmentObjectRepository) GetByID(id int64) (*models.TyDevelopmentObject, error) {
	var obj models.TyDevelopmentObject
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&obj).Error; err != nil {
		return nil, err
	}
	return &obj, nil
}

// Update 更新发展对象记录。
func (r *DevelopmentObjectRepository) Update(obj *models.TyDevelopmentObject) error {
	return r.db.Save(obj).Error
}

// UpdateStatus 仅更新状态字段（用于状态流转）。
func (r *DevelopmentObjectRepository) UpdateStatus(id int64, status string) error {
	return r.db.Model(&models.TyDevelopmentObject{}).
		Where("id = ? AND is_deleted = 0", id).
		Update("status", status).Error
}

// List 列表查询，支持按状态/院系筛选和分页。
func (r *DevelopmentObjectRepository) List(status string, collegeID int64, page, pageSize int) ([]models.TyDevelopmentObject, int64, error) {
	query := r.db.Where("is_deleted = 0")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if collegeID > 0 {
		query = query.Where("application_id IN (SELECT id FROM ty_application WHERE branch_id IN (SELECT id FROM ty_branch WHERE college_id = ?))", collegeID)
	}

	var total int64
	if err := query.Model(&models.TyDevelopmentObject{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.TyDevelopmentObject
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetApplicationByID 查询入团申请详情。
func (r *DevelopmentObjectRepository) GetApplicationByID(id int64) (*models.TyApplication, error) {
	var app models.TyApplication
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// HasPassedRecommendation 检查入团申请是否已通过推优大会（存在 pass 决策的推优会议）。
func (r *DevelopmentObjectRepository) HasPassedRecommendation(applicationID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.TyRecommendationMeeting{}).
		Where("application_id = ? AND decision = 'pass' AND is_deleted = 0", applicationID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

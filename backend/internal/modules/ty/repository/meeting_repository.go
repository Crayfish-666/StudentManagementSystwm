package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// DevelopmentMeetingRepository 发展大会数据访问层。
type DevelopmentMeetingRepository struct {
	db *gorm.DB
}

// NewDevelopmentMeetingRepository 创建发展大会仓储。
func NewDevelopmentMeetingRepository(db *gorm.DB) *DevelopmentMeetingRepository {
	return &DevelopmentMeetingRepository{db: db}
}

// Create 创建发展大会记录。
func (r *DevelopmentMeetingRepository) Create(meeting *models.TyDevelopmentMeeting) error {
	return r.db.Create(meeting).Error
}

// GetByID 按ID查询发展大会。
func (r *DevelopmentMeetingRepository) GetByID(id int64) (*models.TyDevelopmentMeeting, error) {
	var meeting models.TyDevelopmentMeeting
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&meeting).Error; err != nil {
		return nil, err
	}
	return &meeting, nil
}

// ListByDevelopmentID 按发展对象ID查询所有发展大会记录。
func (r *DevelopmentMeetingRepository) ListByDevelopmentID(developmentID int64) ([]models.TyDevelopmentMeeting, error) {
	var meetings []models.TyDevelopmentMeeting
	if err := r.db.Where("development_id = ? AND is_deleted = 0", developmentID).
		Order("id DESC").
		Find(&meetings).Error; err != nil {
		return nil, err
	}
	return meetings, nil
}

// Update 更新发展大会记录。
func (r *DevelopmentMeetingRepository) Update(meeting *models.TyDevelopmentMeeting) error {
	return r.db.Save(meeting).Error
}

// CreateMemberRoster 创建团员花名册记录（事务内调用）。
func (r *DevelopmentMeetingRepository) CreateMemberRoster(roster *models.TyMemberRoster) error {
	return r.db.Create(roster).Error
}

// GetApplicationByID 查询入团申请（用于更新状态）。
func (r *DevelopmentMeetingRepository) GetApplicationByID(id int64) (*models.TyApplication, error) {
	var app models.TyApplication
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// UpdateApplicationStatus 更新入团申请状态（事务内调用）。
func (r *DevelopmentMeetingRepository) UpdateApplicationStatus(applicationID int64, status string) error {
	return r.db.Model(&models.TyApplication{}).
		Where("id = ? AND is_deleted = 0", applicationID).
		Update("status", status).Error
}

// GetStudentByID 查询学生信息（用于更新政治面貌）。
func (r *DevelopmentMeetingRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// UpdateStudentPoliticalStatus 更新学生政治面貌（事务内调用）。
func (r *DevelopmentMeetingRepository) UpdateStudentPoliticalStatus(studentID int64, status string) error {
	return r.db.Model(&models.IdxStudent{}).
		Where("id = ? AND is_deleted = 0", studentID).
		Update("political_status", status).Error
}

// List 列表查询发展大会，支持分页。
func (r *DevelopmentMeetingRepository) List(page, pageSize int) ([]models.TyDevelopmentMeeting, int64, error) {
	query := r.db.Where("is_deleted = 0")

	var total int64
	if err := query.Model(&models.TyDevelopmentMeeting{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.TyDevelopmentMeeting
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

package repository

import (
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// RecommendationRepository 推优大会数据访问层。
type RecommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository 创建推优大会仓储。
func NewRecommendationRepository(db *gorm.DB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

// CreateMeeting 创建推优大会记录。
func (r *RecommendationRepository) CreateMeeting(meeting *models.TyRecommendationMeeting) error {
	return r.db.Create(meeting).Error
}

// CreateVote 创建投票记录。
func (r *RecommendationRepository) CreateVote(vote *models.TyRecommendationVote) error {
	return r.db.Create(vote).Error
}

// GetByApplicationID 按申请ID查询推优大会（含关联投票）。
func (r *RecommendationRepository) GetByApplicationID(applicationID int64) (*models.TyRecommendationMeeting, *models.TyRecommendationVote, error) {
	var meeting models.TyRecommendationMeeting
	if err := r.db.Where("application_id = ? AND is_deleted = 0", applicationID).First(&meeting).Error; err != nil {
		return nil, nil, err
	}

	var vote models.TyRecommendationVote
	voteErr := r.db.Where("meeting_id = ? AND application_id = ?", meeting.ID, applicationID).First(&vote).Error
	if voteErr != nil && voteErr != gorm.ErrRecordNotFound {
		return nil, nil, voteErr
	}
	if voteErr == gorm.ErrRecordNotFound {
		return &meeting, nil, nil
	}
	return &meeting, &vote, nil
}

// ListByBranchID 按团支部查询推优大会列表（分页）。
func (r *RecommendationRepository) ListByBranchID(branchID int64, page, pageSize int) ([]models.TyRecommendationMeeting, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if branchID > 0 {
		// 通过申请单的 branch_id 过滤
		query = query.Where("application_id IN (SELECT id FROM ty_application WHERE branch_id = ? AND is_deleted = 0)", branchID)
	}

	var total int64
	if err := query.Model(&models.TyRecommendationMeeting{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var meetings []models.TyRecommendationMeeting
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&meetings).Error; err != nil {
		return nil, 0, err
	}

	return meetings, total, nil
}

// CheckRecentRecommendation 检查同一申请人3个月内是否已推优过。
// 返回 true 表示存在近期推优记录，不允许再次推优。
func (r *RecommendationRepository) CheckRecentRecommendation(applicationID int64) (bool, error) {
	// 先查该申请对应的 student_id
	var app models.TyApplication
	if err := r.db.Where("id = ? AND is_deleted = 0", applicationID).First(&app).Error; err != nil {
		return false, err
	}

	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	var count int64
	err := r.db.Model(&models.TyRecommendationMeeting{}).
		Joins("JOIN ty_application ON ty_application.id = ty_recommendation_meeting.application_id").
		Where("ty_application.student_id = ? AND ty_recommendation_meeting.is_deleted = 0 AND ty_recommendation_meeting.meeting_at >= ?", app.StudentID, threeMonthsAgo).
		Count(&count).Error

	return count > 0, err
}

// Update 更新推优大会记录。
func (r *RecommendationRepository) Update(meeting *models.TyRecommendationMeeting) error {
	return r.db.Save(meeting).Error
}

// GetByID 按 ID 查询推优大会。
func (r *RecommendationRepository) GetByID(id int64) (*models.TyRecommendationMeeting, error) {
	var meeting models.TyRecommendationMeeting
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&meeting).Error; err != nil {
		return nil, err
	}
	return &meeting, nil
}

// GetVotesByMeetingID 获取某推优大会的投票信息。
func (r *RecommendationRepository) GetVotesByMeetingID(meetingID int64) ([]models.TyRecommendationVote, error) {
	var votes []models.TyRecommendationVote
	if err := r.db.Where("meeting_id = ?", meetingID).Find(&votes).Error; err != nil {
		return nil, err
	}
	return votes, nil
}

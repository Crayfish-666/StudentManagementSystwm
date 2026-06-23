package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// ProbationaryRepository 预备期/转正数据访问层。
type ProbationaryRepository struct {
	db *gorm.DB
}

// NewProbationaryRepository 创建预备期仓储。
func NewProbationaryRepository(db *gorm.DB) *ProbationaryRepository {
	return &ProbationaryRepository{db: db}
}

// ---- 预备期考察记录 ----

// CreateProbationaryRecord 创建预备期考察记录。
func (r *ProbationaryRepository) CreateProbationaryRecord(record *models.TyProbationaryRecord) error {
	return r.db.Create(record).Error
}

// GetProbationaryRecordByID 按ID查询预备期考察记录。
func (r *ProbationaryRepository) GetProbationaryRecordByID(id int64) (*models.TyProbationaryRecord, error) {
	var record models.TyProbationaryRecord
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ListProbationaryRecordsByApplicationID 按申请ID查询所有考察记录。
func (r *ProbationaryRepository) ListProbationaryRecordsByApplicationID(applicationID int64) ([]models.TyProbationaryRecord, error) {
	var records []models.TyProbationaryRecord
	if err := r.db.Where("application_id = ? AND is_deleted = 0", applicationID).
		Order("record_year ASC, record_quarter ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// CheckQuarterlyRecordExists 检查某季度是否已存在考察记录（唯一性约束）。
func (r *ProbationaryRepository) CheckQuarterlyRecordExists(applicationID int64, year, quarter int) (bool, error) {
	var count int64
	err := r.db.Model(&models.TyProbationaryRecord{}).
		Where("application_id = ? AND record_year = ? AND record_quarter = ? AND is_deleted = 0",
			applicationID, year, quarter).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ---- 转正大会 ----

// CreateProbationaryMeeting 创建转正大会记录。
func (r *ProbationaryRepository) CreateProbationaryMeeting(meeting *models.TyProbationaryMeeting) error {
	return r.db.Create(meeting).Error
}

// GetProbationaryMeetingByID 按ID查询转正大会。
func (r *ProbationaryRepository) GetProbationaryMeetingByID(id int64) (*models.TyProbationaryMeeting, error) {
	var meeting models.TyProbationaryMeeting
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&meeting).Error; err != nil {
		return nil, err
	}
	return &meeting, nil
}

// ListProbationaryMeetingsByApplicationID 按申请ID查询转正大会列表。
func (r *ProbationaryRepository) ListProbationaryMeetingsByApplicationID(applicationID int64) ([]models.TyProbationaryMeeting, error) {
	var meetings []models.TyProbationaryMeeting
	if err := r.db.Where("application_id = ? AND is_deleted = 0", applicationID).
		Order("id DESC").
		Find(&meetings).Error; err != nil {
		return nil, err
	}
	return meetings, nil
}

// UpdateProbationaryMeeting 更新转正大会记录。
func (r *ProbationaryRepository) UpdateProbationaryMeeting(meeting *models.TyProbationaryMeeting) error {
	return r.db.Save(meeting).Error
}

// GetMemberRosterByStudentID 按学生ID查询团员花名册（用于检查预备期满）。
func (r *ProbationaryRepository) GetMemberRosterByStudentID(studentID int64) (*models.TyMemberRoster, error) {
	var roster models.TyMemberRoster
	if err := r.db.Where("student_id = ? AND status = 'active' AND is_deleted = 0", studentID).
		First(&roster).Error; err != nil {
		return nil, err
	}
	return &roster, nil
}

// GetMemberRosterByID 按ID查询团员花名册。
func (r *ProbationaryRepository) GetMemberRosterByID(id int64) (*models.TyMemberRoster, error) {
	var roster models.TyMemberRoster
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&roster).Error; err != nil {
		return nil, err
	}
	return &roster, nil
}

// UpdateMemberRoster 更新团员花名册记录（事务内调用）。
func (r *ProbationaryRepository) UpdateMemberRoster(roster *models.TyMemberRoster) error {
	return r.db.Save(roster).Error
}

// UpdateStudentPoliticalStatus 更新学生政治面貌（事务内调用）。
func (r *ProbationaryRepository) UpdateStudentPoliticalStatus(studentID int64, status string) error {
	return r.db.Model(&models.IdxStudent{}).
		Where("id = ? AND is_deleted = 0", studentID).
		Update("political_status", status).Error
}

// ListProbationaryRecords 列表查询预备期考察记录，支持分页与按申请ID过滤。
// applicationID 为 nil 时查询全部；非 nil 时按 application_id 过滤。
func (r *ProbationaryRepository) ListProbationaryRecords(applicationID *int64, page, pageSize int) ([]models.TyProbationaryRecord, int64, error) {
	query := r.db.Model(&models.TyProbationaryRecord{}).Where("is_deleted = 0")
	if applicationID != nil {
		query = query.Where("application_id = ?", *applicationID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.TyProbationaryRecord
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ListProbationaryMeetings 列表查询转正大会，支持分页与按申请ID过滤。
// applicationID 为 nil 时查询全部；非 nil 时按 application_id 过滤。
func (r *ProbationaryRepository) ListProbationaryMeetings(applicationID *int64, page, pageSize int) ([]models.TyProbationaryMeeting, int64, error) {
	query := r.db.Model(&models.TyProbationaryMeeting{}).Where("is_deleted = 0")
	if applicationID != nil {
		query = query.Where("application_id = ?", *applicationID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.TyProbationaryMeeting
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

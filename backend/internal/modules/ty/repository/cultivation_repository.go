package repository

import (
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// CultivationRepository 培养考察数据访问层。
type CultivationRepository struct {
	db *gorm.DB
}

// NewCultivationRepository 创建培养考察仓储。
func NewCultivationRepository(db *gorm.DB) *CultivationRepository {
	return &CultivationRepository{db: db}
}

// ---- 培养联系人 ----

// CreateLink 创建培养联系人记录。
func (r *CultivationRepository) CreateLink(link *models.TyCultivationLink) error {
	return r.db.Create(link).Error
}

// CreateLinksBulk 批量创建培养联系人记录（一次事务，避免半成功）。
// 用于「一次提交 2 位」的业务场景（PRD §4.3.4）。
func (r *CultivationRepository) CreateLinksBulk(links []models.TyCultivationLink) error {
	if len(links) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&links).Error
	})
}

// CountActiveLinks 统计某申请下当前在任的培养联系人数（含 is_deleted=0）。
func (r *CultivationRepository) CountActiveLinks(applicationID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.TyCultivationLink{}).
		Where("application_id = ? AND is_active = 1 AND is_deleted = 0", applicationID).
		Count(&count).Error
	return count, err
}

// CheckMentorInActiveLinks 检查 mentor_student_id 是否已是该申请的在任联系人之一。
func (r *CultivationRepository) CheckMentorInActiveLinks(applicationID, mentorStudentID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.TyCultivationLink{}).
		Where("application_id = ? AND mentor_student_id = ? AND is_active = 1 AND is_deleted = 0",
			applicationID, mentorStudentID).
		Count(&count).Error
	return count > 0, err
}

// CountBranchMembers 统计某团支部下的"在册正式团员"人数（status='active'）。
// 用于「培养联系人优先从支部团员选择」的硬卡控（PRD §4.3.4）。
func (r *CultivationRepository) CountBranchMembers(branchID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.TyMemberRoster{}).
		Where("branch_id = ? AND status = 'active' AND is_deleted = 0", branchID).
		Count(&count).Error
	return count, err
}

// UpdateLink 更新培养联系人记录。
func (r *CultivationRepository) UpdateLink(link *models.TyCultivationLink) error {
	return r.db.Save(link).Error
}

// GetLinkByID 按 ID 查询培养联系人。
func (r *CultivationRepository) GetLinkByID(id int64) (*models.TyCultivationLink, error) {
	var link models.TyCultivationLink
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

// GetActiveLink 获取当前活跃的培养联系人（按申请ID）。
func (r *CultivationRepository) GetActiveLink(applicationID int64) (*models.TyCultivationLink, error) {
	var link models.TyCultivationLink
	if err := r.db.Where("application_id = ? AND is_active = 1 AND is_deleted = 0", applicationID).
		Order("id DESC").First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

// ListLinks 按申请ID查询培养联系人列表。
func (r *CultivationRepository) ListLinks(applicationID int64) ([]models.TyCultivationLink, error) {
	var links []models.TyCultivationLink
	if err := r.db.Where("application_id = ? AND is_deleted = 0", applicationID).
		Order("id ASC").Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

// EndMentor 结束培养关系（设置 end_at 和 is_active=0）。
func (r *CultivationRepository) EndMentor(id int64) error {
	now := time.Now()
	return r.db.Model(&models.TyCultivationLink{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"end_at":    now,
			"is_active": 0,
		}).Error
}

// ---- 培养记录 ----

// CreateRecord 创建培养记录。
func (r *CultivationRepository) CreateRecord(record *models.TyCultivationRecord) error {
	return r.db.Create(record).Error
}

// GetRecordByID 按 ID 查询培养记录。
func (r *CultivationRepository) GetRecordByID(id int64) (*models.TyCultivationRecord, error) {
	var record models.TyCultivationRecord
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ListRecords 按申请ID查询培养记录列表（分页）。
func (r *CultivationRepository) ListRecords(applicationID int64, page, pageSize int) ([]models.TyCultivationRecord, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if applicationID > 0 {
		query = query.Where("application_id = ?", applicationID)
	}

	var total int64
	if err := query.Model(&models.TyCultivationRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []models.TyCultivationRecord
	offset := (page - 1) * pageSize
	if err := query.Order("record_year DESC, record_month DESC, id DESC").
		Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetMonthlyRecords 获取某月培养记录。
func (r *CultivationRepository) GetMonthlyRecords(applicationID int64, year, month int) (*models.TyCultivationRecord, error) {
	var record models.TyCultivationRecord
	if err := r.db.Where("application_id = ? AND record_year = ? AND record_month = ? AND is_deleted = 0",
		applicationID, year, month).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// CheckMonthlyRecordExists 检查当月记录是否存在。
func (r *CultivationRepository) CheckMonthlyRecordExists(applicationID int64, year, month int) (bool, error) {
	var count int64
	err := r.db.Model(&models.TyCultivationRecord{}).
		Where("application_id = ? AND record_year = ? AND record_month = ? AND is_deleted = 0",
			applicationID, year, month).
		Count(&count).Error
	return count > 0, err
}

// ---- 团课记录 ----

// CreateCourse 创建团课记录。
func (r *CultivationRepository) CreateCourse(course *models.TyCourseRecord) error {
	return r.db.Create(course).Error
}

// GetCourseByID 按 ID 查询团课记录。
func (r *CultivationRepository) GetCourseByID(id int64) (*models.TyCourseRecord, error) {
	var course models.TyCourseRecord
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// ListCourses 按学生ID查询团课列表（分页）。
func (r *CultivationRepository) ListCourses(studentID int64, page, pageSize int) ([]models.TyCourseRecord, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if studentID > 0 {
		query = query.Where("student_id = ?", studentID)
	}

	var total int64
	if err := query.Model(&models.TyCourseRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var courses []models.TyCourseRecord
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&courses).Error; err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

// UpdateCoursePassStatus 更新团课结业状态。
func (r *CultivationRepository) UpdateCoursePassStatus(id int64, isPass int) error {
	return r.db.Model(&models.TyCourseRecord{}).
		Where("id = ?", id).
		Update("is_pass", isPass).Error
}

// ---- 思想汇报 ----

// CreateReport 创建思想汇报。
func (r *CultivationRepository) CreateReport(report *models.TyThoughtReport) error {
	return r.db.Create(report).Error
}

// GetReportByID 按 ID 查询思想汇报。
func (r *CultivationRepository) GetReportByID(id int64) (*models.TyThoughtReport, error) {
	var report models.TyThoughtReport
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReports 按申请ID查询思想汇报列表（分页）。
// 支持按 studentIDs（学生白名单）和 majorIDs（学生所属专业白名单）进行数据范围隔离。
func (r *CultivationRepository) ListReports(applicationID int64, studentIDs, majorIDs []int64, page, pageSize int) ([]models.TyThoughtReport, int64, error) {
	query := r.db.Where("is_deleted = 0")
	if applicationID > 0 {
		query = query.Where("application_id = ?", applicationID)
	}
	if len(studentIDs) > 0 {
		query = query.Where("student_id IN ?", studentIDs)
	}
	if len(majorIDs) > 0 {
		// 过滤学生所属专业
		query = query.Where("student_id IN (SELECT id FROM idx_student WHERE major_id IN ? AND is_deleted = 0)", majorIDs)
	}

	var total int64
	if err := query.Model(&models.TyThoughtReport{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var reports []models.TyThoughtReport
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetQuarterlyReports 获取某季度思想汇报。
func (r *CultivationRepository) GetQuarterlyReports(applicationID int64, quarter string) (*models.TyThoughtReport, error) {
	var report models.TyThoughtReport
	if err := r.db.Where("application_id = ? AND quarter = ? AND is_deleted = 0",
		applicationID, quarter).First(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// CheckQuarterlyReportExists 检查当季度思想汇报是否已提交。
func (r *CultivationRepository) CheckQuarterlyReportExists(applicationID int64, quarter string) (bool, error) {
	var count int64
	err := r.db.Model(&models.TyThoughtReport{}).
		Where("application_id = ? AND quarter = ? AND is_deleted = 0", applicationID, quarter).
		Count(&count).Error
	return count > 0, err
}

// GetStudentByID 查询学生信息。
func (r *CultivationRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetUserByID 查询用户（用于回填记录人姓名）。
func (r *CultivationRepository) GetUserByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetApplicationByID 查询入团申请。
func (r *CultivationRepository) GetApplicationByID(id int64) (*models.TyApplication, error) {
	var app models.TyApplication
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

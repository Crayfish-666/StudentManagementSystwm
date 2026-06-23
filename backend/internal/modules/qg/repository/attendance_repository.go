package repository

import (
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// AttendanceRepository 工时打卡数据访问层。
type AttendanceRepository struct {
	db *gorm.DB
}

// NewAttendanceRepository 创建工时打卡仓储。
func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

// List 分页查询工时打卡列表。
// positionTitle 非空时按岗位标题模糊匹配（join qg_position_apply + qg_position）。
func (r *AttendanceRepository) List(applyID, studentID int64, positionTitle string, dateFrom, dateTo *time.Time, page, pageSize int) ([]models.QgAttendance, int64, error) {
	query := r.db.Where("qg_attendance.is_deleted = 0")
	if applyID > 0 {
		query = query.Where("qg_attendance.apply_id = ?", applyID)
	}
	if studentID > 0 {
		query = query.Where("qg_attendance.student_id = ?", studentID)
	}
	if positionTitle != "" {
		query = query.
			Joins("JOIN qg_position_apply pa ON pa.id = qg_attendance.apply_id").
			Joins("JOIN qg_position p ON p.id = pa.position_id").
			Where("p.title LIKE ?", "%"+positionTitle+"%")
	}
	if dateFrom != nil {
		query = query.Where("work_date >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("work_date <= ?", *dateTo)
	}

	var total int64
	if err := query.Model(&models.QgAttendance{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []models.QgAttendance
	offset := (page - 1) * pageSize
	if err := query.Order("qg_attendance.work_date DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetByID 按 ID 查询打卡记录。
func (r *AttendanceRepository) GetByID(id int64) (*models.QgAttendance, error) {
	var rec models.QgAttendance
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

// Create 创建打卡记录。
func (r *AttendanceRepository) Create(rec *models.QgAttendance) error {
	return r.db.Create(rec).Error
}

// Update 更新打卡记录。
func (r *AttendanceRepository) Update(rec *models.QgAttendance) error {
	return r.db.Save(rec).Error
}

// SoftDelete 软删除打卡记录。
func (r *AttendanceRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.QgAttendance{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// ExistsByApplyAndDate 检查指定申请指定日期是否已有打卡记录。
func (r *AttendanceRepository) ExistsByApplyAndDate(applyID int64, workDate time.Time) (bool, error) {
	var count int64
	if err := r.db.Model(&models.QgAttendance{}).
		Where("apply_id = ? AND work_date = ? AND is_deleted = 0", applyID, workDate.Format("2006-01-02")).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// SumMonthlyHours 汇总指定学生指定月份的工时。
func (r *AttendanceRepository) SumMonthlyHours(studentID int64, year, month int) (float64, error) {
	var total float64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)
	if err := r.db.Model(&models.QgAttendance{}).
		Where("student_id = ? AND work_date >= ? AND work_date < ? AND is_deleted = 0", studentID, startDate, endDate).
		Select("COALESCE(SUM(effective_hours), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// SumWeeklyHours 汇总指定学生指定日期所在周的工时。
func (r *AttendanceRepository) SumWeeklyHours(studentID int64, date time.Time) (float64, error) {
	var total float64
	// 计算周一
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := date.AddDate(0, 0, -(weekday - 1))
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.Local)
	sunday := monday.AddDate(0, 0, 7)
	if err := r.db.Model(&models.QgAttendance{}).
		Where("student_id = ? AND work_date >= ? AND work_date < ? AND is_deleted = 0", studentID, monday, sunday).
		Select("COALESCE(SUM(effective_hours), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// SumDailyHours 汇总指定学生指定日期的工时。
func (r *AttendanceRepository) SumDailyHours(studentID int64, date time.Time) (float64, error) {
	var total float64
	dateStr := date.Format("2006-01-02")
	if err := r.db.Model(&models.QgAttendance{}).
		Where("student_id = ? AND work_date = ? AND is_deleted = 0", studentID, dateStr).
		Select("COALESCE(SUM(effective_hours), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// MonthlySummary 月度工时汇总。
func (r *AttendanceRepository) MonthlySummary(studentID int64, year, month int) (float64, int, error) {
	var totalHours float64
	var recordCount int64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	if err := r.db.Model(&models.QgAttendance{}).
		Where("student_id = ? AND work_date >= ? AND work_date < ? AND is_deleted = 0", studentID, startDate, endDate).
		Count(&recordCount).Error; err != nil {
		return 0, 0, err
	}

	if err := r.db.Model(&models.QgAttendance{}).
		Where("student_id = ? AND work_date >= ? AND work_date < ? AND is_deleted = 0", studentID, startDate, endDate).
		Select("COALESCE(SUM(effective_hours), 0)").
		Scan(&totalHours).Error; err != nil {
		return 0, 0, err
	}

	return totalHours, int(recordCount), nil
}

// GetApplyByID 查询岗位申请记录。
func (r *AttendanceRepository) GetApplyByID(id int64) (*models.QgPositionApply, error) {
	var apply models.QgPositionApply
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&apply).Error; err != nil {
		return nil, err
	}
	return &apply, nil
}

// GetPositionByID 查询岗位信息。
func (r *AttendanceRepository) GetPositionByID(id int64) (*models.QgPosition, error) {
	var pos models.QgPosition
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&pos).Error; err != nil {
		return nil, err
	}
	return &pos, nil
}

// GetStudentByID 查询学生信息。
func (r *AttendanceRepository) GetStudentByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

package service

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/qg/repository"
)

// AttendanceService 工时打卡业务服务层。
type AttendanceService struct {
	repo *repository.AttendanceRepository
	db   *gorm.DB
}

// NewAttendanceService 创建工时打卡服务。
func NewAttendanceService(repo *repository.AttendanceRepository, db *gorm.DB) *AttendanceService {
	return &AttendanceService{repo: repo, db: db}
}

// ---- DTO ----

// AttendanceListResult 工时打卡列表结果。
type AttendanceListResult struct {
	Items    []AttendanceView `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// AttendanceView 工时打卡视图。
type AttendanceView struct {
	ID              int64   `json:"id"`
	BizNo           string  `json:"biz_no"`
	ApplyID         int64   `json:"apply_id"`
	PositionTitle   string  `json:"position_title"`
	StudentID       int64   `json:"student_id"`
	StudentName     string  `json:"student_name"`
	WorkDate        string  `json:"work_date"`
	ClockInAt       *string `json:"clock_in_at,omitempty"`
	ClockOutAt      *string `json:"clock_out_at,omitempty"`
	EffectiveHours  float64 `json:"effective_hours"`
	LateMinutes     int     `json:"late_minutes"`
	EarlyMinutes    int     `json:"early_minutes"`
	ClockMethod     string  `json:"clock_method"`
	IP              string  `json:"ip"`
	Geo             string  `json:"geo"`
	IsMakeup        int     `json:"is_makeup"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// ClockInRequest 上班打卡请求。
type ClockInRequest struct {
	ApplyID     int64  `json:"apply_id" binding:"required"`
	WorkDate    string `json:"work_date" binding:"required"`
	ClockMethod string `json:"clock_method" binding:"required"`
	IP          string `json:"ip"`
	Geo         string `json:"geo"`
}

// ClockOutRequest 下班打卡请求。
type ClockOutRequest struct {
	// 无需额外字段，通过路径参数 id 定位记录
}

// MonthlySummaryView 月度工时汇总视图。
type MonthlySummaryView struct {
	StudentID      int64   `json:"student_id"`
	StudentName    string  `json:"student_name"`
	Year           int     `json:"year"`
	Month          int     `json:"month"`
	TotalHours     float64 `json:"total_hours"`
	RecordCount    int     `json:"record_count"`
	PositionTitle  string  `json:"position_title,omitempty"`
}

// ---- 业务方法 ----

// List 分页查询工时打卡列表。
// studentKeyword 支持学号 / 姓名 / 数字主键三合一查询;空字符串表示不过滤。
func (s *AttendanceService) List(applyID, studentID int64, studentKeyword, positionTitle, dateFrom, dateTo string, page, pageSize int) (*AttendanceListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var fromTime, toTime *time.Time
	if dateFrom != "" {
		t, err := parseTime(dateFrom)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误")
		}
		fromTime = &t
	}
	if dateTo != "" {
		t, err := parseTime(dateTo)
		if err != nil {
			return nil, fmt.Errorf("结束日期格式错误")
		}
		// 转换为半开区间：dateTo 自动 +1 天（次日 00:00:00），
		// 配合仓储层 work_date <= 条件，可正确包含 dateTo 当天全量数据。
		// 这样即使历史种子数据 work_date 带有时间部分（如 2026-06-23 01:12:19），
		// 也能被当天的范围搜索命中。
		t = t.AddDate(0, 0, 1)
		toTime = &t
	}

	records, total, err := s.repo.List(applyID, studentID, studentKeyword, positionTitle, fromTime, toTime, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]AttendanceView, 0, len(records))
	for _, rec := range records {
		v := s.toView(rec)
		items = append(items, v)
	}

	return &AttendanceListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取工时打卡详情。
func (s *AttendanceService) Get(id int64) (*AttendanceView, error) {
	rec, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("打卡记录不存在")
	}

	v := s.toView(*rec)
	return &v, nil
}

// ClockIn 上班打卡。
func (s *AttendanceService) ClockIn(userID, studentID int64, req *ClockInRequest) (*AttendanceView, error) {
	// 必须在岗（apply status=on_job）
	apply, err := s.repo.GetApplyByID(req.ApplyID)
	if err != nil {
		return nil, fmt.Errorf("岗位申请记录不存在")
	}
	if apply.Status != "on_job" {
		return nil, fmt.Errorf("当前不在岗，无法打卡")
	}
	if apply.StudentID != studentID {
		return nil, fmt.Errorf("申请记录与学生不匹配")
	}

	// 解析工作日期
	workDate, err := parseTime(req.WorkDate)
	if err != nil {
		return nil, fmt.Errorf("工作日期格式错误")
	}
	workDate = time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 0, 0, 0, 0, workDate.Location())

	// 每日1条记录约束
	exists, err := s.repo.ExistsByApplyAndDate(req.ApplyID, workDate)
	if err != nil {
		return nil, fmt.Errorf("查询打卡记录失败: %w", err)
	}
	if exists {
		return nil, &BizError{Code: 40903, Msg: "当日已有打卡记录"}
	}

	// 核心硬卡控：月≤40h, 周≤20h, 日≤8h
	now := time.Now()

	// 检查月工时
	monthHours, err := s.repo.SumMonthlyHours(studentID, workDate.Year(), int(workDate.Month()))
	if err != nil {
		return nil, fmt.Errorf("查询月工时失败: %w", err)
	}
	if monthHours >= 40 {
		return nil, &BizError{Code: 5401, Msg: "月工时已达上限40小时"}
	}

	// 检查周工时
	weekHours, err := s.repo.SumWeeklyHours(studentID, workDate)
	if err != nil {
		return nil, fmt.Errorf("查询周工时失败: %w", err)
	}
	if weekHours >= 20 {
		return nil, &BizError{Code: 5401, Msg: "周工时已达上限20小时"}
	}

	// 检查日工时（当日已有记录的工时）
	dayHours, err := s.repo.SumDailyHours(studentID, workDate)
	if err != nil {
		return nil, fmt.Errorf("查询日工时失败: %w", err)
	}
	if dayHours >= 8 {
		return nil, &BizError{Code: 5401, Msg: "日工时已达上限8小时"}
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "QG")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	record := &models.QgAttendance{
		BizNo:       bizNo,
		ApplyID:     req.ApplyID,
		StudentID:   studentID,
		WorkDate:    workDate,
		ClockInAt:   &now,
		ClockMethod: req.ClockMethod,
		IP:          req.IP,
		Geo:         req.Geo,
	}

	if err := s.repo.Create(record); err != nil {
		return nil, err
	}

	return s.Get(record.ID)
}

// ClockOut 下班打卡（计算 effective_hours）。
func (s *AttendanceService) ClockOut(id int64) (*AttendanceView, error) {
	rec, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("打卡记录不存在")
	}

	if rec.ClockInAt == nil {
		return nil, fmt.Errorf("未上班打卡，无法下班打卡")
	}
	if rec.ClockOutAt != nil {
		return nil, fmt.Errorf("已下班打卡，不可重复操作")
	}

	now := time.Now()
	rec.ClockOutAt = &now

	// 计算有效工时（小时）
	duration := now.Sub(*rec.ClockInAt).Hours()
	// 扣除迟到/早退时间后四舍五入到0.5小时
	effectiveHours := math.Round(duration*2) / 2
	if effectiveHours < 0 {
		effectiveHours = 0
	}

	// 日工时上限8小时
	dayHours, _ := s.repo.SumDailyHours(rec.StudentID, rec.WorkDate)
	remaining := 8.0 - dayHours
	if remaining < 0 {
		remaining = 0
	}
	if effectiveHours > remaining {
		effectiveHours = remaining
	}

	rec.EffectiveHours = effectiveHours

	if err := s.repo.Update(rec); err != nil {
		return nil, err
	}

	return s.Get(id)
}

// MonthlySummary 月度工时汇总。
func (s *AttendanceService) MonthlySummary(studentID int64, year, month int) (*MonthlySummaryView, error) {
	totalHours, recordCount, err := s.repo.MonthlySummary(studentID, year, month)
	if err != nil {
		return nil, err
	}

	v := &MonthlySummaryView{
		StudentID:   studentID,
		Year:        year,
		Month:       month,
		TotalHours:  totalHours,
		RecordCount: recordCount,
	}

	// 加载学生姓名
	if student, err := s.repo.GetStudentByID(studentID); err == nil {
		v.StudentName = student.Name
	}

	return v, nil
}

// Delete 软删除打卡记录。
func (s *AttendanceService) Delete(id int64) error {
	return s.repo.SoftDelete(id)
}

// ---- 内部方法 ----

func (s *AttendanceService) toView(rec models.QgAttendance) AttendanceView {
	v := AttendanceView{
		ID:             rec.ID,
		BizNo:          rec.BizNo,
		ApplyID:        rec.ApplyID,
		StudentID:      rec.StudentID,
		WorkDate:       rec.WorkDate.Format("2006-01-02T15:04:05+08:00"),
		EffectiveHours: rec.EffectiveHours,
		LateMinutes:    rec.LateMinutes,
		EarlyMinutes:   rec.EarlyMinutes,
		ClockMethod:    rec.ClockMethod,
		IP:             rec.IP,
		Geo:            rec.Geo,
		IsMakeup:       rec.IsMakeup,
		CreatedAt:      rec.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:      rec.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	if rec.ClockInAt != nil {
		t := rec.ClockInAt.Format("2006-01-02T15:04:05+08:00")
		v.ClockInAt = &t
	}
	if rec.ClockOutAt != nil {
		t := rec.ClockOutAt.Format("2006-01-02T15:04:05+08:00")
		v.ClockOutAt = &t
	}

	// 加载岗位标题
	if apply, err := s.repo.GetApplyByID(rec.ApplyID); err == nil {
		if pos, err := s.repo.GetPositionByID(apply.PositionID); err == nil {
			v.PositionTitle = pos.Title
		}
	}

	// 加载学生姓名
	if student, err := s.repo.GetStudentByID(rec.StudentID); err == nil {
		v.StudentName = student.Name
	}

	return v
}

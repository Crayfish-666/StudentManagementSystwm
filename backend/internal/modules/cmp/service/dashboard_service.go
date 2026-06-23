// Package cmp 综合素质 dashboard KPI / 趋势 / 分布服务。
package service

import (
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// DashboardService 综合看板服务。
type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService 创建看板服务。
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// KPIView 关键 KPI。
type KPIView struct {
	ActiveAssoc          int64   `json:"active_assoc"`            // 活跃社团数（status=registered/trial）
	TyPassRate           float64 `json:"ty_pass_rate"`            // 推优通过率（S3 / 已审结）
	L4Incidents30d       int64   `json:"l4_incidents_30d"`        // 近 30 天 L4 事件
	QgPayrollAmountCents int64   `json:"qg_payroll_amount_cents"` // 当月薪酬总额（分）
	ExcellentCount       int64   `json:"excellent_count"`         // 综合分≥85 人数
	StudentCount         int64   `json:"student_count"`           // 在校学生数
	TotalScored          int64   `json:"total_scored"`            // 已计算综合分人数
}

// TrendPoint 趋势点。
type TrendPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// DistributionBucket 分布桶。
type DistributionBucket struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// KPI 关键 KPI。
func (s *DashboardService) KPI(academicYear string) (*KPIView, error) {
	if academicYear == "" {
		academicYear = CurrentAcademicYear()
	}

	view := &KPIView{}

	// 1. 活跃社团数
	if err := s.db.Model(&models.StAssociation{}).
		Where("is_deleted = 0 AND status IN ('registered','trial')").
		Count(&view.ActiveAssoc).Error; err != nil {
		return nil, err
	}

	// 2. 推优通过率
	var tyPassed, tyTotal int64
	s.db.Model(&models.TyApplication{}).Where("is_deleted = 0 AND status = 'S3'").Count(&tyPassed)
	s.db.Model(&models.TyApplication{}).Where("is_deleted = 0 AND status IN ('S3','S4')").Count(&tyTotal)
	if tyTotal > 0 {
		view.TyPassRate = float64(tyPassed) / float64(tyTotal)
	}

	// 3. L4 事件 30 天
	cutoff := time.Now().AddDate(0, 0, -30)
	if err := s.db.Model(&models.SqIncident{}).
		Where("is_deleted = 0 AND incident_level = 'L4' AND occurred_at >= ?", cutoff).
		Count(&view.L4Incidents30d).Error; err != nil {
		return nil, err
	}

	// 4. 当月薪酬总额
	now := time.Now()
	year, month := now.Year(), int(now.Month())
	if err := s.db.Model(&models.QgPayroll{}).
		Select("COALESCE(SUM(net_cents), 0)").
		Where("is_deleted = 0 AND pay_year = ? AND pay_month = ?", year, month).
		Scan(&view.QgPayrollAmountCents).Error; err != nil {
		return nil, err
	}

	// 5. 评优合格（综合分 ≥ 85）
	if err := s.db.Model(&models.CmpScore{}).
		Where("is_deleted = 0 AND academic_year = ? AND total_score >= 85", academicYear).
		Count(&view.ExcellentCount).Error; err != nil {
		return nil, err
	}

	// 6. 在校学生数
	if err := s.db.Model(&models.IdxStudent{}).
		Where("is_deleted = 0 AND status = 'enrolled'").
		Count(&view.StudentCount).Error; err != nil {
		return nil, err
	}

	// 7. 已计算综合分人数
	if err := s.db.Model(&models.CmpScore{}).
		Where("is_deleted = 0 AND academic_year = ?", academicYear).
		Count(&view.TotalScored).Error; err != nil {
		return nil, err
	}

	return view, nil
}

// Trends 趋势图：metric ∈ ty_pass_rate | l4_incidents | qg_payroll | excellent_count | cmp_avg。
func (s *DashboardService) Trends(metric string, rangeKey string) ([]TrendPoint, error) {
	now := time.Now()
	var months int
	switch rangeKey {
	case "6m":
		months = 6
	case "12m":
		months = 12
	case "3m":
		months = 3
	default:
		months = 12
	}

	points := make([]TrendPoint, 0, months)
	for i := months - 1; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		year, month := t.Year(), int(t.Month())
		label := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local).Format("2006-01")
		var v float64
		switch metric {
		case "ty_pass_rate":
			var passed, total int64
			s.db.Model(&models.TyApplication{}).
				Where("is_deleted = 0 AND status = 'S3' AND strftime('%Y', updated_at) = ? AND strftime('%m', updated_at) = ?",
					itoa(year), pad2(month)).
				Count(&passed)
			s.db.Model(&models.TyApplication{}).
				Where("is_deleted = 0 AND status IN ('S3','S4') AND strftime('%Y', updated_at) = ? AND strftime('%m', updated_at) = ?",
					itoa(year), pad2(month)).
				Count(&total)
			if total > 0 {
				v = float64(passed) / float64(total)
			}
		case "l4_incidents":
			var c int64
			s.db.Model(&models.SqIncident{}).
				Where("is_deleted = 0 AND incident_level = 'L4' AND strftime('%Y', occurred_at) = ? AND strftime('%m', occurred_at) = ?",
					itoa(year), pad2(month)).
				Count(&c)
			v = float64(c)
		case "qg_payroll":
			var amt int64
			s.db.Model(&models.QgPayroll{}).
				Select("COALESCE(SUM(net_cents), 0)").
				Where("is_deleted = 0 AND pay_year = ? AND pay_month = ?", year, month).
				Scan(&amt)
			v = float64(amt) / 100 // 转为元
		case "excellent_count":
			var c int64
			academicYear := itoa(year) + "-" + itoa(year+1)
			s.db.Model(&models.CmpScore{}).
				Where("is_deleted = 0 AND academic_year = ? AND total_score >= 85", academicYear).
				Count(&c)
			v = float64(c)
		case "cmp_avg":
			academicYear := itoa(year) + "-" + itoa(year+1)
			var avg float64
			s.db.Model(&models.CmpScore{}).
				Select("COALESCE(AVG(total_score), 0)").
				Where("is_deleted = 0 AND academic_year = ?", academicYear).
				Scan(&avg)
			v = avg
		default:
			v = 0
		}
		points = append(points, TrendPoint{Label: label, Value: v})
	}
	return points, nil
}

// Distribution 分布图：dim ∈ college | gender | grade。
func (s *DashboardService) Distribution(dim string, academicYear string) ([]DistributionBucket, error) {
	if academicYear == "" {
		academicYear = CurrentAcademicYear()
	}

	buckets := make([]DistributionBucket, 0, 8)
	switch dim {
	case "college":
		var cols []models.SysCollege
		s.db.Select("id, name").Where("is_deleted = 0").Find(&cols)
		for _, c := range cols {
			var count int64
			s.db.Model(&models.CmpScore{}).
				Where("is_deleted = 0 AND academic_year = ? AND student_id IN (SELECT id FROM idx_student WHERE college_id = ? AND is_deleted = 0)",
					academicYear, c.ID).
				Count(&count)
			buckets = append(buckets, DistributionBucket{Label: c.Name, Value: count})
		}
	case "gender":
		for _, g := range []struct {
			Code  string
			Label string
		}{
			{"M", "男"},
			{"F", "女"},
		} {
			var count int64
			s.db.Model(&models.CmpScore{}).
				Where("is_deleted = 0 AND academic_year = ? AND student_id IN (SELECT id FROM idx_student WHERE gender = ? AND is_deleted = 0)",
					academicYear, g.Code).
				Count(&count)
			buckets = append(buckets, DistributionBucket{Label: g.Label, Value: count})
		}
	case "grade":
		type row struct {
			Grade  *int
			Count  int64
		}
		var rows []row
		s.db.Raw(`
			SELECT s.grade as grade, COUNT(*) as count
			FROM cmp_score c
			INNER JOIN idx_student s ON s.id = c.student_id AND s.is_deleted = 0
			WHERE c.is_deleted = 0 AND c.academic_year = ?
			GROUP BY s.grade
			ORDER BY s.grade
		`, academicYear).Scan(&rows)
		for _, r := range rows {
			label := "未知年级"
			if r.Grade != nil {
				label = itoa(*r.Grade) + "级"
			}
			buckets = append(buckets, DistributionBucket{Label: label, Value: r.Count})
		}
	case "score_range":
		// 分数段分布：<60, 60-70, 70-80, 80-90, ≥90
		for _, b := range []struct {
			Label string
			Min   float64
			Max   float64
		}{
			{"<60", 0, 60},
			{"60-70", 60, 70},
			{"70-80", 70, 80},
			{"80-90", 80, 90},
			{"≥90", 90, 1000},
		} {
			var c int64
			s.db.Model(&models.CmpScore{}).
				Where("is_deleted = 0 AND academic_year = ? AND total_score >= ? AND total_score < ?",
					academicYear, b.Min, b.Max).
				Count(&c)
			buckets = append(buckets, DistributionBucket{Label: b.Label, Value: c})
		}
	}
	return buckets, nil
}

// ActiveAssocByCollege 各院系活跃社团数（柱状图）。
func (s *DashboardService) ActiveAssocByCollege() ([]DistributionBucket, error) {
	type row struct {
		CollegeID int64
		Count     int64
	}
	var rows []row
	s.db.Raw(`
		SELECT s.college_id as college_id, COUNT(*) as count
		FROM st_association s
		INNER JOIN sys_college c ON c.id = s.college_id AND c.is_deleted = 0
		WHERE s.is_deleted = 0 AND s.status IN ('registered','trial')
		GROUP BY s.college_id
		ORDER BY count DESC
		LIMIT 20
	`).Scan(&rows)

	buckets := make([]DistributionBucket, 0, len(rows))
	for _, r := range rows {
		var col models.SysCollege
		if err := s.db.Select("name").Where("id = ?", r.CollegeID).First(&col).Error; err != nil {
			continue
		}
		buckets = append(buckets, DistributionBucket{Label: col.Name, Value: r.Count})
	}
	return buckets, nil
}

// IncidentLevelDistribution 事件等级分布（饼图）。
func (s *DashboardService) IncidentLevelDistribution() ([]DistributionBucket, error) {
	type row struct {
		Level string
		Count int64
	}
	var rows []row
	s.db.Model(&models.SqIncident{}).
		Select("incident_level as level, count(*) as count").
		Where("is_deleted = 0").
		Group("incident_level").
		Scan(&rows)

	buckets := make([]DistributionBucket, 0, len(rows))
	for _, r := range rows {
		label := r.Level
		switch r.Level {
		case "L1":
			label = "L1-常规报修"
		case "L2":
			label = "L2-一般违规"
		case "L3":
			label = "L3-严重违规"
		case "L4":
			label = "L4-紧急事件"
		}
		buckets = append(buckets, DistributionBucket{Label: label, Value: r.Count})
	}
	return buckets, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

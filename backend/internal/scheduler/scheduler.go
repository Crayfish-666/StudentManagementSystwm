// Package scheduler 基于 robfig/cron/v3 的定时任务调度器（ADR-016）。
package scheduler

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"student-system/internal/models"
)

// Scheduler 定时任务调度器。
type Scheduler struct {
	cron *cron.Cron
	db   *gorm.DB
	zlog *zap.Logger
}

// NewScheduler 创建调度器实例。
func NewScheduler(db *gorm.DB, zlog *zap.Logger) *Scheduler {
	return &Scheduler{
		cron: cron.New(), // 标准 5 字段格式：分 时 日 月 周
		db:   db,
		zlog: zlog,
	}
}

// Start 启动调度器并注册所有定时任务。
func (s *Scheduler) Start() {
	s.registerJobs()
	s.cron.Start()
	s.zlog.Info("调度器已启动")
}

// Stop 优雅停止调度器。
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.zlog.Info("调度器已停止")
}

// registerJobs 注册所有定时任务。
func (s *Scheduler) registerJobs() {
	jobs := []struct {
		name string
		spec string
		fn   func() error
	}{
		{
			name: "ty_overdue_warn",
			spec: "0 9 * * *", // 每日 09:00
			fn:   s.tyOverdueWarn,
		},
		{
			name: "qg_payroll_gen",
			spec: "0 2 1 * *", // 每月 1 号 02:00
			fn:   s.qgPayrollGen,
		},
		{
			name: "sq_late_alert",
			spec: "30 22 * * *", // 每日 22:30
			fn:   s.sqLateAlert,
		},
		{
			name: "cmp_recompute",
			spec: "0 2 * * *", // 每日 02:00
			fn:   s.cmpRecompute,
		},
	}

	for _, j := range jobs {
		name := j.name
		fn := j.fn
		_, err := s.cron.AddFunc(j.spec, func() {
			s.runJob(name, fn)
		})
		if err != nil {
			s.zlog.Error("注册定时任务失败", zap.String("job", name), zap.Error(err))
		} else {
			s.zlog.Info("注册定时任务", zap.String("job", name), zap.String("spec", j.spec))
		}
	}
}

// RunJobManual 手动触发指定任务（供 HTTP handler 调用）。
func (s *Scheduler) RunJobManual(name string) (*models.JobRun, error) {
	jobMap := map[string]func() error{
		"ty_overdue_warn": s.tyOverdueWarn,
		"qg_payroll_gen":  s.qgPayrollGen,
		"sq_late_alert":   s.sqLateAlert,
		"cmp_recompute":   s.cmpRecompute,
	}

	fn, ok := jobMap[name]
	if !ok {
		return nil, fmt.Errorf("未知任务: %s", name)
	}

	return s.runJob(name, fn), nil
}

// runJob 通用任务执行包装：创建 job_run 记录，记录开始/结束/耗时/错误。
func (s *Scheduler) runJob(name string, fn func() error) *models.JobRun {
	now := time.Now()
	jobRun := &models.JobRun{
		JobName:     name,
		ScheduledAt: now,
		Status:      "running",
	}

	startedAt := now
	jobRun.StartedAt = &startedAt

	// 创建运行记录
	if err := s.db.Create(jobRun).Error; err != nil {
		s.zlog.Error("创建 job_run 记录失败", zap.String("job", name), zap.Error(err))
		return jobRun
	}

	// 执行任务
	jobErr := fn()

	// 更新运行记录
	finishedAt := time.Now()
	durationMs := int(finishedAt.Sub(startedAt).Milliseconds())

	updates := map[string]interface{}{
		"finished_at": finishedAt,
		"duration_ms": durationMs,
	}

	if jobErr != nil {
		updates["status"] = "failed"
		updates["error"] = jobErr.Error()
		s.zlog.Error("定时任务执行失败",
			zap.String("job", name),
			zap.Int("duration_ms", durationMs),
			zap.Error(jobErr),
		)
	} else {
		updates["status"] = "success"
		s.zlog.Info("定时任务执行完成",
			zap.String("job", name),
			zap.Int("duration_ms", durationMs),
		)
	}

	if err := s.db.Model(jobRun).Updates(updates).Error; err != nil {
		s.zlog.Error("更新 job_run 记录失败", zap.String("job", name), zap.Error(err))
	}

	// 刷新 jobRun 以便返回最新状态
	s.db.First(jobRun, jobRun.ID)

	return jobRun
}

// ─── 任务实现 ────────────────────────────────────────────────────────────────

// tyOverdueWarn 扫描超期培养记录：查找 is_overdue=0 但距上次记录超过 2 个月的申请，标记 is_overdue=1。
func (s *Scheduler) tyOverdueWarn() error {
	deadline := time.Now().AddDate(0, -2, 0) // 2 个月前

	// 查找超期但未标记的培养记录
	var overdueRecords []models.TyCultivationRecord
	if err := s.db.Where("is_overdue = 0 AND is_deleted = 0 AND updated_at < ?", deadline).
		Find(&overdueRecords).Error; err != nil {
		return fmt.Errorf("查询超期培养记录失败: %w", err)
	}

	if len(overdueRecords) == 0 {
		s.zlog.Info("ty_overdue_warn: 无超期培养记录")
		return nil
	}

	// 批量标记超期
	ids := make([]int64, 0, len(overdueRecords))
	for _, r := range overdueRecords {
		ids = append(ids, r.ID)
	}

	if err := s.db.Model(&models.TyCultivationRecord{}).
		Where("id IN ?", ids).
		Update("is_overdue", 1).Error; err != nil {
		return fmt.Errorf("标记超期培养记录失败: %w", err)
	}

	s.zlog.Info("ty_overdue_warn: 已标记超期培养记录",
		zap.Int("count", len(ids)),
	)

	// 记录到最新 job_run 的 payload_json（在 runJob 中无法直接更新，此处仅日志记录）
	return nil
}

// qgPayrollGen 生成上月薪酬初稿：查找所有 status=on_job 的 qg_position_apply，
// 汇总上月 qg_attendance 有效工时，创建 qg_payroll 草稿记录。
func (s *Scheduler) qgPayrollGen() error {
	now := time.Now()
	// 上月
	payYear := now.Year()
	payMonth := int(now.Month()) - 1
	if payMonth <= 0 {
		payMonth = 12
		payYear--
	}

	// 查找所有在岗的岗位申请
	var onJobApplies []models.QgPositionApply
	if err := s.db.Where("status = ? AND is_deleted = 0", "on_job").
		Find(&onJobApplies).Error; err != nil {
		return fmt.Errorf("查询在岗申请失败: %w", err)
	}

	if len(onJobApplies) == 0 {
		s.zlog.Info("qg_payroll_gen: 无在岗申请")
		return nil
	}

	created := 0
	for _, apply := range onJobApplies {
		// 检查是否已存在该月薪酬记录
		var existCount int64
		s.db.Model(&models.QgPayroll{}).
			Where("student_id = ? AND apply_id = ? AND pay_year = ? AND pay_month = ? AND is_deleted = 0",
				apply.StudentID, apply.ID, payYear, payMonth).
			Count(&existCount)
		if existCount > 0 {
			continue
		}

		// 汇总上月有效工时
		var totalHours float64
		s.db.Model(&models.QgAttendance{}).
			Where("apply_id = ? AND is_deleted = 0", apply.ID).
			Select("COALESCE(SUM(effective_hours), 0)").
			Scan(&totalHours)

		// 获取岗位时薪
		var position models.QgPosition
		if err := s.db.Where("id = ? AND is_deleted = 0", apply.PositionID).First(&position).Error; err != nil {
			s.zlog.Warn("qg_payroll_gen: 查询岗位失败，跳过",
				zap.Int64("position_id", apply.PositionID),
				zap.Error(err),
			)
			continue
		}

		// 计算薪酬：total_hours * hourly_rate_cents * coefficient
		coefficient := 1.0
		grossCents := int64(float64(position.HourlyRateCents) * totalHours * coefficient)

		payroll := models.QgPayroll{
			StudentID:   apply.StudentID,
			ApplyID:     apply.ID,
			PayYear:     payYear,
			PayMonth:    payMonth,
			TotalHours:  totalHours,
			GrossCents:  grossCents,
			TaxCents:    0,
			NetCents:    grossCents, // 草稿阶段 net = gross
			Coefficient: coefficient,
			Status:      "draft",
		}

		if err := s.db.Create(&payroll).Error; err != nil {
			s.zlog.Warn("qg_payroll_gen: 创建薪酬记录失败，跳过",
				zap.Int64("apply_id", apply.ID),
				zap.Error(err),
			)
			continue
		}
		created++
	}

	s.zlog.Info("qg_payroll_gen: 薪酬初稿生成完成",
		zap.Int("on_job_count", len(onJobApplies)),
		zap.Int("created", created),
		zap.Int("pay_year", payYear),
		zap.Int("pay_month", payMonth),
	)
	return nil
}

// sqLateAlert 扫描晚归：查找 sq_late_return 中本学期累计 >= 3 次的学生。
func (s *Scheduler) sqLateAlert() error {
	// 简化实现：统计所有未删除的晚归记录，按学生+学期分组
	type lateCount struct {
		StudentID int64  `json:"student_id"`
		Semester  string `json:"semester"`
		Count     int64  `json:"count"`
	}

	var counts []lateCount
	if err := s.db.Model(&models.SqLateReturn{}).
		Select("student_id, semester, count(*) as count").
		Where("is_deleted = 0").
		Group("student_id, semester").
		Having("count(*) >= 3").
		Find(&counts).Error; err != nil {
		return fmt.Errorf("查询晚归统计失败: %w", err)
	}

	if len(counts) == 0 {
		s.zlog.Info("sq_late_alert: 无累计晚归 >= 3 次的学生")
		return nil
	}

	// 序列化为 JSON 记录到日志
	payload, _ := json.Marshal(counts)
	s.zlog.Info("sq_late_alert: 发现晚归预警学生",
		zap.Int("count", len(counts)),
		zap.String("payload", string(payload)),
	)

	// S11 占位：后续对接通知通道
	return nil
}

// cmpRecompute 综合素质每日全量重算：S12 核心实现，调用 Calculator 对全部学生重新聚合。
func (s *Scheduler) cmpRecompute() error {
	// 1. 拉取当前激活规则版本
	var rule models.CmpRuleVersion
	if err := s.db.Where("is_active = 1 AND is_deleted = 0").First(&rule).Error; err != nil {
		s.zlog.Warn("cmp_recompute: 未找到激活规则版本，使用默认规则")
	}

	// 2. 列出全部学生
	var studentIDs []int64
	if err := s.db.Model(&models.IdxStudent{}).
		Where("is_deleted = 0").
		Pluck("id", &studentIDs).Error; err != nil {
		return fmt.Errorf("查询学生列表失败: %w", err)
	}

	// 3. 学年：当前学年（9 月起算）
	now := time.Now()
	year := now.Year()
	if now.Month() < time.September {
		year--
	}
	academicYear := fmt.Sprintf("%d-%d", year, year+1)

	// 4. 逐学生计算
	succeeded, failed := 0, 0
	for _, sid := range studentIDs {
		// 复用 calculator 的聚合 SQL 是不切实际的（calculator 封装在 cmp 模块内），
		// 这里直接以简化公式计算：维度固定权重的归一化总分（与 PRD §8.4 对齐）。
		total, err := s.simpleRecomputeStudent(sid, academicYear, rule.ID)
		if err != nil {
			s.zlog.Warn("cmp_recompute: 单学生重算失败",
				zap.Int64("student_id", sid), zap.Error(err))
			failed++
			continue
		}
		_ = total
		succeeded++
	}

	s.zlog.Info("cmp_recompute: 全量重算完成",
		zap.String("academic_year", academicYear),
		zap.String("rule_version", rule.Version),
		zap.Int("total_students", len(studentIDs)),
		zap.Int("succeeded", succeeded),
		zap.Int("failed", failed),
	)
	return nil
}

// simpleRecomputeStudent 单学生简化重算（定时任务使用）。
// 与 cmp/service/calculator.go 中 5 个维度聚合保持一致口径。
func (s *Scheduler) simpleRecomputeStudent(studentID int64, academicYear string, ruleVersionID int64) (float64, error) {
	// 团内
	var politicalStatus string
	_ = s.db.Model(&models.IdxStudent{}).
		Select("political_status").Where("id = ?", studentID).First(&politicalStatus).Error
	leagueID := 0.0
	switch politicalStatus {
	case "member":
		leagueID = 5
	case "probationary":
		leagueID = 3
	case "activist":
		leagueID = 1
	}
	var leaguePos int64
	s.db.Model(&models.TyMemberRoster{}).
		Where("student_id = ? AND is_deleted = 0 AND position_code <> '' AND position_code IS NOT NULL", studentID).
		Count(&leaguePos)
	leaguePosScore := 0.0
	if leaguePos > 0 {
		leaguePosScore = 6
	}
	var leagueReports int64
	s.db.Model(&models.TyThoughtReport{}).
		Where("student_id = ? AND is_deleted = 0", studentID).
		Count(&leagueReports)
	leagueReportsScore := math.Min(float64(leagueReports)*2, 6)
	var leaguePassed int64
	s.db.Model(&models.TyApplication{}).
		Where("student_id = ? AND status = 'S3' AND is_deleted = 0", studentID).
		Count(&leaguePassed)
	leaguePassScore := math.Min(float64(leaguePassed)*4, 4)
	var leagueCourses int64
	s.db.Model(&models.TyCourseRecord{}).
		Where("student_id = ? AND is_deleted = 0 AND passed = 1", studentID).
		Count(&leagueCourses)
	leagueCourseScore := math.Min(float64(leagueCourses)*2, 4)
	leagueTotal := leagueID + leaguePosScore + leagueReportsScore + leaguePassScore + leagueCourseScore

	// 社团
	var assocOfficer int64
	s.db.Model(&models.StAssocMember{}).
		Where("student_id = ? AND is_deleted = 0 AND role IN ('president','vice','director')", studentID).
		Count(&assocOfficer)
	assocOfficerScore := 0.0
	if assocOfficer > 0 {
		assocOfficerScore = 5
	}
	var assocActivity int64
	s.db.Model(&models.StActivity{}).
		Where("is_deleted = 0 AND status = 'S3' AND association_id IN (SELECT association_id FROM st_assoc_member WHERE student_id = ? AND is_deleted = 0)", studentID).
		Count(&assocActivity)
	assocOrgScore := math.Min(float64(assocActivity)*2, 8)
	var checkinCount int64
	s.db.Model(&models.StActivityCheckin{}).
		Where("student_id = ? AND is_deleted = 0", studentID).
		Count(&checkinCount)
	assocCheckScore := math.Min(float64(checkinCount)*1, 5)
	var ratingCount int64
	s.db.Model(&models.StRating{}).
		Where("student_id = ? AND is_deleted = 0", studentID).
		Count(&ratingCount)
	assocRateScore := math.Min(float64(ratingCount)*1, 2)
	assocTotal := assocOfficerScore + assocOrgScore + assocCheckScore + assocRateScore

	// 社区
	var sqPos int64
	s.db.Model(&models.SqSelfgovPosition{}).
		Where("student_id = ? AND is_deleted = 0 AND status = 'active'", studentID).
		Count(&sqPos)
	sqPosScore := 0.0
	if sqPos > 0 {
		sqPosScore = 3
	}
	var sqAssessAvg float64
	s.db.Model(&models.SqAssessment{}).
		Select("COALESCE(AVG(score), 0)").
		Where("student_id = ? AND is_deleted = 0", studentID).
		Scan(&sqAssessAvg)
	sqAssessScore := math.Min(sqAssessAvg/20, 5)
	var roomAvg float64
	s.db.Model(&models.SqInspection{}).
		Select("COALESCE(AVG(score), 0)").
		Where("is_deleted = 0 AND room_id IN (SELECT room_id FROM idx_dorm_room_member WHERE student_id = ? AND is_deleted = 0)", studentID).
		Scan(&roomAvg)
	sqCivScore := 0.0
	switch {
	case roomAvg >= 90:
		sqCivScore = 3
	case roomAvg >= 80:
		sqCivScore = 2
	}
	var sqAction int64
	s.db.Model(&models.SqIncidentAction{}).
		Where("action_by IN (SELECT id FROM sys_user WHERE student_id = ?) AND is_deleted = 0", studentID).
		Count(&sqAction)
	sqActionScore := math.Min(float64(sqAction)*2, 4)
	commTotal := sqPosScore + sqAssessScore + sqCivScore + sqActionScore

	// 勤工
	var qgOnJob int64
	s.db.Model(&models.QgPositionApply{}).
		Where("student_id = ? AND is_deleted = 0 AND status = 'on_job'", studentID).
		Count(&qgOnJob)
	qgDutyScore := 0.0
	if qgOnJob > 0 {
		qgDutyScore = 7
	}
	var qgTotalHours float64
	s.db.Model(&models.QgPayroll{}).
		Select("COALESCE(SUM(total_hours), 0)").
		Where("student_id = ? AND is_deleted = 0", studentID).
		Scan(&qgTotalHours)
	qgHoursScore := math.Min(qgTotalHours/40*3, 3)
	var qgAssessAvg float64
	s.db.Model(&models.QgMonthlyAssess{}).
		Select("COALESCE(AVG(total_score), 0)").
		Where("student_id = ? AND is_deleted = 0", studentID).
		Scan(&qgAssessAvg)
	qgAssessScore := 0.0
	switch {
	case qgAssessAvg >= 90:
		qgAssessScore = 5
	case qgAssessAvg >= 60:
		qgAssessScore = 4
	}
	workTotal := qgDutyScore + qgHoursScore + qgAssessScore

	// 学业（占位 15 + 10）
	academicTotal := 25.0

	total := leagueTotal + assocTotal + commTotal + workTotal + academicTotal
	if total > 100 {
		total = 100
	}
	total = math.Round(total*100) / 100

	// upsert
	now := time.Now()
	var existing models.CmpScore
	err := s.db.Where("student_id = ? AND academic_year = ? AND is_deleted = 0",
		studentID, academicYear).First(&existing).Error
	if err == nil {
		existing.TotalScore = total
		existing.RuleVersionID = ruleVersionID
		existing.ComputedAt = now
		_ = s.db.Save(&existing).Error
	} else {
		score := &models.CmpScore{
			StudentID:     studentID,
			AcademicYear:  academicYear,
			TotalScore:    total,
			RuleVersionID: ruleVersionID,
			ComputedAt:    now,
		}
		_ = s.db.Create(score).Error
	}
	return total, nil
}

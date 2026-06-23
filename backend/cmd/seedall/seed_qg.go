package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"student-system/internal/models"
	"student-system/pkg/cryptox"
)

// seedQG 为勤工助学模块灌入测试数据。
// 规模：qg_difficulty_cert +5（覆盖 4 等级）、qg_position +5、qg_position_apply +4、
//       qg_attendance +30、qg_monthly_assess +3、qg_payroll +3。
func (c *ctx) seedQG() error {
	if err := c.seedQgDifficultyCert(); err != nil {
		return err
	}
	if err := c.seedQgPosition(); err != nil {
		return err
	}
	if err := c.seedQgPositionApply(); err != nil {
		return err
	}
	if err := c.seedQgAttendance(); err != nil {
		return err
	}
	if err := c.seedQgMonthlyAssess(); err != nil {
		return err
	}
	if err := c.seedQgPayroll(); err != nil {
		return err
	}
	return nil
}

// seedQgDifficultyCert 困难认定：补 5 条，覆盖 special/hard/normal/none。
func (c *ctx) seedQgDifficultyCert() error {
	const target = 5
	cur := count(c.db, &models.QgDifficultyCert{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] qg_difficulty_cert 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  qg_difficulty_cert 补 %d 条\n", need)

	levels := []string{"special", "hard", "normal", "none", "hard"}
	statuses := []string{S3, S3, S2, S1, S4}

	created := 0
	used := map[int64]struct{}{}
	for i := 0; i < int(need)*2 && created < int(need); i++ {
		stu := c.students[i%len(c.students)]
		if _, ok := used[stu.ID]; ok {
			continue
		}
		used[stu.ID] = struct{}{}

		// 检查同学年是否已存在
		var exists int64
		c.db.Model(&models.QgDifficultyCert{}).Where("student_id = ? AND academic_year = ? AND is_deleted = 0", stu.ID, "2025-2026").Count(&exists)
		if exists > 0 {
			continue
		}

		files, _ := json.Marshal([]int64{1, 2, 3})
		idx := created
		bizNo := nextBizNo(c.db, "QG")
		publicStart := c.now.AddDate(0, 0, -10+idx)
		publicEnd := c.now.AddDate(0, 0, -3+idx)
		cert := models.QgDifficultyCert{
			BizNo:        bizNo,
			StudentID:    stu.ID,
			AcademicYear: "2025-2026",
			Level:        levels[idx%len(levels)],
			CertFiles:    string(files),
			PublicStart:  &publicStart,
			PublicEnd:    &publicEnd,
			Status:       statuses[idx%len(statuses)],
		}
		if statuses[idx%len(statuses)] == S4 {
			cert.RejectReason = "申请表缺少村委会盖章。"
		}
		if err := c.db.Create(&cert).Error; err != nil {
			return fmt.Errorf("创建困难认定失败: %w", err)
		}
		created++
		fmt.Printf("    [OK] 困难认定 level=%s status=%s\n", cert.Level, cert.Status)
	}
	return nil
}

// seedQgPosition 岗位：补 5 个。
func (c *ctx) seedQgPosition() error {
	const target = 5
	cur := count(c.db, &models.QgPosition{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] qg_position 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  qg_position 补 %d 条\n", need)

	specs := []struct {
		DeptType   string
		DeptName   string
		Title      string
		Desc       string
		Headcount  int
		WeeklyHour int
		HourlyRate int64
		Status     string
	}{
		{"admin", "图书馆", "图书管理员助理", "协助图书上架、借还书管理、阅览室秩序维护。", 4, 12, 2500, S3},
		{"teaching", "教务处", "教学助理", "协助教师批改作业、整理课件、教学资料归档。", 6, 15, 3000, S3},
		{"research", "计算机学院", "科研助理", "协助实验室日常事务、数据整理、设备维护。", 3, 16, 3500, S2},
		{"culture", "团委", "学生活动助理", "协助组织校园文化活动、场地协调、物资管理。", 2, 10, 2200, S3},
		{"admin", "后勤处", "校园环境维护助理", "协助校园巡查、设施报修、应急处理。", 5, 14, 2400, S1},
	}
	for i, sp := range specs {
		if i >= int(need) {
			break
		}
		// 标题去重
		var exists int64
		c.db.Model(&models.QgPosition{}).Where("title = ? AND is_deleted = 0", sp.Title).Count(&exists)
		if exists > 0 {
			continue
		}
		bizNo := nextBizNo(c.db, "QG")
		start := c.now.AddDate(0, 0, -30)
		end := c.now.AddDate(0, 6, 0)
		pos := models.QgPosition{
			BizNo:            bizNo,
			DeptType:         sp.DeptType,
			DeptName:         sp.DeptName,
			Title:            sp.Title,
			Description:      sp.Desc,
			Headcount:        sp.Headcount,
			WeeklyHoursLimit: sp.WeeklyHour,
			HourlyRateCents:  sp.HourlyRate,
			StartAt:          start,
			EndAt:            end,
			RiskNotes:        "无危险工种；按要求做好考勤。",
			Status:           sp.Status,
			SupervisorUserID: ptrInt64(1),
		}
		if err := c.db.Create(&pos).Error; err != nil {
			return fmt.Errorf("创建岗位失败: %w", err)
		}
		fmt.Printf("    [OK] 岗位 %s (dept=%s status=%s)\n", pos.Title, pos.DeptName, pos.Status)
	}
	c.reloadPositions()
	return nil
}

// seedQgPositionApply 岗位申请：补 4 条。
func (c *ctx) seedQgPositionApply() error {
	const target = 4
	cur := count(c.db, &models.QgPositionApply{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] qg_position_apply 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  qg_position_apply 补 %d 条\n", need)

	if len(c.positions) == 0 {
		fmt.Println("    [skip] 无可用岗位")
		return nil
	}

	statuses := []string{"on_job", "on_job", "onboarding", "on_job"}
	applyStatuses := []string{"accepted", "accepted", "interview", "accepted"}

	created := 0
	used := map[string]struct{}{}
	for i := 0; i < int(need)*3 && created < int(need); i++ {
		pos := c.positions[i%len(c.positions)]
		stu := c.students[(i*7)%len(c.students)]
		key := fmt.Sprintf("%d-%d", pos.ID, stu.ID)
		if _, ok := used[key]; ok {
			continue
		}
		used[key] = struct{}{}

		// 检查是否已申请过该岗位
		var exists int64
		c.db.Model(&models.QgPositionApply{}).Where("position_id = ? AND student_id = ? AND is_deleted = 0", pos.ID, stu.ID).Count(&exists)
		if exists > 0 {
			continue
		}

		idx := created
		apply := models.QgPositionApply{
			PositionID:  pos.ID,
			StudentID:   stu.ID,
			ApplyStatus: applyStatuses[idx%len(applyStatuses)],
			InterviewAt: ptrTime(c.now.AddDate(0, 0, -10+idx)),
			ConfirmDeadline: ptrTime(c.now.AddDate(0, 0, -7+idx)),
			ConfirmedAt:  ptrTime(c.now.AddDate(0, 0, -5+idx)),
			OnBoardAt:    ptrTime(c.now.AddDate(0, 0, -3+idx)),
			Status:       statuses[idx%len(statuses)],
		}
		if err := c.db.Create(&apply).Error; err != nil {
			return fmt.Errorf("创建岗位申请失败: %w", err)
		}
		created++
		fmt.Printf("    [OK] 岗位申请 student=%s position=%s status=%s\n", stu.StudentNo, pos.Title, apply.Status)
	}
	c.reloadApplys()
	return nil
}

// seedQgAttendance 打卡：补 30 条。
func (c *ctx) seedQgAttendance() error {
	const target = 30
	cur := count(c.db, &models.QgAttendance{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] qg_attendance 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  qg_attendance 补 %d 条\n", need)

	// 取在岗申请
	var apps []models.QgPositionApply
	c.db.Where("status = ? AND is_deleted = 0", "on_job").Order("id ASC").Find(&apps)
	if len(apps) == 0 {
		fmt.Println("    [skip] 无在岗申请")
		return nil
	}

	created := 0
	appIdx := 0
	dayOffset := 1

	// 预加载已存在 (apply_id, work_date) 对，避免 UNIQUE 冲突
	type attKey struct{ ApplyID int64; Date string }
	existAtt := map[attKey]struct{}{}
	var existingAtt []models.QgAttendance
	c.db.Find(&existingAtt)
	for _, a := range existingAtt {
		existAtt[attKey{a.ApplyID, a.WorkDate.Format("2006-01-02")}] = struct{}{}
	}

	for need > 0 {
		app := apps[appIdx%len(apps)]
		if _, ok := c.findPosition(app.PositionID); !ok {
			appIdx++
			continue
		}
		// 每人每天一条
		workDate := c.now.AddDate(0, 0, -dayOffset)
		dateStr := workDate.Format("2006-01-02")
		key := attKey{app.ID, dateStr}
		if _, ok := existAtt[key]; ok {
			appIdx++
			dayOffset++
			if dayOffset > 30 {
				break
			}
			continue
		}
		clockIn := time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 8, 0, 0, 0, workDate.Location())
		clockOut := clockIn.Add(time.Duration(4+rand.Intn(3)) * time.Hour)
		effHours := clockOut.Sub(clockIn).Hours()

		lateMin := 0
		if rand.Intn(4) == 0 {
			lateMin = 5 + rand.Intn(15)
			clockIn = clockIn.Add(time.Duration(lateMin) * time.Minute)
		}

		att := models.QgAttendance{
			ApplyID:        app.ID,
			StudentID:      app.StudentID,
			WorkDate:       workDate,
			ClockInAt:      &clockIn,
			ClockOutAt:     &clockOut,
			EffectiveHours: effHours,
			LateMinutes:    lateMin,
			EarlyMinutes:   0,
			ClockMethod:    []string{"gps_face", "card", "manual"}[appIdx%3],
			IP:             "10.0.0." + fmt.Sprintf("%d", 10+appIdx%200),
			Geo:            "120.12,30.27",
			IsMakeup:       0,
		}
		if err := c.db.Create(&att).Error; err != nil {
			return fmt.Errorf("创建打卡失败: %w", err)
		}
		existAtt[key] = struct{}{}
		created++
		need--
		appIdx++
		if appIdx >= 100 {
			break
		}
	}
	fmt.Printf("    [OK] 打卡 新增 %d 条\n", created)
	return nil
}

// seedQgMonthlyAssess 月度考核：补 3 条。
func (c *ctx) seedQgMonthlyAssess() error {
	const target = 3
	cur := count(c.db, &models.QgMonthlyAssess{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] qg_monthly_assess 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  qg_monthly_assess 补 %d 条\n", need)

	// 取最早一个在岗申请
	var apps []models.QgPositionApply
	c.db.Where("status = ? AND is_deleted = 0", "on_job").Order("id ASC").Find(&apps)
	if len(apps) == 0 {
		return nil
	}

	scores := []struct{ A, W, C int }{
		{40, 35, 15},
		{38, 33, 18},
		{42, 36, 12},
	}
	created := 0
	for i := 0; i < int(need) && i < len(apps) && i < len(scores); i++ {
		app := apps[i]
		// 月度考核 UNIQUE(apply_id, year, month)
		year, month := 2026, time.May
		s := scores[i]
		weighted := float64(s.A) + float64(s.W) + float64(s.C)
		coefficient := 1.0
		if weighted < 60 {
			coefficient = 0.5
		} else if weighted < 75 {
			coefficient = 0.8
		}
		ass := models.QgMonthlyAssess{
			ApplyID:           app.ID,
			StudentID:         app.StudentID,
			AssessYear:        year,
			AssessMonth:       int(month),
			ScoreAttendance:   s.A,
			ScoreWorkComplete: s.W,
			ScoreComprehensive: s.C,
			WeightedScore:     weighted,
			Coefficient:       coefficient,
			IsObservation:     0,
			Note:              "本月份考核合格。",
			Status:            S3,
		}
		if err := c.db.Create(&ass).Error; err != nil {
			// 唯一约束冲突就跳过
			if isUniqueErr(err) {
				continue
			}
			return fmt.Errorf("创建月度考核失败: %w", err)
		}
		created++
	}
	fmt.Printf("    [OK] 月度考核 新增 %d 条\n", created)
	return nil
}

// seedQgPayroll 薪酬发放：补 3 条。
func (c *ctx) seedQgPayroll() error {
	const target = 3
	cur := count(c.db, &models.QgPayroll{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] qg_payroll 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  qg_payroll 补 %d 条\n", need)

	// 找月度考核已通过的申请
	var asss []models.QgMonthlyAssess
	c.db.Where("status = ? AND is_deleted = 0", S3).Order("id ASC").Find(&asss)
	if len(asss) == 0 {
		fmt.Println("    [skip] 无月度考核记录")
		return nil
	}

	created := 0
	for i, ass := range asss {
		if created >= int(need) {
			break
		}
		// 找对应岗位的小时费率
		var apply models.QgPositionApply
		c.db.First(&apply, ass.ApplyID)
		pos, ok := c.findPosition(apply.PositionID)
		if !ok {
			continue
		}

		totalHours := 30.0
		gross := int64(totalHours * float64(pos.HourlyRateCents) * ass.Coefficient)
		tax := int64(0)
		if gross > 80000 {
			tax = gross * 8 / 100 // 简化的个税计算
		}
		net := gross - tax

		bizNo := nextBizNo(c.db, "QG")
		bank := "6228" + fmt.Sprintf("%012d", ass.StudentID*1000+int64(i))
		bankEnc := cryptox.Encrypt(bank)
		paidAt := c.now.AddDate(0, 0, -5)
		payroll := models.QgPayroll{
			BizNo:               bizNo,
			StudentID:           ass.StudentID,
			ApplyID:             ass.ApplyID,
			PayYear:             ass.AssessYear,
			PayMonth:            ass.AssessMonth,
			TotalHours:          totalHours,
			GrossCents:          gross,
			TaxCents:            tax,
			DeductionCents:      0,
			NetCents:            net,
			Coefficient:         ass.Coefficient,
			BankAccountLast4Enc: bankEnc,
			Status:              "paid",
			ReviewedBy:          ptrInt64(1),
			PaidAt:              &paidAt,
		}
		if err := c.db.Create(&payroll).Error; err != nil {
			if isUniqueErr(err) {
				continue
			}
			return fmt.Errorf("创建薪酬失败: %w", err)
		}
		created++
	}
	fmt.Printf("    [OK] 薪酬 新增 %d 条\n", created)
	return nil
}

// findPosition 通过 ID 查找岗位。
func (c *ctx) findPosition(id int64) (models.QgPosition, bool) {
	for _, p := range c.positions {
		if p.ID == id {
			return p, true
		}
	}
	return models.QgPosition{}, false
}

// reloadPositions 重新加载岗位缓存。
func (c *ctx) reloadPositions() {
	c.positions = nil
	c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.positions)
}

// reloadApplys 重新加载岗位申请缓存。
func (c *ctx) reloadApplys() {
	c.applys = nil
	c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.applys)
}

// isUniqueErr 判断是否是唯一约束冲突。
func isUniqueErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "UNIQUE constraint failed") || contains(msg, "unique")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (s[:len(sub)] == sub || s[len(s)-len(sub):] == sub || indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

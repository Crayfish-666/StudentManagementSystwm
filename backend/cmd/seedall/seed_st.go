package main

import (
	"fmt"
	"math/rand"
	"time"

	"student-system/internal/models"
)

// seedST 为社团活动模块灌入测试数据。
// 规模：st_association +4、st_activity +8、st_activity_approval +24、st_activity_checkin +30、
//       st_activity_summary +3、st_expense +2、st_rating +1。
func (c *ctx) seedST() error {
	if err := c.seedStAssociation(); err != nil {
		return err
	}
	if err := c.seedStActivity(); err != nil {
		return err
	}
	if err := c.seedStCheckinAndSummary(); err != nil {
		return err
	}
	if err := c.seedStExpense(); err != nil {
		return err
	}
	if err := c.seedStRating(); err != nil {
		return err
	}
	return nil
}

// seedStAssociation 补足社团到 5 个，覆盖 5 种状态。
func (c *ctx) seedStAssociation() error {
	const target = 5
	cur := count(c.db, &models.StAssociation{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] st_association 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  st_association 补 %d 条\n", need)

	specs := []struct {
		Name          string
		CollegeCode   string
		BusinessScope string
		Status        string
	}{
		{"ACM 程序设计竞赛社", "CS", "程序设计竞赛、算法培训", "registered"},
		{"数学建模协会", "CS", "数学建模培训、竞赛组织", "registered"},
		{"机器人创新社", "EE", "机器人设计、智能硬件开发", "registered"},
		{"演讲与口才社", "CS", "演讲比赛、辩论训练、主持培训", "trial"},
		{"篮球社", "CS", "篮球训练、校际联赛", "preparing"},
	}

	for _, sp := range specs {
		if need <= 0 {
			break
		}
		var exists int64
		c.db.Model(&models.StAssociation{}).Where("name = ? AND is_deleted = 0", sp.Name).Count(&exists)
		if exists > 0 {
			continue
		}
		college := c.colleges[0]
		for _, col := range c.colleges {
			if len(col.Code) > 0 && len(sp.CollegeCode) > 0 && col.Code[0] == sp.CollegeCode[0] {
				college = col
				break
			}
		}
		bizNo := nextBizNo(c.db, "ST")
		assoc := models.StAssociation{
			BizNo:         bizNo,
			Name:          sp.Name,
			CollegeID:     college.ID,
			BusinessScope: sp.BusinessScope,
			Status:        sp.Status,
			FoundedAt:     ptrTime(c.now.AddDate(-1, -rand.Intn(6), 0)),
		}
		switch sp.Status {
		case "trial":
			assoc.TrialStartedAt = ptrTime(c.now.AddDate(0, -3, 0))
		case "registered":
			assoc.TrialStartedAt = ptrTime(c.now.AddDate(0, -6, 0))
			assoc.RegisteredAt = ptrTime(c.now.AddDate(0, -3, 0))
		}
		if err := c.db.Create(&assoc).Error; err != nil {
			return fmt.Errorf("创建社团 %s 失败: %w", sp.Name, err)
		}
		fmt.Printf("    [OK] 社团 %s (status=%s)\n", assoc.Name, assoc.Status)
		need--
	}
	c.reloadAssoc()
	return nil
}

// seedStActivity 补足活动到 8 个，覆盖 A/B/C/D 4 个等级 + 各种状态。
func (c *ctx) seedStActivity() error {
	const target = 8
	cur := count(c.db, &models.StActivity{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] st_activity 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  st_activity 补 %d 条\n", need)

	// 社团不足则跳过
	if len(c.assoc) == 0 {
		fmt.Println("    [skip] 无可用社团")
		return nil
	}

	specs := []struct {
		Title     string
		Level     string
		Status    string
		Budget    int64
		Participants int
	}{
		{"校级程序设计竞赛", "A", S3, 800000, 500},
		{"校园文化节", "A", S2, 1200000, 800},
		{"数学建模校内赛", "B", S3, 300000, 200},
		{"机器人工作坊", "B", S2, 150000, 80},
		{"演讲比赛月赛", "C", S3, 50000, 60},
		{"篮球 3v3 联赛", "C", S1, 30000, 50},
		{"读书分享会", "D", S3, 5000, 30},
		{"编程入门讲座", "D", S0, 2000, 25},
	}

	used := map[int64]struct{}{}
	created := 0
	for _, sp := range specs {
		if need <= 0 {
			break
		}
		// 用模运算选社团
		assoc := c.assoc[created%len(c.assoc)]

		bizNo := nextBizNo(c.db, "ST")
		// 安排时间：A级 > 30 天后；B/C 15-30 天；D 7-15 天
		days := 0
		switch sp.Level {
		case "A":
			days = 45
		case "B":
			days = 25
		case "C":
			days = 15
		case "D":
			days = 7
		}
		start := c.now.AddDate(0, 0, days)
		end := start.Add(time.Hour * 3)
		expCount := sp.Participants
		exp := int(sp.Participants)

		act := models.StActivity{
			BizNo:                bizNo,
			AssociationID:        assoc.ID,
			Title:                sp.Title,
			ActivityLevel:        sp.Level,
			ExpectedParticipants: expCount,
			BudgetCents:          sp.Budget,
			Location:             "大学生活动中心",
			StartedAt:            start,
			EndedAt:              end,
			ExpectedCount:        &exp,
			Status:               sp.Status,
		}
		if err := c.db.Create(&act).Error; err != nil {
			return fmt.Errorf("创建活动 %s 失败: %w", sp.Title, err)
		}
		fmt.Printf("    [OK] 活动 %s (level=%s status=%s)\n", act.Title, act.ActivityLevel, act.Status)
		used[act.ID] = struct{}{}

		// 同步生成审批流
		if err := c.genStApprovals(act); err != nil {
			return err
		}
		need--
		created++
	}
	return nil
}

// genStApprovals 为活动生成审批流（A=5、B=4、C=2、D=1）。
func (c *ctx) genStApprovals(act models.StActivity) error {
	stepMap := map[string][]struct {
		Role string
		Name string
	}{
		"A": {
			{"R-COL-TUTOR", "指导教师"},
			{"R-COL-COUN", "院系辅导员"},
			{"R-SY-COUNCIL", "校社联"},
			{"R-SY-LEAGUE", "校团委"},
			{"R-SY-LEADER", "校领导"},
		},
		"B": {
			{"R-COL-TUTOR", "指导教师"},
			{"R-COL-COUN", "院系辅导员"},
			{"R-SY-COUNCIL", "校社联"},
			{"R-SY-LEAGUE", "校团委"},
		},
		"C": {
			{"R-COL-TUTOR", "指导教师"},
			{"R-SY-COUNCIL", "校社联"},
		},
		"D": {
			{"R-COL-TUTOR", "指导教师"},
		},
	}
	steps := stepMap[act.ActivityLevel]
	// S0 草稿无审批；S1 仅第 1 步完成；S2 部分完成；S3 全部完成
	completedSteps := 0
	switch act.Status {
	case S0, "cancelled":
		return nil
	case S1:
		completedSteps = 0
	case S2:
		completedSteps = len(steps) - 1
	case S3:
		completedSteps = len(steps)
	case S4:
		// 驳回到第 1 步
		completedSteps = 0
	}
	occurred := c.now.AddDate(0, 0, -3)
	passOpinions := []string{
		"活动申报材料齐全，方案完整、安全预案到位、参与人数符合要求，经审核同意按期举办，请落实好现场组织和疫情防控要求。",
		"本活动主题鲜明、立意积极，经评议同意立项，请按活动方案推进执行，注意过程资料的留存归档，活动结束后及时提交总结。",
		"活动内容健康向上、预算合理、风险可控，同意通过审批。建议加强宣传发动，进一步扩大覆盖面和师生参与度。",
		"本活动契合校园文化建设方向，能切实服务学生成长成才，同意批准。须严格按预算执行，并做好财务公开与过程留痕。",
	}
	for i, s := range steps {
		if i >= completedSteps {
			// 尚未推进到的步骤：暂不落库（决策列 CHECK 不允许空字符串）
			continue
		}
		uid := int64(i + 1)
		approval := models.StActivityApproval{
			ActivityID:     act.ID,
			StepNo:         i + 1,
			ApproverRole:   s.Role,
			ApproverUserID: &uid,
			Decision:       "pass",
			Opinion:        passOpinions[i%len(passOpinions)],
			DecidedAt:      ptrTime(occurred),
		}
		occurred = occurred.Add(time.Hour * 12)
		if err := c.db.Create(&approval).Error; err != nil {
			return fmt.Errorf("创建活动审批失败: %w", err)
		}
	}
	return nil
}

// seedStCheckinAndSummary 签到 + 总结。
func (c *ctx) seedStCheckin() error {
	const target = 30
	cur := countNoDel(c.db, &models.StActivityCheckin{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] st_activity_checkin 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  st_activity_checkin 补 %d 条\n", need)

	// 找 S3 已完成的活动
	var acts []models.StActivity
	c.db.Where("status = ? AND is_deleted = 0", S3).Order("id ASC").Find(&acts)
	if len(acts) == 0 {
		fmt.Println("    [skip] 无已完成活动")
		return nil
	}

	// 预加载已存在的 (activity, student) 组合，避免 UNIQUE 冲突
	type pair struct{ ActID, StuID int64 }
	exist := map[pair]struct{}{}
	var existing []models.StActivityCheckin
	c.db.Find(&existing)
	for _, e := range existing {
		exist[pair{e.ActivityID, e.StudentID}] = struct{}{}
	}

	created := 0
	actIdx := 0
	for need > 0 {
		act := acts[actIdx%len(acts)]
		// 给每个活动 6-8 人签到
		studentsPerAct := 6 + actIdx%3
		for i := 0; i < studentsPerAct && need > 0; i++ {
			stu := c.students[(actIdx*7+i)%len(c.students)]
			key := pair{act.ID, stu.ID}
			if _, ok := exist[key]; ok {
				continue
			}
			checkin := act.StartedAt.Add(time.Duration(5+i*2) * time.Minute)
			late := 0
			lateMin := 0
			if i%4 == 3 {
				late = 1
				lateMin = 5 + i
			}
			check := models.StActivityCheckin{
				ActivityID:  act.ID,
				StudentID:   stu.ID,
				CheckinAt:   checkin,
				Method:      []string{"qrcode", "gps", "manual"}[i%3],
				IsLate:      late,
				LateMinutes: lateMin,
				IsPresent:   1,
			}
			if lateMin > 15 {
				check.IsPresent = 0 // 视为缺勤
			}
			if err := c.db.Create(&check).Error; err != nil {
				return fmt.Errorf("创建签到失败: %w", err)
			}
			exist[key] = struct{}{}
			need--
			created++
		}
		actIdx++
		if actIdx > 100 {
			break // 防止死循环
		}
	}
	fmt.Printf("    [OK] 签到 新增 %d 条\n", created)
	return nil
}

// seedStSummary 活动总结。
func (c *ctx) seedStSummary() error {
	const target = 3
	cur := count(c.db, &models.StActivitySummary{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] st_activity_summary 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	fmt.Printf("  st_activity_summary 补 %d 条\n", int64(target)-cur)

	var acts []models.StActivity
	c.db.Where("status = ? AND is_deleted = 0", S3).Order("id ASC").Find(&acts)
	// 已存在 summary 的活动集合
	existing := map[int64]struct{}{}
	var sums []models.StActivitySummary
	c.db.Find(&sums)
	for _, s := range sums {
		existing[s.ActivityID] = struct{}{}
	}
	created := 0
	for _, act := range acts {
		if count(c.db, &models.StActivitySummary{}) >= int64(target) {
			break
		}
		if _, ok := existing[act.ID]; ok {
			continue
		}
		var countCheckin int64
		c.db.Model(&models.StActivityCheckin{}).Where("activity_id = ? AND is_present = 1", act.ID).Count(&countCheckin)
		achievement := 85 + rand.Intn(15)
		suggestions := "活动组织有序，同学参与度高。建议下一届延续形式并增加互动环节，并加强活动宣传预热与跨社团联动。"
		summary := models.StActivitySummary{
			ActivityID:         act.ID,
			ActualParticipants: int(countCheckin),
			AchievementScore:   &achievement,
			Suggestions:        suggestions,
			SubmittedAt:        act.EndedAt.AddDate(0, 0, 2),
		}
		if err := c.db.Create(&summary).Error; err != nil {
			return fmt.Errorf("创建活动总结失败: %w", err)
		}
		created++
	}
	fmt.Printf("    [OK] 活动总结 新增 %d 条\n", created)
	return nil
}

// seedStCheckinAndSummary 同时跑签到和总结。
func (c *ctx) seedStCheckinAndSummary() error {
	if err := c.seedStCheckin(); err != nil {
		return err
	}
	if err := c.seedStSummary(); err != nil {
		return err
	}
	return nil
}

// seedStExpense 经费报销。
func (c *ctx) seedStExpense() error {
	const target = 2
	cur := count(c.db, &models.StExpense{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] st_expense 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	fmt.Printf("  st_expense 补 %d 条\n", int64(target)-cur)

	var acts []models.StActivity
	c.db.Where("status = ? AND is_deleted = 0", S3).Order("id ASC").Find(&acts)
	created := 0
	for _, act := range acts {
		if count(c.db, &models.StExpense{}) >= int64(target) {
			break
		}
		amount := act.BudgetCents / 2
		if amount < 100 {
			continue
		}
		bizNo := nextBizNo(c.db, "ST")
		status := S3
		var coSignedBy *int64
		if amount > 1000000 { // > 1 万元 双签
			co := int64(2)
			coSignedBy = &co
		}
		expense := models.StExpense{
			BizNo:        bizNo,
			ActivityID:   act.ID,
			AmountCents:  amount,
			InvoiceCount: 3,
			InvoiceFiles: "[1,2,3]",
			Status:       status,
			ReviewedBy:   ptrInt64(1),
			ReviewedAt:   ptrTime(c.now.AddDate(0, 0, -1)),
			CoSignedBy:   coSignedBy,
			PaidAt:       ptrTime(c.now),
		}
		if err := c.db.Create(&expense).Error; err != nil {
			return fmt.Errorf("创建经费报销失败: %w", err)
		}
		created++
	}
	fmt.Printf("    [OK] 经费报销 新增 %d 条\n", created)
	return nil
}

// seedStRating 年度评优。
func (c *ctx) seedStRating() error {
	const target = 1
	cur := count(c.db, &models.StRating{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] st_rating 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	fmt.Printf("  st_rating 补 %d 条\n", int64(target)-cur)

	if len(c.assoc) == 0 {
		return nil
	}
	assoc := c.assoc[0]
	rating := models.StRating{
		AssociationID:         assoc.ID,
		AcademicYear:          "2025-2026",
		DimensionActivity:     92,
		DimensionMemberActive: 88,
		DimensionFinance:      85,
		DimensionBrand:        90,
		DimensionSatisfaction: 87,
		WeightedScore:         88.6,
		Star:                  4,
		PublicVoteCount:       ptrInt(0),
		Status:                S3,
	}
	if err := c.db.Create(&rating).Error; err != nil {
		return fmt.Errorf("创建评优失败: %w", err)
	}
	fmt.Printf("    [OK] 评优 %s star=%d\n", assoc.Name, rating.Star)
	return nil
}

// reloadAssoc 重新加载社团缓存。
func (c *ctx) reloadAssoc() {
	c.assoc = nil
	c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.assoc)
}

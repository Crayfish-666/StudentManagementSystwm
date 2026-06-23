package main

import (
	"encoding/json"
	"fmt"
	"time"

	"student-system/internal/models"
)

// seedTY 为团员发展模块灌入测试数据。
// 规模：ty_branch +3、ty_application +8、ty_approval_record +12、ty_recommendation_*+2、
//       ty_cultivation_link +6、ty_cultivation_record +6、ty_course_record +6、ty_thought_report +4、
//       ty_development_object +2、ty_political_review +4、ty_development_meeting +2、
//       ty_probationary_record +3、ty_probationary_meeting +1、ty_member_roster +3。
func (c *ctx) seedTY() error {
	if err := c.seedTyBranch(); err != nil {
		return err
	}
	if err := c.seedTyApplication(); err != nil {
		return err
	}
	if err := c.seedTyRecommendation(); err != nil {
		return err
	}
	if err := c.seedTyCultivation(); err != nil {
		return err
	}
	if err := c.seedTyCourseAndReport(); err != nil {
		return err
	}
	if err := c.seedTyDevelopment(); err != nil {
		return err
	}
	if err := c.seedTyProbationary(); err != nil {
		return err
	}
	if err := c.seedTyMemberRoster(); err != nil {
		return err
	}
	return nil
}

// seedTyBranch 团支部：若不足 5 个，补足。
func (c *ctx) seedTyBranch() error {
	const target = 5
	cur := count(c.db, &models.TyBranch{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] ty_branch 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  ty_branch 补 %d 条\n", need)

	specs := []struct {
		Name                 string
		CollegeID            int64
		SecretaryStudentID   *int64
		ExpectedMemberCount  int
	}{
		{"计算机学院 2023 级第一团支部", c.colleges[0].ID, nil, 45},
		{"计算机学院 2023 级第二团支部", c.colleges[0].ID, nil, 42},
		{"电子工程学院 2024 级第一团支部", c.colleges[1%len(c.colleges)].ID, nil, 38},
		{"计算机学院 2024 级第一团支部", c.colleges[0].ID, nil, 40},
		{"电子工程学院 2023 级第一团支部", c.colleges[1%len(c.colleges)].ID, nil, 36},
	}
	created := 0
	for _, sp := range specs {
		if need <= 0 {
			break
		}
		var exists int64
		c.db.Model(&models.TyBranch{}).Where("name = ? AND is_deleted = 0", sp.Name).Count(&exists)
		if exists > 0 {
			continue
		}
		bizNo := nextBizNo(c.db, "TY")
		branch := models.TyBranch{
			BizNo:                bizNo,
			Name:                 sp.Name,
			CollegeID:            sp.CollegeID,
			SecretaryStudentID:   sp.SecretaryStudentID,
			ExpectedMemberCount:  sp.ExpectedMemberCount,
			EstablishedAt:        ptrTime(c.now.AddDate(-2, 0, 0)),
		}
		if err := c.db.Create(&branch).Error; err != nil {
			return fmt.Errorf("创建团支部 %s 失败: %w", sp.Name, err)
		}
		fmt.Printf("    [OK] 团支部 %s (biz_no=%s)\n", branch.Name, branch.BizNo)
		need--
		created++
	}
	c.reloadBranches()
	return nil
}

// seedTyApplication 入团申请：补 8 条，覆盖 S0/S1/S2/S3/S4 五个状态。
func (c *ctx) seedTyApplication() error {
	const target = 8
	cur := count(c.db, &models.TyApplication{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] ty_application 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  ty_application 补 %d 条\n", need)

	// 每个申请的状态 + 审批意见
	statuses := []string{S0, S1, S1, S2, S2, S3, S3, S4}
	opinions := []string{
		"本人自愿申请加入中国共产主义青年团，恳请组织审批。",
		"申请人思想积极，学习刻苦，符合入团基本条件。",
		"该生学习态度端正，群众基础良好，建议推荐。",
		"经院系团委复核，材料齐全，同意报校团委。",
		"校团委终审通过，同意接收为预备团员。",
		"校团委终审通过，正式成为共青团员。",
		"经审核，材料不完整，本次申请驳回。",
	}

	idx := 0
	used := map[int64]struct{}{}
	created := 0
	for need > 0 && idx < len(c.students)*2 {
		stu, ok := c.pickStudent(used)
		if !ok {
			break
		}
		used[stu.ID] = struct{}{}
		idx++

		// 一个学生不允许同时有 S1/S2 申请（DB 唯一约束 uniq_ty_app_pending）
		var pending int64
		c.db.Model(&models.TyApplication{}).Where("student_id = ? AND status IN ('S1','S2') AND is_deleted = 0", stu.ID).Count(&pending)
		if pending > 0 {
			continue
		}

		// 是否已存在草稿/历史
		var exists int64
		c.db.Model(&models.TyApplication{}).Where("student_id = ? AND is_deleted = 0", stu.ID).Count(&exists)
		if exists > 0 {
			continue
		}

		branchID := c.branches[idx%len(c.branches)].ID
		status := statuses[created%len(statuses)]
		bizNo := nextBizNo(c.db, "TY")
		familyJSON, _ := json.Marshal([]map[string]string{
			{"relation": "父", "name": "张父", "unit": "务工"},
			{"relation": "母", "name": "李母", "unit": "教师"},
		})

		app := models.TyApplication{
			BizNo:         bizNo,
			StudentID:     stu.ID,
			BranchID:      branchID,
			ApplyDate:     c.now.AddDate(0, -1, -idx),
			SelfStatement: genSelfStatement(600 + idx*30),
			FamilyMembers: string(familyJSON),
			RewardsPunish: "高三年级三好学生",
			Status:        status,
		}
		// 按状态回填审批字段 + 审批记录
		switch status {
		case S1:
			// 提交后等待班级初审
		case S2:
			counselorID := int64(1)
			app.CounselorOpinion = "该生思想积极，同意推荐。"
			app.CounselorUserID = &counselorID
			app.CounselorAt = ptrTime(c.now.AddDate(0, 0, -3))
		case S3:
			counselorID := int64(1)
			collegeID := int64(2)
			schoolID := int64(1)
			app.CounselorOpinion = "班级团支部初审通过。"
			app.CounselorUserID = &counselorID
			app.CounselorAt = ptrTime(c.now.AddDate(0, 0, -5))
			app.CollegeOpinion = "院系团委复核通过。"
			app.CollegeUserID = &collegeID
			app.CollegeAt = ptrTime(c.now.AddDate(0, 0, -3))
			app.SchoolOpinion = "校团委终审通过。正式接收为预备团员。"
			app.SchoolUserID = &schoolID
			app.SchoolAt = ptrTime(c.now.AddDate(0, 0, -1))
		case S4:
			collegeID := int64(2)
			app.CounselorOpinion = "初审通过。"
			app.CounselorUserID = &collegeID
			app.CounselorAt = ptrTime(c.now.AddDate(0, 0, -4))
			app.RejectReason = "思想汇报内容不达标，建议下学期重新申请。"
		}

		if err := c.db.Create(&app).Error; err != nil {
			return fmt.Errorf("创建入团申请失败: %w", err)
		}
		fmt.Printf("    [OK] 申请 student=%s status=%s biz_no=%s\n", stu.StudentNo, status, app.BizNo)

		// 同步生成审批记录
		if err := c.genTyApprovalRecords(app, opinions, status); err != nil {
			return err
		}
		need--
		created++
	}
	return nil
}

// genTyApprovalRecords 为申请生成审批流水记录（覆盖 S1→S2→S3 全过程 + S4 驳回）。
func (c *ctx) genTyApprovalRecords(app models.TyApplication, opinions []string, status string) error {
	stepDefs := []struct {
		Step     string
		Role     string
		Approver string
	}{
		{"college_initial", "R-COL-LEAGUE", "班主任王老师"},
		{"college_review", "R-COL-COUN", "院系团委李书记"},
		{"league_final", "R-SY-LEAGUE", "校团委张老师"},
	}
	occurred := c.now.AddDate(0, 0, -5)
	for i, s := range stepDefs {
		var toStatus string
		var result string
		switch status {
		case S0, S1:
			// 草稿 / 刚提交：所有步骤未发生
			return nil
		case S2:
			if i > 0 {
				return nil
			}
			toStatus, result = S2, "approve"
		case S3:
			toStatus, result = S3, "approve"
		case S4:
			if i == 1 {
				toStatus, result = S4, "reject"
			} else {
				return nil
			}
		default:
			return nil
		}
		opinion := opinions[(i+1)%len(opinions)]
		if len(opinion) < 5 {
			opinion = opinion + "（审批意见）"
		}
		approverID := int64(i + 1)
		rec := models.TyApprovalRecord{
			ApplicationID: app.ID,
			Module:        "application",
			TargetID:      app.ID,
			Step:          s.Step,
			ApproverID:    approverID,
			ApproverName:  s.Approver,
			ApproverRole:  s.Role,
			Result:        result,
			Opinion:       opinion,
			FromStatus:    S1,
			ToStatus:      toStatus,
			OccurredAt:    occurred,
		}
		if err := c.db.Create(&rec).Error; err != nil {
			return fmt.Errorf("创建审批记录失败: %w", err)
		}
		occurred = occurred.Add(time.Hour * 24)
	}
	return nil
}

// seedTyRecommendation 推优大会 + 投票。
func (c *ctx) seedTyRecommendation() error {
	const targetMeet = 2
	cur := count(c.db, &models.TyRecommendationMeeting{})
	if cur >= int64(targetMeet) {
		fmt.Printf("  [skip] ty_recommendation_meeting 已 %d 条 ≥ 目标 %d\n", cur, targetMeet)
		return nil
	}
	need := int64(targetMeet) - cur
	fmt.Printf("  ty_recommendation_meeting 补 %d 条\n", need)

	// 找出 S2 状态的申请（待推优）
	var apps []models.TyApplication
	c.db.Where("status = ? AND is_deleted = 0", S2).Order("id ASC").Limit(int(need)).Find(&apps)
	if len(apps) == 0 {
		fmt.Println("    [skip] 无 S2 申请，无法生成推优大会")
		return nil
	}

	for i, app := range apps {
		bizNo := nextBizNo(c.db, "TY")
		meeting := models.TyRecommendationMeeting{
			BizNo:          bizNo,
			ApplicationID:  app.ID,
			MeetingAt:      c.now.AddDate(0, 0, -i-2),
			Location:       "团支部活动室",
			ExpectedCount:  40,
			ActualCount:    35,
			Decision:       "pass",
			DecisionReason: "到会率 87.5%，符合 2/3 多数要求。",
		}
		if err := c.db.Create(&meeting).Error; err != nil {
			return fmt.Errorf("创建推优大会失败: %w", err)
		}

		// 投票明细
		vote := models.TyRecommendationVote{
			MeetingID:     meeting.ID,
			ApplicationID: app.ID,
			ApproveCount:  32,
			AgainstCount:  2,
			AbstainCount:  1,
		}
		if err := c.db.Create(&vote).Error; err != nil {
			return fmt.Errorf("创建投票明细失败: %w", err)
		}
		fmt.Printf("    [OK] 推优大会 biz_no=%s 申请=%d\n", meeting.BizNo, app.ID)
	}
	return nil
}

// seedTyCultivation 培养联系人 + 培养考察记录。
func (c *ctx) seedTyCultivation() error {
	// 培养联系人
	const targetLink = 6
	cur := count(c.db, &models.TyCultivationLink{})
	if cur < int64(targetLink) {
		need := int64(targetLink) - cur
		fmt.Printf("  ty_cultivation_link 补 %d 条\n", need)

		// 找 6 个非社团成员的学生
		var apps []models.TyApplication
		c.db.Where("status IN ? AND is_deleted = 0", []string{S2, S3}).Order("id ASC").Limit(int(need)).Find(&apps)
		for i, app := range apps {
			mentor := c.students[(i+5)%len(c.students)]
			link := models.TyCultivationLink{
				ApplicationID:   app.ID,
				MentorStudentID: mentor.ID,
				MentorType:      []string{"league_member", "party_member"}[i%2],
				StartAt:         c.now.AddDate(0, -3, 0),
				IsActive:        1,
			}
			if err := c.db.Create(&link).Error; err != nil {
				return fmt.Errorf("创建培养联系人失败: %w", err)
			}
		}
		fmt.Printf("    [OK] 培养联系人 新增 %d 条\n", len(apps))
	}

	// 培养考察记录
	const targetRec = 6
	cur = count(c.db, &models.TyCultivationRecord{})
	if cur >= int64(targetRec) {
		fmt.Printf("  [skip] ty_cultivation_record 已 %d 条 ≥ 目标 %d\n", cur, targetRec)
		return nil
	}
	fmt.Printf("  ty_cultivation_record 补 %d 条\n", int64(targetRec)-cur)

	var links []models.TyCultivationLink
	c.db.Where("is_deleted = 0 AND is_active = 1").Order("id ASC").Find(&links)
	if len(links) == 0 {
		return nil
	}

	summary := "本月该同志深入学习党的二十大精神，积极参与团支部组织的各项活动，主动向党组织靠拢。学习上刻苦努力，工作上认真负责，生活上勤俭节约乐于助人。"
	monthIdx := 1
	for i := 0; i < int(int64(targetRec)-cur) && i < len(links)*3; i++ {
		link := links[i%len(links)]
		month := ((monthIdx - 1) % 12) + 1
		year := 2025 + (monthIdx-1)/12
		rec := models.TyCultivationRecord{
			BizNo:            nextBizNo(c.db, "TY-CULT"),
			ApplicationID:    link.ApplicationID,
			RecordYear:       year,
			RecordMonth:      month,
			Summary:          summary,
			PerformanceScore: 80 + (i*3)%20,
			RecordType:       "monthly",
			RecordedBy:       ptrInt64(1),
		}
		if err := c.db.Create(&rec).Error; err != nil {
			return fmt.Errorf("创建培养记录失败: %w", err)
		}
		monthIdx++
	}
	fmt.Printf("    [OK] 培养记录 新增 %d 条\n", int(int64(targetRec)-cur))
	return nil
}

// seedTyCourseAndReport 团课记录 + 思想汇报。
func (c *ctx) seedTyCourseAndReport() error {
	// 团课
	const targetCourse = 6
	cur := count(c.db, &models.TyCourseRecord{})
	if cur < int64(targetCourse) {
		need := int64(targetCourse) - cur
		fmt.Printf("  ty_course_record 补 %d 条\n", need)
		semesters := []string{"2024-2025-1", "2024-2025-2", "2025-2026-1"}
		courses := []string{"共青团基础知识", "党的二十大精神宣讲", "志愿服务与公益实践"}
		for i := 0; i < int(need); i++ {
			stu := c.students[i%len(c.students)]
			score := 75 + i*4
			pass := 0
			if score >= 80 {
				pass = 1
			}
			rec := models.TyCourseRecord{
				StudentID:     stu.ID,
				CourseName:    courses[i%len(courses)],
				Semester:      semesters[i%len(semesters)],
				StudyAt:       c.now.AddDate(0, -i, 0),
				Score:         ptrInt(score),
				CertificateNo: fmt.Sprintf("TY-CERT-%04d", 1000+i),
				IsPass:        pass,
			}
			if err := c.db.Create(&rec).Error; err != nil {
				return fmt.Errorf("创建团课记录失败: %w", err)
			}
		}
		fmt.Printf("    [OK] 团课记录 新增 %d 条\n", need)
	}

	// 思想汇报
	const targetReport = 4
	cur = count(c.db, &models.TyThoughtReport{})
	if cur >= int64(targetReport) {
		fmt.Printf("  [skip] ty_thought_report 已 %d 条 ≥ 目标 %d\n", cur, targetReport)
		return nil
	}
	fmt.Printf("  ty_thought_report 补 %d 条\n", int64(targetReport)-cur)

	var apps []models.TyApplication
	c.db.Where("status IN ? AND is_deleted = 0", []string{S2, S3}).Order("id ASC").Limit(int(int64(targetReport)-cur)).Find(&apps)
	for i, app := range apps {
		content := "本季度我深入学习党的二十大精神，深刻领会习近平新时代中国特色社会主义思想的丰富内涵。在思想上，我积极向党组织靠拢；在学习上，我刻苦努力，成绩名列前茅；在生活上，我乐于助人，积极参与志愿服务。展望下一季度，我将继续努力，以实际行动争取早日加入团组织。团结就是力量，奋斗成就梦想！"
		for len([]rune(content)) < 1100 {
			content += "在团组织的关怀下，我不断成长进步。"
		}
		report := models.TyThoughtReport{
			ApplicationID: app.ID,
			StudentID:     app.StudentID,
			Title:         fmt.Sprintf("思想汇报 2025Q%d", i+1),
			Content:       content,
			Quarter:       fmt.Sprintf("2025Q%d", i%4+1),
			AISimilarity:  ptrFloat(0.08 + float64(i)*0.03),
			IsQualified:   1,
		}
		if err := c.db.Create(&report).Error; err != nil {
			return fmt.Errorf("创建思想汇报失败: %w", err)
		}
	}
	fmt.Printf("    [OK] 思想汇报 新增 %d 条\n", len(apps))
	return nil
}

// seedTyDevelopment 发展对象 + 政审 + 发展大会。
func (c *ctx) seedTyDevelopment() error {
	const targetDev = 2
	cur := count(c.db, &models.TyDevelopmentObject{})
	if cur < int64(targetDev) {
		need := int64(targetDev) - cur
		fmt.Printf("  ty_development_object 补 %d 条\n", need)

		// 找 S3 状态的申请
		var apps []models.TyApplication
		c.db.Where("status = ? AND is_deleted = 0", S3).Order("id ASC").Limit(int(need)).Find(&apps)
		for i, app := range apps {
			bizNo := nextBizNo(c.db, "TY")
			massAt := 15
			dev := models.TyDevelopmentObject{
				BizNo:                bizNo,
				ApplicationID:        app.ID,
				CourseCertNo:         fmt.Sprintf("TY-CERT-%04d", 2000+i),
				MentorOpinion:        "该同志自提出入党申请以来，思想上积极要求进步，认真学习马克思列宁主义、毛泽东思想、邓小平理论、\"三个代表\"重要思想、科学发展观，深入学习习近平新时代中国特色社会主义思想，思想政治素质明显提升。学习上刻苦努力，成绩始终名列班级前茅，多次获得校级奖学金，并积极参加各类学科竞赛取得优异成绩。工作上认真负责，热心为同学服务，得到师生一致好评。生活中勤俭节约、乐于助人，团结同学，积极参与志愿服务活动，每学期志愿服务时长超过 40 小时。群众基础扎实，培养期间表现优异，按时提交思想汇报。综合各方面表现，建议发展为发展对象。",
				CounselorOpinion:     "经辅导员认真审核，该同志自提出入党申请以来，思想政治素质优良，能自觉用习近平新时代中国特色社会主义思想武装头脑，主动学习党的路线方针政策。学习上刻苦努力，成绩始终保持班级前 5 名，获国家奖学金 1 次、校级奖学金 3 次。工作上认真负责，作为班级学习委员能出色完成各项任务，得到师生一致好评。生活中团结同学、乐于助人，积极参与社会实践与志愿服务活动，累计服务时长超过 200 小时。群众基础扎实，在班级民主测评中优秀率 95% 以上。自觉遵守校规校纪，无任何违规违纪记录。该同志符合发展对象标准，同意推荐。",
				MassMeetingAt:        ptrTime(c.now.AddDate(0, 0, -10)),
				MassMeetingAttendees: &massAt,
				PublicStart:          ptrTime(c.now.AddDate(0, 0, -9)),
				PublicEnd:            ptrTime(c.now.AddDate(0, 0, -3)),
				AutobiographyPath:    "/files/ty/autobiography_" + fmt.Sprintf("%d", app.ID) + ".pdf",
				Status:               S3,
			}
			if err := c.db.Create(&dev).Error; err != nil {
				return fmt.Errorf("创建发展对象失败: %w", err)
			}
			fmt.Printf("    [OK] 发展对象 biz_no=%s\n", dev.BizNo)
		}
	}

	// 政审记录（每位发展对象 2 份：本人 + 父母）
	const targetReview = 4
	cur = count(c.db, &models.TyPoliticalReview{})
	if cur >= int64(targetReview) {
		fmt.Printf("  [skip] ty_political_review 已 %d 条 ≥ 目标 %d\n", cur, targetReview)
		return nil
	}
	fmt.Printf("  ty_political_review 补 %d 条\n", int64(targetReview)-cur)

	var devs []models.TyDevelopmentObject
	c.db.Where("is_deleted = 0").Order("id ASC").Limit(2).Find(&devs)
	for _, dev := range devs {
		for _, rel := range []string{"self", "parent"} {
			if count(c.db, &models.TyPoliticalReview{}) >= int64(targetReview) {
				break
			}
			enc, _ := encryptIDCard("31011519900101" + fmt.Sprintf("%04d", dev.ID*10+int64(len(rel))))
			rev := models.TyPoliticalReview{
				DevelopmentID:  dev.ID,
				TargetRelation: rel,
				TargetName:     map[string]string{"self": "本人", "parent": "父母"}[rel],
				TargetIDCardEnc: enc,
				Method:         "letter",
				Conclusion:     "pass",
				DocumentPath:   "/files/ty/political_" + fmt.Sprintf("%d", dev.ID) + "_" + rel + ".pdf",
				IsExtend3M:     0,
			}
			if err := c.db.Create(&rev).Error; err != nil {
				return fmt.Errorf("创建政审记录失败: %w", err)
			}
		}
	}

	// 发展大会
	const targetMeet = 2
	cur = count(c.db, &models.TyDevelopmentMeeting{})
	if cur >= int64(targetMeet) {
		fmt.Printf("  [skip] ty_development_meeting 已 %d 条 ≥ 目标 %d\n", cur, targetMeet)
		return nil
	}
	fmt.Printf("  ty_development_meeting 补 %d 条\n", int64(targetMeet)-cur)

	devs = nil
	c.db.Where("is_deleted = 0 AND status = ?", S3).Order("id ASC").Find(&devs)
	for _, dev := range devs {
		if count(c.db, &models.TyDevelopmentMeeting{}) >= int64(targetMeet) {
			break
		}
		bizNo := nextBizNo(c.db, "TY")
		meet := models.TyDevelopmentMeeting{
			BizNo:             bizNo,
			DevelopmentID:     dev.ID,
			MeetingAt:         c.now.AddDate(0, 0, -2),
			ExpectedCount:     40,
			ActualCount:       38,
			ApproveCount:      36,
			AgainstCount:      1,
			AbstainCount:      1,
			Decision:          "pass",
			VolunteerFormPath: "/files/ty/volunteer_form_" + fmt.Sprintf("%d", dev.ID) + ".pdf",
		}
		if err := c.db.Create(&meet).Error; err != nil {
			return fmt.Errorf("创建发展大会失败: %w", err)
		}
		fmt.Printf("    [OK] 发展大会 biz_no=%s\n", meet.BizNo)
	}
	return nil
}

// seedTyProbationary 预备期考察 + 转正大会。
func (c *ctx) seedTyProbationary() error {
	// 预备期考察
	const targetProb = 3
	cur := count(c.db, &models.TyProbationaryRecord{})
	if cur >= int64(targetProb) {
		fmt.Printf("  [skip] ty_probationary_record 已 %d 条 ≥ 目标 %d\n", cur, targetProb)
	} else {
		fmt.Printf("  ty_probationary_record 补 %d 条\n", int64(targetProb)-cur)
		var apps []models.TyApplication
		c.db.Where("status = ? AND is_deleted = 0", S3).Order("id ASC").Limit(1).Find(&apps)
		for _, app := range apps {
			for q := 1; q <= 3; q++ {
				if count(c.db, &models.TyProbationaryRecord{}) >= int64(targetProb) {
					break
				}
				summary := fmt.Sprintf("预备期第%d季度考察：该同志继续加强理论学习，按时参加团组织生活和主题党日活动，思想政治素质稳步提升。深入学习习近平新时代中国特色社会主义思想，撰写学习笔记%d篇、心得体会%d篇。学习上勤奋刻苦，专业排名稳居班级前 5 名，获校级优秀学生干部、志愿服务先进个人等荣誉。工作中认真履职尽责，作为班级学习委员出色完成教学联络、学风建设等任务。生活上严于律己、团结同学，志愿服务时长累计达 %d 小时，主动参与社区防疫、敬老助残等公益活动，得到师生和群众的一致好评。", q, 2+q, 3+q, 20+q*5)
				rec := models.TyProbationaryRecord{
					ApplicationID: app.ID,
					RecordYear:    2025,
					RecordQuarter: q,
					Summary:       summary,
				}
				if err := c.db.Create(&rec).Error; err != nil {
					return fmt.Errorf("创建预备期记录失败: %w", err)
				}
			}
		}
	}

	// 转正大会
	const targetMeet = 1
	cur = count(c.db, &models.TyProbationaryMeeting{})
	if cur >= int64(targetMeet) {
		fmt.Printf("  [skip] ty_probationary_meeting 已 %d 条 ≥ 目标 %d\n", cur, targetMeet)
		return nil
	}
	fmt.Printf("  ty_probationary_meeting 补 %d 条\n", int64(targetMeet)-cur)

	var apps []models.TyApplication
	c.db.Where("status = ? AND is_deleted = 0", S3).Order("id ASC").Limit(1).Find(&apps)
	for _, app := range apps {
		bizNo := nextBizNo(c.db, "TY")
		meet := models.TyProbationaryMeeting{
			BizNo:               bizNo,
			ApplicationID:       app.ID,
			SelfApplicationPath: "/files/ty/probationary_" + fmt.Sprintf("%d", app.ID) + ".pdf",
			MeetingAt:           c.now.AddDate(0, 0, -1),
			ExpectedCount:       40,
			ActualCount:         38,
			ApproveCount:        37,
			Decision:            "pass",
			FormalJoinAt:        ptrTime(c.now),
		}
		if err := c.db.Create(&meet).Error; err != nil {
			return fmt.Errorf("创建转正大会失败: %w", err)
		}
		fmt.Printf("    [OK] 转正大会 biz_no=%s\n", meet.BizNo)
	}
	return nil
}

// seedTyMemberRoster 团员花名册。
func (c *ctx) seedTyMemberRoster() error {
	const target = 3
	cur := count(c.db, &models.TyMemberRoster{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] ty_member_roster 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	fmt.Printf("  ty_member_roster 补 %d 条\n", int64(target)-cur)

	// 拉取已存在的 student_id，避免唯一冲突
	existingStudent := map[int64]struct{}{}
	var existing []models.TyMemberRoster
	c.db.Where("is_deleted = 0").Find(&existing)
	for _, r := range existing {
		existingStudent[r.StudentID] = struct{}{}
	}

	var meets []models.TyProbationaryMeeting
	c.db.Where("is_deleted = 0 AND decision = 'pass'").Order("id ASC").Find(&meets)
	for _, meet := range meets {
		if count(c.db, &models.TyMemberRoster{}) >= int64(target) {
			break
		}
		var app models.TyApplication
		c.db.First(&app, meet.ApplicationID)
		if _, ok := existingStudent[app.StudentID]; ok {
			continue
		}
		bizNo := nextBizNo(c.db, "TY")
		keepUntil := c.now.AddDate(5, 0, 0)
		roster := models.TyMemberRoster{
			BizNo:                bizNo,
			StudentID:            app.StudentID,
			ApplicationID:        &app.ID,
			BranchID:             app.BranchID,
			JoinAt:               c.now,
			BecomeProbationaryAt: ptrTime(c.now.AddDate(0, -12, 0)),
			ArchiveKeepUntil:     &keepUntil,
			Status:               "active",
		}
		if err := c.db.Create(&roster).Error; err != nil {
			return fmt.Errorf("创建团员花名册失败: %w", err)
		}
		fmt.Printf("    [OK] 团员花名册 biz_no=%s\n", roster.BizNo)
	}
	return nil
}

// reloadBranches 重新加载团支部缓存。
func (c *ctx) reloadBranches() {
	c.branches = nil
	c.db.Where("is_deleted = 0").Order("id ASC").Find(&c.branches)
}

// genSelfStatement 生成 ≥ 500 字的思想汇报。
func genSelfStatement(n int) string {
	parts := []string{
		"我志愿加入中国共产主义青年团，坚决拥护中国共产党的领导，遵守团的章程，执行团的决议，履行团员义务。",
		"在思想上，我认真学习马克思列宁主义、毛泽东思想、邓小平理论、\"三个代表\"重要思想、科学发展观，深入学习习近平新时代中国特色社会主义思想。",
		"在学习上，我刻苦努力、成绩优良，连续多次获得奖学金，积极参加各类学科竞赛并取得优异成绩。",
		"在工作上，我担任班级学习委员，认真履行职责，热心为同学服务，得到老师和同学们的一致好评。",
		"在生活上，我勤俭节约、乐于助人，积极参与志愿服务活动，每学期志愿服务时长超过 30 小时。",
		"我深知自己距离一名合格的共青团员还有一定差距，今后我会更加严格要求自己，争取早日加入团组织。",
		"请团组织考验我，我将以实际行动证明自己的决心。请组织审批。",
	}
	out := ""
	for _, p := range parts {
		out += p
	}
	for len([]rune(out)) < n {
		out += "在团组织的培养和教育下，我不断成长进步。"
	}
	return out
}

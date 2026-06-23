package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"student-system/internal/models"
)

// seedSQ 为学生社区模块灌入测试数据。
// 规模：sq_selfgov_position +5、sq_inspection +6、sq_incident +5（覆盖 L1-L4）、
//       sq_late_return +6、sq_violation +3、sq_assessment +2。
func (c *ctx) seedSQ() error {
	if err := c.seedSqSelfgovPosition(); err != nil {
		return err
	}
	if err := c.seedSqInspection(); err != nil {
		return err
	}
	if err := c.seedSqIncident(); err != nil {
		return err
	}
	if err := c.seedSqLateReturn(); err != nil {
		return err
	}
	if err := c.seedSqViolation(); err != nil {
		return err
	}
	if err := c.seedSqAssessment(); err != nil {
		return err
	}
	return nil
}

// seedSqSelfgovPosition 自治职务：补 5 个，覆盖 building/floor/room 三种 scope + 多种 position。
func (c *ctx) seedSqSelfgovPosition() error {
	const target = 5
	cur := count(c.db, &models.SqSelfgovPosition{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] sq_selfgov_position 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  sq_selfgov_position 补 %d 条\n", need)

	// 楼栋长（楼栋级）
	if need > 0 && len(c.buildings) > 0 {
		stu := c.students[0]
		pos := models.SqSelfgovPosition{
			StudentID:   stu.ID,
			ScopeType:   "building",
			ScopeID:     c.buildings[0].ID,
			Position:    "building_chief",
			StartAt:     c.now.AddDate(0, -6, 0),
			PublicStart: ptrTime(c.now.AddDate(0, -9, 0)),
			PublicEnd:   ptrTime(c.now.AddDate(0, -6, 0)),
			Status:      "formal",
			AppointedBy: ptrInt64(1),
		}
		if err := c.db.Create(&pos).Error; err != nil {
			return fmt.Errorf("创建楼栋长失败: %w", err)
		}
		need--
		fmt.Printf("    [OK] 楼栋长 student=%s\n", stu.StudentNo)
	}

	// 楼层长（每栋 2 个）
	idx := 0
	for need > 0 && idx < 4 {
		b := c.buildings[idx%len(c.buildings)]
		var floor models.IdxDormFloor
		c.db.Where("building_id = ?", b.ID).Order("id ASC").First(&floor)
		stu := c.students[(idx+3)%len(c.students)]
		pos := models.SqSelfgovPosition{
			StudentID:   stu.ID,
			ScopeType:   "floor",
			ScopeID:     floor.ID,
			Position:    "floor_leader",
			StartAt:     c.now.AddDate(0, -3, 0),
			PublicStart: ptrTime(c.now.AddDate(0, -6, 0)),
			PublicEnd:   ptrTime(c.now.AddDate(0, -3, 0)),
			Status:      []string{"formal", "probation", "formal", "renewed"}[idx%4],
			AppointedBy: ptrInt64(1),
		}
		if err := c.db.Create(&pos).Error; err != nil {
			return fmt.Errorf("创建楼层长失败: %w", err)
		}
		need--
		idx++
	}
	fmt.Printf("    [OK] 楼层长 新增 %d 条\n", idx)
	return nil
}

// seedSqInspection 巡查记录：补 6 条 + 扣分项。
func (c *ctx) seedSqInspection() error {
	const target = 6
	cur := count(c.db, &models.SqInspection{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] sq_inspection 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  sq_inspection 补 %d 条\n", need)

	types := []string{"hygiene", "late_return", "appliance", "safety", "fire_lane", "hygiene"}
	items := [][]string{
		{"桌面不整洁", "床铺未叠"},
		{"晚归 1 人"},
		{"违规使用电热水壶"},
		{"插线板串联"},
		{"消防通道堆放杂物"},
		{"地面有垃圾"},
	}
	scores := []int{85, 75, 60, 90, 70, 80}
	created := 0
	for i := 0; i < int(need) && i < len(types); i++ {
		b := c.buildings[i%len(c.buildings)]
		var floor models.IdxDormFloor
		c.db.Where("building_id = ?", b.ID).Order("id ASC").First(&floor)
		var room models.IdxDormRoom
		c.db.Where("floor_id = ?", floor.ID).Order("id ASC").First(&room)

		insp := models.SqInspection{
			BizNo:            nextBizNo(c.db, "SQ"),
			InspectionType:  types[i],
			BuildingID:      b.ID,
			FloorID:         &floor.ID,
			RoomID:          &room.ID,
			InspectorUserID: 1,
			InspectedAt:     c.now.AddDate(0, 0, -i-1),
			Score:           ptrInt(scores[i]),
			Summary:         fmt.Sprintf("本次巡查类型：%s，发现问题 %d 项。", types[i], len(items[i])),
			Status:          "submitted",
		}
		if err := c.db.Create(&insp).Error; err != nil {
			return fmt.Errorf("创建巡查失败: %w", err)
		}
		// 扣分项
		for j, item := range items[i] {
			deduction := models.SqInspectionDeduction{
				InspectionID: insp.ID,
				Item:         item,
				Deduction:    5 + j*3,
			}
			if err := c.db.Create(&deduction).Error; err != nil {
				return fmt.Errorf("创建扣分项失败: %w", err)
			}
		}
		created++
	}
	fmt.Printf("    [OK] 巡查记录 新增 %d 条\n", created)
	return nil
}

// seedSqIncident 异常事件：补 5 条，覆盖 L1-L4 四个等级。
func (c *ctx) seedSqIncident() error {
	const target = 5
	cur := count(c.db, &models.SqIncident{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] sq_incident 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	need := int64(target) - cur
	fmt.Printf("  sq_incident 补 %d 条\n", need)

	specs := []struct {
		Level   string
		Type    string
		Status  string
		Action  string
		Closed  bool
		Desc    string
	}{
		{"L1", "报修", "closed", "已联系维修，当日修复完成。", true, "水管漏水，影响正常生活。"},
		{"L2", "违规电器", "closed", "没收电器，对当事人批评教育。", true, "宿舍内使用电热水壶，存在安全隐患。"},
		{"L3", "打架", "processing", "双方已隔离，辅导员介入调解。", false, "宿舍内两名学生因琐事发生肢体冲突。"},
		{"L3", "晚归", "closed", "已对当事人进行批评教育。", true, "学生深夜未归，超过宿舍关门时间 1 小时。"},
		{"L4", "夜不归宿", "closed", "已联系家长，学生次日返校，辅导员持续关注。", true, "学生整夜未归，存在严重安全隐患。"},
	}

	for i, sp := range specs {
		if i >= int(need) {
			break
		}
		b := c.buildings[i%len(c.buildings)]
		var floor models.IdxDormFloor
		c.db.Where("building_id = ?", b.ID).Order("id ASC").First(&floor)
		var room models.IdxDormRoom
		c.db.Where("floor_id = ?", floor.ID).Order("id ASC").First(&room)

		// 涉及学生（1-3 人）
		involved := []int64{}
		for j := 0; j < 1+rand.Intn(3); j++ {
			stu := c.students[(i*5+j)%len(c.students)]
			involved = append(involved, stu.ID)
		}
		involvedJSON, _ := json.Marshal(involved)

		bizNo := nextBizNo(c.db, "SQ")
		occ := c.now.AddDate(0, 0, -i-2)
		incident := models.SqIncident{
			BizNo:              bizNo,
			IncidentLevel:      sp.Level,
			IncidentType:       sp.Type,
			OccurredAt:         occ,
			BuildingID:         b.ID,
			FloorID:            &floor.ID,
			RoomID:             &room.ID,
			LocationDetail:     fmt.Sprintf("%s栋%d层%s室", b.Name, floor.FloorNo, room.RoomNo),
			ReporterUserID:     1,
			InvolvedStudentIDs: string(involvedJSON),
			InitialAction:      sp.Action,
			Status:             sp.Status,
		}
		if sp.Closed {
			incident.ClosedAt = ptrTime(c.now.AddDate(0, 0, -i-1))
			incident.ClosedBy = ptrInt64(1)
		}
		if err := c.db.Create(&incident).Error; err != nil {
			return fmt.Errorf("创建事件失败: %w", err)
		}

		// 处置记录
		action := models.SqIncidentAction{
			IncidentID: incident.ID,
			ActionText: "初步了解情况，做好记录。",
			ActionAt:   occ.Add(time.Hour),
			ActionBy:   1,
			IsFinal:    0,
		}
		c.db.Create(&action)
		if sp.Closed {
			final := models.SqIncidentAction{
				IncidentID: incident.ID,
				ActionText: sp.Action + "事件已结案。",
				ActionAt:   *incident.ClosedAt,
				ActionBy:   1,
				IsFinal:    1,
			}
			c.db.Create(&final)
		}
		fmt.Printf("    [OK] 事件 level=%s type=%s biz_no=%s\n", sp.Level, sp.Type, incident.BizNo)
	}
	return nil
}

// seedSqLateReturn 晚归：补 6 条。
func (c *ctx) seedSqLateReturn() error {
	const target = 6
	cur := count(c.db, &models.SqLateReturn{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] sq_late_return 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	fmt.Printf("  sq_late_return 补 %d 条\n", int64(target)-cur)

	reasons := []string{"外出就医", "社团活动晚归", "校外实习", "家中临时有事", "公交延误", "外出购物"}
	created := 0
	for i := 0; i < int(int64(target)-cur) && i < len(c.students); i++ {
		stu := c.students[(i*3)%len(c.students)]
		late := models.SqLateReturn{
			StudentID:  stu.ID,
			OccurredAt: c.now.AddDate(0, 0, -i-1),
			ReportedBy: 1,
			Reason:     reasons[i%len(reasons)],
			Semester:   "2025-2026-1",
		}
		if err := c.db.Create(&late).Error; err != nil {
			return fmt.Errorf("创建晚归记录失败: %w", err)
		}
		created++
	}
	fmt.Printf("    [OK] 晚归记录 新增 %d 条\n", created)
	return nil
}

// seedSqViolation 违规电器：补 3 条。
func (c *ctx) seedSqViolation() error {
	const target = 3
	cur := count(c.db, &models.SqViolation{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] sq_violation 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	fmt.Printf("  sq_violation 补 %d 条\n", int64(target)-cur)

	appliances := []string{"电热水壶", "电饭煲", "电吹风（高功率）"}
	statuses := []string{"warned", "reported_to_college", "warned"}
	created := 0
	for i := 0; i < int(int64(target)-cur) && i < len(appliances); i++ {
		stu := c.students[(i*5)%len(c.students)]
		var room models.IdxDormRoom
		c.db.Order("id ASC").First(&room)
		v := models.SqViolation{
			StudentID:     stu.ID,
			RoomID:        room.ID,
			ApplianceName: appliances[i],
			SeizedAt:      c.now.AddDate(0, 0, -i-1),
			ReportedBy:    1,
			Status:        statuses[i],
		}
		if err := c.db.Create(&v).Error; err != nil {
			return fmt.Errorf("创建违规记录失败: %w", err)
		}
		created++
	}
	fmt.Printf("    [OK] 违规电器 新增 %d 条\n", created)
	return nil
}

// seedSqAssessment 考核：补 2 条。
func (c *ctx) seedSqAssessment() error {
	const target = 2
	cur := count(c.db, &models.SqAssessment{})
	if cur >= int64(target) {
		fmt.Printf("  [skip] sq_assessment 已 %d 条 ≥ 目标 %d\n", cur, target)
		return nil
	}
	fmt.Printf("  sq_assessment 补 %d 条\n", int64(target)-cur)

	specs := []struct {
		Cycle   string
		Key     string
		Ins      int
		Inc      int
		Act      int
		Sat      int
		Bonus    int
		Rating   string
	}{
		{"monthly", "2026-05", 88, 75, 90, 85, 5, "excellent"},
		{"semester", "2025-2026-1", 82, 80, 88, 90, 0, "good"},
	}
	created := 0
	for _, sp := range specs {
		if count(c.db, &models.SqAssessment{}) >= int64(target) {
			break
		}
		stu := c.students[0]
		weighted := float64(sp.Ins)*0.3 + float64(sp.Inc)*0.2 + float64(sp.Act)*0.3 + float64(sp.Sat)*0.2 + float64(sp.Bonus)
		ass := models.SqAssessment{
			CycleType:         sp.Cycle,
			CycleKey:          sp.Key,
			TargetUserID:      stu.ID,
			ScoreInspection:   sp.Ins,
			ScoreIncident:     sp.Inc,
			ScoreActivity:     sp.Act,
			ScoreSatisfaction: sp.Sat,
			ScoreBonus:        sp.Bonus,
			WeightedScore:     weighted,
			Rating:            sp.Rating,
		}
		if err := c.db.Create(&ass).Error; err != nil {
			return fmt.Errorf("创建考核失败: %w", err)
		}
		created++
	}
	fmt.Printf("    [OK] 考核 新增 %d 条\n", created)
	return nil
}

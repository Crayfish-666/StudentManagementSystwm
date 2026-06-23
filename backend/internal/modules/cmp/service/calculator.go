// Package cmp 综合素质评分计算器：订阅 4 大模块事件 + 手动/定时重算入口。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/models"
	cmprepo "student-system/internal/modules/cmp/repository"
)

// Calculator 负责按规则版本聚合学生综合素质量化分。
type Calculator struct {
	db   *gorm.DB
	repo *cmprepo.ScoreRepository
	bus  *eventx.Bus
}

// NewCalculator 创建计算器。
func NewCalculator(db *gorm.DB, repo *cmprepo.ScoreRepository, bus *eventx.Bus) *Calculator {
	return &Calculator{db: db, repo: repo, bus: bus}
}

// DimensionRule 单条子项规则（rubric）。
type DimensionRule struct {
	SubItem string  `json:"sub_item"`
	Score   float64 `json:"score"`
	Weight  float64 `json:"weight"`
	Max     float64 `json:"max"`
}

// DimensionRuleGroup 单个维度的子项集合。
type DimensionRuleGroup struct {
	Dimension string          `json:"dimension"`
	Rules     []DimensionRule `json:"rules"`
}

// Rubric 完整规则集（与 cmp_rule_version.rules_json 对应）。
type Rubric struct {
	Weights    map[string]float64       `json:"weights"`    // 维度权重
	Dimensions []DimensionRuleGroup     `json:"dimensions"` // 子项打分规则
	Academic   map[string]float64       `json:"academic"`   // 学业占位（GPA→分值）
}

// ---- 默认评分规则（与 PRD §8.4 对齐）----

// defaultRubric 默认规则（与 cmp_rule_version.rules_json 默认内容一致）。
// 维度权重：团内 30 / 社团 25 / 社区 20 / 勤工 15 / 学业 10（合计 100）。
// 子项加总：30 + 25 + 20 + 15 + 10 = 100。
func defaultRubric() Rubric {
	return Rubric{
		Weights: map[string]float64{
			"league":    0.30,
			"assoc":     0.25,
			"community": 0.20,
			"workstudy": 0.15,
			"academic":  0.10,
		},
		Dimensions: []DimensionRuleGroup{
			{Dimension: "league", Rules: []DimensionRule{
				{SubItem: "团内身份", Score: 5, Weight: 0.05, Max: 5},
				{SubItem: "团内任职", Score: 10, Weight: 0.10, Max: 10},
				{SubItem: "团内活动参与", Score: 15, Weight: 0.15, Max: 15},
			}},
			{Dimension: "assoc", Rules: []DimensionRule{
				{SubItem: "社团任职", Score: 10, Weight: 0.10, Max: 10},
				{SubItem: "活动组织", Score: 10, Weight: 0.10, Max: 10},
				{SubItem: "评优获奖", Score: 5, Weight: 0.05, Max: 5},
			}},
			{Dimension: "community", Rules: []DimensionRule{
				{SubItem: "自治职务", Score: 5, Weight: 0.05, Max: 5},
				{SubItem: "巡查与事件处置", Score: 10, Weight: 0.10, Max: 10},
				{SubItem: "文明寝室", Score: 5, Weight: 0.05, Max: 5},
			}},
			{Dimension: "workstudy", Rules: []DimensionRule{
				{SubItem: "岗位履职", Score: 10, Weight: 0.10, Max: 10},
				{SubItem: "工时完成度", Score: 5, Weight: 0.05, Max: 5},
			}},
			{Dimension: "academic", Rules: []DimensionRule{
				{SubItem: "GPA/排名", Score: 10, Weight: 0.10, Max: 10},
			}},
		},
		Academic: map[string]float64{
			"gpa_per_100":  0.10, // 4.0 GPA → 100 分
			"rank_top_5":   10.0,
			"rank_top_20":  8.0,
			"rank_default": 5.0,
		},
	}
}

// CurrentAcademicYear 取当前学年字符串（如 "2025-2026"）。
// 按 9 月作为新学年开始点。
func CurrentAcademicYear() string {
	now := time.Now()
	year := now.Year()
	if now.Month() < time.September {
		year--
	}
	return fmt.Sprintf("%d-%d", year, year+1)
}

// Recompute 计算并持久化单个学生的综合分快照。
func (c *Calculator) Recompute(ctx context.Context, studentID int64, academicYear string) (*models.CmpScore, error) {
	if academicYear == "" {
		academicYear = CurrentAcademicYear()
	}

	// 1. 拉取当前激活的规则版本
	rule, err := c.repo.GetActiveRuleVersion()
	if err != nil {
		// 没有激活规则 → 用默认规则内存版
		rule = &models.CmpRuleVersion{
			Version:   "v-default",
			RulesJSON: mustJSON(defaultRubric()),
			IsActive:  1,
		}
	}

	var rubric Rubric
	if err := json.Unmarshal([]byte(rule.RulesJSON), &rubric); err != nil {
		return nil, fmt.Errorf("解析规则版本失败: %w", err)
	}

	// 2. 收集各维度子项明细
	details, total := c.aggregate(studentID, academicYear, rubric)
	// 注：不再做 total>100 截断，PRD §8.4 满分 = 100，由各子项 max 共同约束
	total = math.Round(total*100) / 100

	// 3. 写入 cmp_score + cmp_score_detail
	now := time.Now()
	score := &models.CmpScore{
		StudentID:     studentID,
		AcademicYear:  academicYear,
		TotalScore:    total,
		RuleVersionID: rule.ID,
		ComputedAt:    now,
	}
	if err := c.repo.UpsertScore(score); err != nil {
		return nil, fmt.Errorf("写入 cmp_score 失败: %w", err)
	}

	// 重新拉一次确保拿到 ID
	fresh, err := c.repo.GetByStudentYear(studentID, academicYear)
	if err != nil {
		return nil, err
	}
	score = fresh

	// 4. 替换明细
	if err := c.repo.ReplaceDetails(score.ID, details); err != nil {
		return nil, fmt.Errorf("写入 cmp_score_detail 失败: %w", err)
	}

	// 5. 写一条 CMP 事件（事件溯源）
	if c.bus != nil {
		_ = c.bus.Publish(&eventx.Event{
			Aggregate:   "cmp.score",
			AggregateID: fmt.Sprintf("%d-%s", studentID, academicYear),
			EventType:   "CmpScoreRecomputed",
			Module:      "CMP",
			ActorID:     0,
			ActorRole:   "system",
			Payload: map[string]interface{}{
				"student_id":     studentID,
				"academic_year":  academicYear,
				"total_score":    total,
				"rule_version":   rule.Version,
				"detail_count":   len(details),
				"recomputed_at":  now.Format(time.RFC3339),
			},
		})
	}

	return score, nil
}

// aggregate 聚合 4 大模块事件，返回明细切片 + 总分。
// 各维度子项数：league 3 + assoc 3 + community 3 + workstudy 2 + academic 1 = 12
func (c *Calculator) aggregate(studentID int64, academicYear string, rubric Rubric) ([]models.CmpScoreDetail, float64) {
	details := make([]models.CmpScoreDetail, 0, 12)
	now := time.Now()

	// === league 团内 ===
	leagueDetails, leagueTotal := c.scoreLeague(studentID, academicYear, rubric)
	details = append(details, leagueDetails...)

	// === assoc 社团 ===
	assocDetails, assocTotal := c.scoreAssoc(studentID, academicYear, rubric)
	details = append(details, assocDetails...)

	// === community 社区 ===
	commDetails, commTotal := c.scoreCommunity(studentID, academicYear, rubric)
	details = append(details, commDetails...)

	// === workstudy 勤工 ===
	workDetails, workTotal := c.scoreWorkstudy(studentID, academicYear, rubric)
	details = append(details, workDetails...)

	// === academic 学业（占位：未对接教务系统 → 取默认值）===
	acaDetails, acaTotal := c.scoreAcademic(studentID, academicYear, rubric)
	details = append(details, acaDetails...)

	total := leagueTotal + assocTotal + commTotal + workTotal + acaTotal

	// 关联 source_event_id = nil（聚合自多事件，不指单一）
	for i := range details {
		details[i].CreatedAt = now
		// ScoreID 在写入明细时再回填（score.ID 此刻尚未生成）
	}
	return details, total
}

// scoreLeague 团内维度：基于 idx_student.political_status + ty_member_roster.position_code + ty_thought_report 数量。
// 子项：团内身份 5 / 团内任职 10 / 团内活动参与 15（合计 30）。
func (c *Calculator) scoreLeague(studentID int64, academicYear string, rubric Rubric) ([]models.CmpScoreDetail, float64) {
	details := make([]models.CmpScoreDetail, 0, 3)

	// 子项 1: 团内身份（5/3/1/0）
	politicalStatus := ""
	var stu models.IdxStudent
	if err := c.db.Select("political_status").Where("id = ?", studentID).First(&stu).Error; err == nil {
		politicalStatus = stu.PoliticalStatus
	}
	idScore := 0.0
	idRaw := "非团员"
	switch politicalStatus {
	case "member":
		idScore, idRaw = 5.0, "共青团员"
	case "probationary":
		idScore, idRaw = 3.0, "预备团员"
	case "activist":
		idScore, idRaw = 1.0, "入团积极分子"
	}
	details = append(details, models.CmpScoreDetail{
		Dimension:    "league",
		SubItem:      "团内身份",
		SourceModule: "IDX",
		RawValue:     idRaw,
		Score:        idScore,
		Weight:       0.05,
	})

	// 子项 2: 团内任职（支部书记 10 / 委员 6 / 普通 0；schema 无 position_code 字段，简化为：在团 status=active 给 10）
	var rosterCount int64
	c.db.Model(&models.TyMemberRoster{}).
		Where("student_id = ? AND is_deleted = 0 AND status = 'active'", studentID).
		Count(&rosterCount)
	posScore := 0.0
	posRaw := "无任职"
	if rosterCount > 0 {
		posScore, posRaw = 10.0, fmt.Sprintf("在团组织 %d 个", rosterCount)
	}
	details = append(details, models.CmpScoreDetail{
		Dimension:    "league",
		SubItem:      "团内任职",
		SourceModule: "TY",
		RawValue:     posRaw,
		Score:        posScore,
		Weight:       0.10,
	})

	// 子项 3: 团内活动参与（思想汇报每篇 3 分，max 15）
	var reportCount int64
	c.db.Model(&models.TyThoughtReport{}).
		Where("student_id = ? AND is_deleted = 0", studentID).
		Count(&reportCount)
	reportScore := math.Min(float64(reportCount)*3, 15)
	details = append(details, models.CmpScoreDetail{
		Dimension:    "league",
		SubItem:      "团内活动参与",
		SourceModule: "TY",
		RawValue:     fmt.Sprintf("思想汇报 %d 篇", reportCount),
		Score:        reportScore,
		Weight:       0.15,
	})

	total := idScore + posScore + reportScore
	return details, total
}

// scoreAssoc 社团维度：基于 st_assoc_member + st_activity + st_rating。
// 子项：社团任职 10 / 活动组织 10 / 评优获奖 5（合计 25）。
func (c *Calculator) scoreAssoc(studentID int64, academicYear string, rubric Rubric) ([]models.CmpScoreDetail, float64) {
	details := make([]models.CmpScoreDetail, 0, 3)

	// 子项 1: 社团任职（PRD：会长 10 / 副会长 7 / 理事 4 / 会员 1；schema role='president'/'vice_president'/'director'，简化为有任职 10/0）
	var assocOfficerCount int64
	c.db.Model(&models.StAssocMember{}).
		Where("student_id = ? AND is_deleted = 0 AND role IN ('president','vice_president','director')", studentID).
		Count(&assocOfficerCount)
	officerScore := 0.0
	officerRaw := "普通成员"
	if assocOfficerCount > 0 {
		officerScore, officerRaw = 10.0, fmt.Sprintf("干部 %d 个社团", assocOfficerCount)
	}
	details = append(details, models.CmpScoreDetail{
		Dimension:    "assoc",
		SubItem:      "社团任职",
		SourceModule: "ST",
		RawValue:     officerRaw,
		Score:        officerScore,
		Weight:       0.10,
	})

	// 子项 2: 活动组织：本人所在社团（任一身份）发起的活动（PRD：立项数 × 1.5 + 总结评分均值）
	var activityCount int64
	c.db.Model(&models.StActivity{}).
		Where("is_deleted = 0 AND status = 'S3'").
		Where("association_id IN (SELECT association_id FROM st_assoc_member WHERE student_id = ? AND is_deleted = 0)", studentID).
		Count(&activityCount)
	orgScore := math.Min(float64(activityCount)*1.5, 10)
	details = append(details, models.CmpScoreDetail{
		Dimension:    "assoc",
		SubItem:      "活动组织",
		SourceModule: "ST",
		RawValue:     fmt.Sprintf("参与组织 %d 场", activityCount),
		Score:        orgScore,
		Weight:       0.10,
	})

	// 子项 3: 评优获奖：st_rating 是社团级别（无 student_id 列），无法精确到个人
	// 兜底：本人参加社团数 × 1，max 5
	var memberCount int64
	c.db.Model(&models.StAssocMember{}).
		Where("student_id = ? AND is_deleted = 0", studentID).
		Count(&memberCount)
	rateScore := math.Min(float64(memberCount), 5)
	details = append(details, models.CmpScoreDetail{
		Dimension:    "assoc",
		SubItem:      "评优获奖",
		SourceModule: "ST",
		RawValue:     fmt.Sprintf("参加 %d 个社团（st_rating 无 student_id 字段，临时兜底）", memberCount),
		Score:        rateScore,
		Weight:       0.05,
	})

	total := officerScore + orgScore + rateScore
	return details, total
}

// scoreCommunity 社区维度：sq_selfgov_position + sq_assessment + sq_inspection。
// 子项：自治职务 5 / 巡查与事件处置 10 / 文明寝室 5（合计 20）。
func (c *Calculator) scoreCommunity(studentID int64, academicYear string, rubric Rubric) ([]models.CmpScoreDetail, float64) {
	details := make([]models.CmpScoreDetail, 0, 3)

	// 子项 1: 自治职务（PRD：楼长 5 / 楼层长 3 / 寝室长 1；schema status='formal'/'probation'，简化为有任职 5/0）
	var posCount int64
	c.db.Model(&models.SqSelfgovPosition{}).
		Where("student_id = ? AND is_deleted = 0 AND status IN ('formal','probation')", studentID).
		Count(&posCount)
	posScore := 0.0
	posRaw := "无职务"
	if posCount > 0 {
		posScore, posRaw = 5.0, fmt.Sprintf("任职 %d 个自治岗位", posCount)
	}
	details = append(details, models.CmpScoreDetail{
		Dimension:    "community",
		SubItem:      "自治职务",
		SourceModule: "SQ",
		RawValue:     posRaw,
		Score:        posScore,
		Weight:       0.05,
	})

	// 子项 2: 巡查与事件处置：sq_assessment 月度平均考核 × 10 / 100
	// schema 用 target_user_id，需通过 sys_user.student_id 间接关联
	var avgScore float64
	row := c.db.Model(&models.SqAssessment{}).
		Select("COALESCE(AVG(weighted_score), 0)").
		Where("target_user_id IN (SELECT id FROM sys_user WHERE student_id = ? AND is_deleted = 0) AND is_deleted = 0", studentID).
		Row()
	if row != nil {
		_ = row.Scan(&avgScore)
	}
	assessScore := math.Min(avgScore/10, 10) // 100 分制 → 10 分
	details = append(details, models.CmpScoreDetail{
		Dimension:    "community",
		SubItem:      "巡查与事件处置",
		SourceModule: "SQ",
		RawValue:     fmt.Sprintf("平均考核 %.1f", avgScore),
		Score:        assessScore,
		Weight:       0.10,
	})

	// 子项 3: 文明寝室：基于 sq_inspection 平均分（PRD：优秀 5 / 合格 3 / 不合格 0）
	// schema：本人通过 idx_dorm_bed.occupant_student_id 关联到 room_id
	var roomAvg float64
	c.db.Model(&models.SqInspection{}).
		Select("COALESCE(AVG(score), 0)").
		Where("is_deleted = 0 AND room_id IN (SELECT room_id FROM idx_dorm_bed WHERE occupant_student_id = ? AND (move_out_at IS NULL OR move_out_at > date('now')) AND is_deleted = 0)", studentID).
		Scan(&roomAvg)
	civScore := 0.0
	civRaw := "无记录"
	switch {
	case roomAvg >= 90:
		civScore, civRaw = 5.0, fmt.Sprintf("优秀(%.1f)", roomAvg)
	case roomAvg >= 80:
		civScore, civRaw = 3.0, fmt.Sprintf("合格(%.1f)", roomAvg)
	case roomAvg > 0:
		civScore, civRaw = 0.0, fmt.Sprintf("不合格(%.1f)", roomAvg)
	}
	details = append(details, models.CmpScoreDetail{
		Dimension:    "community",
		SubItem:      "文明寝室",
		SourceModule: "SQ",
		RawValue:     civRaw,
		Score:        civScore,
		Weight:       0.05,
	})

	total := posScore + assessScore + civScore
	return details, total
}

// scoreWorkstudy 勤工维度：qg_position_apply + qg_payroll。
// 子项：岗位履职 10 / 工时完成度 5（合计 15）。
func (c *Calculator) scoreWorkstudy(studentID int64, academicYear string, rubric Rubric) ([]models.CmpScoreDetail, float64) {
	details := make([]models.CmpScoreDetail, 0, 2)

	// 子项 1: 岗位履职：qg_position_apply.on_job（PRD：月度考核均值 × 10 → 简化为在岗 10/0）
	// schema status='on_job' 或 'onboarding' 都算在岗
	var onJobCount int64
	c.db.Model(&models.QgPositionApply{}).
		Where("student_id = ? AND is_deleted = 0 AND status IN ('on_job','onboarding')", studentID).
		Count(&onJobCount)
	dutyScore := 0.0
	dutyRaw := "无在岗岗位"
	if onJobCount > 0 {
		dutyScore, dutyRaw = 10.0, fmt.Sprintf("在岗 %d 个", onJobCount)
	}
	details = append(details, models.CmpScoreDetail{
		Dimension:    "workstudy",
		SubItem:      "岗位履职",
		SourceModule: "QG",
		RawValue:     dutyRaw,
		Score:        dutyScore,
		Weight:       0.10,
	})

	// 子项 2: 工时完成度：实际工时 / 应达工时 × 5，max 5（PRD 公式）
	var totalHours float64
	c.db.Model(&models.QgPayroll{}).
		Select("COALESCE(SUM(total_hours), 0)").
		Where("student_id = ? AND is_deleted = 0", studentID).
		Scan(&totalHours)
	hoursScore := math.Min(totalHours/40*5, 5) // 假设月应达 40h
	details = append(details, models.CmpScoreDetail{
		Dimension:    "workstudy",
		SubItem:      "工时完成度",
		SourceModule: "QG",
		RawValue:     fmt.Sprintf("累计 %.1f h", totalHours),
		Score:        hoursScore,
		Weight:       0.05,
	})

	total := dutyScore + hoursScore
	return details, total
}

// scoreAcademic 学业维度：当前 V1 未对接教务系统，固定默认值。
// 子项：GPA/排名 10（合计 10）。
func (c *Calculator) scoreAcademic(studentID int64, academicYear string, rubric Rubric) ([]models.CmpScoreDetail, float64) {
	details := make([]models.CmpScoreDetail, 0, 1)

	// 占位：未对接教务 → 取默认值 10
	acaScore := 10.0
	acaRaw := "未对接教务系统（占位 10）"
	details = append(details, models.CmpScoreDetail{
		Dimension:    "academic",
		SubItem:      "GPA/排名",
		SourceModule: "EXTERNAL",
		RawValue:     acaRaw,
		Score:        acaScore,
		Weight:       0.10,
	})

	return details, acaScore
}

// RecomputeAll 全量重算（管理端 + 定时任务使用）。
func (c *Calculator) RecomputeAll(ctx context.Context, academicYear string) (int, error) {
	if academicYear == "" {
		academicYear = CurrentAcademicYear()
	}
	ids, err := c.repo.ListAllStudents()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, sid := range ids {
		if _, err := c.Recompute(ctx, sid, academicYear); err != nil {
			continue
		}
		count++
	}
	return count, nil
}

// RecomputeByCollege 院系级全量重算。
func (c *Calculator) RecomputeByCollege(ctx context.Context, collegeID int64, academicYear string) (int, error) {
	if academicYear == "" {
		academicYear = CurrentAcademicYear()
	}
	ids, err := c.repo.ListStudentsByCollege(collegeID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, sid := range ids {
		if _, err := c.Recompute(ctx, sid, academicYear); err != nil {
			continue
		}
		count++
	}
	return count, nil
}

// mustJSON 序列化失败回退为空 JSON。
func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

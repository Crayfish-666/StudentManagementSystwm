// Package cmp 综合素质（CMP）模块业务服务层。
package service

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/models"
	cmprepo "student-system/internal/modules/cmp/repository"
)

// ScoreService 综合分查询与重算服务。
type ScoreService struct {
	db         *gorm.DB
	repo       *cmprepo.ScoreRepository
	calc       *Calculator
	bus        *eventx.Bus
}

// NewScoreService 创建综合分服务。
func NewScoreService(db *gorm.DB, repo *cmprepo.ScoreRepository, calc *Calculator, bus *eventx.Bus) *ScoreService {
	return &ScoreService{db: db, repo: repo, calc: calc, bus: bus}
}

// ---- DTO ----

// ScoreListItem 排行/列表条目。
type ScoreListItem struct {
	ID            int64   `json:"id"`
	StudentID     int64   `json:"student_id"`
	StudentNo     string  `json:"student_no"`
	StudentName   string  `json:"student_name"`
	CollegeID     *int64  `json:"college_id,omitempty"`
	CollegeName   string  `json:"college_name"`
	ClassID       *int64  `json:"class_id,omitempty"`
	ClassName     string  `json:"college_class_name"`
	AcademicYear  string  `json:"academic_year"`
	TotalScore    float64 `json:"total_score"`
	RankInClass   *int    `json:"rank_in_class,omitempty"`
	RankInCollege *int    `json:"rank_in_college,omitempty"`
	ComputedAt    string  `json:"computed_at"`
}

// ScoreListResult 列表结果。
type ScoreListResult struct {
	Items    []ScoreListItem `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// ScoreDetailItem 子项明细。
type ScoreDetailItem struct {
	Dimension    string  `json:"dimension"`
	DimensionZh  string  `json:"dimension_zh"`
	SubItem      string  `json:"sub_item"`
	RawValue     string  `json:"raw_value"`
	Score        float64 `json:"score"`
	Max          float64 `json:"max"`
	Weight       float64 `json:"weight"`
	SourceModule string  `json:"source_module"`
}

// MyScoreView 学生本人综合分视图。
type MyScoreView struct {
	AcademicYear string           `json:"academic_year"`
	TotalScore   float64          `json:"total_score"`
	Dimensions   map[string]float64 `json:"dimensions"` // league/assoc/community/workstudy/academic → 分值
	Details      []ScoreDetailItem `json:"details"`
	RuleVersion  string           `json:"rule_version"`
	ComputedAt   string           `json:"computed_at"`
}

// ScoreDetailView 详情视图。
type ScoreDetailView struct {
	ID           int64            `json:"id"`
	StudentID    int64            `json:"student_id"`
	StudentNo    string           `json:"student_no"`
	StudentName  string           `json:"student_name"`
	CollegeName  string           `json:"college_name"`
	ClassName    string           `json:"college_class_name"`
	AcademicYear string           `json:"academic_year"`
	TotalScore   float64          `json:"total_score"`
	RankInClass  *int             `json:"rank_in_class,omitempty"`
	RankInCollege *int            `json:"rank_in_college,omitempty"`
	RuleVersion  string           `json:"rule_version"`
	Details      []ScoreDetailItem `json:"details"`
	ComputedAt   string           `json:"computed_at"`
}

var dimZh = map[string]string{
	"league":    "团内表现",
	"assoc":     "社团活动",
	"community": "社区履职",
	"workstudy": "勤工表现",
	"academic":  "学业",
}

// ---- 业务方法 ----

// MyScore 学生本人综合分（自动按当前学生身份查询）。
func (s *ScoreService) MyScore(userID int64, academicYear string) (*MyScoreView, error) {
	if academicYear == "" {
		academicYear = CurrentAcademicYear()
	}
	stu, err := s.getStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	score, err := s.repo.GetByStudentYear(stu.ID, academicYear)
	if err != nil {
		// 无记录则自动重算
		if _, recErr := s.calc.Recompute(context.Background(), stu.ID, academicYear); recErr != nil {
			return nil, recErr
		}
		score, err = s.repo.GetByStudentYear(stu.ID, academicYear)
		if err != nil {
			return nil, err
		}
	}
	details, err := s.repo.GetDetails(score.ID)
	if err != nil {
		return nil, err
	}
	// 兜底：历史脏数据可能导致 score 存在但 details 缺失（如 score_id 未回填的事故数据），
	// 此时前端会看到总分正常但维度分/雷达图为空。检测到 details 为空时自动重算一次。
	if len(details) == 0 {
		if _, recErr := s.calc.Recompute(context.Background(), stu.ID, academicYear); recErr != nil {
			return nil, recErr
		}
		score, err = s.repo.GetByStudentYear(stu.ID, academicYear)
		if err != nil {
			return nil, err
		}
		details, err = s.repo.GetDetails(score.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.toMyView(score, details, stu.ID), nil
}

// Get 详情（含明细）。
func (s *ScoreService) Get(studentID int64, academicYear string) (*ScoreDetailView, error) {
	if academicYear == "" {
		academicYear = CurrentAcademicYear()
	}
	score, err := s.repo.GetByStudentYear(studentID, academicYear)
	if err != nil {
		return nil, err
	}
	details, err := s.repo.GetDetails(score.ID)
	if err != nil {
		return nil, err
	}
	view := &ScoreDetailView{
		ID:           score.ID,
		StudentID:    score.StudentID,
		AcademicYear: score.AcademicYear,
		TotalScore:   score.TotalScore,
		RankInClass:  score.RankInClass,
		RankInCollege: score.RankInCollege,
		Details:      s.toDetailItems(details),
		ComputedAt:   score.ComputedAt.Format(time.RFC3339),
	}
	// 关联学生信息
	if stu, err := s.getStudentByID(score.StudentID); err == nil {
		view.StudentName = stu.Name
		view.StudentNo = stu.StudentNo
		if stu.CollegeID != nil {
			var col models.SysCollege
			if err := s.db.Select("name").Where("id = ?", *stu.CollegeID).First(&col).Error; err == nil {
				view.CollegeName = col.Name
			}
		}
		if stu.ClassID != nil {
			var cls models.IdxClass
			if err := s.db.Select("name").Where("id = ?", *stu.ClassID).First(&cls).Error; err == nil {
				view.ClassName = cls.Name
			}
		}
	}
	// 规则版本号
	if rule, err := s.repo.GetActiveRuleVersion(); err == nil {
		view.RuleVersion = rule.Version
	}
	return view, nil
}

// List 排行/列表。
func (s *ScoreService) List(academicYear string, collegeID, classID int64, page, pageSize int) (*ScoreListResult, error) {
	if academicYear == "" {
		academicYear = CurrentAcademicYear()
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 取原始分数列表
	scores, _, err := s.repo.ListByYear(academicYear, collegeID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 加载学生/院系/班级信息
	studentIDs := make([]int64, 0, len(scores))
	for _, sc := range scores {
		studentIDs = append(studentIDs, sc.StudentID)
	}
	studentMap := s.batchGetStudents(studentIDs)

	items := make([]ScoreListItem, 0, len(scores))
	for i, sc := range scores {
		item := ScoreListItem{
			ID:           sc.ID,
			StudentID:    sc.StudentID,
			AcademicYear: sc.AcademicYear,
			TotalScore:   sc.TotalScore,
			RankInClass:  sc.RankInClass,
			RankInCollege: sc.RankInCollege,
			ComputedAt:   sc.ComputedAt.Format(time.RFC3339),
		}
		if stu, ok := studentMap[sc.StudentID]; ok {
			item.StudentNo = stu.StudentNo
			item.StudentName = stu.Name
			item.CollegeID = stu.CollegeID
			item.ClassID = stu.ClassID
			if stu.CollegeID != nil {
				if col, ok := s.collegeMap()[*stu.CollegeID]; ok {
					item.CollegeName = col
				}
			}
			if stu.ClassID != nil {
				if cls, ok := s.classMap()[*stu.ClassID]; ok {
					item.ClassName = cls
				}
			}
		}
		// class 过滤
		if classID > 0 && (item.ClassID == nil || *item.ClassID != classID) {
			continue
		}
		_ = i
		items = append(items, item)
	}
	// 过滤后重算 total
	filteredTotal := int64(len(items))

	// 班级/院系内排名（轻量：内存内）
	s.computeRanks(items)

	return &ScoreListResult{
		Items:    items,
		Total:    filteredTotal,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// RecomputeOne 手动重算单学生。
func (s *ScoreService) RecomputeOne(ctx context.Context, studentID int64, academicYear string) (*ScoreDetailView, error) {
	if _, err := s.calc.Recompute(ctx, studentID, academicYear); err != nil {
		return nil, err
	}
	return s.Get(studentID, academicYear)
}

// RecomputeBatch 批量重算（admin 触发）。
func (s *ScoreService) RecomputeBatch(ctx context.Context, collegeID int64, academicYear string) (int, error) {
	if collegeID > 0 {
		return s.calc.RecomputeByCollege(ctx, collegeID, academicYear)
	}
	return s.calc.RecomputeAll(ctx, academicYear)
}

// RecomputeFromEvent 由事件订阅器调用：触发指定学生增量重算。
func (s *ScoreService) RecomputeFromEvent(ctx context.Context, studentID int64) error {
	_, err := s.calc.Recompute(ctx, studentID, CurrentAcademicYear())
	return err
}

// ---- 内部方法 ----

func (s *ScoreService) toMyView(score *models.CmpScore, details []models.CmpScoreDetail, studentID int64) *MyScoreView {
	dims := map[string]float64{
		"league":    0,
		"assoc":     0,
		"community": 0,
		"workstudy": 0,
		"academic":  0,
	}
	for _, d := range details {
		dims[d.Dimension] += d.Score
	}
	view := &MyScoreView{
		AcademicYear: score.AcademicYear,
		TotalScore:   score.TotalScore,
		Dimensions:   dims,
		Details:      s.toDetailItems(details),
		ComputedAt:   score.ComputedAt.Format(time.RFC3339),
	}
	if rule, err := s.repo.GetActiveRuleVersion(); err == nil {
		view.RuleVersion = rule.Version
	}
	return view
}

func (s *ScoreService) toDetailItems(details []models.CmpScoreDetail) []ScoreDetailItem {
	items := make([]ScoreDetailItem, 0, len(details))
	for _, d := range details {
		max := 0.0
		switch d.SubItem {
		case "团内身份":
			max = 5
		case "团内任职":
			max = 10
		case "团内活动参与":
			max = 15
		case "社团任职":
			max = 10
		case "活动组织":
			max = 10
		case "评优获奖":
			max = 5
		case "自治职务":
			max = 5
		case "巡查与事件处置":
			max = 10
		case "文明寝室":
			max = 5
		case "岗位履职":
			max = 10
		case "工时完成度":
			max = 5
		case "GPA/排名":
			max = 10
		}
		items = append(items, ScoreDetailItem{
			Dimension:    d.Dimension,
			DimensionZh:  dimZh[d.Dimension],
			SubItem:      d.SubItem,
			RawValue:     d.RawValue,
			Score:        d.Score,
			Max:          max,
			Weight:       d.Weight,
			SourceModule: d.SourceModule,
		})
	}
	return items
}

func (s *ScoreService) computeRanks(items []ScoreListItem) {
	// 按 (class_id, total_score DESC) 排名
	classMap := make(map[int64][]int)
	for i, it := range items {
		if it.ClassID != nil {
			classMap[*it.ClassID] = append(classMap[*it.ClassID], i)
		}
	}
	for _, idxs := range classMap {
		for r, idx := range idxs {
			rank := r + 1
			items[idx].RankInClass = &rank
		}
	}
	// 按 (college_id, total_score DESC) 排名
	collegeMap := make(map[int64][]int)
	for i, it := range items {
		if it.CollegeID != nil {
			collegeMap[*it.CollegeID] = append(collegeMap[*it.CollegeID], i)
		}
	}
	for _, idxs := range collegeMap {
		for r, idx := range idxs {
			rank := r + 1
			items[idx].RankInCollege = &rank
		}
	}
}

func (s *ScoreService) getStudentByUserID(userID int64) (*models.IdxStudent, error) {
	var u models.SysUser
	if err := s.db.Select("student_id").Where("id = ? AND is_deleted = 0", userID).First(&u).Error; err != nil {
		return nil, err
	}
	if u.StudentID == nil {
		return nil, errStudentNotLinked
	}
	return s.getStudentByID(*u.StudentID)
}

func (s *ScoreService) getStudentByID(id int64) (*models.IdxStudent, error) {
	var stu models.IdxStudent
	if err := s.db.Where("id = ? AND is_deleted = 0", id).First(&stu).Error; err != nil {
		return nil, err
	}
	return &stu, nil
}

func (s *ScoreService) batchGetStudents(ids []int64) map[int64]*models.IdxStudent {
	if len(ids) == 0 {
		return nil
	}
	var stus []models.IdxStudent
	if err := s.db.Where("id IN ? AND is_deleted = 0", ids).Find(&stus).Error; err != nil {
		return nil
	}
	res := make(map[int64]*models.IdxStudent, len(stus))
	for i := range stus {
		res[stus[i].ID] = &stus[i]
	}
	return res
}

func (s *ScoreService) collegeMap() map[int64]string {
	var cols []models.SysCollege
	if err := s.db.Select("id, name").Where("is_deleted = 0").Find(&cols).Error; err != nil {
		return nil
	}
	res := make(map[int64]string, len(cols))
	for _, c := range cols {
		res[c.ID] = c.Name
	}
	return res
}

func (s *ScoreService) classMap() map[int64]string {
	var cls []models.IdxClass
	if err := s.db.Select("id, name").Where("is_deleted = 0").Find(&cls).Error; err != nil {
		return nil
	}
	res := make(map[int64]string, len(cls))
	for _, c := range cls {
		res[c.ID] = c.Name
	}
	return res
}

var errStudentNotLinked = &strErr{msg: "当前账户未关联学生身份"}

// strErr 简易错误。
type strErr struct{ msg string }

func (e *strErr) Error() string { return e.msg }

// 抑制未使用
var _ = strconv.Itoa

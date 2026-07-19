package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/ty/repository"
)

// CultivationService 培养考察业务服务层。
type CultivationService struct {
	repo    *repository.CultivationRepository
	appRepo *repository.ApplicationRepository
	db      *gorm.DB
	bus     *eventx.Bus
}

// NewCultivationService 创建培养考察服务。
func NewCultivationService(
	repo *repository.CultivationRepository,
	appRepo *repository.ApplicationRepository,
	db *gorm.DB,
	bus *eventx.Bus,
) *CultivationService {
	return &CultivationService{
		repo:    repo,
		appRepo: appRepo,
		db:      db,
		bus:     bus,
	}
}

// ==================== 培养联系人管理 ====================

// MentorItem 单个培养联系人提交项。
type MentorItem struct {
	MentorStudentID int64  `json:"mentor_student_id" binding:"required"`
	MentorType      string `json:"mentor_type" binding:"required"` // league_member / party_member
}

// AssignMentorsRequest 分配培养联系人请求（PRD §4.3.4：必须为 2 位）。
// 注：mentors 数组长度校验放在 Service 层（错误码 2541），不放在 binding，
//     便于前端拿到"培养联系人数量须为 2 位"的明确业务错误码。
type AssignMentorsRequest struct {
	ApplicationID int64        `json:"application_id" binding:"required"`
	Mentors       []MentorItem `json:"mentors" binding:"required"`
	StartAt       string       `json:"start_at" binding:"required"`
}

// CultivationLinkView 培养联系人视图。
type CultivationLinkView struct {
	ID              int64   `json:"id"`
	ApplicationID   int64   `json:"application_id"`
	MentorStudentID int64   `json:"mentor_student_id"`
	MentorName      string  `json:"mentor_name"`
	MentorType      string  `json:"mentor_type"`
	StartAt         string  `json:"start_at"`
	EndAt           *string `json:"end_at,omitempty"`
	IsActive        int     `json:"is_active"`
	CreatedAt       string  `json:"created_at"`
}

// isLeagueMember 判断政治面貌是否属于"正式团员"。
func isLeagueMember(politicalStatus string) bool {
	switch strings.TrimSpace(politicalStatus) {
	case "member", "league_member", "共青团员":
		return true
	}
	return false
}

// isPartyMember 判断政治面貌是否属于"党员/预备党员"。
func isPartyMember(politicalStatus string) bool {
	switch strings.TrimSpace(politicalStatus) {
	case "party_member", "party_probationary", "中共党员", "预备党员":
		return true
	}
	return false
}

// AssignMentors 批量分配 2 位培养联系人（PRD §4.3.4）。
//
// 硬卡控规则：
//  1. 数量必须为 2 位（一次性提交，不支持单条覆盖）；
//  2. 优先从申请人所在团支部的"在册正式团员"中选任；
//  3. 当支部团员数 < 2 时，剩余名额可由党员补足；
//  4. 2 位培养联系人 mentor_student_id 不可重复；
//  5. 每位 mentor 的 political_status 必须与 mentor_type 对应（团员/党员）；
//  6. 申请单已存在在任联系人不允许重新分配（须先结束）。
func (s *CultivationService) AssignMentors(userID int64, req *AssignMentorsRequest) ([]CultivationLinkView, error) {
	// 卡控 1：数量必须为 2
	if len(req.Mentors) != 2 {
		return nil, fmt.Errorf("培养联系人数量须为 2 位")
	}

	// 校验申请存在
	app, err := s.appRepo.GetByID(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("入团申请不存在")
	}

	// 卡控 6：已存在在任联系人 → 不允许重新分配
	activeCount, err := s.repo.CountActiveLinks(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("查询在任联系人失败: %w", err)
	}
	if activeCount > 0 {
		return nil, fmt.Errorf("该申请已存在在任培养联系人，请先结束既有关系")
	}

	// 卡控 4：两位联系人不能为同一人
	if req.Mentors[0].MentorStudentID == req.Mentors[1].MentorStudentID {
		return nil, fmt.Errorf("两位培养联系人不能为同一人")
	}

	// 解析开始日期
	startAt, err := parseDate(req.StartAt)
	if err != nil {
		return nil, fmt.Errorf("开始日期格式错误")
	}

	// 卡控 2/3：统计支部在册正式团员数
	branchMembers, err := s.repo.CountBranchMembers(app.BranchID)
	if err != nil {
		return nil, fmt.Errorf("查询支部团员数失败: %w", err)
	}

	// 验证每位 mentor
	leagueChosen := 0
	partyChosen := 0
	mentorCache := make(map[int64]string) // student_id -> name，复用

	for idx, m := range req.Mentors {
		if m.MentorType != "league_member" && m.MentorType != "party_member" {
			return nil, fmt.Errorf("第 %d 位培养联系人类型无效，须为 league_member 或 party_member", idx+1)
		}

		mentor, err := s.repo.GetStudentByID(m.MentorStudentID)
		if err != nil {
			return nil, fmt.Errorf("第 %d 位培养联系人不存在", idx+1)
		}

		// 卡控 5：political_status 与 mentor_type 对应
		if m.MentorType == "league_member" && !isLeagueMember(mentor.PoliticalStatus) {
			return nil, fmt.Errorf("第 %d 位培养联系人民主党派为 %s，与「团员」类型不匹配", idx+1, mentor.PoliticalStatus)
		}
		if m.MentorType == "party_member" && !isPartyMember(mentor.PoliticalStatus) {
			return nil, fmt.Errorf("第 %d 位培养联系人民主党派为 %s，与「党员」类型不匹配", idx+1, mentor.PoliticalStatus)
		}

		if m.MentorType == "league_member" {
			leagueChosen++
		} else {
			partyChosen++
		}
		mentorCache[m.MentorStudentID] = mentor.Name
	}

	// 卡控 2/3：优先团员 → 党员名额 ≤ max(0, 2 - 支部团员数)
	maxPartyAllowed := int(2 - branchMembers)
	if maxPartyAllowed < 0 {
		maxPartyAllowed = 0
	}
	if partyChosen > maxPartyAllowed {
		return nil, fmt.Errorf("支部有 %d 名在册正式团员，培养联系人须优先从中选任（最多可选 %d 名党员）", branchMembers, maxPartyAllowed)
	}

	// 批量创建
	links := make([]models.TyCultivationLink, 0, 2)
	for _, m := range req.Mentors {
		links = append(links, models.TyCultivationLink{
			ApplicationID:   req.ApplicationID,
			MentorStudentID: m.MentorStudentID,
			MentorType:      m.MentorType,
			StartAt:         startAt,
			IsActive:        1,
		})
	}
	if err := s.repo.CreateLinksBulk(links); err != nil {
		return nil, fmt.Errorf("创建培养联系人失败: %w", err)
	}

	// 查回已创建的 2 条 link（按 created_at 倒序取最近 2 条 active）
	created, err := s.repo.ListLinks(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("查询已创建联系人失败: %w", err)
	}
	activeLinks := make([]models.TyCultivationLink, 0, 2)
	for _, l := range created {
		if l.IsActive == 1 {
			activeLinks = append(activeLinks, l)
		}
	}

	views := make([]CultivationLinkView, 0, 2)
	for _, l := range activeLinks {
		name := mentorCache[l.MentorStudentID]
		views = append(views, *s.toLinkView(l, name))
	}

	s.publishCultivationEvent("TyCultivationLinkAssigned", userID, map[string]interface{}{
		"application_id":      req.ApplicationID,
		"mentor_student_ids":  []int64{req.Mentors[0].MentorStudentID, req.Mentors[1].MentorStudentID},
		"mentor_types":        []string{req.Mentors[0].MentorType, req.Mentors[1].MentorType},
		"branch_id":           app.BranchID,
		"branch_member_count": branchMembers,
	})

	return views, nil
}

// EndMentor 结束培养关系。
func (s *CultivationService) EndMentor(id int64, userID int64) (*CultivationLinkView, error) {
	link, err := s.repo.GetLinkByID(id)
	if err != nil {
		return nil, fmt.Errorf("培养联系人记录不存在")
	}
	if link.IsActive == 0 {
		return nil, fmt.Errorf("该培养关系已结束")
	}

	if err := s.repo.EndMentor(id); err != nil {
		return nil, fmt.Errorf("结束培养关系失败: %w", err)
	}

	// 重新获取更新后的记录
	updated, err := s.repo.GetLinkByID(id)
	if err != nil {
		return nil, err
	}

	mentorName := ""
	if mentor, err := s.repo.GetStudentByID(updated.MentorStudentID); err == nil {
		mentorName = mentor.Name
	}

	s.publishCultivationEvent("TyCultivationLinkEnded", userID, map[string]interface{}{
		"link_id":        id,
		"application_id": updated.ApplicationID,
	})

	return s.toLinkView(*updated, mentorName), nil
}

// ListLinks 查询培养联系人列表。
func (s *CultivationService) ListLinks(applicationID int64) ([]CultivationLinkView, error) {
	links, err := s.repo.ListLinks(applicationID)
	if err != nil {
		return nil, err
	}

	views := make([]CultivationLinkView, 0, len(links))
	for _, link := range links {
		mentorName := ""
		if mentor, err := s.repo.GetStudentByID(link.MentorStudentID); err == nil {
			mentorName = mentor.Name
			}
		views = append(views, *s.toLinkView(link, mentorName))
	}
	return views, nil
}

// toLinkView 将培养联系人模型转为视图。
func (s *CultivationService) toLinkView(link models.TyCultivationLink, mentorName string) *CultivationLinkView {
	v := &CultivationLinkView{
		ID:              link.ID,
		ApplicationID:   link.ApplicationID,
		MentorStudentID: link.MentorStudentID,
		MentorName:      mentorName,
		MentorType:      link.MentorType,
		StartAt:         link.StartAt.Format("2006-01-02"),
		IsActive:        link.IsActive,
		CreatedAt:       link.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if link.EndAt != nil {
		t := link.EndAt.Format("2006-01-02")
		v.EndAt = &t
	}
	return v
}

// ==================== 培养记录管理 ====================

// CreateRecordRequest 创建培养记录请求。
type CreateRecordRequest struct {
	ApplicationID    int64  `json:"application_id" binding:"required"`
	RecordYear       int    `json:"record_year" binding:"required"`
	RecordMonth      int    `json:"record_month" binding:"required"`
	Summary          string `json:"summary" binding:"required"`
	PerformanceScore int    `json:"performance_score" binding:"required"`
	RecordType       string `json:"record_type"` // monthly / quarterly，默认 monthly
}

// RecordListResult 培养记录列表结果。
type RecordListResult struct {
	Items    []RecordView `json:"items"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// RecordView 培养记录视图。
// 培养记录通过 application_id 间接关联学生（SSOT §5.2.5），
// 此处展开 student_id/student_no/student_name 供前端直接展示，避免再次跨表查询。
type RecordView struct {
	ID               int64  `json:"id"`
	BizNo            string `json:"biz_no"`
	ApplicationID    int64  `json:"application_id"`
	StudentID        int64  `json:"student_id,omitempty"`
	StudentNo        string `json:"student_no,omitempty"`
	StudentName      string `json:"student_name,omitempty"`
	RecordYear       int    `json:"record_year"`
	RecordMonth      int    `json:"record_month"`
	Summary          string `json:"summary"`
	PerformanceScore int    `json:"performance_score"`
	RecordType       string `json:"record_type"`
	IsOverdue        int    `json:"is_overdue"`
	RecordedBy       *int64 `json:"recorded_by,omitempty"`
	RecordedByName   string `json:"recorded_by_name,omitempty"`
	CreatedAt        string `json:"created_at"`
}

// CreateRecord 创建月度/季度培养记录。
//
// 校验规则：
//   - summary ≥ 50 字
//   - performance_score 0-100
//   - 同月不重复
func (s *CultivationService) CreateRecord(userID int64, req *CreateRecordRequest) (*RecordView, error) {
	// 校验摘要字数 ≥ 50
	if utf8.RuneCountInString(req.Summary) < 50 {
		return nil, fmt.Errorf("培养记录摘要不足 50 字")
	}

	// 校验成绩范围 0-100
	if req.PerformanceScore < 0 || req.PerformanceScore > 100 {
		return nil, fmt.Errorf("成绩须在 0-100 之间")
	}

	// 校验月份范围
	if req.RecordMonth < 1 || req.RecordMonth > 12 {
		return nil, fmt.Errorf("月份须在 1-12 之间")
	}

	// 校验同月不重复
	exists, err := s.repo.CheckMonthlyRecordExists(req.ApplicationID, req.RecordYear, req.RecordMonth)
	if err != nil {
		return nil, fmt.Errorf("检查当月记录失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("该月份的培养记录已存在")
	}

	recordType := req.RecordType
	if recordType == "" {
		recordType = "monthly"
	}
	if recordType != "monthly" && recordType != "quarterly" {
		return nil, fmt.Errorf("记录类型无效，须为 monthly 或 quarterly")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "TY-CULT")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	recordedByID := userID
	record := models.TyCultivationRecord{
		BizNo:            bizNo,
		ApplicationID:    req.ApplicationID,
		RecordYear:       req.RecordYear,
		RecordMonth:      req.RecordMonth,
		Summary:          req.Summary,
		PerformanceScore: req.PerformanceScore,
		RecordType:       recordType,
		RecordedBy:       &recordedByID,
	}

	if err := s.repo.CreateRecord(&record); err != nil {
		return nil, fmt.Errorf("创建培养记录失败: %w", err)
	}

	s.publishCultivationEvent("TyCultivationRecordCreated", userID, map[string]interface{}{
		"record_id":      record.ID,
		"application_id": req.ApplicationID,
		"record_year":    req.RecordYear,
		"record_month":   req.RecordMonth,
	})

	return s.toRecordView(record), nil
}

// ListRecords 查询培养记录列表。
func (s *CultivationService) ListRecords(applicationID int64, page, pageSize int) (*RecordListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := s.repo.ListRecords(applicationID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]RecordView, 0, len(records))
	for _, r := range records {
		items = append(items, *s.toRecordView(r))
	}

	return &RecordListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// toRecordView 将培养记录模型转为视图。
// 顺路通过 application → student 补全学生信息（与 ListReports 保持一致的 N+1 模式）。
func (s *CultivationService) toRecordView(r models.TyCultivationRecord) *RecordView {
	v := &RecordView{
		ID:               r.ID,
		BizNo:            r.BizNo,
		ApplicationID:    r.ApplicationID,
		RecordYear:       r.RecordYear,
		RecordMonth:      r.RecordMonth,
		Summary:          r.Summary,
		PerformanceScore: r.PerformanceScore,
		RecordType:       r.RecordType,
		IsOverdue:        r.IsOverdue,
		RecordedBy:       r.RecordedBy,
		CreatedAt:        r.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if r.RecordedBy != nil {
		if user, err := s.repo.GetUserByID(*r.RecordedBy); err == nil {
			v.RecordedByName = user.DisplayName
		}
	}
	// 补全学生信息：application.student_id → idx_student
	if app, err := s.appRepo.GetByID(r.ApplicationID); err == nil {
		v.StudentID = app.StudentID
		if student, err := s.repo.GetStudentByID(app.StudentID); err == nil {
			v.StudentName = student.Name
			v.StudentNo = student.StudentNo
		}
	}
	return v
}

// ==================== 团课记录管理 ====================

// CreateCourseRequest 创建团课记录请求。
// 注意：StudentID 提交者必填；学生本人提交时由后端从 user.student_id 注入，
// 管理员/团支书/院系/校级可显式传入 student_id 代填。
type CreateCourseRequest struct {
	StudentID     *int64 `json:"student_id"` // 学生本人由后端注入；管理员/教师可显式代填
	CourseName    string `json:"course_name" binding:"required"`
	Semester      string `json:"semester" binding:"required"`
	StudyAt       string `json:"study_at" binding:"required"`
	Score         *int   `json:"score"`
	CertificateNo string `json:"certificate_no"`
}

// CourseListResult 团课记录列表结果。
type CourseListResult struct {
	Items    []CourseView `json:"items"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// CourseView 团课记录视图。
type CourseView struct {
	ID            int64  `json:"id"`
	StudentID     int64  `json:"student_id"`
	StudentNo     string `json:"student_no,omitempty"`
	StudentName   string `json:"student_name"`
	CourseName    string `json:"course_name"`
	Semester      string `json:"semester"`
	StudyAt       string `json:"study_at"`
	Score         *int   `json:"score,omitempty"`
	CertificateNo string `json:"certificate_no"`
	IsPass        int    `json:"is_pass"`
	CreatedAt     string `json:"created_at"`
}

// CreateCourse 创建团课记录。
//
// 数据范围：
//   - 学生 (R-STU-NORM / R-STU-LEAGUE)：只能为自己添加，student_id 由后端从 user 注入；
//     若前端显式传入与本人不一致 → 拒绝。
//   - 管理员/团支书/院系/校级 (R-SY-* / R-COL-LEAGUE / R-COL-COUN / R-STU-LEAGUE)：
//     可显式传入 student_id 代他人录入；不传则按 user.student_id 注入。
func (s *CultivationService) CreateCourse(userID int64, req *CreateCourseRequest) (*CourseView, error) {
	// 解析角色
	roles, _ := s.findUserRoles(userID)
	isAdmin := hasAny(roles, "R-SY-ADMIN", "R-SY-LEAGUE", "R-COL-LEAGUE", "R-COL-COUN", "R-STU-LEAGUE")
	isStudent := hasAny(roles, "R-STU-NORM", "R-STU-LEAGUE") && !isAdmin
	// 注：R-STU-LEAGUE 既是学生又是团支书 → 当作管理员优先级更高

	// 解析"绑定学生"targetStudentID
	targetStudentID := int64(0)
	if req.StudentID != nil && *req.StudentID > 0 {
		// 学生显式传：必须等于本人
		if isStudent {
			user, err := s.repo.GetUserByID(userID)
			if err != nil || user == nil || user.StudentID == nil {
				return nil, fmt.Errorf("当前账号未关联学生身份，无法添加团课记录")
			}
			if *req.StudentID != *user.StudentID {
				return nil, fmt.Errorf("学生只能为自己添加团课记录")
			}
			targetStudentID = *user.StudentID
		} else {
			// 管理员/教师：直接用
			targetStudentID = *req.StudentID
		}
	} else {
		// 未传：从登录用户注入
		user, err := s.repo.GetUserByID(userID)
		if err != nil || user == nil || user.StudentID == nil {
			return nil, fmt.Errorf("当前账号未关联学生身份，无法添加团课记录")
		}
		targetStudentID = *user.StudentID
	}

	// 校验学生存在
	student, err := s.repo.GetStudentByID(targetStudentID)
	if err != nil {
		return nil, fmt.Errorf("学生信息不存在")
	}

	// 校验成绩范围（如有提供）
	if req.Score != nil && (*req.Score < 0 || *req.Score > 100) {
		return nil, fmt.Errorf("团课成绩须在 0-100 之间")
	}

	studyAt, err := parseDate(req.StudyAt)
	if err != nil {
		return nil, fmt.Errorf("学习日期格式错误")
	}

	course := models.TyCourseRecord{
		StudentID:     targetStudentID,
		CourseName:    req.CourseName,
		Semester:      req.Semester,
		StudyAt:       studyAt,
		Score:         req.Score,
		CertificateNo: req.CertificateNo,
	}

	if err := s.repo.CreateCourse(&course); err != nil {
		return nil, fmt.Errorf("创建团课记录失败: %w", err)
	}

	s.publishCultivationEvent("TyCourseRecordCreated", userID, map[string]interface{}{
		"course_id":  course.ID,
		"student_id": targetStudentID,
		"course_name": req.CourseName,
	})

	view := s.toCourseView(course)
	view.StudentName = student.Name
	view.StudentNo = student.StudentNo
	return view, nil
}

// ListCourses 查询团课列表。
//
// 数据范围隔离（与 ListReports 对齐）：
//   - 学生 (R-STU-NORM / R-STU-LEAGUE)：仅可查看自己的
//   - 辅导员 (R-COL-COUN)（未兼校/院系高级角色）：仅可查看本专业学生的
//   - 校/院系管理员及以上：可查看全部
// 前端可显式传 student_id 进一步收窄。
func (s *CultivationService) ListCourses(userID int64, studentID int64, page, pageSize int) (*CourseListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析角色与数据范围
	roles, _ := s.findUserRoles(userID)
	isStudent := hasAny(roles, "R-STU-NORM", "R-STU-LEAGUE")

	// 学生身份：强制锁死为本人（防越权查询）
	if isStudent {
		user, _ := s.repo.GetUserByID(userID)
		if user == nil || user.StudentID == nil {
			return &CourseListResult{Items: []CourseView{}, Total: 0, Page: page, PageSize: pageSize}, nil
		}
		studentID = *user.StudentID
	}

	// TODO(master bug): service 已实现 studentIDs/majorIDs 角色数据范围隔离，
	// 但 repository.ListCourses 尚未同步扩展签名。当前最小回退：仅传 studentID。
	// 完整修复需同步扩展 repo.ListCourses 接收 studentIDs/majorIDs。
	courses, total, err := s.repo.ListCourses(studentID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]CourseView, 0, len(courses))
	for _, c := range courses {
		view := s.toCourseView(c)
		if student, err := s.repo.GetStudentByID(c.StudentID); err == nil {
			view.StudentName = student.Name
			view.StudentNo = student.StudentNo
		}
		items = append(items, *view)
	}

	return &CourseListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdatePassStatus 更新团课结业状态（score >= 80 自动通过）。
// 权限：仅校/院系管理员、辅导员、团支书可标记结业；学生本人不能改自己的结业状态。
func (s *CultivationService) UpdatePassStatus(id int64, userID int64) (*CourseView, error) {
	roles, _ := s.findUserRoles(userID)
	if !hasAny(roles, "R-SY-ADMIN", "R-SY-LEAGUE", "R-COL-LEAGUE", "R-COL-COUN", "R-STU-LEAGUE") {
		return nil, fmt.Errorf("仅管理员或教师可标记团课结业")
	}

	course, err := s.repo.GetCourseByID(id)
	if err != nil {
		return nil, fmt.Errorf("团课记录不存在")
	}

	isPass := 0
	if course.Score != nil && *course.Score >= 80 {
		isPass = 1
	}

	if err := s.repo.UpdateCoursePassStatus(id, isPass); err != nil {
		return nil, fmt.Errorf("更新结业状态失败: %w", err)
	}

	updated, err := s.repo.GetCourseByID(id)
	if err != nil {
		return nil, err
	}

	s.publishCultivationEvent("TyCourseRecordPassed", userID, map[string]interface{}{
		"course_id": id,
		"is_pass":   isPass,
	})

	view := s.toCourseView(*updated)
	if student, err := s.repo.GetStudentByID(updated.StudentID); err == nil {
		view.StudentName = student.Name
		view.StudentNo = student.StudentNo
	}
	return view, nil
}

// UpdateCourseRequest 更新团课记录请求。
type UpdateCourseRequest struct {
	Score         *int   `json:"score"`
	CertificateNo string `json:"certificate_no"`
	IsPass        *int   `json:"is_pass"`
}

// UpdateCourse 更新团课记录（成绩、证书编号、结业状态）。
// 权限：仅管理员/教师可编辑；学生本人不能改。
func (s *CultivationService) UpdateCourse(id int64, userID int64, req *UpdateCourseRequest) (*CourseView, error) {
	roles, _ := s.findUserRoles(userID)
	if !hasAny(roles, "R-SY-ADMIN", "R-SY-LEAGUE", "R-COL-LEAGUE", "R-COL-COUN", "R-STU-LEAGUE") {
		return nil, fmt.Errorf("仅管理员或教师可编辑团课记录")
	}

	_, err := s.repo.GetCourseByID(id)
	if err != nil {
		return nil, fmt.Errorf("团课记录不存在")
	}

	updates := map[string]interface{}{}
	if req.Score != nil {
		if *req.Score < 0 || *req.Score > 100 {
			return nil, fmt.Errorf("团课成绩须在 0-100 之间")
		}
		updates["score"] = *req.Score
	}
	if req.CertificateNo != "" {
		updates["certificate_no"] = req.CertificateNo
	}
	if req.IsPass != nil {
		updates["is_pass"] = *req.IsPass
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("没有需要更新的字段")
	}

	if err := s.repo.UpdateCourse(id, updates); err != nil {
		return nil, fmt.Errorf("更新团课记录失败: %w", err)
	}

	s.publishCultivationEvent("TyCourseRecordUpdated", userID, map[string]interface{}{
		"course_id": id,
	})

	updated, err := s.repo.GetCourseByID(id)
	if err != nil {
		return nil, err
	}

	view := s.toCourseView(*updated)
	if student, err := s.repo.GetStudentByID(updated.StudentID); err == nil {
		view.StudentName = student.Name
		view.StudentNo = student.StudentNo
	}
	return view, nil
}

// toCourseView 将团课记录模型转为视图。
func (s *CultivationService) toCourseView(c models.TyCourseRecord) *CourseView {
	v := &CourseView{
		ID:            c.ID,
		StudentID:     c.StudentID,
		CourseName:    c.CourseName,
		Semester:      c.Semester,
		StudyAt:       c.StudyAt.Format("2006-01-02"),
		Score:         c.Score,
		CertificateNo: c.CertificateNo,
		IsPass:        c.IsPass,
		CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if student, err := s.repo.GetStudentByID(c.StudentID); err == nil {
		v.StudentName = student.Name
		v.StudentNo = student.StudentNo
	}
	return v
}

// ==================== 思想汇报管理 ====================

// CreateReportRequest 提交思想汇报请求。
// 注意：StudentID 不再由前端传入，提交者必须登录且绑定学生身份，后端自动注入。
// 若提交者具有 R-SY-ADMIN / R-SY-LEAGUE / R-COL-LEAGUE 角色（如团支书代填），可显式传入 StudentID。
type CreateReportRequest struct {
	ApplicationID int64   `json:"application_id" binding:"required"`
	StudentID     *int64  `json:"student_id,omitempty"` // 可选：仅管理员/院系团委可代填
	Title         string  `json:"title" binding:"required"`
	Content       string  `json:"content" binding:"required"`
	Quarter       string  `json:"quarter" binding:"required"` // 格式：2026Q1
}

// ReportListResult 思想汇报列表结果。
type ReportListResult struct {
	Items    []ReportView `json:"items"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// ReportView 思想汇报视图。
type ReportView struct {
	ID            int64    `json:"id"`
	BizNo         string   `json:"biz_no"`
	ApplicationID int64    `json:"application_id"`
	StudentID     int64    `json:"student_id"`
	StudentName   string   `json:"student_name"`
	StudentNo     string   `json:"student_no,omitempty"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Quarter       string   `json:"quarter"`
	AISimilarity  *float64 `json:"ai_similarity,omitempty"`
	IsQualified   int      `json:"is_qualified"`
	CreatedAt     string   `json:"created_at"`
}

// CreateReport 提交思想汇报。
//
// 校验规则：
//   - content ≥ 1000 字
//   - quarter 格式为 YYYYQn（如 2026Q1）
//   - 同季度不重复
//   - AI 查重预留字段 V1 设为 0 即合格
//   - 提交者必须登录，且若为学生则 student_id 由后端从 user 注入；申请单必须属于该学生
//   - 校/院系管理员可显式指定 student_id 代填
func (s *CultivationService) CreateReport(userID int64, req *CreateReportRequest) (*ReportView, error) {
	// 校验字数 ≥ 1000
	if utf8.RuneCountInString(req.Content) < 1000 {
		return nil, fmt.Errorf("思想汇报内容不足 1000 字")
	}

	// 校验 quarter 格式（YYYYQn）
	if !isValidQuarter(req.Quarter) {
		return nil, fmt.Errorf("季度格式错误，须为 YYYYQn 格式（如 2026Q1）")
	}

	// 校验同季度不重复
	exists, err := s.repo.CheckQuarterlyReportExists(req.ApplicationID, req.Quarter)
	if err != nil {
		return nil, fmt.Errorf("检查季度汇报失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("该季度的思想汇报已提交")
	}

	// 解析"汇报人"studentID：学生本人提交时由后端从 user 注入；管理员可显式传入
	roles, _ := s.findUserRoles(userID)
	isAdmin := hasAny(roles, "R-SY-ADMIN", "R-SY-LEAGUE", "R-COL-LEAGUE", "R-STU-LEAGUE")
	targetStudentID := int64(0)
	if req.StudentID != nil && *req.StudentID > 0 {
		if !isAdmin {
			return nil, fmt.Errorf("仅校/院系管理员或团支书可代他人提交思想汇报")
		}
		targetStudentID = *req.StudentID
	} else {
		// 学生本人提交：从登录用户的 student_id 字段注入
		user, err := s.repo.GetUserByID(userID)
		if err != nil || user == nil || user.StudentID == nil {
			return nil, fmt.Errorf("当前账号未关联学生身份，无法提交思想汇报")
		}
		targetStudentID = *user.StudentID
	}

	// 校验学生存在
	student, err := s.repo.GetStudentByID(targetStudentID)
	if err != nil {
		return nil, fmt.Errorf("汇报人学生信息不存在")
	}

	// 校验入团申请存在且属于该学生（防越权代填）
	app, err := s.appRepo.GetByID(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("入团申请不存在")
	}
	if app.StudentID != targetStudentID {
		return nil, fmt.Errorf("该入团申请不属于当前汇报人，请检查入团申请单")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "TY-RPT")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// V1 版本：AI 查重占位，默认设为 0 表示合格
	aiSimilarity := 0.0

	report := models.TyThoughtReport{
		BizNo:         bizNo,
		ApplicationID: req.ApplicationID,
		StudentID:     targetStudentID,
		Title:         req.Title,
		Content:       req.Content,
		Quarter:       req.Quarter,
		AISimilarity:  &aiSimilarity,
		IsQualified:   1, // V1 默认合格
	}

	if err := s.repo.CreateReport(&report); err != nil {
		return nil, fmt.Errorf("提交思想汇报失败: %w", err)
	}

	s.publishCultivationEvent("TyThoughtReportSubmitted", userID, map[string]interface{}{
		"report_id":      report.ID,
		"application_id": req.ApplicationID,
		"student_id":     targetStudentID,
		"quarter":        req.Quarter,
	})

	view := s.toReportView(report)
	view.StudentName = student.Name
	view.StudentNo = student.StudentNo
	return view, nil
}

// ListReports 查询思想汇报列表。
//
// 数据范围隔离（BR-TY-06 扩展）：
//   - 学生 (R-STU-NORM / R-STU-LEAGUE)：仅可查看自己提交的思想汇报
//   - 辅导员 (R-COL-COUN)（未兼校/院系高级角色）：仅可查看本专业学生提交的思想汇报
//   - 校/院系管理员及以上：可查看全部
func (s *CultivationService) ListReports(userID int64, applicationID int64, page, pageSize int) (*ReportListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析角色与数据范围
	roles, _ := s.findUserRoles(userID)
	var studentIDs, majorIDs []int64

	isSuperAdmin := hasAny(roles, "R-SY-ADMIN", "R-SY-LEAGUE", "R-COL-LEAGUE")
	isCounselor := hasAny(roles, "R-COL-COUN")
	isStudent := hasAny(roles, "R-STU-NORM", "R-STU-LEAGUE")

	if isSuperAdmin {
		// 全部可见，无需过滤
	} else if isCounselor {
		// 辅导员：仅可见本专业学生提交的
		majorIDs, _ = s.findCounselorMajorIDs(userID)
	} else if isStudent {
		// 学生：仅可见自己提交的
		user, err := s.repo.GetUserByID(userID)
		if err != nil || user == nil || user.StudentID == nil {
			// 未关联学生身份 → 返回空结果
			return &ReportListResult{Items: []ReportView{}, Total: 0, Page: page, PageSize: pageSize}, nil
		}
		studentIDs = []int64{*user.StudentID}
	} else {
		// 其它未识别角色：默认仅看自己（保守策略）
		user, err := s.repo.GetUserByID(userID)
		if err == nil && user != nil && user.StudentID != nil {
			studentIDs = []int64{*user.StudentID}
		} else {
			return &ReportListResult{Items: []ReportView{}, Total: 0, Page: page, PageSize: pageSize}, nil
		}
	}

	reports, total, err := s.repo.ListReports(applicationID, studentIDs, majorIDs, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]ReportView, 0, len(reports))
	for _, r := range reports {
		items = append(items, *s.toReportView(r))
	}

	return &ReportListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetReport 获取思想汇报详情。
func (s *CultivationService) GetReport(id int64) (*ReportView, error) {
	report, err := s.repo.GetReportByID(id)
	if err != nil {
		return nil, fmt.Errorf("思想汇报不存在")
	}
	return s.toReportView(*report), nil
}

// toReportView 将思想汇报模型转为视图。
func (s *CultivationService) toReportView(r models.TyThoughtReport) *ReportView {
	v := &ReportView{
		ID:            r.ID,
		BizNo:         r.BizNo,
		ApplicationID: r.ApplicationID,
		StudentID:     r.StudentID,
		Title:         r.Title,
		Content:       r.Content,
		Quarter:       r.Quarter,
		AISimilarity:  r.AISimilarity,
		IsQualified:   r.IsQualified,
		CreatedAt:     r.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if student, err := s.repo.GetStudentByID(r.StudentID); err == nil {
		v.StudentName = student.Name
		v.StudentNo = student.StudentNo
	}
	// 兜底：历史数据可能因早期 seed 直 INSERT 导致 biz_no 为空，
	// 此时前端"编号"列会空白。同步时按 "TY-RPT-YYYY-NNNN" 格式补一个展示值。
	if v.BizNo == "" {
		year := v.CreatedAt[:4]
		v.BizNo = fmt.Sprintf("TY-RPT-%s-%04d", year, v.ID)
	}
	return v
}

// ---- 内部方法 ----

// findUserRoles 查询当前用户的角色码列表。
// 委托 ApplicationRepository（已实现跨仓储的角色解析），不直接持有 sys_user 表。
func (s *CultivationService) findUserRoles(userID int64) ([]string, error) {
	if s.appRepo == nil {
		return nil, nil
	}
	return s.appRepo.FindUserRoles(userID)
}

// findCounselorMajorIDs 查询当前用户作为辅导员所负责的专业 ID 列表。
func (s *CultivationService) findCounselorMajorIDs(userID int64) ([]int64, error) {
	if s.appRepo == nil {
		return nil, nil
	}
	return s.appRepo.FindCounselorMajorIDs(userID)
}

// publishCultivationEvent 发布培养考察事件。
func (s *CultivationService) publishCultivationEvent(evtType string, actorID int64, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	_ = s.bus.Publish(&eventx.Event{
		Aggregate: "ty.cultivation",
		EventType: evtType,
		Module:    "TY",
		ActorID:   actorID,
		Payload:   payload,
	})
}

// parseDate 解析日期字符串（支持 2006-01-02 格式）。
func parseDate(d string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return time.Time{}, fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD")
	}
	return t, nil
}

// isValidQuarter 校验季度格式是否合法（YYYYQn）。
func isValidQuarter(q string) bool {
	if len(q) != 6 {
		return false
	}
	yearStr := q[:4]
	qChar := q[4]
	qNum := q[5]

	// 年份必须是数字
	for _, c := range yearStr {
		if c < '0' || c > '9' {
			return false
		}
	}
	// 第5位必须是 Q
	if qChar != 'Q' {
		return false
	}
	// 第6位必须是 1-4
	if qNum < '1' || qNum > '4' {
		return false
	}
	return true
}

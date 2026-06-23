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

// AssignMentorRequest 分配培养联系人请求。
type AssignMentorRequest struct {
	ApplicationID   int64  `json:"application_id" binding:"required"`
	MentorStudentID int64  `json:"mentor_student_id" binding:"required"`
	MentorType      string `json:"mentor_type" binding:"required"` // league_member / party_member
	StartAt         string `json:"start_at" binding:"required"`
}

// CultivationLinkView 培养联系人视图。
type CultivationLinkView struct {
	ID              int64  `json:"id"`
	ApplicationID   int64  `json:"application_id"`
	MentorStudentID int64  `json:"mentor_student_id"`
	MentorName      string `json:"mentor_name"`
	MentorType      string `json:"mentor_type"`
	StartAt         string `json:"start_at"`
	EndAt           *string `json:"end_at,omitempty"`
	IsActive        int    `json:"is_active"`
	CreatedAt       string `json:"created_at"`
}

// AssignMentor 分配培养联系人。
//
// 校验规则：
//   - 培养联系人必须是正式团员（member）或党员（party_member）
func (s *CultivationService) AssignMentor(userID int64, req *AssignMentorRequest) (*CultivationLinkView, error) {
	// 校验申请存在
	if _, err := s.appRepo.GetByID(req.ApplicationID); err != nil {
		return nil, fmt.Errorf("入团申请不存在")
	}

	// 校验培养联系人身份：必须是正式团员或党员
	mentor, err := s.repo.GetStudentByID(req.MentorStudentID)
	if err != nil {
		return nil, fmt.Errorf("培养联系人学生信息不存在")
	}
	status := strings.TrimSpace(mentor.PoliticalStatus)
	if status != "member" && status != "party_member" &&
		status != "共青团员" && status != "中共党员" &&
		status != "party_probationary" && status != "预备党员" {
		return nil, fmt.Errorf("培养联系人须为正式团员或党员")
	}

	// 校验 mentor_type 合法性
	if req.MentorType != "league_member" && req.MentorType != "party_member" {
		return nil, fmt.Errorf("培养联系人类型无效，须为 league_member 或 party_member")
	}

	// 解析开始日期
	startAt, err := parseDate(req.StartAt)
	if err != nil {
		return nil, fmt.Errorf("开始日期格式错误")
	}

	link := models.TyCultivationLink{
		ApplicationID:   req.ApplicationID,
		MentorStudentID: req.MentorStudentID,
		MentorType:      req.MentorType,
		StartAt:         startAt,
		IsActive:        1,
	}

	if err := s.repo.CreateLink(&link); err != nil {
		return nil, fmt.Errorf("创建培养联系人失败: %w", err)
	}

	s.publishCultivationEvent("TyCultivationLinkAssigned", userID, map[string]interface{}{
		"application_id":   req.ApplicationID,
		"mentor_student_id": req.MentorStudentID,
		"mentor_type":       req.MentorType,
	})

	return s.toLinkView(link, mentor.Name), nil
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
type RecordView struct {
	ID               int64  `json:"id"`
	BizNo            string `json:"biz_no"`
	ApplicationID    int64  `json:"application_id"`
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
	return v
}

// ==================== 团课记录管理 ====================

// CreateCourseRequest 创建团课记录请求。
type CreateCourseRequest struct {
	StudentID     int64  `json:"student_id" binding:"required"`
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
	ID             int64  `json:"id"`
	StudentID      int64  `json:"student_id"`
	StudentName    string `json:"student_name"`
	CourseName     string `json:"course_name"`
	Semester       string `json:"semester"`
	StudyAt        string `json:"study_at"`
	Score          *int   `json:"score,omitempty"`
	CertificateNo  string `json:"certificate_no"`
	IsPass         int    `json:"is_pass"`
	CreatedAt      string `json:"created_at"`
}

// CreateCourse 创建团课记录。
func (s *CultivationService) CreateCourse(userID int64, req *CreateCourseRequest) (*CourseView, error) {
	// 校验学生存在
	_, err := s.repo.GetStudentByID(req.StudentID)
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
		StudentID:     req.StudentID,
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
		"course_id": course.ID,
		"student_id": req.StudentID,
		"course_name": req.CourseName,
	})

	return s.toCourseView(course), nil
}

// ListCourses 查询团课列表。
func (s *CultivationService) ListCourses(studentID int64, page, pageSize int) (*CourseListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	courses, total, err := s.repo.ListCourses(studentID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]CourseView, 0, len(courses))
	for _, c := range courses {
		items = append(items, *s.toCourseView(c))
	}

	return &CourseListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdatePassStatus 更新团课结业状态（score >= 80 自动通过）。
func (s *CultivationService) UpdatePassStatus(id int64, userID int64) (*CourseView, error) {
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

	return s.toCourseView(*updated), nil
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

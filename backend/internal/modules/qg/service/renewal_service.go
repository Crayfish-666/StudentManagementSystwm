package service

import (
	"fmt"

	"gorm.io/gorm"

	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/qg/repository"
)

// RenewalService 续聘/解聘+申诉业务服务层。
type RenewalService struct {
	repo *repository.RenewalRepository
	db   *gorm.DB
}

// NewRenewalService 创建续聘/解聘+申诉服务。
func NewRenewalService(repo *repository.RenewalRepository, db *gorm.DB) *RenewalService {
	return &RenewalService{repo: repo, db: db}
}

// ---- DTO ----

// RenewalListResult 续聘/解聘列表结果。
type RenewalListResult struct {
	Items    []RenewalView `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// RenewalView 续聘/解聘视图。
type RenewalView struct {
	ID                     int64    `json:"id"`
	BizNo                  string   `json:"biz_no"`
	ApplyID                int64    `json:"apply_id"`
	PositionTitle          string   `json:"position_title"`
	StudentID              int64    `json:"student_id"`
	StudentName            string   `json:"student_name"`
	Type                   string   `json:"type"`
	TypeText               string   `json:"type_text"`
	Reason                 string   `json:"reason"`
	EffectiveAt            string   `json:"effective_at"`
	SemesterAvgScore       *float64 `json:"semester_avg_score,omitempty"`
	InitiatedBy            int64    `json:"initiated_by"`
	InitiatedByName        string   `json:"initiated_by_name"`
	CounselorSignedBy      *int64   `json:"counselor_signed_by,omitempty"`
	CounselorSignedName    string   `json:"counselor_signed_name,omitempty"`
	StudentAffairsSignedBy *int64   `json:"student_affairs_signed_by,omitempty"`
	StudentAffairsSignedName string `json:"student_affairs_signed_name,omitempty"`
	Status                 string   `json:"status"`
	StatusText             string   `json:"status_text"`
	CreatedAt              string   `json:"created_at"`
	UpdatedAt              string   `json:"updated_at"`
}

// CreateRenewalRequest 创建续聘/解聘请求。
type CreateRenewalRequest struct {
	ApplyID          int64   `json:"apply_id" binding:"required"`
	Type             string  `json:"type" binding:"required"`
	Reason           string  `json:"reason" binding:"required"`
	EffectiveAt      string  `json:"effective_at" binding:"required"`
	SemesterAvgScore *float64 `json:"semester_avg_score"`
}

// ComplaintListResult 申诉列表结果。
type ComplaintListResult struct {
	Items    []ComplaintView `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// ComplaintView 申诉视图。
type ComplaintView struct {
	ID                int64   `json:"id"`
	BizNo             string  `json:"biz_no"`
	StudentID         int64   `json:"student_id"`
	StudentName       string  `json:"student_name"`
	TargetType        string  `json:"target_type"`
	TargetTypeText    string  `json:"target_type_text"`
	TargetID          int64   `json:"target_id"`
	Reason            string  `json:"reason"`
	ExpectedReplyDays int     `json:"expected_reply_days"`
	Status            string  `json:"status"`
	StatusText        string  `json:"status_text"`
	Result            string  `json:"result"`
	HandledBy         *int64  `json:"handled_by,omitempty"`
	HandledByName     string  `json:"handled_by_name,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// CreateComplaintRequest 创建申诉请求。
type CreateComplaintRequest struct {
	TargetType        string `json:"target_type" binding:"required"`
	TargetID          int64  `json:"target_id" binding:"required"`
	Reason            string `json:"reason" binding:"required"`
	ExpectedReplyDays int    `json:"expected_reply_days"`
}

// ---- 状态映射 ----

var renewalStatusTextMap = map[string]string{
	"S1": "待辅导员签字",
	"S2": "辅导员已签",
	"S3": "学生处已签",
	"S4": "已驳回",
}

var renewalTypeTextMap = map[string]string{
	"renewal":     "续聘",
	"termination": "解聘",
}

var complaintStatusTextMap = map[string]string{
	"S1": "待处理",
	"S2": "处理中",
	"S3": "已回复",
	"S4": "已驳回",
}

var complaintTargetTypeTextMap = map[string]string{
	"attendance": "工时打卡",
	"assess":     "月度考核",
	"payroll":    "薪酬发放",
}

// ---- 续聘/解聘业务方法 ----

// CreateRenewal 创建续聘/解聘。
func (s *RenewalService) CreateRenewal(userID int64, req *CreateRenewalRequest) (*RenewalView, error) {
	// 获取申请记录
	apply, err := s.repo.GetApplyByID(req.ApplyID)
	if err != nil {
		return nil, fmt.Errorf("岗位申请记录不存在")
	}

	// 解析生效日期
	effectiveAt, err := parseTime(req.EffectiveAt)
	if err != nil {
		return nil, fmt.Errorf("生效日期格式错误")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "QG")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	renewal := &models.QgRenewalTerm{
		BizNo:            bizNo,
		ApplyID:          req.ApplyID,
		StudentID:        apply.StudentID,
		Type:             req.Type,
		Reason:           req.Reason,
		EffectiveAt:      effectiveAt,
		SemesterAvgScore: req.SemesterAvgScore,
		InitiatedBy:      userID,
		Status:           "S1",
	}

	if err := s.repo.CreateRenewal(renewal); err != nil {
		return nil, err
	}

	return s.GetRenewal(renewal.ID)
}

// ListRenewal 分页查询续聘/解聘列表。
func (s *RenewalService) ListRenewal(applyID int64, renewalType string, page, pageSize int) (*RenewalListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	renewals, total, err := s.repo.ListRenewals(applyID, renewalType, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]RenewalView, 0, len(renewals))
	for _, r := range renewals {
		v := s.toRenewalView(r)
		items = append(items, v)
	}

	return &RenewalListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetRenewal 获取续聘/解聘详情。
func (s *RenewalService) GetRenewal(id int64) (*RenewalView, error) {
	renewal, err := s.repo.GetRenewalByID(id)
	if err != nil {
		return nil, fmt.Errorf("续聘/解聘记录不存在")
	}

	v := s.toRenewalView(*renewal)
	return &v, nil
}

// CounselorSign 辅导员签字（S1→S2）。
func (s *RenewalService) CounselorSign(id, userID int64) (*RenewalView, error) {
	renewal, err := s.repo.GetRenewalByID(id)
	if err != nil {
		return nil, fmt.Errorf("续聘/解聘记录不存在")
	}

	if renewal.Status != "S1" {
		return nil, fmt.Errorf("当前状态不允许辅导员签字")
	}

	renewal.CounselorSignedBy = &userID
	renewal.Status = "S2"

	if err := s.repo.UpdateRenewal(renewal); err != nil {
		return nil, err
	}

	return s.GetRenewal(id)
}

// AffairsSign 学生处签字（S2→S3）。
func (s *RenewalService) AffairsSign(id, userID int64) (*RenewalView, error) {
	renewal, err := s.repo.GetRenewalByID(id)
	if err != nil {
		return nil, fmt.Errorf("续聘/解聘记录不存在")
	}

	if renewal.Status != "S2" {
		return nil, fmt.Errorf("当前状态不允许学生处签字")
	}

	renewal.StudentAffairsSignedBy = &userID
	renewal.Status = "S3"

	// 事务：更新续聘/解聘状态 + 更新岗位申请状态
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(renewal).Error; err != nil {
			return err
		}

		// 解聘生效时更新岗位申请状态
		if renewal.Type == "termination" {
			if err := tx.Model(&models.QgPositionApply{}).
				Where("id = ?", renewal.ApplyID).
				Update("status", "terminated").Error; err != nil {
				return fmt.Errorf("更新岗位申请状态失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.GetRenewal(id)
}

// ---- 申诉业务方法 ----

// CreateComplaint 创建申诉。
func (s *RenewalService) CreateComplaint(userID, studentID int64, req *CreateComplaintRequest) (*ComplaintView, error) {
	expectedReplyDays := req.ExpectedReplyDays
	if expectedReplyDays <= 0 {
		expectedReplyDays = 10
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "QG")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	complaint := &models.QgComplaint{
		BizNo:             bizNo,
		StudentID:         studentID,
		TargetType:        req.TargetType,
		TargetID:          req.TargetID,
		Reason:            req.Reason,
		ExpectedReplyDays: expectedReplyDays,
		Status:            "S1",
	}

	if err := s.repo.CreateComplaint(complaint); err != nil {
		return nil, err
	}

	return s.GetComplaint(complaint.ID)
}

// ListComplaint 分页查询申诉列表。
func (s *RenewalService) ListComplaint(studentID int64, targetType, status string, page, pageSize int) (*ComplaintListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	complaints, total, err := s.repo.ListComplaints(studentID, targetType, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]ComplaintView, 0, len(complaints))
	for _, c := range complaints {
		v := s.toComplaintView(c)
		items = append(items, v)
	}

	return &ComplaintListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetComplaint 获取申诉详情。
func (s *RenewalService) GetComplaint(id int64) (*ComplaintView, error) {
	complaint, err := s.repo.GetComplaintByID(id)
	if err != nil {
		return nil, fmt.Errorf("申诉记录不存在")
	}

	v := s.toComplaintView(*complaint)
	return &v, nil
}

// ReplyComplaint 回复申诉。
func (s *RenewalService) ReplyComplaint(id, userID int64, result, decision string) (*ComplaintView, error) {
	complaint, err := s.repo.GetComplaintByID(id)
	if err != nil {
		return nil, fmt.Errorf("申诉记录不存在")
	}

	if complaint.Status != "S1" && complaint.Status != "S2" {
		return nil, fmt.Errorf("当前状态不允许回复")
	}

	complaint.Result = result
	complaint.HandledBy = &userID

	// decision: approve→S3, reject→S4
	switch decision {
	case "approve":
		complaint.Status = "S3"
	case "reject":
		complaint.Status = "S4"
	default:
		complaint.Status = "S3"
	}

	if err := s.repo.UpdateComplaint(complaint); err != nil {
		return nil, err
	}

	return s.GetComplaint(id)
}

// ---- 内部方法 ----

func (s *RenewalService) toRenewalView(r models.QgRenewalTerm) RenewalView {
	v := RenewalView{
		ID:                   r.ID,
		BizNo:                r.BizNo,
		ApplyID:              r.ApplyID,
		StudentID:            r.StudentID,
		Type:                 r.Type,
		TypeText:             renewalTypeTextMap[r.Type],
		Reason:               r.Reason,
		EffectiveAt:          r.EffectiveAt.Format("2006-01-02T15:04:05+08:00"),
		SemesterAvgScore:     r.SemesterAvgScore,
		InitiatedBy:          r.InitiatedBy,
		CounselorSignedBy:    r.CounselorSignedBy,
		StudentAffairsSignedBy: r.StudentAffairsSignedBy,
		Status:               r.Status,
		StatusText:           renewalStatusTextMap[r.Status],
		CreatedAt:            r.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:            r.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 加载岗位标题
	if apply, err := s.repo.GetApplyByID(r.ApplyID); err == nil {
		if pos, err := s.repo.GetPositionByID(apply.PositionID); err == nil {
			v.PositionTitle = pos.Title
		}
	}

	// 加载学生姓名
	if student, err := s.repo.GetStudentByID(r.StudentID); err == nil {
		v.StudentName = student.Name
	}

	// 加载发起人姓名
	if u, err := s.repo.GetUserByID(r.InitiatedBy); err == nil {
		v.InitiatedByName = u.DisplayName
	}

	// 加载辅导员签字姓名
	if r.CounselorSignedBy != nil {
		if u, err := s.repo.GetUserByID(*r.CounselorSignedBy); err == nil {
			v.CounselorSignedName = u.DisplayName
		}
	}

	// 加载学生处签字姓名
	if r.StudentAffairsSignedBy != nil {
		if u, err := s.repo.GetUserByID(*r.StudentAffairsSignedBy); err == nil {
			v.StudentAffairsSignedName = u.DisplayName
		}
	}

	return v
}

func (s *RenewalService) toComplaintView(c models.QgComplaint) ComplaintView {
	v := ComplaintView{
		ID:                c.ID,
		BizNo:             c.BizNo,
		StudentID:         c.StudentID,
		TargetType:        c.TargetType,
		TargetTypeText:    complaintTargetTypeTextMap[c.TargetType],
		TargetID:          c.TargetID,
		Reason:            c.Reason,
		ExpectedReplyDays: c.ExpectedReplyDays,
		Status:            c.Status,
		StatusText:        complaintStatusTextMap[c.Status],
		Result:            c.Result,
		HandledBy:         c.HandledBy,
		CreatedAt:         c.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:         c.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 加载学生姓名
	if student, err := s.repo.GetStudentByID(c.StudentID); err == nil {
		v.StudentName = student.Name
	}

	// 加载处理人姓名
	if c.HandledBy != nil {
		if u, err := s.repo.GetUserByID(*c.HandledBy); err == nil {
			v.HandledByName = u.DisplayName
		}
	}

	return v
}

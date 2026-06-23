package service

import (
	"fmt"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/ty/repository"
)

// DevelopmentObjectService 发展对象业务服务层。
type DevelopmentObjectService struct {
	repo *repository.DevelopmentObjectRepository
	appRepo *repository.ApplicationRepository
	db   *gorm.DB
	bus  *eventx.Bus
}

// NewDevelopmentObjectService 创建发展对象服务。
func NewDevelopmentObjectService(
	repo *repository.DevelopmentObjectRepository,
	appRepo *repository.ApplicationRepository,
	db *gorm.DB,
	bus *eventx.Bus,
) *DevelopmentObjectService {
	return &DevelopmentObjectService{
		repo:    repo,
		appRepo: appRepo,
		db:      db,
		bus:     bus,
	}
}

// ---- DTO 定义 ----

// CreateDevelopmentObjectRequest 创建发展对象申请请求。
type CreateDevelopmentObjectRequest struct {
	ApplicationID        int64  `json:"application_id" binding:"required"`
	CourseCertNo         string `json:"course_cert_no" binding:"required"`
	MentorOpinion        string `json:"mentor_opinion" binding:"required"`
	CounselorOpinion     string `json:"counselor_opinion" binding:"required"`
	MassMeetingAt        string `json:"mass_meeting_at"`
	MassMeetingAttendees int    `json:"mass_meeting_attendees" binding:"required"`
	AutobiographyPath    string `json:"autobiography_path" binding:"required"`
}

// PublicizeRequest 公示请求。
type PublicizeRequest struct {
	PublicStart string `json:"public_start" binding:"required"` // 公示开始日期 YYYY-MM-DD
}

// ApproveDevelopmentRequest 发展对象审批请求。
type ApproveDevelopmentRequest struct {
	Step    string `json:"step" binding:"required"`    // branch | college | school
	Result  string `json:"result" binding:"required"`  // approve | reject
	Opinion string `json:"opinion" binding:"required"` // 审批意见
}

// DevelopmentObjectView 发展对象视图。
type DevelopmentObjectView struct {
	ID                   int64      `json:"id"`
	BizNo                string     `json:"biz_no"`
	ApplicationID        int64      `json:"application_id"`
	StudentID            int64      `json:"student_id"`
	StudentNo            string     `json:"student_no,omitempty"`
	StudentName          string     `json:"student_name,omitempty"`
	BranchID             int64      `json:"branch_id"`
	BranchName           string     `json:"branch_name,omitempty"`
	CourseCertNo         string     `json:"course_cert_no"`
	MentorOpinion        string     `json:"mentor_opinion"`
	CounselorOpinion     string     `json:"counselor_opinion"`
	MassMeetingAt        *string    `json:"mass_meeting_at,omitempty"`
	MassMeetingAttendees *int       `json:"mass_meeting_attendees,omitempty"`
	PublicStart          *string    `json:"public_start,omitempty"`
	PublicEnd            *string    `json:"public_end,omitempty"`
	AutobiographyPath    string     `json:"autobiography_path"`
	Status               string     `json:"status"`
	StatusText           string     `json:"status_text"`
	CreatedAt            string     `json:"created_at"`
	UpdatedAt            string     `json:"updated_at"`
}

// DevelopmentObjectListResult 发展对象列表结果。
type DevelopmentObjectListResult struct {
	Items    []DevelopmentObjectView `json:"items"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// ---- 状态映射 ----

var devObjStatusTextMap = map[string]string{
	"S0": "待提交",
	"S1": "待审",
	"S2": "公示中",
	"S3": "可进入政审/发展大会",
	"S4": "已终止",
}

// ---- 业务方法 ----

// Submit 提交发展对象申请（S0 → S1）。
//
// 校验规则：
//   - 申请必须已通过推优大会（入团积极分子状态）
//   - 必填字段完整性校验
//   - 培养联系人意见 ≥ 200字
//   - 辅导员意见 ≥ 200字
//   - 群众座谈人数 ≥ 10人
func (s *DevelopmentObjectService) Submit(userID int64, req *CreateDevelopmentObjectRequest, actorName, actorRole, ip, ua string) (*DevelopmentObjectView, error) {
	// 校验入团申请是否存在
	_, err := s.appRepo.GetByID(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("入团申请不存在")
	}

	// 校验前置条件：申请必须已通过推优大会（即已是"入团积极分子"状态）
	passed, err := s.repo.HasPassedRecommendation(req.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("校验推优状态失败")
	}
	if !passed {
		return nil, fmt.Errorf("申请人尚未完成推优流程，错误码:2602")
	}

	// 校验必填字段
	if req.CourseCertNo == "" {
		return nil, fmt.Errorf("团课证书编号不能为空，错误码:2601")
	}
	if utf8.RuneCountInString(req.MentorOpinion) < 200 {
		return nil, fmt.Errorf("培养联系人意见须 ≥ 200字，错误码:2601")
	}
	if utf8.RuneCountInString(req.CounselorOpinion) < 200 {
		return nil, fmt.Errorf("辅导员意见须 ≥ 200字，错误码:2601")
	}
	if req.MassMeetingAttendees < 10 {
		return nil, fmt.Errorf("群众座谈人数不足10人，错误码:2603")
	}
	if req.AutobiographyPath == "" {
		return nil, fmt.Errorf("自传路径不能为空，错误码:2601")
	}

	// 检查是否已存在发展对象记录（同一申请仅允许一条）
	existing, _ := s.repo.GetByApplicationID(req.ApplicationID)
	if existing != nil {
		return nil, fmt.Errorf("该申请已创建发展对象记录")
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "TY")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	// 解析群众座谈时间
	var massMeetingAt *time.Time
	if req.MassMeetingAt != "" {
		t, parseErr := time.Parse("2006-01-02", req.MassMeetingAt)
		if parseErr != nil {
			return nil, fmt.Errorf("群众座谈日期格式错误")
		}
		massMeetingAt = &t
	}

	obj := models.TyDevelopmentObject{
		BizNo:                bizNo,
		ApplicationID:        req.ApplicationID,
		CourseCertNo:         req.CourseCertNo,
		MentorOpinion:        req.MentorOpinion,
		CounselorOpinion:     req.CounselorOpinion,
		MassMeetingAt:        massMeetingAt,
		MassMeetingAttendees: &req.MassMeetingAttendees,
		AutobiographyPath:    req.AutobiographyPath,
		Status:               "S1", // 提交后直接进入待审状态
	}

	if err := s.repo.Create(&obj); err != nil {
		return nil, fmt.Errorf("创建发展对象记录失败: %w", err)
	}

	s.publishDevEvent(&obj, "TyDevelopmentObjectSubmitted", userID, actorRole, ip, ua, map[string]interface{}{
		"from": "S0",
		"to":   "S1",
	})

	return s.Get(obj.ID)
}

// Publicize 设置公示期（S1 → S2）。
//
// 规则：公示结束日期 = 开始日期 + 5个工作日
func (s *DevelopmentObjectService) Publicize(id int64, userID int64, req *PublicizeRequest, actorName, actorRole, ip, ua string) (*DevelopmentObjectView, error) {
	obj, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("发展对象记录不存在")
	}

	if obj.Status != "S1" {
		return nil, fmt.Errorf("仅待审状态可设置公示")
	}

	// 解析公示开始时间
	publicStart, err := time.Parse("2006-01-02", req.PublicStart)
	if err != nil {
		return nil, fmt.Errorf("公示开始日期格式错误")
	}

	// 计算5个工作日后的结束日期（简单实现：+7天）
	publicEnd := publicStart.AddDate(0, 0, 7)

	obj.PublicStart = &publicStart
	obj.PublicEnd = &publicEnd
	obj.Status = "S2"
	if err := s.repo.Update(obj); err != nil {
		return nil, err
	}

	s.publishDevEvent(obj, "TyDevelopmentObjectPublicized", userID, actorRole, ip, ua, map[string]interface{}{
		"from":         "S1",
		"to":           "S2",
		"public_start": publicStart.Format("2006-01-02"),
		"public_end":   publicEnd.Format("2006-01-02"),
	})

	return s.Get(obj.ID)
}

// Approve 审批发展对象（支持三级审批：branch → college → school）。
//
// 规则：
//   - 通过：branch 通过 → 保持 S2；college 通过 → 保持 S2；school 通过 → S3
//   - 驳回：S* → S4（终止）
//   - 步骤前置：college 要求 branch 已通过；school 要求 college 已通过
//   - 同一 step 不允许重复通过
func (s *DevelopmentObjectService) Approve(id int64, userID int64, req *ApproveDevelopmentRequest, actorName, actorRole, ip, ua string) (*DevelopmentObjectView, error) {
	if req == nil {
		return nil, fmt.Errorf("参数不能为空")
	}
	if req.Step != "branch" && req.Step != "college" && req.Step != "school" {
		return nil, fmt.Errorf("无效的审批步骤")
	}
	if req.Result != "approve" && req.Result != "reject" {
		return nil, fmt.Errorf("无效的审批结果")
	}
	if utf8.RuneCountInString(req.Opinion) < 5 {
		return nil, fmt.Errorf("审批意见至少 5 字")
	}

	obj, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("发展对象记录不存在")
	}

	fromStatus := obj.Status

	// 步骤前置校验
	if req.Step == "college" {
		ok, err := s.appRepo.HasApprovedStepByModule(obj.ApplicationID, "development_object", id, "branch")
		if err != nil {
			return nil, fmt.Errorf("校验前置步骤失败")
		}
		if !ok {
			return nil, fmt.Errorf("团支部大会尚未通过，院系不可复核")
		}
	}
	if req.Step == "school" {
		ok, err := s.appRepo.HasApprovedStepByModule(obj.ApplicationID, "development_object", id, "college")
		if err != nil {
			return nil, fmt.Errorf("校验前置步骤失败")
		}
		if !ok {
			return nil, fmt.Errorf("院系复核尚未通过，校级不可终审")
		}
	}

	// 同一 step 不允许重复通过
	if req.Result == "approve" {
		ok, err := s.appRepo.HasApprovedStepByModule(obj.ApplicationID, "development_object", id, req.Step)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, fmt.Errorf("该步骤已审批通过，不可重复审批")
		}
	}

	var toStatus string
	if req.Result == "approve" {
		if req.Step == "school" {
			toStatus = "S3"
		} else {
			toStatus = obj.Status // 保持当前状态
		}
	} else {
		toStatus = "S4"
	}

	obj.Status = toStatus
	if err := s.repo.Update(obj); err != nil {
		return nil, err
	}

	// 写入审批记录
	rec := &models.TyApprovalRecord{
		ApplicationID: obj.ApplicationID,
		Module:        "development_object",
		TargetID:      obj.ID,
		Step:          req.Step,
		ApproverID:    userID,
		ApproverName:  actorName,
		ApproverRole:  actorRole,
		Result:        req.Result,
		Opinion:       req.Opinion,
		FromStatus:    fromStatus,
		ToStatus:      toStatus,
		IP:            ip,
	}
	_ = s.appRepo.CreateApprovalRecord(rec)

	eventType := "TyDevelopmentObjectApproved"
	if req.Result == "reject" {
		eventType = "TyDevelopmentObjectRejected"
	}
	s.publishDevEvent(obj, eventType, userID, actorRole, ip, ua, map[string]interface{}{
		"step":    req.Step,
		"result":  req.Result,
		"opinion": req.Opinion,
		"from":    fromStatus,
		"to":      toStatus,
	})

	return s.Get(obj.ID)
}

// Get 获取发展对象详情。
func (s *DevelopmentObjectService) Get(id int64) (*DevelopmentObjectView, error) {
	obj, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toView(*obj), nil
}

// List 列表查询发展对象。
func (s *DevelopmentObjectService) List(status string, collegeID int64, page, pageSize int) (*DevelopmentObjectListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := s.repo.List(status, collegeID, page, pageSize)
	if err != nil {
		return nil, err
	}

	views := make([]DevelopmentObjectView, 0, len(items))
	for _, item := range items {
		views = append(views, *s.toView(item))
	}

	return &DevelopmentObjectListResult{
		Items:    views,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ---- 内部方法 ----

// toView 将模型转为视图。
func (s *DevelopmentObjectService) toView(obj models.TyDevelopmentObject) *DevelopmentObjectView {
	v := &DevelopmentObjectView{
		ID:                obj.ID,
		BizNo:             obj.BizNo,
		ApplicationID:     obj.ApplicationID,
		CourseCertNo:      obj.CourseCertNo,
		MentorOpinion:     obj.MentorOpinion,
		CounselorOpinion:  obj.CounselorOpinion,
		AutobiographyPath: obj.AutobiographyPath,
		Status:            obj.Status,
		StatusText:        devObjStatusTextMap[obj.Status],
		CreatedAt:         obj.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:         obj.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	// 通过 application 关联回填申请人 / 团支部信息
	if app, err := s.repo.GetApplicationByID(obj.ApplicationID); err == nil && app != nil {
		v.StudentID = app.StudentID
		v.BranchID = app.BranchID
		if stu, err := s.appRepo.GetStudentByID(app.StudentID); err == nil && stu != nil {
			v.StudentName = stu.Name
			v.StudentNo = stu.StudentNo
		}
		if br, err := s.appRepo.GetBranchByID(app.BranchID); err == nil && br != nil {
			v.BranchName = br.Name
		}
	}

	if obj.MassMeetingAt != nil {
		t := obj.MassMeetingAt.Format("2006-01-02")
		v.MassMeetingAt = &t
	}
	if obj.MassMeetingAttendees != nil {
		v.MassMeetingAttendees = obj.MassMeetingAttendees
	}
	if obj.PublicStart != nil {
		t := obj.PublicStart.Format("2006-01-02")
		v.PublicStart = &t
	}
	if obj.PublicEnd != nil {
		t := obj.PublicEnd.Format("2006-01-02")
		v.PublicEnd = &t
	}

	return v
}

// publishDevEvent 发布发展对象相关事件。
func (s *DevelopmentObjectService) publishDevEvent(obj *models.TyDevelopmentObject, evtType string, actorID int64, actorRole, ip, ua string, payload map[string]interface{}) {
	if s.bus == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["development_object_id"] = obj.ID
	payload["biz_no"] = obj.BizNo
	payload["status"] = obj.Status

	_ = s.bus.Publish(&eventx.Event{
		Aggregate:   "ty.development_object",
		AggregateID: obj.BizNo,
		EventType:   evtType,
		Module:      "TY",
		ActorID:     actorID,
		ActorRole:   actorRole,
		Payload:     payload,
		BizNo:       obj.BizNo,
		IP:          ip,
		UA:          ua,
	})
}

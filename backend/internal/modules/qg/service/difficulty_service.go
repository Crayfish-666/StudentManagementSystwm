package service

import (
	"fmt"

	"gorm.io/gorm"

	"student-system/internal/eventx"
	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/qg/repository"
)

// DifficultyService 困难认定业务服务层。
type DifficultyService struct {
	repo *repository.DifficultyRepository
	db   *gorm.DB
	bus  *eventx.Bus
}

// NewDifficultyService 创建困难认定服务。
func NewDifficultyService(repo *repository.DifficultyRepository, db *gorm.DB, bus *eventx.Bus) *DifficultyService {
	return &DifficultyService{repo: repo, db: db, bus: bus}
}

// ---- DTO ----

// DifficultyListResult 困难认定列表结果。
type DifficultyListResult struct {
	Items    []DifficultyView `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// DifficultyView 困难认定视图。
type DifficultyView struct {
	ID              int64   `json:"id"`
	BizNo           string  `json:"biz_no"`
	StudentID       int64   `json:"student_id"`
	StudentName     string  `json:"student_name"`
	StudentNo       string  `json:"student_no"`
	AcademicYear    string  `json:"academic_year"`
	Level           string  `json:"level"`
	LevelText       string  `json:"level_text"`
	CertFiles       string  `json:"cert_files"`
	PublicStart     *string `json:"public_start,omitempty"`
	PublicEnd       *string `json:"public_end,omitempty"`
	Status          string  `json:"status"`
	StatusText      string  `json:"status_text"`
	RejectReason    string  `json:"reject_reason"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// CreateDifficultyRequest 创建困难认定请求。
type CreateDifficultyRequest struct {
	StudentID    int64  `json:"student_id" binding:"required"`
	AcademicYear string `json:"academic_year" binding:"required"`
	Level        string `json:"level" binding:"required"`
	CertFiles    string `json:"cert_files"`
	PublicStart  string `json:"public_start"`
	PublicEnd    string `json:"public_end"`
}

// ApproveDifficultyRequest 审批困难认定请求。
type ApproveDifficultyRequest struct {
	Level         string `json:"level" binding:"required"`
	RejectReason  string `json:"reject_reason"`
}

// ---- 状态映射 ----

var difficultyStatusTextMap = map[string]string{
	"S0": "草稿",
	"S1": "待审",
	"S2": "院系通过",
	"S3": "终审通过",
	"S4": "已驳回",
}

var difficultyLevelTextMap = map[string]string{
	"special": "特别困难",
	"hard":    "困难",
	"normal":  "一般困难",
	"none":    "不困难",
}

// ---- 业务方法 ----

// List 分页查询困难认定列表。
func (s *DifficultyService) List(level, status string, studentID int64, page, pageSize int) (*DifficultyListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	certs, total, err := s.repo.List(level, status, studentID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]DifficultyView, 0, len(certs))
	for _, cert := range certs {
		v := s.toView(cert)
		items = append(items, v)
	}

	return &DifficultyListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取困难认定详情。
func (s *DifficultyService) Get(id int64) (*DifficultyView, error) {
	cert, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("困难认定不存在")
	}

	v := s.toView(*cert)
	return &v, nil
}

// Create 创建困难认定。
func (s *DifficultyService) Create(userID int64, req *CreateDifficultyRequest) (*DifficultyView, error) {
	// BR: 同一学生同一学年只能认定1次
	count, err := s.repo.CountByStudentAndYear(req.StudentID, req.AcademicYear)
	if err != nil {
		return nil, fmt.Errorf("查询困难认定记录失败: %w", err)
	}
	if count > 0 {
		return nil, &BizError{Code: 40905, Msg: "同一学生同一学年只能认定1次"}
	}

	// 生成业务编号
	bizNo, err := idgen.NextBizNo(s.db, "QG-DIF")
	if err != nil {
		return nil, fmt.Errorf("生成业务编号失败: %w", err)
	}

	cert := &models.QgDifficultyCert{
		BizNo:        bizNo,
		StudentID:    req.StudentID,
		AcademicYear: req.AcademicYear,
		Level:        req.Level,
		CertFiles:    req.CertFiles,
		Status:       "S0",
		CreatedBy:    &userID,
		UpdatedBy:    &userID,
	}

	// 解析公示时间
	if req.PublicStart != "" {
		t, err := parseTime(req.PublicStart)
		if err != nil {
			return nil, fmt.Errorf("公示开始时间格式错误")
		}
		cert.PublicStart = &t
	}
	if req.PublicEnd != "" {
		t, err := parseTime(req.PublicEnd)
		if err != nil {
			return nil, fmt.Errorf("公示结束时间格式错误")
		}
		cert.PublicEnd = &t
	}

	if err := s.repo.Create(cert); err != nil {
		return nil, err
	}

	return s.Get(cert.ID)
}

// Submit 提交困难认定（S0→S1）。
func (s *DifficultyService) Submit(id int64) (*DifficultyView, error) {
	cert, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("困难认定不存在")
	}

	if cert.Status != "S0" {
		return nil, fmt.Errorf("当前状态不允许提交")
	}

	cert.Status = "S1"
	if err := s.repo.Update(cert); err != nil {
		return nil, err
	}

	return s.Get(id)
}

// Approve 审批困难认定。
// level=college: S1→S2, level=school: S2→S3，设置 difficulty_level。
func (s *DifficultyService) Approve(id, userID int64, req *ApproveDifficultyRequest) (*DifficultyView, error) {
	cert, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("困难认定不存在")
	}

	// 根据审批层级判断状态流转
	switch {
	case cert.Status == "S1" && req.Level == "college":
		cert.Status = "S2"
	case cert.Status == "S2" && req.Level == "school":
		cert.Status = "S3"
	default:
		return nil, fmt.Errorf("当前状态不允许审批")
	}

	// 终审（S3）时把认定的 level 同步到学生档案（不修改认定的 level 字段）
	updatedBy := userID
	cert.UpdatedBy = &updatedBy

	// 事务：更新认定 + 更新学生困难标记
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(cert).Error; err != nil {
			return err
		}
		// 终审通过后更新 idx_student.is_difficulty=1 和 difficulty_level
		if cert.Status == "S3" {
			if err := tx.Model(&models.IdxStudent{}).
				Where("id = ?", cert.StudentID).
				Updates(map[string]interface{}{
					"is_difficulty":    1,
					"difficulty_level": cert.Level,
				}).Error; err != nil {
				return fmt.Errorf("更新学生困难标记失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 终审通过触发事件
	if cert.Status == "S3" && s.bus != nil {
		_ = s.bus.Publish(&eventx.Event{
			Aggregate:   "qg.difficulty",
			AggregateID: cert.BizNo,
			EventType:   "QgDifficultyApproved",
			Module:      "QG",
			ActorID:     userID,
			Payload: map[string]interface{}{
				"cert_id":    cert.ID,
				"biz_no":     cert.BizNo,
				"student_id": cert.StudentID,
				"level":      cert.Level,
			},
			BizNo: cert.BizNo,
		})
	}

	return s.Get(id)
}

// Reject 驳回困难认定（→S4）。
func (s *DifficultyService) Reject(id, userID int64, req *ApproveDifficultyRequest) (*DifficultyView, error) {
	cert, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("困难认定不存在")
	}

	if cert.Status != "S1" && cert.Status != "S2" {
		return nil, fmt.Errorf("当前状态不允许驳回")
	}

	cert.Status = "S4"
	cert.RejectReason = req.RejectReason
	updatedBy := userID
	cert.UpdatedBy = &updatedBy

	if err := s.repo.Update(cert); err != nil {
		return nil, err
	}

	return s.Get(id)
}

// Delete 软删除困难认定。
func (s *DifficultyService) Delete(id int64) error {
	return s.repo.SoftDelete(id)
}

// ---- 内部方法 ----

func (s *DifficultyService) toView(cert models.QgDifficultyCert) DifficultyView {
	v := DifficultyView{
		ID:           cert.ID,
		BizNo:        cert.BizNo,
		StudentID:    cert.StudentID,
		AcademicYear: cert.AcademicYear,
		Level:        cert.Level,
		LevelText:    difficultyLevelTextMap[cert.Level],
		CertFiles:    cert.CertFiles,
		Status:       cert.Status,
		StatusText:   difficultyStatusTextMap[cert.Status],
		RejectReason: cert.RejectReason,
		CreatedAt:    cert.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:    cert.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}

	if cert.PublicStart != nil {
		t := cert.PublicStart.Format("2006-01-02T15:04:05+08:00")
		v.PublicStart = &t
	}
	if cert.PublicEnd != nil {
		t := cert.PublicEnd.Format("2006-01-02T15:04:05+08:00")
		v.PublicEnd = &t
	}

	// 加载学生姓名和学号
	if student, err := s.repo.GetStudentByID(cert.StudentID); err == nil {
		v.StudentName = student.Name
		v.StudentNo = student.StudentNo
	}

	return v
}

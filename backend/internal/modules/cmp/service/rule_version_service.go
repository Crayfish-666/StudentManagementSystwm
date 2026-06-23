// Package cmp 综合素质量化规则版本服务。
package service

import (
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
	cmprepo "student-system/internal/modules/cmp/repository"
)

// RuleVersionService 规则版本服务。
type RuleVersionService struct {
	db   *gorm.DB
	repo *cmprepo.ScoreRepository
}

// NewRuleVersionService 创建规则版本服务。
func NewRuleVersionService(db *gorm.DB, repo *cmprepo.ScoreRepository) *RuleVersionService {
	return &RuleVersionService{db: db, repo: repo}
}

// RuleVersionView 规则版本视图。
type RuleVersionView struct {
	ID          int64  `json:"id"`
	Version     string `json:"version"`
	RulesJSON   string `json:"rules_json"`
	EffectiveAt string `json:"effective_at"`
	ExpiredAt   string `json:"expired_at,omitempty"`
	IsActive    int    `json:"is_active"`
	CreatedBy   *int64 `json:"created_by,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// List 列出全部版本。
func (s *RuleVersionService) List() ([]RuleVersionView, error) {
	versions, err := s.repo.ListRuleVersions()
	if err != nil {
		return nil, err
	}
	items := make([]RuleVersionView, 0, len(versions))
	for _, v := range versions {
		items = append(items, s.toView(v))
	}
	return items, nil
}

// Create 新建规则版本。
type CreateRuleVersionRequest struct {
	Version     string `json:"version" binding:"required"`
	RulesJSON   string `json:"rules_json" binding:"required"`
	EffectiveAt string `json:"effective_at" binding:"required"`
	ExpiredAt   string `json:"expired_at"`
}

func (s *RuleVersionService) Create(uid int64, req *CreateRuleVersionRequest) (*RuleVersionView, error) {
	effectiveAt, err := time.Parse("2006-01-02", req.EffectiveAt)
	if err != nil {
		return nil, err
	}
	var expiredAt *time.Time
	if req.ExpiredAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiredAt)
		if err != nil {
			return nil, err
		}
		expiredAt = &t
	}
	v := &models.CmpRuleVersion{
		Version:     req.Version,
		RulesJSON:   req.RulesJSON,
		EffectiveAt: effectiveAt,
		ExpiredAt:   expiredAt,
		IsActive:    0,
		CreatedBy:   &uid,
	}
	if err := s.repo.CreateRuleVersion(v); err != nil {
		return nil, err
	}
	view := s.toView(*v)
	return &view, nil
}

// Activate 激活指定版本（同时取消其它激活位）。
func (s *RuleVersionService) Activate(id int64) error {
	return s.repo.ActivateRuleVersion(id)
}

func (s *RuleVersionService) toView(v models.CmpRuleVersion) RuleVersionView {
	view := RuleVersionView{
		ID:          v.ID,
		Version:     v.Version,
		RulesJSON:   v.RulesJSON,
		EffectiveAt: v.EffectiveAt.Format("2006-01-02"),
		IsActive:    v.IsActive,
		CreatedBy:   v.CreatedBy,
		CreatedAt:   v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   v.UpdatedAt.Format(time.RFC3339),
	}
	if v.ExpiredAt != nil {
		view.ExpiredAt = v.ExpiredAt.Format("2006-01-02")
	}
	return view
}

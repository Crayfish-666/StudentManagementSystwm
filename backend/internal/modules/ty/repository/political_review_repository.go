package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// PoliticalReviewRepository 政审数据访问层。
type PoliticalReviewRepository struct {
	db *gorm.DB
}

// NewPoliticalReviewRepository 创建政审仓储。
func NewPoliticalReviewRepository(db *gorm.DB) *PoliticalReviewRepository {
	return &PoliticalReviewRepository{db: db}
}

// Create 创建政审记录。
func (r *PoliticalReviewRepository) Create(review *models.TyPoliticalReview) error {
	return r.db.Create(review).Error
}

// ListByDevelopmentID 按发展对象ID查询所有政审记录。
func (r *PoliticalReviewRepository) ListByDevelopmentID(developmentID int64) ([]models.TyPoliticalReview, error) {
	var reviews []models.TyPoliticalReview
	if err := r.db.Where("development_id = ? AND is_deleted = 0", developmentID).
		Order("id ASC").
		Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

// GetByID 按ID查询政审记录。
func (r *PoliticalReviewRepository) GetByID(id int64) (*models.TyPoliticalReview, error) {
	var review models.TyPoliticalReview
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// UpdateConclusion 更新政审结论。
func (r *PoliticalReviewRepository) UpdateConclusion(id int64, conclusion string) error {
	return r.db.Model(&models.TyPoliticalReview{}).
		Where("id = ? AND is_deleted = 0", id).
		Update("conclusion", conclusion).Error
}

// Update 更新政审记录。
func (r *PoliticalReviewRepository) Update(review *models.TyPoliticalReview) error {
	return r.db.Save(review).Error
}

// CheckAllPassed 检查指定发展对象的政审是否全部通过。
//
// 返回值：
//   - allPassed: 是否全部 pass（无 fail 和 basic_pass）
//   - hasBasicPass: 是否存在 basic_pass 结论
//   - hasFail: 是否存在 fail 结论
//   - err: 数据库错误
func (r *PoliticalReviewRepository) CheckAllPassed(developmentID int64) (allPassed bool, hasBasicPass bool, hasFail bool, err error) {
	var reviews []models.TyPoliticalReview
	if err := r.db.Where("development_id = ? AND is_deleted = 0", developmentID).
		Find(&reviews).Error; err != nil {
		return false, false, false, err
	}

	for _, rev := range reviews {
		switch rev.Conclusion {
		case "pass":
			continue
		case "basic_pass":
			hasBasicPass = true
		case "fail":
			hasFail = true
		}
	}

	allPassed = !hasBasicPass && !hasFail && len(reviews) > 0
	return allPassed, hasBasicPass, hasFail, nil
}

// CountByDevelopmentID 统计发展对象的政审记录数量。
func (r *PoliticalReviewRepository) CountByDevelopmentID(developmentID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.TyPoliticalReview{}).
		Where("development_id = ? AND is_deleted = 0", developmentID).
		Count(&count).Error
	return count, err
}

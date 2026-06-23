// Package cmp 综合素质（CMP）模块仓储层：负责 GORM 数据访问。
package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// ScoreRepository 综合素质得分数据访问层。
type ScoreRepository struct {
	db *gorm.DB
}

// NewScoreRepository 创建得分仓储。
func NewScoreRepository(db *gorm.DB) *ScoreRepository {
	return &ScoreRepository{db: db}
}

// GetByStudentYear 按 (student_id, academic_year) 查总分快照。
func (r *ScoreRepository) GetByStudentYear(studentID int64, academicYear string) (*models.CmpScore, error) {
	var s models.CmpScore
	if err := r.db.Where("student_id = ? AND academic_year = ? AND is_deleted = 0", studentID, academicYear).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// ListByYear 按年度分页查询（可按 college_id 过滤）。
func (r *ScoreRepository) ListByYear(academicYear string, collegeID int64, page, pageSize int) ([]models.CmpScore, int64, error) {
	q := r.db.Model(&models.CmpScore{}).Where("is_deleted = 0 AND academic_year = ?", academicYear)
	if collegeID > 0 {
		q = q.Where("student_id IN (SELECT id FROM idx_student WHERE college_id = ? AND is_deleted = 0)", collegeID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var scores []models.CmpScore
	offset := (page - 1) * pageSize
	if err := q.Order("total_score DESC, id ASC").Offset(offset).Limit(pageSize).Find(&scores).Error; err != nil {
		return nil, 0, err
	}
	return scores, total, nil
}

// ListAllStudents 列出全部学生（用于全量重算）。
func (r *ScoreRepository) ListAllStudents() ([]int64, error) {
	var ids []int64
	if err := r.db.Model(&models.IdxStudent{}).
		Where("is_deleted = 0").
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// ListStudentsByCollege 列出指定院系的学生 ID。
func (r *ScoreRepository) ListStudentsByCollege(collegeID int64) ([]int64, error) {
	var ids []int64
	if err := r.db.Model(&models.IdxStudent{}).
		Where("is_deleted = 0 AND college_id = ?", collegeID).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// UpsertScore 写入或更新学生总分快照（按 student_id + academic_year 唯一）。
func (r *ScoreRepository) UpsertScore(score *models.CmpScore) error {
	var existing models.CmpScore
	err := r.db.Where("student_id = ? AND academic_year = ? AND is_deleted = 0",
		score.StudentID, score.AcademicYear).First(&existing).Error
	if err == nil {
		// 更新
		score.ID = existing.ID
		return r.db.Save(score).Error
	}
	return r.db.Create(score).Error
}

// ReplaceDetails 替换某 score 的全部明细（事务内先删后插）。
// 注意：写入前必须给每条明细回填 score_id，否则 NOT NULL 约束会失败、整事务回滚，
// 导致 cmp_score 总分已写但 cmp_score_detail 永远为空（前端雷达图 / 维度分也会全为 0）。
func (r *ScoreRepository) ReplaceDetails(scoreID int64, details []models.CmpScoreDetail) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("score_id = ?", scoreID).Delete(&models.CmpScoreDetail{}).Error; err != nil {
			return err
		}
		if len(details) == 0 {
			return nil
		}
		// 回填 score_id（calculator 聚合时 score.ID 尚未生成，这里补齐外键）
		for i := range details {
			details[i].ScoreID = scoreID
		}
		return tx.Create(&details).Error
	})
}

// GetDetails 按 score_id 查明细，按 dimension, id 排序。
func (r *ScoreRepository) GetDetails(scoreID int64) ([]models.CmpScoreDetail, error) {
	var details []models.CmpScoreDetail
	if err := r.db.Where("score_id = ? AND is_deleted = 0", scoreID).
		Order("dimension ASC, id ASC").
		Find(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

// GetByID 按主键查总分。
func (r *ScoreRepository) GetByID(id int64) (*models.CmpScore, error) {
	var s models.CmpScore
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// ---- 规则版本 ----

// GetActiveRuleVersion 取当前激活的规则版本（仅 1 个）。
func (r *ScoreRepository) GetActiveRuleVersion() (*models.CmpRuleVersion, error) {
	var v models.CmpRuleVersion
	if err := r.db.Where("is_active = 1 AND is_deleted = 0").First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

// ListRuleVersions 列出全部规则版本（按 id DESC）。
func (r *ScoreRepository) ListRuleVersions() ([]models.CmpRuleVersion, error) {
	var vs []models.CmpRuleVersion
	if err := r.db.Where("is_deleted = 0").Order("id DESC").Find(&vs).Error; err != nil {
		return nil, err
	}
	return vs, nil
}

// CreateRuleVersion 创建规则版本。
func (r *ScoreRepository) CreateRuleVersion(v *models.CmpRuleVersion) error {
	return r.db.Create(v).Error
}

// ActivateRuleVersion 激活指定版本（事务：先清空其它激活位）。
func (r *ScoreRepository) ActivateRuleVersion(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.CmpRuleVersion{}).
			Where("is_deleted = 0").
			Update("is_active", 0).Error; err != nil {
			return err
		}
		return tx.Model(&models.CmpRuleVersion{}).
			Where("id = ? AND is_deleted = 0", id).
			Update("is_active", 1).Error
	})
}

// ---- 维度统计 ----

// CountByDimension 统计各维度子项数量（用于检视明细）。
func (r *ScoreRepository) CountByDimension(scoreID int64) (map[string]int, error) {
	type row struct {
		Dimension string
		Count     int
	}
	var rows []row
	if err := r.db.Model(&models.CmpScoreDetail{}).
		Select("dimension, count(*) as count").
		Where("score_id = ? AND is_deleted = 0", scoreID).
		Group("dimension").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	res := make(map[string]int, len(rows))
	for _, r := range rows {
		res[r.Dimension] = r.Count
	}
	return res, nil
}

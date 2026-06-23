// Package idgen 提供业务编号生成服务（ADR-013）。
// 格式：<MODULE>-<YYYY>-<4位流水>，年内最大 9999，跨年从 0 重置。
package idgen

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// NextBizNo 事务原子操作 biz_seq 表，生成下一个业务编号。
// module 参数如 "TY"、"ST"、"SQ"、"QG"。
func NextBizNo(db *gorm.DB, module string) (string, error) {
	year := time.Now().Year()

	// 尝试更新已有记录
	result := db.Model(&models.BizSeq{}).
		Where("module = ? AND year = ?", module, year).
		Update("cur", gorm.Expr("cur + 1"))

	if result.Error != nil {
		return "", result.Error
	}

	if result.RowsAffected == 0 {
		// 首次生成该模块该年度的编号
		seq := models.BizSeq{
			Module: module,
			Year:   year,
			Cur:    1,
		}
		if err := db.Create(&seq).Error; err != nil {
			// 并发竞争：另一个事务可能已创建
			// 回退到更新
			result = db.Model(&models.BizSeq{}).
				Where("module = ? AND year = ?", module, year).
				Update("cur", gorm.Expr("cur + 1"))
			if result.Error != nil {
				return "", result.Error
			}
			// 重新查询当前值
			var seq models.BizSeq
			if err := db.Where("module = ? AND year = ?", module, year).First(&seq).Error; err != nil {
				return "", err
			}
			return fmt.Sprintf("%s-%d-%04d", module, year, seq.Cur), nil
		}
		return fmt.Sprintf("%s-%d-%04d", module, year, 1), nil
	}

	// 查询更新后的值
	var seq models.BizSeq
	if err := db.Where("module = ? AND year = ?", module, year).First(&seq).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%d-%04d", module, year, seq.Cur), nil
}

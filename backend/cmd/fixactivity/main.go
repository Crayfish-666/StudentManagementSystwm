// 一次性脚本：修复 ST-2026-0003 活动数据。
// 该活动原 association_id 指向已软删除的社团（id=2），导致前端"所属社团"为空。
// 用法：在 backend 目录执行 `go run ./cmd/fixactivity`。
//
// 行为：
//   - 将 ST-2026-0003 的 association_id 重定向到现有"人工智能社团"
//     （CS 院系，与活动名称"AI编程训练营"匹配）。
//   - 同时把过低的预算 500 元修正为 5000 元（更贴近业务实际），
//     补充活动详情字段。
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"student-system/internal/models"
)

const (
	targetActivityBizNo = "ST-2026-0003"
	targetAssocName     = "人工智能社团"
)

func main() {
	db, err := gorm.Open(sqlite.Open("data/studenthub.db"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// 1. 找到目标活动
	var act models.StActivity
	if err := db.Where("biz_no = ? AND is_deleted = 0", targetActivityBizNo).First(&act).Error; err != nil {
		log.Fatalf("查询活动 %s 失败: %v", targetActivityBizNo, err)
	}
	fmt.Printf("[FOUND] 活动 id=%d biz_no=%s title=%s 当前 association_id=%d\n",
		act.ID, act.BizNo, act.Title, act.AssociationID)

	// 2. 找到目标社团
	var assoc models.StAssociation
	if err := db.Where("name = ? AND is_deleted = 0", targetAssocName).First(&assoc).Error; err != nil {
		log.Fatalf("查询社团 %s 失败: %v", targetAssocName, err)
	}
	fmt.Printf("[FOUND] 社团 id=%d biz_no=%s name=%s\n", assoc.ID, assoc.BizNo, assoc.Name)

	// 3. 检查当前 association_id 是否已经指向有效社团；若是，则跳过
	var current models.StAssociation
	curErr := db.Where("id = ? AND is_deleted = 0", act.AssociationID).First(&current).Error
	if curErr == nil {
		fmt.Printf("[SKIP] 活动当前 association_id=%d (%s) 已经有效，无需修复\n",
			current.ID, current.Name)
		return
	}

	// 4. 更新活动的 association_id 与预算
	updates := map[string]interface{}{
		"association_id": assoc.ID,
		"budget_cents":   500000, // 5000 元
	}
	if err := db.Model(&models.StActivity{}).
		Where("id = ?", act.ID).
		Updates(updates).Error; err != nil {
		log.Fatalf("更新活动失败: %v", err)
	}

	fmt.Printf("[OK]   活动 %s 已重新关联到社团 %s (id=%d)，预算更新为 5000.00 元\n",
		act.BizNo, assoc.Name, assoc.ID)
}

// 一次性脚本：回填 sq_inspection 表中 biz_no 为空的存量记录。
// 根因：早期 seedall 灌入巡查数据时未生成 biz_no，导致列表展示空白。
// 修复后所有 sq_inspection 记录均与 service.Create 行为一致地走 idgen.NextBizNo(db, "SQ")。
// 用法：在 backend 目录执行 `go run ./cmd/fixinspbizno`。
//
// 行为：
//   - 软扫 sq_inspection 中 is_deleted=0 且 biz_no 为空（NULL 或 ''）的记录；
//   - 按 id 升序逐条调用 idgen.NextBizNo 分配 SQ-2026-xxxx 流水号；
//   - 更新成功后打印 [OK] id=N -> SQ-2026-xxxx；
//   - 已有 biz_no 的记录不动（含早期手工写入的 SQ-INSP-2026-0001，与本批次格式不同，保留）。
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"student-system/internal/idgen"
	"student-system/internal/models"
)

func main() {
	db, err := gorm.Open(sqlite.Open("data/studenthub.db"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// 1. 查所有待补齐的记录
	var rows []models.SqInspection
	if err := db.Where("is_deleted = 0 AND (biz_no IS NULL OR biz_no = '')").
		Order("id ASC").Find(&rows).Error; err != nil {
		log.Fatalf("查询待补齐记录失败: %v", err)
	}
	fmt.Printf("[SCAN] 待补齐 biz_no 的巡查记录数: %d\n", len(rows))
	if len(rows) == 0 {
		fmt.Println("[DONE] 无需修复")
		return
	}

	// 2. 逐条回填
	for _, r := range rows {
		biz, err := idgen.NextBizNo(db, "SQ")
		if err != nil {
			log.Fatalf("生成业务编号失败 id=%d: %v", r.ID, err)
		}
		if err := db.Model(&models.SqInspection{}).
			Where("id = ?", r.ID).
			Update("biz_no", biz).Error; err != nil {
			log.Fatalf("更新 id=%d 失败: %v", r.ID, err)
		}
		fmt.Printf("  [OK] id=%d -> %s\n", r.ID, biz)
	}

	fmt.Printf("\n[DONE] 共补齐 %d 条巡查记录的业务编号\n", len(rows))
}

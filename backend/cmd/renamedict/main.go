// 一次性脚本：对 sys_dict 中 political_status 分类按 Sort 顺序重新分配 ID，
// 并把 204/206 的中文名更新为「共青团员 / 中共党员」。
// 用法：在 backend 目录执行 `go run ./cmd/renamedict`。
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// 旧 ID -> 新 (新 ID, 新编码, 新中文名, 新 Sort)
var remap = []struct {
	OldID     int
	NewID     int
	NewCode   string
	NewNameZh string
	NewSort   int
}{
	{201, 206, "masses", "群众", 6},               // 原 masses 后置为 206
	{202, 205, "activist", "入团积极分子", 5},     // activist -> 205
	{203, 204, "probationary", "预备团员", 4},     // probationary -> 204
	{204, 203, "member", "共青团员", 3},           // member -> 203, 名称修正
	{205, 202, "party_probationary", "预备党员", 2}, // 预备党员 -> 202
	{206, 201, "party_member", "中共党员", 1},     // 中共党员 -> 201, 名称修正
}

func main() {
	db, err := gorm.Open(sqlite.Open("data/studenthub.db?_pragma=foreign_keys(1)"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// 0. 检查现状
	var oldRows []struct {
		ID      int
		Code    string
		NameZh  string
		Sort    int
	}
	if err := db.Raw("SELECT id, code, name_zh, sort FROM sys_dict WHERE category='political_status' ORDER BY id").Scan(&oldRows).Error; err != nil {
		log.Fatalf("query old dicts: %v", err)
	}
	fmt.Println("--- 变更前 ---")
	for _, r := range oldRows {
		fmt.Printf("  id=%d code=%s name=%s sort=%d\n", r.ID, r.Code, r.NameZh, r.Sort)
	}

	// 1. 关闭外键
	if err := db.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		log.Fatalf("disable fk: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		log.Fatalf("begin tx: %v", tx.Error)
	}

	// 2. 先把 6 条记录搬到临时 ID 9001-9006，规避主键冲突
	for _, m := range remap {
		if err := tx.Exec("UPDATE sys_dict SET id = ? WHERE id = ?", m.OldID+9000, m.OldID).Error; err != nil {
			tx.Rollback()
			log.Fatalf("stage row old=%d: %v", m.OldID, err)
		}
	}

	// 3. 写回正式新 ID
	for _, m := range remap {
		if err := tx.Exec(
			"UPDATE sys_dict SET id = ?, code = ?, name_zh = ?, sort = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			m.NewID, m.NewCode, m.NewNameZh, m.NewSort, m.OldID+9000,
		).Error; err != nil {
			tx.Rollback()
			log.Fatalf("apply new old=%d -> new=%d: %v", m.OldID, m.NewID, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Fatalf("commit: %v", err)
	}

	// 4. 恢复外键
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		log.Fatalf("enable fk: %v", err)
	}

	// 5. 打印变更后状态
	var newRows []struct {
		ID     int
		Code   string
		NameZh string
		Sort   int
	}
	if err := db.Raw("SELECT id, code, name_zh, sort FROM sys_dict WHERE category='political_status' ORDER BY sort").Scan(&newRows).Error; err != nil {
		log.Fatalf("query new dicts: %v", err)
	}
	fmt.Println("--- 变更后（按 Sort 排序） ---")
	for _, r := range newRows {
		fmt.Printf("  id=%d code=%s name=%s sort=%d\n", r.ID, r.Code, r.NameZh, r.Sort)
	}
	fmt.Println("=== 完成 ===")
}

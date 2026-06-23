package main

// 一次性脚本：列出 SQLite 数据库中所有表名 + 数量，用于 S01 验收。
// 用法：在 backend 目录执行 `go run ./cmd/dbcheck`。
import (
	"fmt"
	"log"
	"sort"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(sqlite.Open("data/studenthub.db"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	var names []string
	if err := db.Raw(
		"SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name",
	).Scan(&names).Error; err != nil {
		log.Fatalf("query: %v", err)
	}
	sort.Strings(names)
	fmt.Printf("table_count=%d\n", len(names))
	for _, n := range names {
		fmt.Println(" -", n)
	}
}

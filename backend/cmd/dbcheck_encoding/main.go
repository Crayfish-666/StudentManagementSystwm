// dbcheck_encoding 直接读 DB 序列化 JSON 看输出
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TyRecommendationMeeting struct {
	ID             int64  `gorm:"column:id" json:"id"`
	BizNo          string `gorm:"column:biz_no" json:"biz_no"`
	DecisionReason string `gorm:"column:decision_reason" json:"decision_reason"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("data/studenthub.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("open err:", err)
		os.Exit(1)
	}
	var m TyRecommendationMeeting
	now := time.Now()
	if err := db.Raw("SELECT id, biz_no, decision_reason FROM ty_recommendation_meeting WHERE id=1").Scan(&m).Error; err != nil {
		fmt.Println("scan err:", err)
		os.Exit(1)
	}
	fmt.Println("scan took:", time.Since(now))
	fmt.Println("ID:", m.ID)
	fmt.Println("BizNo:", m.BizNo)
	fmt.Printf("DecisionReason string: %q\n", m.DecisionReason)
	fmt.Printf("DecisionReason bytes (Go internal): %s\n", hex.EncodeToString([]byte(m.DecisionReason)))

	// serialize to JSON
	b, _ := json.Marshal(&m)
	fmt.Println("JSON output hex:", hex.EncodeToString(b))
	fmt.Println("JSON output:", string(b))
}

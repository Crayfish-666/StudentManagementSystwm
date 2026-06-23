package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SysMenu struct {
	ID       int64  `gorm:"primaryKey;column:id"`
	ParentID *int64 `gorm:"column:parent_id"`
	Code     string `gorm:"column:code"`
	Title    string `gorm:"column:title"`
	Icon     string `gorm:"column:icon"`
	Path     string `gorm:"column:path"`
	Sort     int    `gorm:"column:sort"`
	Roles    string `gorm:"column:roles"`
}

func (SysMenu) TableName() string { return "sys_menu" }

func main() {
	db, _ := gorm.Open(sqlite.Open("data/studenthub.db?_pragma=foreign_keys(1)"), &gorm.Config{})

	// 1) 查 sys 顶级菜单的 id（作为 sys-job 的父）
	var sysParent SysMenu
	db.Where("code = ? AND parent_id IS NULL", "sys").First(&sysParent)
	fmt.Printf("sys id=%d\n", sysParent.ID)

	// 2) 插入 noti 顶级菜单
	noti := SysMenu{
		Code:  "noti",
		Title: "通知中心",
		Icon:  "Bell",
		Path:  "/notifications",
		Sort:  8,
		Roles: "[]",
	}
	if err := db.Where("code = ?", "noti").FirstOrCreate(&noti).Error; err != nil {
		fmt.Println("insert noti err:", err)
	} else {
		fmt.Printf("noti inserted/updated: id=%d\n", noti.ID)
	}

	// 3) 插入 sys-job 子菜单
	sysJob := SysMenu{
		ParentID: &sysParent.ID,
		Code:     "sys-job",
		Title:    "任务监控",
		Icon:     "",
		Path:     "/sys/job",
		Sort:     4,
		Roles:    `["R-SY-ADMIN"]`,
	}
	if err := db.Where("code = ?", "sys-job").FirstOrCreate(&sysJob).Error; err != nil {
		fmt.Println("insert sys-job err:", err)
	} else {
		fmt.Printf("sys-job inserted/updated: id=%d\n", sysJob.ID)
	}

	// 4) 同时补全 component 字段
	db.Model(&SysMenu{}).Where("code = ?", "noti").Update("component", "views/notifications/NotificationCenter.vue")
	db.Model(&SysMenu{}).Where("code = ?", "sys-job").Update("component", "views/sys/JobMonitor.vue")

	// 5) 打印最终结果
	var all []SysMenu
	db.Where("code = ? OR code = ?", "noti", "sys-job").Find(&all)
	for _, m := range all {
		fmt.Printf("  id=%d, parent=%v, code=%s, title=%s, path=%s, comp=%s, roles=%s\n",
			m.ID, fmt.Sprint(m.ParentID), m.Code, m.Title, m.Path, "", m.Roles)
	}
}

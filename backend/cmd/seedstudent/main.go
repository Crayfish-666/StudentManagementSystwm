// 一性次脚本：为数据库中每个 idx_student（is_deleted=0 且未绑定 sys_user）开通测试账号。
// 用法：在 backend 目录执行 `go run ./cmd/seedstudent`。
//
// 规则（与项目铁律对齐 docs/03 §4.2 / docs/05 S06）：
// - username 沿用现有命名风格 student{NN}，NN 取现有最大编号 +1；
// - password_hash 使用 authservice.HashPassword（bcrypt cost=12）；
// - 关联 R-STU-NORM 普通学生角色；
// - 已绑定的学生账号保持不变（不重置密码、不重命名）。
package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"student-system/internal/models"
	authservice "student-system/internal/modules/auth/service"
)

const (
	testPassword = "student@123"
	roleCode     = "R-STU-NORM"
)

var studentUserRe = regexp.MustCompile(`^student(\d+)$`)

func main() {
	db, err := gorm.Open(sqlite.Open("data/studenthub.db"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// 1. 查找 R-STU-NORM 角色 ID
	var role models.SysRole
	if err := db.Where("code = ? AND is_deleted = 0", roleCode).First(&role).Error; err != nil {
		log.Fatalf("查询角色 %s 失败: %v", roleCode, err)
	}

	// 2. 拉取所有有效学生
	var students []models.IdxStudent
	if err := db.Where("is_deleted = 0").Order("id ASC").Find(&students).Error; err != nil {
		log.Fatalf("查询学生失败: %v", err)
	}
	if len(students) == 0 {
		fmt.Println("无学生可处理，退出")
		return
	}

	// 3. 计算下一个 student{NN} 编号
	nextSeq, err := nextStudentSeq(db)
	if err != nil {
		log.Fatalf("计算学生账号序号失败: %v", err)
	}

	created := 0
	skipped := 0
	for _, stu := range students {
		// 检查是否已存在有效账号
		var existing models.SysUser
		err := db.Where("student_id = ? AND is_deleted = 0", stu.ID).First(&existing).Error
		if err == nil {
			fmt.Printf("[SKIP] student_id=%d student_no=%s name=%s 已有账号 username=%s\n",
				stu.ID, stu.StudentNo, stu.Name, existing.Username)
			skipped++
			continue
		}
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("查询学生 %d 的账号失败: %v", stu.ID, err)
		}

		username := fmt.Sprintf("student%02d", nextSeq)
		nextSeq++

		hash, hashErr := authservice.HashPassword(testPassword)
		if hashErr != nil {
			log.Fatalf("生成密码哈希失败: %v", hashErr)
		}

		sid := stu.ID
		user := models.SysUser{
			Username:     username,
			PasswordHash: hash,
			DisplayName:  stu.Name,
			Status:       "active",
			StudentID:    &sid,
		}
		if createErr := db.Create(&user).Error; createErr != nil {
			log.Fatalf("创建 sys_user 失败 student_no=%s: %v", stu.StudentNo, createErr)
		}

		// 关联普通学生角色
		ur := models.SysUserRole{
			UserID:    user.ID,
			RoleID:    role.ID,
			GrantedBy: &user.ID,
		}
		if urErr := db.Create(&ur).Error; urErr != nil {
			log.Fatalf("关联角色失败 user_id=%d: %v", user.ID, urErr)
		}

		fmt.Printf("[OK]   student_id=%d student_no=%s name=%s -> username=%s password=%s role=%s\n",
			stu.ID, stu.StudentNo, stu.Name, username, testPassword, roleCode)
		created++
	}

	fmt.Printf("\n=== 处理完成 ===  新建=%d  跳过=%d  总计=%d\n", created, skipped, len(students))
}

// nextStudentSeq 计算下一个可用的 student{NN} 编号。
func nextStudentSeq(db *gorm.DB) (int, error) {
	var usernames []string
	if err := db.Model(&models.SysUser{}).
		Where("is_deleted = 0").
		Pluck("username", &usernames).Error; err != nil {
		return 0, err
	}

	maxSeq := 0
	for _, u := range usernames {
		m := studentUserRe.FindStringSubmatch(u)
		if len(m) != 2 {
			continue
		}
		n, convErr := strconv.Atoi(m[1])
		if convErr != nil {
			continue
		}
		if n > maxSeq {
			maxSeq = n
		}
	}
	// 与现有命名风格一致：若已有 student01 则下一个为 student02；若无则从 01 开始
	if maxSeq == 0 {
		return 1, nil
	}
	return maxSeq + 1, nil
}

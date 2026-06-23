// 一次性脚本：为 ST 模块补充测试数据。
// 用法：在 backend 目录执行 `go run ./cmd/seedst`。
//
// 行为：
//   - 若 idx_student 不足 20 条，补足到 20 条（沿用现有院系，2023/2024 级各半）。
//   - 若 st_association 不足 4 条，按预设清单补足，覆盖 4 种状态 + 2 个院系。
//   - 同时为每个社团创建发起人（5 人）和成员记录。
package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/pkg/cryptox"
)

const targetStudentCount = 20

// 学生姓名池（足量）。
var studentNames = []string{
	"赵子轩", "钱欣怡", "孙思源", "李梓萱", "周浩然",
	"吴佳怡", "郑俊杰", "王晨曦", "冯雅婷", "陈博文",
	"褚雨桐", "卫子豪", "蒋哲瀚", "沈嘉怡", "韩沐辰",
	"杨若曦", "朱俊熙", "秦欣妍", "尤宇航", "许梓涵",
}

type assocSpec struct {
	Name          string
	CollegeCode   string // CS / EE
	BusinessScope string
	Status        string // preparing / trial / registered / rectifying
}

// 预设社团清单（4 个，覆盖 4 种状态 + 2 个院系）。
var assocSpecs = []assocSpec{
	{Name: "人工智能社团", CollegeCode: "CS", BusinessScope: "AI技术学习、算法竞赛、学术沙龙", Status: "preparing"},
	{Name: "开源软件社团", CollegeCode: "CS", BusinessScope: "开源项目实践、代码贡献、技术分享", Status: "trial"},
	{Name: "嵌入式开发社团", CollegeCode: "EE", BusinessScope: "嵌入式系统开发、硬件设计、机器人竞赛", Status: "registered"},
	{Name: "电子DIY社团", CollegeCode: "EE", BusinessScope: "电子电路设计、焊接实践、创客活动", Status: "rectifying"},
}

func main() {
	db, err := gorm.Open(sqlite.Open("data/studenthub.db"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// 1. 确保学生数量充足
	if err := ensureStudents(db); err != nil {
		log.Fatalf("补充学生失败: %v", err)
	}

	// 2. 拉取学生 + 院系映射
	var students []models.IdxStudent
	if err := db.Where("is_deleted = 0").Order("id ASC").Find(&students).Error; err != nil {
		log.Fatalf("拉取学生失败: %v", err)
	}
	if len(students) < 20 {
		log.Fatalf("学生数量不足 20，无法批量创建社团")
	}

	colleges := map[string]models.SysCollege{}
	var colList []models.SysCollege
	if err := db.Where("is_deleted = 0").Find(&colList).Error; err != nil {
		log.Fatalf("拉取院系失败: %v", err)
	}
	for _, c := range colList {
		colleges[c.Code] = c
	}

	// 3. 按 college_id 分组学生，便于挑选发起人
	studentsByCollege := map[int64][]models.IdxStudent{}
	for _, s := range students {
		if s.CollegeID == nil {
			continue
		}
		studentsByCollege[*s.CollegeID] = append(studentsByCollege[*s.CollegeID], s)
	}

	// 4. 创建社团（若名称已存在则跳过）
	created := 0
	for _, spec := range assocSpecs {
		col, ok := colleges[spec.CollegeCode]
		if !ok {
			log.Printf("[SKIP] 院系不存在: %s", spec.CollegeCode)
			continue
		}

		var exists int64
		db.Model(&models.StAssociation{}).Where("name = ? AND is_deleted = 0", spec.Name).Count(&exists)
		if exists > 0 {
			fmt.Printf("[SKIP] 社团已存在: %s\n", spec.Name)
			continue
		}

		pool := studentsByCollege[col.ID]
		if len(pool) < 5 {
			// 院系学生不足时，从全体学生池补
			pool = students
		}
		if len(pool) < 5 {
			log.Printf("[SKIP] 学生池不足 5 人，无法创建: %s", spec.Name)
			continue
		}

		founders := pool[:5]

		if err := createAssociation(db, spec, col.ID, founders); err != nil {
			log.Printf("[FAIL] 创建社团 %s 失败: %v", spec.Name, err)
			continue
		}
		fmt.Printf("[OK]   创建社团: %s (%s, %s)\n", spec.Name, col.Name, spec.Status)
		created++
	}

	// 5. 输出汇总
	var totalAssoc int64
	db.Model(&models.StAssociation{}).Where("is_deleted = 0").Count(&totalAssoc)
	var totalStu int64
	db.Model(&models.IdxStudent{}).Where("is_deleted = 0").Count(&totalStu)
	fmt.Printf("\n=== 处理完成 ===  本次新建社团=%d  当前社团总数=%d  学生总数=%d\n", created, totalAssoc, totalStu)
}

// ensureStudents 若学生不足 20，补足到 20。
func ensureStudents(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.IdxStudent{}).Where("is_deleted = 0").Count(&count).Error; err != nil {
		return err
	}
	if count >= targetStudentCount {
		fmt.Printf("学生数量已达 %d 人，无需补充\n", count)
		return nil
	}

	var colleges []models.SysCollege
	if err := db.Where("is_deleted = 0").Order("id ASC").Find(&colleges).Error; err != nil {
		return err
	}
	if len(colleges) == 0 {
		return fmt.Errorf("无可用院系")
	}

	// 找出现有最大学号序号，避免重复
	var existingNos []string
	db.Model(&models.IdxStudent{}).Pluck("student_no", &existingNos)
	used := map[string]struct{}{}
	for _, no := range existingNos {
		used[no] = struct{}{}
	}

	birthBase, _ := time.Parse("2006-01-02", "2005-01-01")
	enroll2023, _ := time.Parse("2006-01-02", "2023-09-01")
	enroll2024, _ := time.Parse("2006-01-02", "2024-09-01")

	created := 0
	idx := 0
	for count < targetStudentCount {
		if idx >= len(studentNames) {
			break
		}
		name := studentNames[idx]
		idx++

		// 生成不重复的学号：2023XXXX / 2024XXXX
		var studentNo string
		var grade int
		var enroll time.Time
		for seq := 100 + idx; seq < 9999; seq++ {
			if count%2 == 0 {
				studentNo = fmt.Sprintf("2023%04d", seq)
				grade = 2023
				enroll = enroll2023
			} else {
				studentNo = fmt.Sprintf("2024%04d", seq)
				grade = 2024
				enroll = enroll2024
			}
			if _, ok := used[studentNo]; !ok {
				used[studentNo] = struct{}{}
				break
			}
		}

		// 身份证号 / 手机号也要保证 hash 唯一
		idCard := fmt.Sprintf("310115%s%04d", studentNo[:8], idx)
		phone := fmt.Sprintf("138%08d", 10000000+idx)
		birth := birthBase.AddDate(0, idx, 0)
		col := colleges[idx%len(colleges)]
		colID := col.ID
		gradePtr := grade
		gender := "M"
		if idx%2 == 0 {
			gender = "F"
		}

		stu := models.IdxStudent{
			StudentNo:       studentNo,
			Name:            name,
			IDCardEnc:       cryptox.Encrypt(idCard),
			IDCardHash:      fmt.Sprintf("%x", sha256.Sum256([]byte(idCard))),
			Gender:          gender,
			BirthDate:       &birth,
			Ethnicity:       "汉族",
			PoliticalStatus: "member",
			CollegeID:       &colID,
			PhoneEnc:        cryptox.Encrypt(phone),
			PhoneHash:       fmt.Sprintf("%x", sha256.Sum256([]byte(phone))),
			Email:           fmt.Sprintf("stu%s@example.com", studentNo),
			EnrollmentAt:    &enroll,
			Grade:           &gradePtr,
			Status:          "enrolled",
		}
		if err := db.Create(&stu).Error; err != nil {
			return fmt.Errorf("创建学生 %s 失败: %w", name, err)
		}
		fmt.Printf("[OK]   补充学生: %s %s %s\n", studentNo, name, col.Name)
		created++
		count++
	}

	fmt.Printf("学生补充完成：本次新建=%d, 当前总数=%d\n", created, count)
	return nil
}

// createAssociation 创建一条社团及其发起人/成员记录。
func createAssociation(
	db *gorm.DB,
	spec assocSpec,
	collegeID int64,
	founders []models.IdxStudent,
) error {
	bizNo, err := idgen.NextBizNo(db, "ST")
	if err != nil {
		return fmt.Errorf("生成业务编号: %w", err)
	}

	now := time.Now()
	assoc := models.StAssociation{
		BizNo:         bizNo,
		Name:          spec.Name,
		CollegeID:     collegeID,
		BusinessScope: spec.BusinessScope,
		Status:        spec.Status,
		FoundedAt:     &now,
	}

	// 不同状态填上对应时间戳，便于前端展示
	switch spec.Status {
	case "trial":
		assoc.TrialStartedAt = &now
	case "registered":
		assoc.TrialStartedAt = ptrTime(now.AddDate(0, -3, 0))
		assoc.RegisteredAt = &now
	case "rectifying":
		assoc.TrialStartedAt = ptrTime(now.AddDate(0, -6, 0))
		assoc.RegisteredAt = ptrTime(now.AddDate(0, -3, 0))
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&assoc).Error; err != nil {
			return err
		}
		// 发起人
		for _, f := range founders {
			if err := tx.Create(&models.StFounder{
				AssociationID: assoc.ID,
				StudentID:     f.ID,
				JoinedAt:      now,
			}).Error; err != nil {
				return fmt.Errorf("创建发起人: %w", err)
			}
		}
		// 成员（首位作为社长，其余为会员）
		for i, f := range founders {
			role := "member"
			isCore := 0
			if i == 0 {
				role = "president"
				isCore = 1
			}
			if err := tx.Create(&models.StAssocMember{
				AssociationID: assoc.ID,
				StudentID:     f.ID,
				Role:          role,
				JoinedAt:      now,
				IsCoreOfficer: isCore,
			}).Error; err != nil {
				return fmt.Errorf("创建成员: %w", err)
			}
		}
		// registered 状态社团回填社长字段
		if spec.Status == "registered" || spec.Status == "rectifying" {
			presID := founders[0].ID
			if err := tx.Model(&models.StAssociation{}).
				Where("id = ?", assoc.ID).
				Update("president_student_id", presID).Error; err != nil {
				return fmt.Errorf("回填社长: %w", err)
			}
		}
		return nil
	})
}

func ptrTime(t time.Time) *time.Time { return &t }

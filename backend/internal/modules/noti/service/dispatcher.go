package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"student-system/internal/idgen"
	"student-system/internal/models"
	"student-system/internal/modules/noti/repository"
)

// Dispatcher 通知调度器，将事件映射为通知。
type Dispatcher struct {
	repo *repository.NotificationRepository
	db   *gorm.DB
}

// NewDispatcher 创建通知调度器。
func NewDispatcher(repo *repository.NotificationRepository, db *gorm.DB) *Dispatcher {
	return &Dispatcher{repo: repo, db: db}
}

// Dispatch 为每个接收人创建 channel=site 的通知。
func (d *Dispatcher) Dispatch(ctx context.Context, eventType string, recipientUserIDs []int64, title, content, linkURL string, priority string) error {
	if priority == "" {
		priority = "normal"
	}

	for _, uid := range recipientUserIDs {
		bizNo, err := idgen.NextBizNo(d.db, "NOTI")
		if err != nil {
			return fmt.Errorf("生成通知编号失败: %w", err)
		}

		now := time.Now()
		n := &models.Notification{
			BizNo:           bizNo,
			RecipientUserID: uid,
			Channel:         "site",
			Title:           title,
			Content:         content,
			LinkURL:         linkURL,
			Priority:        priority,
			IsRead:          0,
			SendStatus:      "sent",
			SentAt:          &now,
		}
		if err := d.repo.Create(n); err != nil {
			return fmt.Errorf("创建通知失败: %w", err)
		}
	}
	return nil
}

// DispatchL4Incident L4事件通知：查找楼栋管理员+辅导员，发送紧急通知。
func (d *Dispatcher) DispatchL4Incident(ctx context.Context, incidentID int64, buildingName string) error {
	// 查找事件对应的楼栋
	var incident models.SqIncident
	if err := d.db.Where("id = ? AND is_deleted = 0", incidentID).First(&incident).Error; err != nil {
		return fmt.Errorf("事件不存在: %w", err)
	}

	var recipientIDs []int64

	// 1. 查找楼栋管理员（TutorUserID）
	var building models.IdxDormBuilding
	if err := d.db.Where("id = ? AND is_deleted = 0", incident.BuildingID).First(&building).Error; err == nil {
		if building.TutorUserID != nil && *building.TutorUserID > 0 {
			recipientIDs = append(recipientIDs, *building.TutorUserID)
		}
	}

	// 2. 查找楼栋长（building_chief）
	var positions []models.SqSelfgovPosition
	if err := d.db.Where("scope_type = ? AND scope_id = ? AND position = ? AND status IN (?) AND is_deleted = 0",
		"building", incident.BuildingID, "building_chief", []string{"formal", "renewed"}).
		Find(&positions).Error; err == nil {
		for _, p := range positions {
			// 楼栋长是学生，需要找到对应的用户ID
			var student models.IdxStudent
			if err := d.db.Where("id = ? AND is_deleted = 0", p.StudentID).First(&student).Error; err == nil {
				var user models.SysUser
				if err := d.db.Where("student_id = ? AND is_deleted = 0", student.ID).First(&user).Error; err == nil {
					recipientIDs = append(recipientIDs, user.ID)
				}
			}
		}
	}

	// 3. 查找辅导员：通过楼栋内学生的班级找到辅导员
	var roomIDs []int64
	d.db.Model(&models.IdxDormRoom{}).Where("building_id = ? AND is_deleted = 0", incident.BuildingID).Pluck("id", &roomIDs)
	if len(roomIDs) > 0 {
		var studentIDs []int64
		d.db.Model(&models.IdxDormBed{}).Where("room_id IN (?) AND occupant_student_id IS NOT NULL AND is_deleted = 0", roomIDs).Pluck("occupant_student_id", &studentIDs)
		if len(studentIDs) > 0 {
			var classIDs []int64
			d.db.Model(&models.IdxStudent{}).Where("id IN (?) AND class_id IS NOT NULL AND is_deleted = 0", studentIDs).Pluck("class_id", &classIDs)
			if len(classIDs) > 0 {
				var counselorIDs []int64
				d.db.Model(&models.IdxClass{}).Where("id IN (?) AND counselor_id IS NOT NULL AND is_deleted = 0", classIDs).Pluck("counselor_id", &counselorIDs)
				recipientIDs = append(recipientIDs, counselorIDs...)
			}
		}
	}

	// 去重
	recipientIDs = uniqueInt64(recipientIDs)

	if len(recipientIDs) == 0 {
		return nil
	}

	title := fmt.Sprintf("【紧急】%s 发生 L4 级紧急事件", buildingName)
	content := fmt.Sprintf("楼栋 %s 发生 L4 级紧急事件，请立即处理。", buildingName)
	linkURL := fmt.Sprintf("/sq/incident/%d", incidentID)

	return d.Dispatch(ctx, "SqIncidentL4Raised", recipientIDs, title, content, linkURL, "urgent")
}

// DispatchOverdueCultivation 培养记录超期预警。
func (d *Dispatcher) DispatchOverdueCultivation(ctx context.Context, studentID int64, studentName string) error {
	// 查找学生对应的用户
	var user models.SysUser
	if err := d.db.Where("student_id = ? AND is_deleted = 0", studentID).First(&user).Error; err != nil {
		return nil // 学生无账号则跳过
	}

	title := "培养记录超期提醒"
	content := fmt.Sprintf("同学 %s 的培养记录已超期，请尽快完成。", studentName)

	return d.Dispatch(ctx, "OverdueCultivation", []int64{user.ID}, title, content, "", "normal")
}

// DispatchLateReturn 晚归累计告警。
func (d *Dispatcher) DispatchLateReturn(ctx context.Context, studentID int64, studentName string, count int) error {
	// 查找学生对应的用户
	var user models.SysUser
	if err := d.db.Where("student_id = ? AND is_deleted = 0", studentID).First(&user).Error; err != nil {
		return nil
	}

	// 查找辅导员
	var student models.IdxStudent
	var counselorUserID int64
	if err := d.db.Where("id = ? AND is_deleted = 0", studentID).First(&student).Error; err == nil && student.ClassID != nil {
		var class models.IdxClass
		if err := d.db.Where("id = ? AND is_deleted = 0", *student.ClassID).First(&class).Error; err == nil && class.CounselorID != nil {
			counselorUserID = *class.CounselorID
		}
	}

	recipientIDs := []int64{user.ID}
	if counselorUserID > 0 {
		recipientIDs = append(recipientIDs, counselorUserID)
	}

	title := "晚归累计告警"
	content := fmt.Sprintf("同学 %s 本学期晚归已达 %d 次，请关注。", studentName, count)

	return d.Dispatch(ctx, "LateReturnAccumulated", recipientIDs, title, content, "", "high")
}

// uniqueInt64 对 int64 切片去重。
func uniqueInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{})
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

// Package cmp 综合素质事件订阅器：订阅 4 大模块事件触发增量重算。
package event

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"student-system/internal/eventx"
	cmpservice "student-system/internal/modules/cmp/service"
)

// EventSubscriber 监听 4 大模块关键事件，触发对应学生的综合分增量重算。
type EventSubscriber struct {
	db    *gorm.DB
	svc   *cmpservice.ScoreService
	zlog  *zap.Logger
}

// NewEventSubscriber 创建事件订阅器。
func NewEventSubscriber(db *gorm.DB, svc *cmpservice.ScoreService, zlog *zap.Logger) *EventSubscriber {
	return &EventSubscriber{db: db, svc: svc, zlog: zlog}
}

// RegisterBusSubscriptions 在事件总线上注册订阅。
//
// 监听以下事件：
//   - TyApplicationApproved      团员发展审批通过
//   - TyApplicationRejected      团员发展审批驳回
//   - StActivityCheckin          社团活动签到
//   - SqIncidentClosed           社区事件结案
//   - QgPayrollIssued            勤工薪酬生成
//   - QgDifficultyApproved       困难认定终审
func (s *EventSubscriber) RegisterBusSubscriptions(bus *eventx.Bus) {
	bus.Subscribe("TyApplicationApproved", s.handleStudentScopedEvent)
	bus.Subscribe("TyApplicationRejected", s.handleStudentScopedEvent)
	bus.Subscribe("StActivityCheckin", s.handleStudentScopedEvent)
	bus.Subscribe("SqIncidentClosed", s.handleStudentScopedEvent)
	bus.Subscribe("QgPayrollIssued", s.handleStudentScopedEvent)
	bus.Subscribe("QgDifficultyApproved", s.handleStudentScopedEvent)
}

// handleStudentScopedEvent 从 payload 中提取 student_id，触发重算。
//
// 事件 payload 中须包含 student_id（int64 或 string）。重算失败仅记录日志。
func (s *EventSubscriber) handleStudentScopedEvent(evt *eventx.Event) error {
	if evt == nil || evt.Payload == nil {
		return nil
	}
	studentID, ok := extractStudentID(evt.Payload)
	if !ok || studentID <= 0 {
		return nil
	}
	ctx := context.Background()
	if err := s.svc.RecomputeFromEvent(ctx, studentID); err != nil {
		s.zlog.Warn("事件触发综合分重算失败",
			zap.String("event_type", evt.EventType),
			zap.Int64("student_id", studentID),
			zap.Error(err))
		return err
	}
	s.zlog.Info("事件触发综合分重算成功",
		zap.String("event_type", evt.EventType),
		zap.Int64("student_id", studentID))
	return nil
}

// extractStudentID 兼容 int64 / float64 / string / json.Number。
func extractStudentID(payload map[string]interface{}) (int64, bool) {
	v, ok := payload["student_id"]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

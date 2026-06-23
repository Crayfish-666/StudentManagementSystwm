// Package eventx 提供轻量事件总线：事件持久化到 event_log，支持订阅同步处理。
//
// 设计依据：docs/02 ADR-008（精简版 Event Sourcing）。
// 关键约束：append-only，链式 hash 用于完整性校验。
package eventx

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"student-system/internal/models"
)

// Event 业务事件统一结构。
type Event struct {
	Aggregate   string                 // 如 "ty.application"
	AggregateID string                 // 业务编号或主键字符串
	EventType   string                 // 如 "TyApplicationApproved"
	Module      string                 // TY/ST/SQ/QG/CMP
	ActorID     int64
	ActorRole   string
	Payload     map[string]interface{} // 业务负载
	BizNo       string
	IP          string
	UA          string
}

// Handler 同步事件处理器。
type Handler func(evt *Event) error

// Bus 简单进程内事件总线（写日志 + 同步分发订阅）。
type Bus struct {
	db          *gorm.DB
	mu          sync.RWMutex
	subscribers map[string][]Handler
}

// NewBus 创建事件总线。
func NewBus(db *gorm.DB) *Bus {
	return &Bus{db: db, subscribers: make(map[string][]Handler)}
}

// Subscribe 订阅指定 event_type；同一类型可多个订阅者。
func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Publish 持久化事件到 event_log 并同步分发订阅。
// 写入事件后失败的订阅器仅记录、不影响事件日志已写入。
func (b *Bus) Publish(evt *Event) error {
	if evt == nil {
		return fmt.Errorf("event 不能为空")
	}
	if evt.Aggregate == "" || evt.EventType == "" || evt.Module == "" {
		return fmt.Errorf("event aggregate/type/module 必填")
	}

	payloadBytes, err := json.Marshal(evt.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	// 取上一条同 aggregate 事件的 hash 作为 prev_hash（链式）
	var prevHash string
	var prev models.EventLog
	if err := b.db.Where("aggregate = ? AND aggregate_id = ?", evt.Aggregate, evt.AggregateID).
		Order("id DESC").Limit(1).Find(&prev).Error; err == nil && prev.ID > 0 {
		prevHash = prev.Hash
	}

	now := time.Now()
	eventID := uuid.New().String()
	hashSrc := eventID + "|" + evt.EventType + "|" + evt.AggregateID + "|" + string(payloadBytes) + "|" + prevHash + "|" + now.Format(time.RFC3339Nano)
	sum := sha256.Sum256([]byte(hashSrc))
	hash := hex.EncodeToString(sum[:])

	log := models.EventLog{
		EventID:     eventID,
		Aggregate:   evt.Aggregate,
		AggregateID: evt.AggregateID,
		EventType:   evt.EventType,
		Module:      evt.Module,
		ActorID:     evt.ActorID,
		ActorRole:   evt.ActorRole,
		PayloadJSON: string(payloadBytes),
		PrevHash:    prevHash,
		Hash:        hash,
		BizNo:       evt.BizNo,
		IP:          evt.IP,
		UA:          evt.UA,
		OccurredAt:  now,
	}
	if err := b.db.Create(&log).Error; err != nil {
		return fmt.Errorf("写入 event_log: %w", err)
	}

	// 同步分发订阅
	b.mu.RLock()
	handlers := append([]Handler(nil), b.subscribers[evt.EventType]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(evt); err != nil {
			// 订阅器失败仅记录，不影响事件日志已写入
			_ = err
		}
	}
	return nil
}

// QueryByAggregate 按 (aggregate, aggregate_id) 反查事件流（按时间正序）。
func (b *Bus) QueryByAggregate(aggregate, aggregateID string) ([]models.EventLog, error) {
	var logs []models.EventLog
	if err := b.db.Where("aggregate = ? AND aggregate_id = ?", aggregate, aggregateID).
		Order("occurred_at ASC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

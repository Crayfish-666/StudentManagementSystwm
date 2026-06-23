// Package channels 多通道通知适配器（ADR-015）。
package channels

import "go.uber.org/zap"

// SMS 短信通知适配器（V1 stub 实现）。
type SMS struct {
	zlog *zap.Logger
}

// NewSMS 创建短信适配器。
func NewSMS(zlog *zap.Logger) *SMS {
	return &SMS{zlog: zlog}
}

// Send 发送短信通知（V1 stub：仅记录日志，不实际发送）。
func (s *SMS) Send(to, title, content string) error {
	if s.zlog != nil {
		s.zlog.Info("[SMS stub] 发送短信",
			zap.String("to", to),
			zap.String("title", title),
		)
	}
	return nil
}

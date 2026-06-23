package channels

import "go.uber.org/zap"

// Email 邮件通知适配器（V1 stub 实现）。
type Email struct {
	zlog *zap.Logger
}

// NewEmail 创建邮件适配器。
func NewEmail(zlog *zap.Logger) *Email {
	return &Email{zlog: zlog}
}

// Send 发送邮件通知（V1 stub：仅记录日志，不实际发送）。
func (e *Email) Send(to, title, content string) error {
	if e.zlog != nil {
		e.zlog.Info("[Email stub] 发送邮件",
			zap.String("to", to),
			zap.String("title", title),
		)
	}
	return nil
}

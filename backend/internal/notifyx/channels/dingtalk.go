package channels

import "go.uber.org/zap"

// DingTalk 钉钉通知适配器（V1 stub 实现）。
type DingTalk struct {
	zlog *zap.Logger
}

// NewDingTalk 创建钉钉适配器。
func NewDingTalk(zlog *zap.Logger) *DingTalk {
	return &DingTalk{zlog: zlog}
}

// Send 发送钉钉通知（V1 stub：仅记录日志，不实际发送）。
func (d *DingTalk) Send(to, title, content string) error {
	if d.zlog != nil {
		d.zlog.Info("[DingTalk stub] 发送钉钉消息",
			zap.String("to", to),
			zap.String("title", title),
		)
	}
	return nil
}

package channels

import "go.uber.org/zap"

// WeCom 企业微信通知适配器（V1 stub 实现）。
type WeCom struct {
	zlog *zap.Logger
}

// NewWeCom 创建企微适配器。
func NewWeCom(zlog *zap.Logger) *WeCom {
	return &WeCom{zlog: zlog}
}

// Send 发送企微通知（V1 stub：仅记录日志，不实际发送）。
func (w *WeCom) Send(to, title, content string) error {
	if w.zlog != nil {
		w.zlog.Info("[WeCom stub] 发送企微消息",
			zap.String("to", to),
			zap.String("title", title),
		)
	}
	return nil
}

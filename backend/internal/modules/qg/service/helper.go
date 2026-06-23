package service

import (
	"fmt"
	"time"
)

// BizError 业务错误（携带错误码）。
type BizError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Error 实现 error 接口。
func (e *BizError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

// parseTime 解析时间字符串。
func parseTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05+08:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", s)
}

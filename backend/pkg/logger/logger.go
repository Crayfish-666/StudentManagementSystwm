// Package logger 提供基于 zap 的结构化日志（ADR-020）。
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 根据环境创建 zap Logger。
// env=dev 使用 ConsoleEncoder，便于本地观察；其他环境输出 JSON 结构化日志。
func New(env string) (*zap.Logger, error) {
	var cfg zap.Config
	if env == "prod" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build()
}

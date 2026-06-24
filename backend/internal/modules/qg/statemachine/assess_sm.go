// Package statemachine 定义 QG 模块各业务对象的状态机。
//
// 设计依据：docs/01 §7.3.6 月度考核、docs/05 §10；docs/02 ADR-007/008。
// 月度考核两态模型：S1 待确认（用人部门提交） → S3 已确认（学生处/财务复核）。
// 本文件仅声明状态与动作常量 + 跃迁规则；Guard/Effect/持久化由 Service 层负责。
package statemachine

import (
	"student-system/internal/statem"
)

// 状态常量（与 docs/03 §8.2.7 status check 约束保持一致）。
const (
	StatePending  = "S1" // 待确认
	StateConfirmed = "S3" // 已确认
)

// 动作常量。
const (
	ActionConfirm = "confirm" // 复核/确认：S1 → S3
)

// NewAssessSM 构建月度考核的状态机。
func NewAssessSM() *statem.Engine {
	sm := statem.New("qg.monthly_assess")

	// 学生处 / 财务管理员确认月度考核
	sm.Allow(StatePending, ActionConfirm, StateConfirmed)

	return sm
}

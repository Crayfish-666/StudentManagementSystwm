// Package statemachine 定义 TY 模块各业务对象的状态机。
//
// 设计依据：docs/01 §3.2、§4.3.2；docs/05 §6.2；docs/02 ADR-007/008。
// 入团申请三级审批链：辅导员初审 → 院系团委复核 → 校团委终审。
package statemachine

import (
	"student-system/internal/statem"
)

// 状态常量（与 docs/01 §3.2 5 态模型对齐）。
const (
	StateDraft    = "S0" // 草稿
	StatePending  = "S1" // 待审（辅导员待审）
	StateInReview = "S2" // 审批中（院系/校级审批中）
	StatePassed   = "S3" // 终审通过
	StateRejected = "S4" // 驳回/终止
)

// 动作常量。
const (
	ActionSubmit            = "submit"
	ActionWithdraw          = "withdraw"
	ActionApproveCounselor  = "counselor_approve"
	ActionApproveCollege    = "college_approve"
	ActionApproveSchool     = "school_approve"
	ActionRejectCounselor   = "counselor_reject"
	ActionRejectCollege     = "college_reject"
	ActionRejectSchool      = "school_reject"
)

// 审批步骤标签（写入 ty_approval_record.step）。
const (
	StepCounselor = "counselor"
	StepCollege   = "college"
	StepSchool    = "school"
)

// NewApplicationSM 构建入团申请的状态机。
// 不在此处绑定 Guard/Effect（业务规则与持久化由 Service 层负责）。
func NewApplicationSM() *statem.Engine {
	sm := statem.New("ty.application")

	// 学生侧
	sm.Allow(StateDraft, ActionSubmit, StatePending)
	sm.Allow(StatePending, ActionWithdraw, StateDraft)

	// 三级审批：通过链路 S1 → S2 → S2 → S3
	sm.Allow(StatePending, ActionApproveCounselor, StateInReview)
	sm.Allow(StateInReview, ActionApproveCollege, StateInReview)
	sm.Allow(StateInReview, ActionApproveSchool, StatePassed)

	// 三级审批：驳回链路（任意环节驳回 → S4）
	sm.Allow(StatePending, ActionRejectCounselor, StateRejected)
	sm.Allow(StateInReview, ActionRejectCollege, StateRejected)
	sm.Allow(StateInReview, ActionRejectSchool, StateRejected)

	return sm
}

// StepOfAction 由动作还原审批步骤。
func StepOfAction(action string) string {
	switch action {
	case ActionApproveCounselor, ActionRejectCounselor:
		return StepCounselor
	case ActionApproveCollege, ActionRejectCollege:
		return StepCollege
	case ActionApproveSchool, ActionRejectSchool:
		return StepSchool
	}
	return ""
}

// ResolveAction 根据 (step, result) 还原动作。
func ResolveAction(step, result string) string {
	switch step {
	case StepCounselor:
		if result == "approve" {
			return ActionApproveCounselor
		}
		return ActionRejectCounselor
	case StepCollege:
		if result == "approve" {
			return ActionApproveCollege
		}
		return ActionRejectCollege
	case StepSchool:
		if result == "approve" {
			return ActionApproveSchool
		}
		return ActionRejectSchool
	}
	return ""
}

// Package statemachine 定义 ST 模块招新计划的状态机。
//
// 设计依据：docs/01 §5.3.4 招新子流程；docs/02 ADR-007/008；docs/03 §6.2.5。
// 招新计划状态机四态模型：S0 草稿 → S1 待审 → S3 通过（已发布）/ S4 驳回。
package statemachine

import (
	"student-system/internal/statem"
)

// 招新计划状态常量（与 docs/03 §6.2.5 check 约束对齐）。
const (
	PlanStateDraft    = "S0" // 草稿
	PlanStatePending  = "S1" // 待审
	PlanStatePassed   = "S3" // 通过（已发布可投递）
	PlanStateRejected = "S4" // 驳回
)

// 招新计划动作常量。
const (
	PlanActionSubmit   = "submit"
	PlanActionWithdraw = "withdraw"
	PlanActionApprove  = "approve"
	PlanActionReject   = "reject"
	PlanActionPublish  = "publish"
)

// NewRecruitPlanSM 构建招新计划状态机。
//
// 状态转移：
//   S0 --submit--> S1
//   S1 --withdraw--> S0
//   S1 --approve--> S3
//   S1 --reject--> S4
//   S3 --publish--> S3（发布动作仅刷新 result_deadline，不改变状态）
func NewRecruitPlanSM() *statem.Engine {
	sm := statem.New("st.recruit_plan")
	sm.Allow(PlanStateDraft, PlanActionSubmit, PlanStatePending)
	sm.Allow(PlanStatePending, PlanActionWithdraw, PlanStateDraft)
	sm.Allow(PlanStatePending, PlanActionApprove, PlanStatePassed)
	sm.Allow(PlanStatePending, PlanActionReject, PlanStateRejected)
	// 已通过的计划再次发布：保持 S3
	sm.Allow(PlanStatePassed, PlanActionPublish, PlanStatePassed)
	return sm
}

// Package statemachine 定义 ST 模块各业务对象的状态机。
//
// 设计依据：docs/01 §5.3.5；docs/02 ADR-007/008。
// 活动分级审批链：按活动等级动态生成 1~5 级审批节点。
package statemachine

import (
	"fmt"

	"student-system/internal/statem"
)

// 状态常量（与 docs/01 §5.3.5 五态模型对齐）。
const (
	StateDraft    = "S0" // 草稿
	StatePending  = "S1" // 待审
	StateInReview = "S2" // 审批中（中间步骤）
	StatePassed   = "S3" // 终审通过
	StateRejected = "S4" // 驳回
)

// 动作常量。
const (
	ActionSubmit   = "submit"
	ActionWithdraw = "withdraw"
)

// 审批步骤编号常量。
const (
	StepTutor  = 1 // 指导教师
	StepCollege = 2 // 院系
	StepUnion   = 3 // 校社联
	StepLeague  = 4 // 校团委
	StepLeader  = 5 // 校领导
)

// 审批步骤标签。
const (
	StepTextTutor   = "tutor"
	StepTextCollege = "college"
	StepTextUnion   = "union"
	StepTextLeague  = "league"
	StepTextLeader  = "leader"
)

// 审批步骤角色映射。
var stepRoleMap = map[int][]string{
	StepTutor:   {"R-COL-TUTOR", "R-SY-ADMIN"},
	StepCollege: {"R-COL-COUN", "R-COL-LEAGUE", "R-SY-ADMIN"},
	StepUnion:   {"R-SY-ADMIN"},
	StepLeague:  {"R-SY-LEAGUE", "R-SY-ADMIN"},
	StepLeader:  {"R-SY-ADMIN"},
}

// 审批步骤中文映射。
var stepTextCNMap = map[int]string{
	StepTutor:   "指导教师审批",
	StepCollege: "院系审批",
	StepUnion:   "校社联审批",
	StepLeague:  "校团委审批",
	StepLeader:  "校领导审批",
}

// 审批步骤编号到标签映射。
var stepNoToTextMap = map[int]string{
	StepTutor:   StepTextTutor,
	StepCollege: StepTextCollege,
	StepUnion:   StepTextUnion,
	StepLeague:  StepTextLeague,
	StepLeader:  StepTextLeader,
}

// 审批链定义：按活动等级返回步骤编号序列。
func approvalChain(level string) ([]int, error) {
	switch level {
	case "A":
		return []int{StepTutor, StepCollege, StepUnion, StepLeague, StepLeader}, nil
	case "B":
		return []int{StepTutor, StepCollege, StepUnion, StepLeague}, nil
	case "C":
		return []int{StepTutor, StepCollege}, nil
	case "D":
		return []int{StepTutor}, nil
	default:
		return nil, fmt.Errorf("无效的活动等级: %s", level)
	}
}

// NewActivitySM 根据活动等级构建分级审批状态机。
// 不在此处绑定 Guard/Effect（业务规则与持久化由 Service 层负责）。
func NewActivitySM(level string) (*statem.Engine, error) {
	chain, err := approvalChain(level)
	if err != nil {
		return nil, err
	}

	sm := statem.New("st.activity")

	// 学生侧
	sm.Allow(StateDraft, ActionSubmit, StatePending)
	sm.Allow(StatePending, ActionWithdraw, StateDraft)

	// 按审批链动态注册转移
	for i, stepNo := range chain {
		approveAction := fmt.Sprintf("approve_step%d", stepNo)
		rejectAction := fmt.Sprintf("reject_step%d", stepNo)

		if i == len(chain)-1 {
			// 最后一步通过 → S3
			sm.Allow(StatePending, approveAction, StatePassed)
			sm.Allow(StateInReview, approveAction, StatePassed)
		} else {
			// 中间步骤通过 → S2
			sm.Allow(StatePending, approveAction, StateInReview)
			sm.Allow(StateInReview, approveAction, StateInReview)
		}

		// 任意步骤驳回 → S4
		sm.Allow(StatePending, rejectAction, StateRejected)
		sm.Allow(StateInReview, rejectAction, StateRejected)
	}

	return sm, nil
}

// StepNoOfAction 由动作还原审批步骤编号。
// 例如 "approve_step2" → 2, "reject_step3" → 3。
func StepNoOfAction(action string) int {
	var stepNo int
	if _, err := fmt.Sscanf(action, "approve_step%d", &stepNo); err == nil {
		return stepNo
	}
	if _, err := fmt.Sscanf(action, "reject_step%d", &stepNo); err == nil {
		return stepNo
	}
	return 0
}

// ResolveAction 根据 (stepNo, result) 还原动作。
// 例如 (2, "approve") → "approve_step2"。
func ResolveAction(stepNo int, result string) string {
	if result == "approve" {
		return fmt.Sprintf("approve_step%d", stepNo)
	}
	return fmt.Sprintf("reject_step%d", stepNo)
}

// StepTextOfNo 根据步骤编号返回步骤标签。
func StepTextOfNo(stepNo int) string {
	if text, ok := stepNoToTextMap[stepNo]; ok {
		return text
	}
	return ""
}

// StepCNTextOfNo 根据步骤编号返回中文标签。
func StepCNTextOfNo(stepNo int) string {
	if text, ok := stepTextCNMap[stepNo]; ok {
		return text
	}
	return ""
}

// StepRolesOfNo 根据步骤编号返回允许审批的角色列表。
func StepRolesOfNo(stepNo int) []string {
	if roles, ok := stepRoleMap[stepNo]; ok {
		return roles
	}
	return nil
}

// MaxStepNo 根据活动等级返回最大步骤编号。
func MaxStepNo(level string) int {
	chain, err := approvalChain(level)
	if err != nil {
		return 0
	}
	return chain[len(chain)-1]
}

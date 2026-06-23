// Package models 集中定义所有 GORM 实体（基础层 + TY/ST/SQ/QG/CMP）。
//
// 设计依据：docs/03_database_design_spec.md。
// 命名规范：表名 `{module}_{entity}`；通过 TableName() 显式声明，避免 GORM 复数化推断。
// 字段规范：与设计文档保持完全一致（字段名、类型、长度、CHECK、索引）。
package models

import "time"

// BaseModel 通用字段（每张业务表必须）。
// 见 docs/03_database_design_spec.md §2.1。
type BaseModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy *int64    `gorm:"column:created_by" json:"created_by,omitempty"`
	UpdatedBy *int64    `gorm:"column:updated_by" json:"updated_by,omitempty"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
}

// AllModels 返回需要交给 GORM AutoMigrate 的全部实体指针。
// 顺序：先基础层（被外键引用的父表），再业务模块。
func AllModels() []interface{} {
	return []interface{}{
		// 基础层 - 组织
		&SysCollege{},
		&SysMajor{},
		&IdxClass{},
		&IdxStudent{},
		&IdxDormBuilding{},
		&IdxDormFloor{},
		&IdxDormRoom{},
		&IdxDormBed{},

		// 基础层 - 用户与权限
		&SysUser{},
		&SysRole{},
		&SysUserRole{},

		// 基础层 - 字典 / 菜单 / 流水 / 文件 / 事件 / 审计 / 通知 / 任务
		&SysDict{},
		&SysMenu{},
		&BizSeq{},
		&FileMeta{},
		&EventLog{},
		&AuditLog{},
		&Notification{},
		&JobRun{},

		// TY 团员发展
		&TyBranch{},
		&TyApplication{},
		&TyRecommendationMeeting{},
		&TyRecommendationVote{},
		&TyCultivationLink{},
		&TyCultivationRecord{},
		&TyCourseRecord{},
		&TyThoughtReport{},
		&TyDevelopmentObject{},
		&TyPoliticalReview{},
		&TyDevelopmentMeeting{},
		&TyProbationaryRecord{},
		&TyProbationaryMeeting{},
		&TyMemberRoster{},
		&TyApprovalRecord{},

		// ST 社团活动
		&StAssociation{},
		&StCharter{},
		&StFounder{},
		&StAssocMember{},
		&StRecruitPlan{},
		&StRecruitApply{},
		&StActivity{},
		&StActivityApproval{},
		&StActivityCheckin{},
		&StActivitySummary{},
		&StActivityPhoto{},
		&StExpense{},
		&StElection{},
		&StRating{},
		&StBlacklist{},

		// SQ 学生社区
		&SqSelfgovPosition{},
		&SqInspection{},
		&SqInspectionDeduction{},
		&SqIncident{},
		&SqIncidentAttach{},
		&SqIncidentAction{},
		&SqActivity{},
		&SqAssessment{},
		&SqLateReturn{},
		&SqViolation{},
		&SqVacationStay{},
		&SqRoomChange{},

		// QG 勤工助学
		&QgDifficultyCert{},
		&QgPosition{},
		&QgPositionApply{},
		&QgAttendance{},
		&QgMakeupAttend{},
		&QgLeave{},
		&QgMonthlyAssess{},
		&QgPayroll{},
		&QgPayrollDetail{},
		&QgRenewalTerm{},
		&QgComplaint{},

		// CMP 综合素质量化
		&CmpRuleVersion{},
		&CmpScore{},
		&CmpScoreDetail{},
	}
}

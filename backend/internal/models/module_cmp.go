package models

import "time"

// CmpRuleVersion 量化规则版本。docs/03 §9.2。
type CmpRuleVersion struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Version     string     `gorm:"column:version;type:text;not null;uniqueIndex:uniq_cmp_rule_version" json:"version"`
	RulesJSON   string     `gorm:"column:rules_json;type:text;not null" json:"rules_json"`
	EffectiveAt time.Time  `gorm:"column:effective_at;type:date;not null" json:"effective_at"`
	ExpiredAt   *time.Time `gorm:"column:expired_at;type:date" json:"expired_at,omitempty"`
	IsActive    int        `gorm:"column:is_active;not null;default:0" json:"is_active"`
	CreatedBy   *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
	IsDeleted   int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CmpRuleVersion) TableName() string { return "cmp_rule_version" }

// CmpScore 学生总分快照。docs/03 §9.2。
type CmpScore struct {
	ID             int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StudentID      int64     `gorm:"column:student_id;not null;uniqueIndex:uniq_cmp_score_stu_year,priority:1" json:"student_id"`
	AcademicYear   string    `gorm:"column:academic_year;type:text;not null;uniqueIndex:uniq_cmp_score_stu_year,priority:2;index:idx_cmp_score_year_score,priority:1" json:"academic_year"`
	TotalScore     float64   `gorm:"column:total_score;not null;check:total_score BETWEEN 0 AND 100;index:idx_cmp_score_year_score,priority:2,sort:desc" json:"total_score"`
	RankInClass    *int      `gorm:"column:rank_in_class" json:"rank_in_class,omitempty"`
	RankInCollege  *int      `gorm:"column:rank_in_college" json:"rank_in_college,omitempty"`
	RuleVersionID  int64     `gorm:"column:rule_version_id;not null" json:"rule_version_id"`
	ComputedAt     time.Time `gorm:"column:computed_at;not null;default:CURRENT_TIMESTAMP" json:"computed_at"`
	IsDeleted      int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CmpScore) TableName() string { return "cmp_score" }

// CmpScoreDetail 分维度明细。docs/03 §9.2。
type CmpScoreDetail struct {
	ID             int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ScoreID        int64     `gorm:"column:score_id;not null;index:idx_cmp_detail_score,priority:1" json:"score_id"`
	Dimension      string    `gorm:"column:dimension;type:text;not null;check:dimension IN ('league','assoc','community','workstudy','academic');index:idx_cmp_detail_score,priority:2" json:"dimension"`
	SubItem        string    `gorm:"column:sub_item;type:text;not null" json:"sub_item"`
	SourceEventID  *int64    `gorm:"column:source_event_id" json:"source_event_id,omitempty"`
	SourceModule   string    `gorm:"column:source_module;type:text" json:"source_module"`
	RawValue       string    `gorm:"column:raw_value;type:text" json:"raw_value"`
	Score          float64   `gorm:"column:score;not null;check:score >= 0" json:"score"`
	Weight         float64   `gorm:"column:weight;not null;check:weight >= 0 AND weight <= 1" json:"weight"`
	IsDeleted      int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (CmpScoreDetail) TableName() string { return "cmp_score_detail" }

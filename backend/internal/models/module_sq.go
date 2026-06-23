package models

import "time"

// SqSelfgovPosition 自治职务。docs/03 §7.2.1。
type SqSelfgovPosition struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo        string     `gorm:"column:biz_no;type:text" json:"biz_no"`
	StudentID    int64      `gorm:"column:student_id;not null;index:idx_sq_position_student,priority:1" json:"student_id"`
	ScopeType    string     `gorm:"column:scope_type;type:text;not null;check:scope_type IN ('building','floor','room');index:idx_sq_position_scope,priority:1" json:"scope_type"`
	ScopeID      int64      `gorm:"column:scope_id;not null;index:idx_sq_position_scope,priority:2" json:"scope_id"`
	Position     string     `gorm:"column:position;type:text;not null;check:position IN ('building_chief','floor_leader','room_leader','council_member')" json:"position"`
	StartAt      time.Time  `gorm:"column:start_at;type:date;not null" json:"start_at"`
	EndAt        *time.Time `gorm:"column:end_at;type:date" json:"end_at,omitempty"`
	Status       string     `gorm:"column:status;type:text;not null;default:candidate;check:status IN ('candidate','probation','formal','renewed','dismissed','resigned');index:idx_sq_position_student,priority:2;index:idx_sq_position_scope,priority:3" json:"status"`
	PublicStart  *time.Time `gorm:"column:public_start;type:date" json:"public_start,omitempty"`
	PublicEnd    *time.Time `gorm:"column:public_end;type:date" json:"public_end,omitempty"`
	AppointedBy  *int64     `gorm:"column:appointed_by" json:"appointed_by,omitempty"`
	IsDeleted    int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqSelfgovPosition) TableName() string { return "sq_selfgov_position" }

// SqInspection 巡查记录。docs/03 §7.2.2。
type SqInspection struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo            string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	InspectionType   string    `gorm:"column:inspection_type;type:text;not null;check:inspection_type IN ('hygiene','late_return','appliance','safety','fire_lane')" json:"inspection_type"`
	BuildingID       int64     `gorm:"column:building_id;not null;index:idx_sq_inspection_building,priority:1" json:"building_id"`
	FloorID          *int64    `gorm:"column:floor_id" json:"floor_id,omitempty"`
	RoomID           *int64    `gorm:"column:room_id" json:"room_id,omitempty"`
	InspectorUserID  int64     `gorm:"column:inspector_user_id;not null" json:"inspector_user_id"`
	InspectedAt      time.Time `gorm:"column:inspected_at;not null;index:idx_sq_inspection_building,priority:2" json:"inspected_at"`
	Score            *int      `gorm:"column:score;check:score IS NULL OR (score BETWEEN 0 AND 100)" json:"score,omitempty"`
	Summary          string    `gorm:"column:summary;type:text" json:"summary"`
	Status           string    `gorm:"column:status;type:text;not null;default:submitted;check:status IN ('draft','submitted')" json:"status"`
	IsDeleted        int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqInspection) TableName() string { return "sq_inspection" }

// SqInspectionDeduction 巡查扣分项。docs/03 §7.2.3。
type SqInspectionDeduction struct {
	ID           int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	InspectionID int64  `gorm:"column:inspection_id;not null;index" json:"inspection_id"`
	Item         string `gorm:"column:item;type:text;not null" json:"item"`
	Deduction    int    `gorm:"column:deduction;not null;check:deduction > 0" json:"deduction"`
	PhotoFileID  *int64 `gorm:"column:photo_file_id" json:"photo_file_id,omitempty"`
}

func (SqInspectionDeduction) TableName() string { return "sq_inspection_deduction" }

// SqIncident 异常事件。docs/03 §7.2.4。
type SqIncident struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo               string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_sq_incident_biz_no" json:"biz_no"`
	IncidentLevel       string     `gorm:"column:incident_level;type:text;not null;check:incident_level IN ('L1','L2','L3','L4');index:idx_sq_incident_level_status,priority:1" json:"incident_level"`
	IncidentType        string     `gorm:"column:incident_type;type:text;not null" json:"incident_type"`
	OccurredAt          time.Time  `gorm:"column:occurred_at;not null;index:idx_sq_incident_level_status,priority:3;index:idx_sq_incident_building,priority:2" json:"occurred_at"`
	BuildingID          int64      `gorm:"column:building_id;not null;index:idx_sq_incident_building,priority:1" json:"building_id"`
	FloorID             *int64     `gorm:"column:floor_id" json:"floor_id,omitempty"`
	RoomID              *int64     `gorm:"column:room_id" json:"room_id,omitempty"`
	LocationDetail      string     `gorm:"column:location_detail;type:text" json:"location_detail"`
	ReporterUserID      int64      `gorm:"column:reporter_user_id;not null" json:"reporter_user_id"`
	InvolvedStudentIDs  string     `gorm:"column:involved_student_ids;type:text" json:"involved_student_ids"`
	WitnessUserIDs      string     `gorm:"column:witness_user_ids;type:text" json:"witness_user_ids"`
	InitialAction       string     `gorm:"column:initial_action;type:text" json:"initial_action"`
	Status              string     `gorm:"column:status;type:text;not null;default:open;check:status IN ('open','processing','closed','cancelled');index:idx_sq_incident_level_status,priority:2" json:"status"`
	ClosedAt            *time.Time `gorm:"column:closed_at" json:"closed_at,omitempty"`
	ClosedBy            *int64     `gorm:"column:closed_by" json:"closed_by,omitempty"`
	IsDeleted           int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqIncident) TableName() string { return "sq_incident" }

// SqIncidentAttach 事件附件。docs/03 §7.2.5。
type SqIncidentAttach struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	IncidentID int64     `gorm:"column:incident_id;not null;index" json:"incident_id"`
	FileID     int64     `gorm:"column:file_id;not null" json:"file_id"`
	Caption    string    `gorm:"column:caption;type:text" json:"caption"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (SqIncidentAttach) TableName() string { return "sq_incident_attach" }

// SqIncidentAction 事件处置。docs/03 §7.2.6。
type SqIncidentAction struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	IncidentID int64     `gorm:"column:incident_id;not null;index" json:"incident_id"`
	ActionText string    `gorm:"column:action_text;type:text;not null" json:"action_text"`
	ActionAt   time.Time `gorm:"column:action_at;not null" json:"action_at"`
	ActionBy   int64     `gorm:"column:action_by;not null" json:"action_by"`
	IsFinal    int       `gorm:"column:is_final;not null;default:0" json:"is_final"`
}

func (SqIncidentAction) TableName() string { return "sq_incident_action" }

// SqActivity 自治活动。docs/03 §7.2.7。
type SqActivity struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo                string    `gorm:"column:biz_no;type:text;uniqueIndex:uniq_sq_activity_biz_no" json:"biz_no"`
	BuildingID           int64     `gorm:"column:building_id;not null;index" json:"building_id"`
	Title                string    `gorm:"column:title;type:text;not null" json:"title"`
	ActivityType         string    `gorm:"column:activity_type;type:text;not null" json:"activity_type"`
	ExpectedParticipants int       `gorm:"column:expected_participants;not null" json:"expected_participants"`
	BudgetCents          int64     `gorm:"column:budget_cents;not null;default:0" json:"budget_cents"`
	StartedAt            time.Time `gorm:"column:started_at;not null" json:"started_at"`
	EndedAt              time.Time `gorm:"column:ended_at;not null" json:"ended_at"`
	Summary              string    `gorm:"column:summary;type:text" json:"summary"`
	Status               string    `gorm:"column:status;type:text;not null;default:S0;check:status IN ('S0','S1','S2','S3','S4')" json:"status"`
	CoSignedBy           *int64    `gorm:"column:co_signed_by" json:"co_signed_by,omitempty"`
	IsDeleted            int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt            time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqActivity) TableName() string { return "sq_activity" }

// SqAssessment 考核。docs/03 §7.2.8。
type SqAssessment struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo              string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	CycleType          string    `gorm:"column:cycle_type;type:text;not null;check:cycle_type IN ('monthly','semester');uniqueIndex:uniq_sq_assess_cycle_target,priority:1" json:"cycle_type"`
	CycleKey           string    `gorm:"column:cycle_key;type:text;not null;uniqueIndex:uniq_sq_assess_cycle_target,priority:2" json:"cycle_key"`
	TargetUserID       int64     `gorm:"column:target_user_id;not null;uniqueIndex:uniq_sq_assess_cycle_target,priority:3" json:"target_user_id"`
	TargetPositionID   *int64    `gorm:"column:target_position_id" json:"target_position_id,omitempty"`
	ScoreInspection    int       `gorm:"column:score_inspection;not null;check:score_inspection BETWEEN 0 AND 100" json:"score_inspection"`
	ScoreIncident      int       `gorm:"column:score_incident;not null;check:score_incident BETWEEN 0 AND 100" json:"score_incident"`
	ScoreActivity      int       `gorm:"column:score_activity;not null;check:score_activity BETWEEN 0 AND 100" json:"score_activity"`
	ScoreSatisfaction  int       `gorm:"column:score_satisfaction;not null;check:score_satisfaction BETWEEN 0 AND 100" json:"score_satisfaction"`
	ScoreBonus         int       `gorm:"column:score_bonus;not null;default:0" json:"score_bonus"`
	WeightedScore      float64   `gorm:"column:weighted_score;not null" json:"weighted_score"`
	Rating             string    `gorm:"column:rating;type:text;not null;check:rating IN ('excellent','good','qualified','unqualified')" json:"rating"`
	RectificationNote  string    `gorm:"column:rectification_note;type:text" json:"rectification_note"`
	IsDeleted          int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt          time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqAssessment) TableName() string { return "sq_assessment" }

// SqLateReturn 晚归。docs/03 §7.2.9。
type SqLateReturn struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StudentID  int64     `gorm:"column:student_id;not null;index:idx_sq_late_student_semester,priority:1" json:"student_id"`
	OccurredAt time.Time `gorm:"column:occurred_at;not null" json:"occurred_at"`
	ReportedBy int64     `gorm:"column:reported_by;not null" json:"reported_by"`
	Reason     string    `gorm:"column:reason;type:text" json:"reason"`
	Semester   string    `gorm:"column:semester;type:text;not null;index:idx_sq_late_student_semester,priority:2" json:"semester"`
	IsDeleted  int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqLateReturn) TableName() string { return "sq_late_return" }

// SqViolation 违规电器。docs/03 §7.2.10。
type SqViolation struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StudentID       int64     `gorm:"column:student_id;not null;index" json:"student_id"`
	RoomID          int64     `gorm:"column:room_id;not null" json:"room_id"`
	ApplianceName   string    `gorm:"column:appliance_name;type:text;not null" json:"appliance_name"`
	SeizedAt        time.Time `gorm:"column:seized_at;not null" json:"seized_at"`
	PhotoFileID     *int64    `gorm:"column:photo_file_id" json:"photo_file_id,omitempty"`
	SignatureFileID *int64    `gorm:"column:signature_file_id" json:"signature_file_id,omitempty"`
	ReportedBy      int64     `gorm:"column:reported_by;not null" json:"reported_by"`
	Status          string    `gorm:"column:status;type:text;not null;default:warned;check:status IN ('warned','reported_to_college','cancelled')" json:"status"`
	IsDeleted       int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqViolation) TableName() string { return "sq_violation" }

// SqVacationStay 寒暑假留校。docs/03 §7.2.11。
type SqVacationStay struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo       string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	StudentID   int64     `gorm:"column:student_id;not null;index" json:"student_id"`
	Semester    string    `gorm:"column:semester;type:text;not null" json:"semester"`
	StartAt     time.Time `gorm:"column:start_at;type:date;not null" json:"start_at"`
	EndAt       time.Time `gorm:"column:end_at;type:date;not null" json:"end_at"`
	Reason      string    `gorm:"column:reason;type:text;not null" json:"reason"`
	Status      string    `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S3','S4')" json:"status"`
	SubmittedAt time.Time `gorm:"column:submitted_at;not null;default:CURRENT_TIMESTAMP" json:"submitted_at"`
	IsDeleted   int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqVacationStay) TableName() string { return "sq_vacation_stay" }

// SqRoomChange 寝室调整。docs/03 §7.2.12。
type SqRoomChange struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo               string     `gorm:"column:biz_no;type:text" json:"biz_no"`
	StudentID           int64      `gorm:"column:student_id;not null;index" json:"student_id"`
	FromRoomID          int64      `gorm:"column:from_room_id;not null" json:"from_room_id"`
	ToRoomID            int64      `gorm:"column:to_room_id;not null" json:"to_room_id"`
	ToBedID             *int64     `gorm:"column:to_bed_id" json:"to_bed_id,omitempty"`
	Reason              string     `gorm:"column:reason;type:text;not null" json:"reason"`
	CounselorSignedBy   *int64     `gorm:"column:counselor_signed_by" json:"counselor_signed_by,omitempty"`
	CouncilSignedBy     *int64     `gorm:"column:council_signed_by" json:"council_signed_by,omitempty"`
	MovedAt             *time.Time `gorm:"column:moved_at;type:date" json:"moved_at,omitempty"`
	Status              string     `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S3','S4')" json:"status"`
	IsDeleted           int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SqRoomChange) TableName() string { return "sq_room_change" }

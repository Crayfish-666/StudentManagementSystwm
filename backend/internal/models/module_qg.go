package models

import "time"

// QgDifficultyCert 困难认定。docs/03 §8.2.1。
type QgDifficultyCert struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo        string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_qg_diff_biz_no" json:"biz_no"`
	StudentID    int64      `gorm:"column:student_id;not null" json:"student_id"`
	AcademicYear string     `gorm:"column:academic_year;type:text;not null;index:idx_qg_diff_level_year,priority:2" json:"academic_year"`
	Level        string     `gorm:"column:level;type:text;not null;check:level IN ('special','hard','normal','none');index:idx_qg_diff_level_year,priority:1" json:"level"`
	CertFiles    string     `gorm:"column:cert_files;type:text" json:"cert_files"`
	PublicStart  *time.Time `gorm:"column:public_start;type:date" json:"public_start,omitempty"`
	PublicEnd    *time.Time `gorm:"column:public_end;type:date" json:"public_end,omitempty"`
	Status       string     `gorm:"column:status;type:text;not null;default:S0;check:status IN ('S0','S1','S2','S3','S4')" json:"status"`
	RejectReason string     `gorm:"column:reject_reason;type:text" json:"reject_reason"`
	IsDeleted    int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy    *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
	UpdatedBy    *int64     `gorm:"column:updated_by" json:"updated_by,omitempty"`
}

func (QgDifficultyCert) TableName() string { return "qg_difficulty_cert" }

// QgPosition 岗位。docs/03 §8.2.2。
type QgPosition struct {
	ID                int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo             string    `gorm:"column:biz_no;type:text;uniqueIndex:uniq_qg_position_biz_no" json:"biz_no"`
	DeptType          string    `gorm:"column:dept_type;type:text;not null;check:dept_type IN ('admin','teaching','research','culture');index:idx_qg_position_dept,priority:1" json:"dept_type"`
	DeptName          string    `gorm:"column:dept_name;type:text;not null" json:"dept_name"`
	Title             string    `gorm:"column:title;type:text;not null" json:"title"`
	Description       string    `gorm:"column:description;type:text;not null" json:"description"`
	Headcount         int       `gorm:"column:headcount;not null;check:headcount > 0" json:"headcount"`
	WeeklyHoursLimit  int       `gorm:"column:weekly_hours_limit;not null;check:weekly_hours_limit > 0 AND weekly_hours_limit <= 20" json:"weekly_hours_limit"`
	HourlyRateCents   int64     `gorm:"column:hourly_rate_cents;not null;check:hourly_rate_cents > 0" json:"hourly_rate_cents"`
	StartAt           time.Time `gorm:"column:start_at;type:date;not null;index:idx_qg_position_status,priority:2" json:"start_at"`
	EndAt             time.Time `gorm:"column:end_at;type:date;not null" json:"end_at"`
	RiskNotes         string    `gorm:"column:risk_notes;type:text" json:"risk_notes"`
	KpiJSON           string    `gorm:"column:kpi_json;type:text" json:"kpi_json"`
	Status            string    `gorm:"column:status;type:text;not null;default:S0;check:status IN ('S0','S1','S2','S3','S4','closed');index:idx_qg_position_status,priority:1;index:idx_qg_position_dept,priority:2" json:"status"`
	SupervisorUserID  *int64    `gorm:"column:supervisor_user_id" json:"supervisor_user_id,omitempty"`
	IsDeleted         int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy         *int64    `gorm:"column:created_by" json:"created_by,omitempty"`
	UpdatedBy         *int64    `gorm:"column:updated_by" json:"updated_by,omitempty"`
}

func (QgPosition) TableName() string { return "qg_position" }

// QgPositionApply 岗位申请。docs/03 §8.2.3。
type QgPositionApply struct {
	ID               int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo            string     `gorm:"column:biz_no;type:text" json:"biz_no"`
	PositionID       int64      `gorm:"column:position_id;not null;uniqueIndex:uniq_qg_apply_pos_stu,priority:1" json:"position_id"`
	StudentID        int64      `gorm:"column:student_id;not null;uniqueIndex:uniq_qg_apply_pos_stu,priority:2;index:idx_qg_apply_student_status,priority:1" json:"student_id"`
	ResumeFileID     *int64     `gorm:"column:resume_file_id" json:"resume_file_id,omitempty"`
	ApplyStatus      string     `gorm:"column:apply_status;type:text;not null;default:pending;check:apply_status IN ('pending','interview','accepted','rejected','abandoned','expired')" json:"apply_status"`
	InterviewAt      *time.Time `gorm:"column:interview_at" json:"interview_at,omitempty"`
	InterviewNote    string     `gorm:"column:interview_note;type:text" json:"interview_note"`
	ConfirmDeadline  *time.Time `gorm:"column:confirm_deadline" json:"confirm_deadline,omitempty"`
	ConfirmedAt      *time.Time `gorm:"column:confirmed_at" json:"confirmed_at,omitempty"`
	OnBoardAt        *time.Time `gorm:"column:on_board_at;type:date" json:"on_board_at,omitempty"`
	OffBoardAt       *time.Time `gorm:"column:off_board_at;type:date" json:"off_board_at,omitempty"`
	Status           string     `gorm:"column:status;type:text;not null;default:onboarding;check:status IN ('onboarding','on_job','renewal','terminated','closed');index:idx_qg_apply_student_status,priority:2" json:"status"`
	IsDeleted        int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (QgPositionApply) TableName() string { return "qg_position_apply" }

// QgAttendance 工时打卡。docs/03 §8.2.4。
type QgAttendance struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo           string     `gorm:"column:biz_no;type:text" json:"biz_no"`
	ApplyID         int64      `gorm:"column:apply_id;not null;uniqueIndex:uniq_qg_attend_day,priority:1" json:"apply_id"`
	StudentID       int64      `gorm:"column:student_id;not null;index:idx_qg_attend_student_month,priority:1" json:"student_id"`
	WorkDate        time.Time  `gorm:"column:work_date;type:date;not null;uniqueIndex:uniq_qg_attend_day,priority:2;index:idx_qg_attend_student_month,priority:2" json:"work_date"`
	ClockInAt       *time.Time `gorm:"column:clock_in_at" json:"clock_in_at,omitempty"`
	ClockOutAt      *time.Time `gorm:"column:clock_out_at" json:"clock_out_at,omitempty"`
	EffectiveHours  float64    `gorm:"column:effective_hours;not null;default:0;check:effective_hours >= 0" json:"effective_hours"`
	LateMinutes     int        `gorm:"column:late_minutes;not null;default:0" json:"late_minutes"`
	EarlyMinutes    int        `gorm:"column:early_minutes;not null;default:0" json:"early_minutes"`
	ClockMethod     string     `gorm:"column:clock_method;type:text;not null;check:clock_method IN ('card','gps_face','manual')" json:"clock_method"`
	IP              string     `gorm:"column:ip;type:text" json:"ip"`
	Geo             string     `gorm:"column:geo;type:text" json:"geo"`
	IsMakeup        int        `gorm:"column:is_makeup;not null;default:0" json:"is_makeup"`
	MakeupID        *int64     `gorm:"column:makeup_id" json:"makeup_id,omitempty"`
	IsDeleted       int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (QgAttendance) TableName() string { return "qg_attendance" }

// QgMakeupAttend 补卡申请。docs/03 §8.2.5。
type QgMakeupAttend struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo               string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	ApplyID             int64     `gorm:"column:apply_id;not null;index" json:"apply_id"`
	StudentID           int64     `gorm:"column:student_id;not null;index" json:"student_id"`
	WorkDate            time.Time `gorm:"column:work_date;type:date;not null" json:"work_date"`
	Reason              string    `gorm:"column:reason;type:text;not null;check:length(reason) >= 20" json:"reason"`
	CounselorSignedBy   *int64    `gorm:"column:counselor_signed_by" json:"counselor_signed_by,omitempty"`
	SupervisorSignedBy  *int64    `gorm:"column:supervisor_signed_by" json:"supervisor_signed_by,omitempty"`
	Status              string    `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S3','S4')" json:"status"`
	IsDeleted           int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt           time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (QgMakeupAttend) TableName() string { return "qg_makeup_attend" }

// QgLeave 请假。docs/03 §8.2.6。
type QgLeave struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ApplyID   int64     `gorm:"column:apply_id;not null;index" json:"apply_id"`
	StudentID int64     `gorm:"column:student_id;not null;index" json:"student_id"`
	StartAt   time.Time `gorm:"column:start_at;not null" json:"start_at"`
	EndAt     time.Time `gorm:"column:end_at;not null" json:"end_at"`
	Reason    string    `gorm:"column:reason;type:text;not null" json:"reason"`
	Status    string    `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S3','S4')" json:"status"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (QgLeave) TableName() string { return "qg_leave" }

// QgMonthlyAssess 月度考核。docs/03 §8.2.7。
type QgMonthlyAssess struct {
	ID                int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo             string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	ApplyID           int64     `gorm:"column:apply_id;not null;uniqueIndex:uniq_qg_assess_apply_month,priority:1" json:"apply_id"`
	StudentID         int64     `gorm:"column:student_id;not null;index" json:"student_id"`
	AssessYear        int       `gorm:"column:assess_year;not null;uniqueIndex:uniq_qg_assess_apply_month,priority:2" json:"assess_year"`
	AssessMonth       int       `gorm:"column:assess_month;not null;check:assess_month BETWEEN 1 AND 12;uniqueIndex:uniq_qg_assess_apply_month,priority:3" json:"assess_month"`
	ScoreAttendance   int       `gorm:"column:score_attendance;not null;check:score_attendance BETWEEN 0 AND 100" json:"score_attendance"`
	ScoreWorkComplete int       `gorm:"column:score_work_complete;not null;check:score_work_complete BETWEEN 0 AND 100" json:"score_work_complete"`
	ScoreComprehensive int      `gorm:"column:score_comprehensive;not null;check:score_comprehensive BETWEEN 0 AND 100" json:"score_comprehensive"`
	WeightedScore     float64   `gorm:"column:weighted_score;not null" json:"weighted_score"`
	Coefficient       float64   `gorm:"column:coefficient;not null;check:coefficient IN (1.0, 0.8, 0.5, 0.0)" json:"coefficient"`
	IsObservation     int       `gorm:"column:is_observation;not null;default:0" json:"is_observation"`
	Note              string    `gorm:"column:note;type:text" json:"note"`
	Status            string    `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S3')" json:"status"`
	IsDeleted         int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (QgMonthlyAssess) TableName() string { return "qg_monthly_assess" }

// QgPayroll 薪酬发放。docs/03 §8.2.8。
type QgPayroll struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo               string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_qg_payroll_biz_no" json:"biz_no"`
	StudentID           int64      `gorm:"column:student_id;not null;uniqueIndex:uniq_qg_payroll_stu_apply_month,priority:1" json:"student_id"`
	ApplyID             int64      `gorm:"column:apply_id;not null;uniqueIndex:uniq_qg_payroll_stu_apply_month,priority:2" json:"apply_id"`
	PayYear             int        `gorm:"column:pay_year;not null;uniqueIndex:uniq_qg_payroll_stu_apply_month,priority:3;index:idx_qg_payroll_status_month,priority:1" json:"pay_year"`
	PayMonth            int        `gorm:"column:pay_month;not null;check:pay_month BETWEEN 1 AND 12;uniqueIndex:uniq_qg_payroll_stu_apply_month,priority:4;index:idx_qg_payroll_status_month,priority:2" json:"pay_month"`
	TotalHours          float64    `gorm:"column:total_hours;not null;check:total_hours >= 0" json:"total_hours"`
	GrossCents          int64      `gorm:"column:gross_cents;not null;check:gross_cents >= 0" json:"gross_cents"`
	TaxCents            int64      `gorm:"column:tax_cents;not null;default:0" json:"tax_cents"`
	DeductionCents      int64      `gorm:"column:deduction_cents;not null;default:0" json:"deduction_cents"`
	NetCents            int64      `gorm:"column:net_cents;not null;check:net_cents >= 0" json:"net_cents"`
	Coefficient         float64    `gorm:"column:coefficient;not null" json:"coefficient"`
	BankAccountLast4Enc string     `gorm:"column:bank_account_last4_enc;type:text" json:"-"`
	Status              string     `gorm:"column:status;type:text;not null;default:draft;check:status IN ('draft','reviewed','paid','failed');index:idx_qg_payroll_status_month,priority:3" json:"status"`
	ReviewedBy          *int64     `gorm:"column:reviewed_by" json:"reviewed_by,omitempty"`
	PaidAt              *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	FailureReason       string     `gorm:"column:failure_reason;type:text" json:"failure_reason"`
	IsDeleted           int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (QgPayroll) TableName() string { return "qg_payroll" }

// QgPayrollDetail 薪酬明细。docs/03 §8.2.9。
type QgPayrollDetail struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PayrollID     int64     `gorm:"column:payroll_id;not null;uniqueIndex:uniq_qg_payroll_detail,priority:1" json:"payroll_id"`
	AttendanceID  int64     `gorm:"column:attendance_id;not null;uniqueIndex:uniq_qg_payroll_detail,priority:2" json:"attendance_id"`
	WorkDate      time.Time `gorm:"column:work_date;type:date;not null" json:"work_date"`
	Hours         float64   `gorm:"column:hours;not null" json:"hours"`
	RateCents     int64     `gorm:"column:rate_cents;not null" json:"rate_cents"`
	AmountCents   int64     `gorm:"column:amount_cents;not null" json:"amount_cents"`
}

func (QgPayrollDetail) TableName() string { return "qg_payroll_detail" }

// QgRenewalTerm 续聘 / 解聘。docs/03 §8.2.10。
type QgRenewalTerm struct {
	ID                       int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo                    string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	ApplyID                  int64     `gorm:"column:apply_id;not null;index" json:"apply_id"`
	StudentID                int64     `gorm:"column:student_id;not null;index" json:"student_id"`
	Type                     string    `gorm:"column:type;type:text;not null;check:type IN ('renewal','termination')" json:"type"`
	Reason                   string    `gorm:"column:reason;type:text;not null" json:"reason"`
	EffectiveAt              time.Time `gorm:"column:effective_at;type:date;not null" json:"effective_at"`
	SemesterAvgScore         *float64  `gorm:"column:semester_avg_score" json:"semester_avg_score,omitempty"`
	InitiatedBy              int64     `gorm:"column:initiated_by;not null" json:"initiated_by"`
	CounselorSignedBy        *int64    `gorm:"column:counselor_signed_by" json:"counselor_signed_by,omitempty"`
	StudentAffairsSignedBy   *int64    `gorm:"column:student_affairs_signed_by" json:"student_affairs_signed_by,omitempty"`
	Status                   string    `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S2','S3','S4')" json:"status"`
	IsDeleted                int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt                time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (QgRenewalTerm) TableName() string { return "qg_renewal_term" }

// QgComplaint 申诉。docs/03 §8.2.11。
type QgComplaint struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo              string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	StudentID          int64     `gorm:"column:student_id;not null;index" json:"student_id"`
	TargetType         string    `gorm:"column:target_type;type:text;not null;check:target_type IN ('attendance','assess','payroll')" json:"target_type"`
	TargetID           int64     `gorm:"column:target_id;not null" json:"target_id"`
	Reason             string    `gorm:"column:reason;type:text;not null;check:length(reason) >= 30" json:"reason"`
	ExpectedReplyDays  int       `gorm:"column:expected_reply_days;not null;default:10" json:"expected_reply_days"`
	Status             string    `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S2','S3','S4')" json:"status"`
	Result             string    `gorm:"column:result;type:text" json:"result"`
	HandledBy          *int64    `gorm:"column:handled_by" json:"handled_by,omitempty"`
	IsDeleted          int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt          time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (QgComplaint) TableName() string { return "qg_complaint" }

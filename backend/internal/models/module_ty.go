package models

import "time"

// TyBranch 团支部。docs/03 §5.2.1。
type TyBranch struct {
	ID                   int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo                string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_ty_branch_biz_no" json:"biz_no"`
	Name                 string     `gorm:"column:name;type:text;not null" json:"name"`
	CollegeID            int64      `gorm:"column:college_id;not null;index" json:"college_id"`
	SecretaryStudentID   *int64     `gorm:"column:secretary_student_id" json:"secretary_student_id,omitempty"`
	ExpectedMemberCount  int        `gorm:"column:expected_member_count;not null;default:0" json:"expected_member_count"`
	EstablishedAt        *time.Time `gorm:"column:established_at;type:date" json:"established_at,omitempty"`
	IsDeleted            int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy            *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
	UpdatedBy            *int64     `gorm:"column:updated_by" json:"updated_by,omitempty"`
}

func (TyBranch) TableName() string { return "ty_branch" }

// TyApplication 入团申请。docs/03 §5.2.2。
type TyApplication struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo             string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_ty_application_biz_no" json:"biz_no"`
	StudentID         int64      `gorm:"column:student_id;not null;index:idx_ty_app_student_status,priority:1" json:"student_id"`
	BranchID          int64      `gorm:"column:branch_id;not null;index:idx_ty_app_branch_status,priority:1" json:"branch_id"`
	ApplyDate         time.Time  `gorm:"column:apply_date;type:date;not null" json:"apply_date"`
	SelfStatement     string     `gorm:"column:self_statement;type:text;not null;check:length(self_statement) >= 500" json:"self_statement"`
	FamilyMembers     string     `gorm:"column:family_members_json;type:text" json:"family_members_json"`
	RewardsPunish     string     `gorm:"column:rewards_punishments;type:text" json:"rewards_punishments"`
	Status            string     `gorm:"column:status;type:text;not null;default:S0;check:status IN ('S0','S1','S2','S3','S4');index:idx_ty_app_student_status,priority:2;index:idx_ty_app_branch_status,priority:2" json:"status"`
	CounselorOpinion  string     `gorm:"column:counselor_opinion;type:text" json:"counselor_opinion"`
	CounselorUserID   *int64     `gorm:"column:counselor_user_id" json:"counselor_user_id,omitempty"`
	CounselorAt       *time.Time `gorm:"column:counselor_at" json:"counselor_at,omitempty"`
	CollegeOpinion    string     `gorm:"column:college_opinion;type:text" json:"college_opinion"`
	CollegeUserID     *int64     `gorm:"column:college_user_id" json:"college_user_id,omitempty"`
	CollegeAt         *time.Time `gorm:"column:college_at" json:"college_at,omitempty"`
	SchoolOpinion     string     `gorm:"column:school_opinion;type:text" json:"school_opinion"`
	SchoolUserID      *int64     `gorm:"column:school_user_id" json:"school_user_id,omitempty"`
	SchoolAt          *time.Time `gorm:"column:school_at" json:"school_at,omitempty"`
	RejectReason      string     `gorm:"column:reject_reason;type:text" json:"reject_reason"`
	IsDeleted         int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;index:idx_ty_app_created_at" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy         *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
	UpdatedBy         *int64     `gorm:"column:updated_by" json:"updated_by,omitempty"`
}

func (TyApplication) TableName() string { return "ty_application" }

// TyRecommendationMeeting 推优大会。docs/03 §5.2.3。
type TyRecommendationMeeting struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo           string    `gorm:"column:biz_no;type:text;uniqueIndex:uniq_ty_rec_meeting_biz_no" json:"biz_no"`
	ApplicationID   int64     `gorm:"column:application_id;not null;index" json:"application_id"`
	MeetingAt       time.Time `gorm:"column:meeting_at;not null" json:"meeting_at"`
	Location        string    `gorm:"column:location;type:text;not null" json:"location"`
	ExpectedCount   int       `gorm:"column:expected_count;not null;check:expected_count > 0" json:"expected_count"`
	ActualCount     int       `gorm:"column:actual_count;not null;check:actual_count > 0" json:"actual_count"`
	PhotoOverallID  *int64    `gorm:"column:photo_overall_id" json:"photo_overall_id,omitempty"`
	PhotoVoteID     *int64    `gorm:"column:photo_vote_id" json:"photo_vote_id,omitempty"`
	Decision        string    `gorm:"column:decision;type:text;not null;check:decision IN ('pass','reject')" json:"decision"`
	DecisionReason  string    `gorm:"column:decision_reason;type:text" json:"decision_reason"`
	RecorderUserID  *int64    `gorm:"column:recorder_user_id" json:"recorder_user_id,omitempty"`
	IsDeleted       int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyRecommendationMeeting) TableName() string { return "ty_recommendation_meeting" }

// TyRecommendationVote 推优大会投票。docs/03 §5.2.3。
type TyRecommendationVote struct {
	ID            int64 `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MeetingID     int64 `gorm:"column:meeting_id;not null;uniqueIndex:uniq_ty_rec_vote_meeting_app,priority:1" json:"meeting_id"`
	ApplicationID int64 `gorm:"column:application_id;not null;uniqueIndex:uniq_ty_rec_vote_meeting_app,priority:2" json:"application_id"`
	ApproveCount  int   `gorm:"column:approve_count;not null;default:0;check:approve_count >= 0" json:"approve_count"`
	AgainstCount  int   `gorm:"column:against_count;not null;default:0;check:against_count >= 0" json:"against_count"`
	AbstainCount  int   `gorm:"column:abstain_count;not null;default:0;check:abstain_count >= 0" json:"abstain_count"`
}

func (TyRecommendationVote) TableName() string { return "ty_recommendation_vote" }

// TyCultivationLink 培养联系人。docs/03 §5.2.4。
type TyCultivationLink struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ApplicationID   int64      `gorm:"column:application_id;not null;index" json:"application_id"`
	MentorStudentID int64      `gorm:"column:mentor_student_id;not null" json:"mentor_student_id"`
	MentorType      string     `gorm:"column:mentor_type;type:text;not null;check:mentor_type IN ('league_member','party_member')" json:"mentor_type"`
	StartAt         time.Time  `gorm:"column:start_at;type:date;not null" json:"start_at"`
	EndAt           *time.Time `gorm:"column:end_at;type:date" json:"end_at,omitempty"`
	IsActive        int        `gorm:"column:is_active;not null;default:1" json:"is_active"`
	IsDeleted       int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyCultivationLink) TableName() string { return "ty_cultivation_link" }

// TyCultivationRecord 培养考察记录。docs/03 §5.2.5。
type TyCultivationRecord struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo            string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	ApplicationID    int64     `gorm:"column:application_id;not null;index:idx_ty_cultivation_app_month,priority:1" json:"application_id"`
	RecordYear       int       `gorm:"column:record_year;not null;index:idx_ty_cultivation_app_month,priority:2" json:"record_year"`
	RecordMonth      int       `gorm:"column:record_month;not null;check:record_month BETWEEN 1 AND 12;index:idx_ty_cultivation_app_month,priority:3" json:"record_month"`
	Summary          string    `gorm:"column:summary;type:text;not null;check:length(summary) >= 50" json:"summary"`
	PerformanceScore int       `gorm:"column:performance_score;not null;check:performance_score BETWEEN 0 AND 100" json:"performance_score"`
	RecordType       string    `gorm:"column:record_type;type:text;not null;default:monthly;check:record_type IN ('monthly','quarterly')" json:"record_type"`
	IsOverdue        int       `gorm:"column:is_overdue;not null;default:0" json:"is_overdue"`
	RecordedBy       *int64    `gorm:"column:recorded_by" json:"recorded_by,omitempty"`
	IsDeleted        int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyCultivationRecord) TableName() string { return "ty_cultivation_record" }

// TyCourseRecord 团课记录。docs/03 §5.2.6。
type TyCourseRecord struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StudentID     int64     `gorm:"column:student_id;not null;index:idx_ty_course_student,priority:1" json:"student_id"`
	CourseName    string    `gorm:"column:course_name;type:text;not null" json:"course_name"`
	Semester      string    `gorm:"column:semester;type:text;not null;index:idx_ty_course_student,priority:2" json:"semester"`
	StudyAt       time.Time `gorm:"column:study_at;type:date;not null" json:"study_at"`
	Score         *int      `gorm:"column:score;check:score BETWEEN 0 AND 100" json:"score,omitempty"`
	CertificateNo string    `gorm:"column:certificate_no;type:text" json:"certificate_no"`
	IsPass        int       `gorm:"column:is_pass;not null;default:0" json:"is_pass"`
	IsDeleted     int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyCourseRecord) TableName() string { return "ty_course_record" }

// TyThoughtReport 思想汇报。docs/03 §5.2.7。
type TyThoughtReport struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo         string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	ApplicationID int64     `gorm:"column:application_id;not null;index:idx_ty_report_app_quarter,priority:1" json:"application_id"`
	StudentID     int64     `gorm:"column:student_id;not null;index" json:"student_id"`
	Title         string    `gorm:"column:title;type:text;not null" json:"title"`
	Content       string    `gorm:"column:content;type:text;not null;check:length(content) >= 1000" json:"content"`
	Quarter       string    `gorm:"column:quarter;type:text;not null;index:idx_ty_report_app_quarter,priority:2" json:"quarter"`
	AISimilarity  *float64  `gorm:"column:ai_similarity" json:"ai_similarity,omitempty"`
	IsQualified   int       `gorm:"column:is_qualified;not null;default:0" json:"is_qualified"`
	IsDeleted     int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyThoughtReport) TableName() string { return "ty_thought_report" }

// TyDevelopmentObject 发展对象。docs/03 §5.2.8。
type TyDevelopmentObject struct {
	ID                   int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo                string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_ty_dev_obj_biz_no" json:"biz_no"`
	ApplicationID        int64      `gorm:"column:application_id;not null;uniqueIndex:uniq_ty_dev_obj_app" json:"application_id"`
	CourseCertNo         string     `gorm:"column:course_cert_no;type:text" json:"course_cert_no"`
	MentorOpinion        string     `gorm:"column:mentor_opinion;type:text;check:mentor_opinion IS NULL OR length(mentor_opinion) >= 200" json:"mentor_opinion"`
	CounselorOpinion     string     `gorm:"column:counselor_opinion;type:text;check:counselor_opinion IS NULL OR length(counselor_opinion) >= 200" json:"counselor_opinion"`
	MassMeetingAt        *time.Time `gorm:"column:mass_meeting_at" json:"mass_meeting_at,omitempty"`
	MassMeetingAttendees *int       `gorm:"column:mass_meeting_attendees;check:mass_meeting_attendees IS NULL OR mass_meeting_attendees >= 10" json:"mass_meeting_attendees,omitempty"`
	PublicStart          *time.Time `gorm:"column:public_start;type:date" json:"public_start,omitempty"`
	PublicEnd            *time.Time `gorm:"column:public_end;type:date" json:"public_end,omitempty"`
	AutobiographyPath    string     `gorm:"column:autobiography_path;type:text" json:"autobiography_path"`
	Status               string     `gorm:"column:status;type:text;not null;default:S0;check:status IN ('S0','S1','S2','S3','S4')" json:"status"`
	IsDeleted            int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyDevelopmentObject) TableName() string { return "ty_development_object" }

// TyPoliticalReview 政审。docs/03 §5.2.9。
type TyPoliticalReview struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	DevelopmentID   int64     `gorm:"column:development_id;not null;index:idx_ty_pol_review_dev" json:"development_id"`
	TargetRelation  string    `gorm:"column:target_relation;type:text;not null;check:target_relation IN ('self','parent','spouse')" json:"target_relation"`
	TargetName      string    `gorm:"column:target_name;type:text;not null" json:"target_name"`
	TargetIDCardEnc string    `gorm:"column:target_id_card_enc;type:text" json:"-"`
	Method          string    `gorm:"column:method;type:text;not null;check:method IN ('letter','interview')" json:"method"`
	Conclusion      string    `gorm:"column:conclusion;type:text;not null;check:conclusion IN ('pass','basic_pass','fail')" json:"conclusion"`
	DocumentPath    string    `gorm:"column:document_path;type:text" json:"document_path"`
	IsExtend3M      int       `gorm:"column:is_extend_3m;not null;default:0" json:"is_extend_3m"`
	IsDeleted       int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyPoliticalReview) TableName() string { return "ty_political_review" }

// TyDevelopmentMeeting 发展大会。docs/03 §5.2.10。
type TyDevelopmentMeeting struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo              string    `gorm:"column:biz_no;type:text;uniqueIndex:uniq_ty_dev_meeting_biz_no" json:"biz_no"`
	DevelopmentID      int64     `gorm:"column:development_id;not null;index" json:"development_id"`
	MeetingAt          time.Time `gorm:"column:meeting_at;not null" json:"meeting_at"`
	ExpectedCount      int       `gorm:"column:expected_count;not null" json:"expected_count"`
	ActualCount        int       `gorm:"column:actual_count;not null" json:"actual_count"`
	ApproveCount       int       `gorm:"column:approve_count;not null" json:"approve_count"`
	AgainstCount       int       `gorm:"column:against_count;not null" json:"against_count"`
	AbstainCount       int       `gorm:"column:abstain_count;not null" json:"abstain_count"`
	Decision           string    `gorm:"column:decision;type:text;not null;check:decision IN ('pass','reject')" json:"decision"`
	VolunteerFormPath  string    `gorm:"column:volunteer_form_path;type:text" json:"volunteer_form_path"`
	IsDeleted          int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt          time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyDevelopmentMeeting) TableName() string { return "ty_development_meeting" }

// TyProbationaryRecord 预备期考察。docs/03 §5.2.11。
type TyProbationaryRecord struct {
	ID             int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ApplicationID  int64     `gorm:"column:application_id;not null;uniqueIndex:uniq_ty_prob_record_app_year_q,priority:1" json:"application_id"`
	RecordYear     int       `gorm:"column:record_year;not null;uniqueIndex:uniq_ty_prob_record_app_year_q,priority:2" json:"record_year"`
	RecordQuarter  int       `gorm:"column:record_quarter;not null;check:record_quarter BETWEEN 1 AND 4;uniqueIndex:uniq_ty_prob_record_app_year_q,priority:3" json:"record_quarter"`
	Summary        string    `gorm:"column:summary;type:text;not null;check:length(summary) >= 100" json:"summary"`
	IsDeleted      int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyProbationaryRecord) TableName() string { return "ty_probationary_record" }

// TyProbationaryMeeting 转正大会。docs/03 §5.2.12。
type TyProbationaryMeeting struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo               string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_ty_prob_meeting_biz_no" json:"biz_no"`
	ApplicationID       int64      `gorm:"column:application_id;not null;index" json:"application_id"`
	SelfApplicationPath string     `gorm:"column:self_application_path;type:text" json:"self_application_path"`
	MeetingAt           time.Time  `gorm:"column:meeting_at;not null" json:"meeting_at"`
	ExpectedCount       int        `gorm:"column:expected_count;not null" json:"expected_count"`
	ActualCount         int        `gorm:"column:actual_count;not null" json:"actual_count"`
	ApproveCount        int        `gorm:"column:approve_count;not null" json:"approve_count"`
	Decision            string     `gorm:"column:decision;type:text;not null;check:decision IN ('pass','reject')" json:"decision"`
	FormalJoinAt        *time.Time `gorm:"column:formal_join_at;type:date" json:"formal_join_at,omitempty"`
	IsDeleted           int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyProbationaryMeeting) TableName() string { return "ty_probationary_meeting" }

// TyMemberRoster 团员花名册。docs/03 §5.2.13。
type TyMemberRoster struct {
	ID                   int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo                string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_ty_roster_biz_no" json:"biz_no"`
	StudentID            int64      `gorm:"column:student_id;not null;uniqueIndex:uniq_ty_roster_student" json:"student_id"`
	ApplicationID        *int64     `gorm:"column:application_id" json:"application_id,omitempty"`
	BranchID             int64      `gorm:"column:branch_id;not null;index:idx_ty_roster_branch_status,priority:1" json:"branch_id"`
	JoinAt               time.Time  `gorm:"column:join_at;type:date;not null" json:"join_at"`
	BecomeProbationaryAt *time.Time `gorm:"column:become_probationary_at;type:date" json:"become_probationary_at,omitempty"`
	IsOvertime           int        `gorm:"column:is_overtime;not null;default:0" json:"is_overtime"`
	TransferredAt        *time.Time `gorm:"column:transferred_at;type:date" json:"transferred_at,omitempty"`
	ArchiveKeepUntil     *time.Time `gorm:"column:archive_keep_until;type:date" json:"archive_keep_until,omitempty"`
	Status               string     `gorm:"column:status;type:text;not null;default:active;check:status IN ('active','transferred','overtime','archived');index:idx_ty_roster_branch_status,priority:2" json:"status"`
	IsDeleted            int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyMemberRoster) TableName() string { return "ty_member_roster" }

// TyApprovalRecord 团员发展全流程审批记录。
// 支持入团申请、发展对象、转正三个阶段的三级审批流。
// Module 字段区分阶段：application / development_object / probationary。
// TargetID 关联到对应阶段的主记录 ID（与 ApplicationID 互为补充，ApplicationID 仅用于 application 模块）。
type TyApprovalRecord struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ApplicationID int64     `gorm:"column:application_id;not null;index:idx_ty_approval_app_step,priority:1" json:"application_id"`
	Module        string    `gorm:"column:module;type:text;not null;default:'application';index:idx_ty_approval_module_target,priority:1" json:"module"`
	TargetID      int64     `gorm:"column:target_id;not null;default:0;index:idx_ty_approval_module_target,priority:2" json:"target_id"`
	Step          string    `gorm:"column:step;type:text;not null;index:idx_ty_approval_app_step,priority:2" json:"step"`
	ApproverID    int64     `gorm:"column:approver_id;not null;index:idx_ty_approval_approver" json:"approver_id"`
	ApproverName  string    `gorm:"column:approver_name;type:text;not null" json:"approver_name"`
	ApproverRole  string    `gorm:"column:approver_role;type:text;not null" json:"approver_role"`
	Result        string    `gorm:"column:result;type:text;not null;check:result IN ('approve','reject')" json:"result"`
	Opinion       string    `gorm:"column:opinion;type:text;not null;check:length(opinion) >= 5" json:"opinion"`
	FromStatus    string    `gorm:"column:from_status;type:text;not null" json:"from_status"`
	ToStatus      string    `gorm:"column:to_status;type:text;not null" json:"to_status"`
	OccurredAt    time.Time `gorm:"column:occurred_at;not null;default:CURRENT_TIMESTAMP;index:idx_ty_approval_occurred_at" json:"occurred_at"`
	IP            string    `gorm:"column:ip;type:text" json:"ip"`
	IsDeleted     int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TyApprovalRecord) TableName() string { return "ty_approval_record" }

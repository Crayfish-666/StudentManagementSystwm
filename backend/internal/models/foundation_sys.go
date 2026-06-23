package models

import "time"

// SysUser 用户。docs/03 §4.2。
type SysUser struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username       string     `gorm:"column:username;type:text;not null;uniqueIndex:uniq_sys_user_username" json:"username"`
	PasswordHash   string     `gorm:"column:password_hash;type:text;not null" json:"-"`
	StudentID      *int64     `gorm:"column:student_id;index:idx_sys_user_student_id" json:"student_id,omitempty"`
	StaffNo        string     `gorm:"column:staff_no;type:text;index:idx_sys_user_staff_no" json:"staff_no"`
	DisplayName    string     `gorm:"column:display_name;type:text;not null" json:"display_name"`
	AvatarURL      string     `gorm:"column:avatar_url;type:text" json:"avatar_url"`
	Status         string     `gorm:"column:status;type:text;not null;default:active;check:status IN ('active','locked','disabled')" json:"status"`
	LastLoginAt    *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	FailedAttempts int        `gorm:"column:failed_attempts;not null;default:0" json:"failed_attempts"`
	LockUntil      *time.Time `gorm:"column:lock_until" json:"lock_until,omitempty"`
	TokenVersion   int        `gorm:"column:token_version;not null;default:0" json:"token_version"`
	IsDeleted      int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SysUser) TableName() string { return "sys_user" }

// SysRole 角色。docs/03 §4.2。
type SysRole struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Code        string    `gorm:"column:code;type:text;not null;uniqueIndex:uniq_sys_role_code" json:"code"`
	Name        string    `gorm:"column:name;type:text;not null" json:"name"`
	Scope       string    `gorm:"column:scope;type:text;not null;check:scope IN ('school','college','student')" json:"scope"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	IsDeleted   int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SysRole) TableName() string { return "sys_role" }

// SysUserRole 用户-角色。docs/03 §4.2。
type SysUserRole struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID         int64      `gorm:"column:user_id;not null;index" json:"user_id"`
	RoleID         int64      `gorm:"column:role_id;not null;index" json:"role_id"`
	ScopeCollegeID *int64     `gorm:"column:scope_college_id" json:"scope_college_id,omitempty"`
	ScopeOrgType   string     `gorm:"column:scope_org_type;type:text" json:"scope_org_type"`
	ScopeOrgID     *int64     `gorm:"column:scope_org_id" json:"scope_org_id,omitempty"`
	GrantedAt      time.Time  `gorm:"column:granted_at;not null;default:CURRENT_TIMESTAMP" json:"granted_at"`
	GrantedBy      *int64     `gorm:"column:granted_by" json:"granted_by,omitempty"`
	ExpiresAt      *time.Time `gorm:"column:expires_at" json:"expires_at,omitempty"`
	IsDeleted      int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SysUserRole) TableName() string { return "sys_user_role" }

// SysDict 字典。docs/03 §4.3。
type SysDict struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	Category  string    `gorm:"column:category;type:text;not null;index:idx_sys_dict_category" json:"category"`
	Code      string    `gorm:"column:code;type:text;not null" json:"code"`
	NameZh    string    `gorm:"column:name_zh;type:text;not null" json:"name_zh"`
	NameEn    string    `gorm:"column:name_en;type:text" json:"name_en"`
	Sort      int       `gorm:"column:sort;not null;default:0" json:"sort"`
	ExtraJSON string    `gorm:"column:extra_json;type:text" json:"extra_json"`
	IsActive  int       `gorm:"column:is_active;not null;default:1" json:"is_active"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SysDict) TableName() string { return "sys_dict" }

// SysMenu 菜单项。S03 核心通用布局所需。
type SysMenu struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ParentID  *int64    `gorm:"column:parent_id;index:idx_sys_menu_parent_id" json:"parent_id,omitempty"`
	Code      string    `gorm:"column:code;type:text;not null;uniqueIndex:uniq_sys_menu_code" json:"code"`
	Title     string    `gorm:"column:title;type:text;not null" json:"title"`
	Icon      string    `gorm:"column:icon;type:text" json:"icon"`
	Path      string    `gorm:"column:path;type:text" json:"path"`
	Component string    `gorm:"column:component;type:text" json:"component"`
	Sort      int       `gorm:"column:sort;not null;default:0" json:"sort"`
	Visible   int       `gorm:"column:visible;not null;default:1" json:"visible"`
	// Roles 可见角色列表，JSON 数组存储，空数组表示所有角色可见
	Roles     string    `gorm:"column:roles;type:text" json:"roles"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SysMenu) TableName() string { return "sys_menu" }

// BizSeq 业务编号流水。docs/03 §4.3 / ADR-013。
// 复合主键 (module, year)；GORM 通过 primaryKey 标记两列。
type BizSeq struct {
	Module string `gorm:"column:module;type:text;not null;primaryKey" json:"module"`
	Year   int    `gorm:"column:year;not null;primaryKey" json:"year"`
	Cur    int    `gorm:"column:cur;not null;default:0" json:"cur"`
}

func (BizSeq) TableName() string { return "biz_seq" }

// FileMeta 文件元数据。docs/03 §4.4。
type FileMeta struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo        string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	Module       string    `gorm:"column:module;type:text;not null;index:idx_file_meta_module,priority:1" json:"module"`
	BizType      string    `gorm:"column:biz_type;type:text;not null;index:idx_file_meta_module,priority:2" json:"biz_type"`
	OriginalName string    `gorm:"column:original_name;type:text;not null" json:"original_name"`
	StorageKey   string    `gorm:"column:storage_key;type:text;not null;uniqueIndex:uniq_file_meta_storage_key" json:"storage_key"`
	MimeType     string    `gorm:"column:mime_type;type:text;not null" json:"mime_type"`
	SizeBytes    int64     `gorm:"column:size_bytes;not null" json:"size_bytes"`
	SHA256       string    `gorm:"column:sha256;type:text;not null" json:"sha256"`
	UploaderID   int64     `gorm:"column:uploader_id;not null;index:idx_file_meta_uploader" json:"uploader_id"`
	Visibility   string    `gorm:"column:visibility;type:text;not null;default:private;check:visibility IN ('private','org','public')" json:"visibility"`
	IsDeleted    int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (FileMeta) TableName() string { return "file_meta" }

// EventLog 业务事件日志（append-only）。docs/03 §4.4 / ADR-008。
type EventLog struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EventID     string    `gorm:"column:event_id;type:text;not null;uniqueIndex:uniq_event_log_event_id" json:"event_id"`
	Aggregate   string    `gorm:"column:aggregate;type:text;not null;index:idx_event_aggregate,priority:1" json:"aggregate"`
	AggregateID string    `gorm:"column:aggregate_id;type:text;not null;index:idx_event_aggregate,priority:2" json:"aggregate_id"`
	EventType   string    `gorm:"column:event_type;type:text;not null;index:idx_event_type" json:"event_type"`
	Module      string    `gorm:"column:module;type:text;not null;index:idx_event_module,priority:1" json:"module"`
	ActorID     int64     `gorm:"column:actor_id;not null;index:idx_event_actor,priority:1" json:"actor_id"`
	ActorRole   string    `gorm:"column:actor_role;type:text;not null" json:"actor_role"`
	PayloadJSON string    `gorm:"column:payload_json;type:text;not null" json:"payload_json"`
	PrevHash    string    `gorm:"column:prev_hash;type:text" json:"prev_hash"`
	Hash        string    `gorm:"column:hash;type:text;not null" json:"hash"`
	BizNo       string    `gorm:"column:biz_no;type:text" json:"biz_no"`
	IP          string    `gorm:"column:ip;type:text" json:"ip"`
	UA          string    `gorm:"column:ua;type:text" json:"ua"`
	OccurredAt  time.Time `gorm:"column:occurred_at;not null;default:CURRENT_TIMESTAMP;index:idx_event_aggregate,priority:3;index:idx_event_module,priority:2;index:idx_event_actor,priority:2" json:"occurred_at"`
}

func (EventLog) TableName() string { return "event_log" }

// AuditLog API 访问审计。docs/03 §4.4 / ADR-012。
type AuditLog struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TS              time.Time `gorm:"column:ts;not null;default:CURRENT_TIMESTAMP;index:idx_audit_actor,priority:2;index:idx_audit_path,priority:2" json:"ts"`
	ActorID         *int64    `gorm:"column:actor_id;index:idx_audit_actor,priority:1" json:"actor_id,omitempty"`
	Role            string    `gorm:"column:role;type:text" json:"role"`
	Method          string    `gorm:"column:method;type:text" json:"method"`
	Path            string    `gorm:"column:path;type:text;index:idx_audit_path,priority:1" json:"path"`
	Status          int       `gorm:"column:status" json:"status"`
	LatencyMs       int       `gorm:"column:latency_ms" json:"latency_ms"`
	IP              string    `gorm:"column:ip;type:text" json:"ip"`
	UA              string    `gorm:"column:ua;type:text" json:"ua"`
	RequestID       string    `gorm:"column:request_id;type:text" json:"request_id"`
	BizNo           string    `gorm:"column:biz_no;type:text;index:idx_audit_biz_no" json:"biz_no"`
	PayloadRedacted string    `gorm:"column:payload_redacted;type:text" json:"payload_redacted"`
}

func (AuditLog) TableName() string { return "audit_log" }

// Notification 站内信。docs/03 §4.4 / ADR-015。
type Notification struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo           string     `gorm:"column:biz_no;type:text" json:"biz_no"`
	RecipientUserID int64      `gorm:"column:recipient_user_id;not null;index:idx_noti_recipient,priority:1" json:"recipient_user_id"`
	Channel         string     `gorm:"column:channel;type:text;not null;check:channel IN ('site','sms','email','wecom','dingtalk')" json:"channel"`
	Title           string     `gorm:"column:title;type:text;not null" json:"title"`
	Content         string     `gorm:"column:content;type:text;not null" json:"content"`
	LinkURL         string     `gorm:"column:link_url;type:text" json:"link_url"`
	Priority        string     `gorm:"column:priority;type:text;not null;default:normal;check:priority IN ('low','normal','high','urgent')" json:"priority"`
	IsRead          int        `gorm:"column:is_read;not null;default:0;index:idx_noti_recipient,priority:2" json:"is_read"`
	ReadAt          *time.Time `gorm:"column:read_at" json:"read_at,omitempty"`
	SendStatus      string     `gorm:"column:send_status;type:text;not null;default:pending;check:send_status IN ('pending','sent','failed')" json:"send_status"`
	SentAt          *time.Time `gorm:"column:sent_at" json:"sent_at,omitempty"`
	RetryCount      int        `gorm:"column:retry_count;not null;default:0" json:"retry_count"`
	LastError       string     `gorm:"column:last_error;type:text" json:"last_error"`
	IsDeleted       int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;index:idx_noti_recipient,priority:3" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Notification) TableName() string { return "notification" }

// JobRun 定时任务运行日志。docs/03 §4.4 / ADR-016。
type JobRun struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	JobName     string     `gorm:"column:job_name;type:text;not null;index:idx_job_run_name,priority:1" json:"job_name"`
	ScheduledAt time.Time  `gorm:"column:scheduled_at;not null" json:"scheduled_at"`
	StartedAt   *time.Time `gorm:"column:started_at;index:idx_job_run_name,priority:2" json:"started_at,omitempty"`
	FinishedAt  *time.Time `gorm:"column:finished_at" json:"finished_at,omitempty"`
	Status      string     `gorm:"column:status;type:text;not null;check:status IN ('running','success','failed','skipped')" json:"status"`
	DurationMs  int        `gorm:"column:duration_ms" json:"duration_ms"`
	Error       string     `gorm:"column:error;type:text" json:"error"`
	PayloadJSON string     `gorm:"column:payload_json;type:text" json:"payload_json"`
}

func (JobRun) TableName() string { return "job_run" }

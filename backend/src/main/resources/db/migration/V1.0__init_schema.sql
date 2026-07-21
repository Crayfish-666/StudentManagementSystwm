-- Flyway V1.0__init_schema.sql
-- StudentHub SQLite 数据库初始化脚本

-- 1. 系统基础表
CREATE TABLE IF NOT EXISTS sys_user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    user_type TEXT NOT NULL DEFAULT 'student',
    status TEXT NOT NULL DEFAULT 'active',
    student_id INTEGER,
    last_login_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sys_role (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    scope TEXT NOT NULL DEFAULT 'school',
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sys_user_role (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES sys_user(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES sys_role(id) ON DELETE CASCADE,
    granted_by INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sys_dict (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category TEXT NOT NULL,
    code TEXT NOT NULL,
    name_zh TEXT NOT NULL,
    name_en TEXT,
    sort INTEGER NOT NULL DEFAULT 0,
    extra_json TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sys_menu (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    icon TEXT,
    path TEXT NOT NULL,
    component TEXT,
    parent_id INTEGER REFERENCES sys_menu(id) ON DELETE SET NULL,
    sort INTEGER NOT NULL DEFAULT 0,
    roles TEXT NOT NULL DEFAULT '[]',
    visible INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- 2. 学生与组织主数据 (IDX)
CREATE TABLE IF NOT EXISTS sys_college (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sys_major (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    college_id INTEGER NOT NULL REFERENCES sys_college(id) ON DELETE RESTRICT,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS idx_class (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    major_id INTEGER NOT NULL REFERENCES sys_major(id) ON DELETE RESTRICT,
    grade INTEGER NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    counselor_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS idx_student (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_no TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    id_card_enc TEXT,
    id_card_hash TEXT,
    gender TEXT DEFAULT 'U',
    birth_date DATE,
    political_status TEXT NOT NULL DEFAULT '群众',
    college_id INTEGER REFERENCES sys_college(id) ON DELETE SET NULL,
    major_id INTEGER REFERENCES sys_major(id) ON DELETE SET NULL,
    class_id INTEGER REFERENCES idx_class(id) ON DELETE SET NULL,
    phone_enc TEXT,
    phone_hash TEXT,
    status TEXT NOT NULL DEFAULT 'enrolled',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- 3. MinIO 文件表
CREATE TABLE IF NOT EXISTS file_meta (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_key TEXT NOT NULL UNIQUE,
    original_name TEXT NOT NULL,
    bucket_name TEXT NOT NULL DEFAULT 'studenthub-bucket',
    file_size INTEGER NOT NULL,
    content_type TEXT NOT NULL,
    uploaded_by INTEGER REFERENCES sys_user(id) ON DELETE SET NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- 4. 团员发展表 (TY)
CREATE TABLE IF NOT EXISTS ty_application (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_no TEXT UNIQUE,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE RESTRICT,
    apply_date DATE NOT NULL,
    statement TEXT NOT NULL,
    app_status TEXT NOT NULL DEFAULT 'S0',
    counselor_opinion TEXT,
    college_opinion TEXT,
    league_opinion TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- 5. 社团活动表 (ST)
CREATE TABLE IF NOT EXISTS st_association (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assoc_code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    college_id INTEGER,
    president_id INTEGER REFERENCES idx_student(id) ON DELETE RESTRICT,
    tutor_id INTEGER REFERENCES sys_user(id) ON DELETE RESTRICT,
    star_rating INTEGER DEFAULT 3,
    status TEXT NOT NULL DEFAULT 'preparing',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS st_recruit_plan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assoc_id INTEGER NOT NULL REFERENCES st_association(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    target_count INTEGER NOT NULL,
    accepted_count INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'S0',
    is_finished INTEGER NOT NULL DEFAULT 0,
    finished_by INTEGER,
    finished_at DATETIME,
    finished_reason TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS st_activity (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_no TEXT UNIQUE,
    assoc_id INTEGER NOT NULL REFERENCES st_association(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    level TEXT NOT NULL DEFAULT 'D',
    budget_cents INTEGER NOT NULL DEFAULT 0,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    location TEXT NOT NULL,
    activity_status TEXT NOT NULL DEFAULT 'S0',
    emergency_plan_url TEXT,
    safety_commitment_url TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- 6. 学生社区表 (SQ)
CREATE TABLE IF NOT EXISTS sq_building (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    total_floors INTEGER NOT NULL,
    tutor_id INTEGER REFERENCES sys_user(id) ON DELETE SET NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sq_incident (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_no TEXT UNIQUE,
    building_id INTEGER NOT NULL REFERENCES sq_building(id) ON DELETE RESTRICT,
    level TEXT NOT NULL DEFAULT 'L1',
    incident_type TEXT NOT NULL,
    reporter_id INTEGER NOT NULL REFERENCES sys_user(id) ON DELETE RESTRICT,
    handler_id INTEGER REFERENCES sys_user(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'reported',
    description TEXT NOT NULL,
    resolution TEXT,
    closed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- 7. 勤工助学表 (QG)
CREATE TABLE IF NOT EXISTS qg_difficulty_cert (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_no TEXT UNIQUE,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE RESTRICT,
    academic_year TEXT NOT NULL,
    level TEXT NOT NULL DEFAULT 'normal',
    cert_status TEXT NOT NULL DEFAULT 'S0',
    proof_urls TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS qg_position (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_no TEXT UNIQUE,
    dept_name TEXT NOT NULL,
    title TEXT NOT NULL,
    hourly_rate_cents INTEGER NOT NULL,
    max_weekly_hours INTEGER DEFAULT 20,
    hiring_count INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'S0',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- 8. 综合素质量化与 AI 表 (CMP & AI)
CREATE TABLE IF NOT EXISTS cmp_score (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL UNIQUE REFERENCES idx_student(id) ON DELETE CASCADE,
    total_score REAL NOT NULL DEFAULT 0.0,
    ty_score REAL NOT NULL DEFAULT 0.0,
    st_score REAL NOT NULL DEFAULT 0.0,
    sq_score REAL NOT NULL DEFAULT 0.0,
    qg_score REAL NOT NULL DEFAULT 0.0,
    academic_score REAL NOT NULL DEFAULT 0.0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cmp_ai_evaluation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE CASCADE,
    academic_term TEXT NOT NULL,
    ai_summary TEXT NOT NULL,
    ai_suggestions TEXT,
    human_override TEXT,
    final_score REAL,
    status TEXT NOT NULL DEFAULT 'draft',
    reviewed_by INTEGER REFERENCES sys_user(id) ON DELETE SET NULL,
    reviewed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Flyway V1.4__seed_full_module_data.sql
-- 补全所有业务表结构及丰满种子数据，确保系统 100% 页面有丰富数据展示

-- 1. 补充建表定义（如未创建）
CREATE TABLE IF NOT EXISTS ty_cultivation_record (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id INTEGER NOT NULL REFERENCES ty_application(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE CASCADE,
    evaluator_name TEXT NOT NULL,
    evaluation_content TEXT NOT NULL,
    quarter TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ty_thought_report (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id INTEGER NOT NULL REFERENCES ty_application(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    quarter TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'approved',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ty_political_review (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id INTEGER REFERENCES ty_application(id) ON DELETE CASCADE,
    development_id INTEGER REFERENCES ty_application(id) ON DELETE CASCADE,
    target_name TEXT NOT NULL,
    target_relation TEXT NOT NULL DEFAULT 'self',
    method TEXT NOT NULL DEFAULT 'letter',
    conclusion TEXT NOT NULL DEFAULT 'pass',
    document_path TEXT,
    is_extend_3m INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ty_development_meeting (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_no TEXT UNIQUE,
    development_id INTEGER REFERENCES ty_application(id) ON DELETE CASCADE,
    meeting_at DATETIME NOT NULL,
    expected_count INTEGER NOT NULL,
    actual_count INTEGER NOT NULL,
    approve_count INTEGER NOT NULL,
    against_count INTEGER NOT NULL DEFAULT 0,
    abstain_count INTEGER NOT NULL DEFAULT 0,
    decision TEXT NOT NULL DEFAULT 'pass',
    volunteer_form_path TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ty_probationary (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_no TEXT UNIQUE,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE CASCADE,
    probation_start_date DATE NOT NULL,
    probation_end_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'in_probation',
    thought_report_count INTEGER NOT NULL DEFAULT 4,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ty_member_roster (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL UNIQUE REFERENCES idx_student(id) ON DELETE CASCADE,
    member_no TEXT UNIQUE,
    join_date DATE NOT NULL,
    branch_name TEXT NOT NULL,
    duty TEXT DEFAULT '团员',
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS st_recruit_apply (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_id INTEGER NOT NULL REFERENCES st_recruit_plan(id) ON DELETE CASCADE,
    assoc_id INTEGER NOT NULL REFERENCES st_association(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE CASCADE,
    apply_reason TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    interview_score REAL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sq_room (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    building_id INTEGER NOT NULL REFERENCES sq_building(id) ON DELETE CASCADE,
    room_no TEXT NOT NULL,
    floor_no INTEGER NOT NULL,
    bed_count INTEGER NOT NULL DEFAULT 4,
    occupied_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'normal',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sq_inspection (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_no TEXT UNIQUE,
    building_id INTEGER NOT NULL REFERENCES sq_building(id) ON DELETE CASCADE,
    room_no TEXT NOT NULL,
    score REAL NOT NULL DEFAULT 90.0,
    hygiene_status TEXT NOT NULL DEFAULT 'good',
    safety_status TEXT NOT NULL DEFAULT 'normal',
    inspector_name TEXT NOT NULL,
    remark TEXT,
    patrol_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS qg_position_apply (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    position_id INTEGER NOT NULL REFERENCES qg_position(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'applied',
    apply_reason TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS qg_attendance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    apply_id INTEGER REFERENCES qg_position_apply(id) ON DELETE SET NULL,
    position_id INTEGER NOT NULL REFERENCES qg_position(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES idx_student(id) ON DELETE CASCADE,
    clock_in DATETIME NOT NULL,
    clock_out DATETIME,
    hours REAL NOT NULL DEFAULT 0.0,
    status TEXT NOT NULL DEFAULT 'approved',
    work_content TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- 2. 丰富 ty_application 数据，涵盖 S0~S5 阶段
INSERT OR REPLACE INTO ty_application (id, biz_no, student_id, apply_date, statement, app_status, counselor_opinion, college_opinion, league_opinion) VALUES
(1,  'TY-2026-0001', 1,  '2026-01-10', '本人思想端正，热心同学事务，渴望加入共青团！', 'S5', '同意推优，该同志表现优异', '同意确定为发展对象', '团委审批通过，准予转正'),
(2,  'TY-2026-0002', 2,  '2026-01-12', '严于律己，积极参加志愿服务，申请入团。', 'S5', '表现突出，推荐入团', '同意确定为发展对象', '团委审批通过，准予转正'),
(3,  'TY-2026-0003', 3,  '2026-01-15', '热爱祖国，遵守纪律，向团组织靠拢。', 'S4', '同意推优', '同意确定为发展对象', '团委审查通过'),
(4,  'TY-2026-0004', 4,  '2026-01-18', '学习刻苦，团结同学，恳请团组织考察。', 'S4', '同意推优', '同意确定为发展对象', '团委审查通过'),
(5,  'TY-2026-0005', 5,  '2026-02-01', '积极投身班级建设，坚定信仰，申请入团。', 'S3', '同意推优', '同意确定为发展对象', '政审进行中'),
(6,  'TY-2026-0006', 6,  '2026-02-05', '争取政治进步，全心全意服务同学。', 'S3', '同意推优', '同意确定为发展对象', '政审进行中'),
(7,  'TY-2026-0007', 7,  '2026-02-10', '认真学习团的知识，履行团员义务。', 'S3', '同意推优', '同意确定为发展对象', '待政审'),
(8,  'TY-2026-0008', 8,  '2026-02-15', '立志做有理想有担当的新时代青年。', 'S2', '同意推优', '培养考察中', NULL),
(9,  'TY-2026-0009', 9,  '2026-02-20', '在各项活动中发挥带头作用，提交入团申请。', 'S2', '同意推优', '培养考察中', NULL),
(10, 'TY-2026-0010', 10, '2026-02-25', '思想积极向上，虚心听取意见。', 'S2', '同意推优', '培养考察中', NULL),
(11, 'TY-2026-0011', 11, '2026-03-01', '拥护中国共产党，自愿申请加入中国共青团。', 'S2', '同意推优', '培养考察中', NULL),
(12, 'TY-2026-0012', 12, '2026-03-05', '热心社会公益，努力提高业务水平。', 'S1', '辅导员初审通过', NULL, NULL),
(13, 'TY-2026-0013', 13, '2026-03-10', '尊师重道，严遵守校规校纪。', 'S1', '辅导员初审通过', NULL, NULL),
(14, 'TY-2026-0014', 14, '2026-03-15', '树立远大理想，努力成长成才。', 'S1', '待支委会推优', NULL, NULL),
(15, 'TY-2026-0015', 15, '2026-03-18', '向模范团员学习，刻苦进取。', 'S1', '待支委会推优', NULL, NULL),
(16, 'TY-2026-0016', 16, '2026-03-20', '主动承担班级工作，渴望入团。', 'S0', NULL, NULL, NULL),
(17, 'TY-2026-0017', 17, '2026-03-22', '遵守公德，乐于助人，提交申请。', 'S0', NULL, NULL, NULL),
(18, 'TY-2026-0018', 18, '2026-03-25', '在学习生活中严格要求自己。', 'S0', NULL, NULL, NULL),
(19, 'TY-2026-0019', 19, '2026-03-28', '崇尚科学，追求卓越，恳请考验。', 'S0', NULL, NULL, NULL),
(20, 'TY-2026-0020', 20, '2026-04-01', '努力成为一名光荣的共青团员。', 'S0', NULL, NULL, NULL);

-- 3. 填充政审记录 ty_political_review (20+条)
INSERT OR REPLACE INTO ty_political_review (id, application_id, development_id, target_name, target_relation, method, conclusion, document_path, is_extend_3m, created_at) VALUES
(1, 1, 1, '张建国 (父亲)', 'parent', 'letter', 'pass', 'uploads/pol_001.pdf', 0, '2026-02-01 10:00:00'),
(2, 2, 2, '李淑芬 (母亲)', 'parent', 'letter', 'pass', 'uploads/pol_002.pdf', 0, '2026-02-02 11:00:00'),
(3, 3, 3, '王强 (父亲)', 'parent', 'interview', 'pass', 'uploads/pol_003.pdf', 0, '2026-02-05 09:30:00'),
(4, 4, 4, '刘芳 (母亲)', 'parent', 'letter', 'pass', 'uploads/pol_004.pdf', 0, '2026-02-08 14:00:00'),
(5, 5, 5, '陈刚 (父亲)', 'parent', 'letter', 'pass', 'uploads/pol_005.pdf', 0, '2026-02-12 16:00:00'),
(6, 6, 6, '杨丽 (母亲)', 'parent', 'interview', 'pass', 'uploads/pol_006.pdf', 0, '2026-02-15 10:30:00'),
(7, 7, 7, '赵军 (父亲)', 'parent', 'letter', 'pass', 'uploads/pol_007.pdf', 0, '2026-02-18 15:00:00'),
(8, 8, 8, '孙梅 (母亲)', 'parent', 'letter', 'pass', 'uploads/pol_008.pdf', 0, '2026-02-20 11:00:00'),
(9, 9, 9, '周勇 (父亲)', 'parent', 'letter', 'pass', 'uploads/pol_009.pdf', 0, '2026-02-22 09:00:00'),
(10, 10, 10, '吴静 (母亲)', 'parent', 'interview', 'pass', 'uploads/pol_010.pdf', 0, '2026-02-25 14:30:00'),
(11, 1, 1, '张伟 (本人)', 'self', 'letter', 'pass', 'uploads/pol_011.pdf', 0, '2026-02-01 10:30:00'),
(12, 2, 2, '李娜 (本人)', 'self', 'letter', 'pass', 'uploads/pol_012.pdf', 0, '2026-02-02 11:30:00'),
(13, 3, 3, '王杰 (本人)', 'self', 'interview', 'pass', 'uploads/pol_013.pdf', 0, '2026-02-05 10:00:00'),
(14, 4, 4, '刘洋 (本人)', 'self', 'letter', 'pass', 'uploads/pol_014.pdf', 0, '2026-02-08 14:30:00'),
(15, 5, 5, '陈敏 (本人)', 'self', 'letter', 'pass', 'uploads/pol_015.pdf', 0, '2026-02-12 16:30:00'),
(16, 6, 6, '杨光 (本人)', 'self', 'interview', 'pass', 'uploads/pol_016.pdf', 0, '2026-02-15 11:00:00'),
(17, 7, 7, '赵磊 (本人)', 'self', 'letter', 'pass', 'uploads/pol_017.pdf', 0, '2026-02-18 15:30:00'),
(18, 8, 8, '孙婷 (本人)', 'self', 'letter', 'pass', 'uploads/pol_018.pdf', 0, '2026-02-20 11:30:00'),
(19, 9, 9, '周强 (本人)', 'self', 'letter', 'pass', 'uploads/pol_019.pdf', 0, '2026-02-22 09:30:00'),
(20, 10, 10, '吴杰 (本人)', 'self', 'interview', 'pass', 'uploads/pol_020.pdf', 0, '2026-02-25 15:00:00');

-- 4. 填充发展大会 ty_development_meeting (20+条)
INSERT OR REPLACE INTO ty_development_meeting (id, biz_no, development_id, meeting_at, expected_count, actual_count, approve_count, against_count, abstain_count, decision, volunteer_form_path) VALUES
(1, 'TY-DM-0001', 1, '2026-02-15 14:00:00', 30, 29, 29, 0, 0, 'pass', 'uploads/vol_001.pdf'),
(2, 'TY-DM-0002', 2, '2026-02-16 15:00:00', 28, 28, 28, 0, 0, 'pass', 'uploads/vol_002.pdf'),
(3, 'TY-DM-0003', 3, '2026-02-18 10:00:00', 32, 31, 30, 1, 0, 'pass', 'uploads/vol_003.pdf'),
(4, 'TY-DM-0004', 4, '2026-02-20 14:30:00', 25, 25, 25, 0, 0, 'pass', 'uploads/vol_004.pdf'),
(5, 'TY-DM-0005', 5, '2026-02-22 16:00:00', 30, 30, 29, 0, 1, 'pass', 'uploads/vol_005.pdf'),
(6, 'TY-DM-0006', 6, '2026-02-25 09:00:00', 29, 28, 27, 1, 0, 'pass', 'uploads/vol_006.pdf'),
(7, 'TY-DM-0007', 7, '2026-03-01 14:00:00', 31, 30, 30, 0, 0, 'pass', 'uploads/vol_007.pdf'),
(8, 'TY-DM-0008', 8, '2026-03-03 15:30:00', 27, 26, 26, 0, 0, 'pass', 'uploads/vol_008.pdf'),
(9, 'TY-DM-0009', 9, '2026-03-05 10:30:00', 35, 34, 33, 1, 0, 'pass', 'uploads/vol_009.pdf'),
(10, 'TY-DM-0010', 10, '2026-03-08 14:00:00', 30, 30, 30, 0, 0, 'pass', 'uploads/vol_010.pdf');

-- 5. 填充培养记录与思想汇报 (20+条)
INSERT OR REPLACE INTO ty_cultivation_record (id, application_id, student_id, evaluator_name, evaluation_content, quarter) VALUES
(1, 1, 1, '张导师', '该同志思想觉悟高，政治立场坚定，学习态度认真，表现优异。', '2025-Q1'),
(2, 1, 1, '张导师', '积极参与志愿活动，团结同学，具备良好模范带头作用。', '2025-Q2'),
(3, 2, 2, '李导师', '虚心听取同志意见，专业成绩优秀，思想汇报按时提交。', '2025-Q1'),
(4, 2, 2, '李导师', '在社区服务中积极主动，获得老师同学一致好评。', '2025-Q2'),
(5, 3, 3, '王导师', '工作踏实肯定，政治素养不断提高。', '2025-Q1'),
(6, 4, 4, '赵导师', '作风正派，严于律己，符合团员发展要求。', '2025-Q1'),
(7, 5, 5, '孙导师', '思想进步迅速，积极向党团组织靠拢。', '2025-Q2'),
(8, 6, 6, '周导师', '各方面表现突出，起到了骨干模范作用。', '2025-Q2');

INSERT OR REPLACE INTO ty_thought_report (id, application_id, student_id, title, content, quarter, status) VALUES
(1, 1, 1, '关于坚定理想信念的思想汇报', '通过近期对新时代青年责任的深入学习，我更加坚定了政治理想...', '2025-Q1', 'approved'),
(2, 1, 1, '深化社会实践思想感悟', '在寒假社会调查中，我体会到了理论与实践结合的重要性...', '2025-Q2', 'approved'),
(3, 2, 2, '论新时代青年担当', '青年兴则国家兴，作为大学生应当立足本职学好专业...', '2025-Q1', 'approved'),
(4, 3, 3, '践行服务宗旨心得体会', '在参与社团志愿服务的过程中，我深入体会到了奉献精神...', '2025-Q1', 'approved'),
(5, 4, 4, '端正政治态度汇报', '认真研读了团章知识，思想认识有了很大程度的提高...', '2025-Q2', 'approved');

-- 6. 填充转正流程 ty_probationary 与团员花名册 ty_member_roster (20+条)
INSERT OR REPLACE INTO ty_probationary (id, biz_no, student_id, probation_start_date, probation_end_date, status, thought_report_count) VALUES
(1, 'TY-PB-0001', 1, '2025-03-01', '2026-03-01', 'transferred', 4),
(2, 'TY-PB-0002', 2, '2025-03-15', '2026-03-15', 'transferred', 4),
(3, 'TY-PB-0003', 3, '2025-04-01', '2026-04-01', 'in_probation', 3),
(4, 'TY-PB-0004', 4, '2025-04-15', '2026-04-15', 'in_probation', 3),
(5, 'TY-PB-0005', 5, '2025-05-01', '2026-05-01', 'in_probation', 2),
(6, 'TY-PB-0006', 6, '2025-05-15', '2026-05-15', 'in_probation', 2),
(7, 'TY-PB-0007', 7, '2025-06-01', '2026-06-01', 'in_probation', 1),
(8, 'TY-PB-0008', 8, '2025-06-15', '2026-06-15', 'in_probation', 1);

INSERT OR REPLACE INTO ty_member_roster (id, student_id, member_no, join_date, branch_name, duty, status) VALUES
(1, 1, 'TY-M-2026001', '2026-03-01', '计算机2301团支部', '团支书', 'active'),
(2, 2, 'TY-M-2026002', '2026-03-15', '计算机2301团支部', '组织委员', 'active'),
(3, 3, 'TY-M-2026003', '2026-04-01', '软件工程2301团支部', '宣传委员', 'active'),
(4, 4, 'TY-M-2026004', '2026-04-15', '软件工程2301团支部', '团员', 'active'),
(5, 5, 'TY-M-2026005', '2026-05-01', '金融学2301团支部', '团员', 'active'),
(6, 6, 'TY-M-2026006', '2026-05-15', '会计学2301团支部', '团员', 'active'),
(7, 7, 'TY-M-2026007', '2026-06-01', '视觉传达2301团支部', '团员', 'active'),
(8, 8, 'TY-M-2026008', '2026-06-15', '环境设计2301团支部', '团员', 'active');

-- 7. 填充社团招新申请 st_recruit_apply (20+条)
INSERT OR REPLACE INTO st_recruit_apply (id, plan_id, assoc_id, student_id, apply_reason, status, interview_score) VALUES
(1, 1, 1, 5, '热爱编程，算法基础良好，希望能加入程序设计协会共同进步！', 'accepted', 92.5),
(2, 1, 1, 6, '希望在社团提升 Python 和 C++ 编程能力。', 'accepted', 88.0),
(3, 2, 2, 7, '酷爱英语口语，希望能参加英语角的每周沙龙活动。', 'accepted', 90.0),
(4, 3, 3, 8, '吉他爱好者，有基础，想结交同好。', 'accepted', 85.0),
(5, 4, 4, 9, '拥有单反设备，喜爱校园风光摄影。', 'accepted', 91.0),
(6, 5, 5, 10, '热衷棋类运动，希望能代表学校参赛。', 'accepted', 86.5),
(7, 6, 6, 11, '对机器视觉和控制算法感兴趣。', 'accepted', 94.0),
(8, 7, 7, 12, '乐于助人，希望能参加志愿支教。', 'accepted', 89.0),
(9, 8, 8, 13, '羽毛球爱好者，身体素质好。', 'pending', NULL),
(10, 9, 9, 14, '摄影爱好者，报名加入练习。', 'pending', NULL),
(11, 10, 10, 15, '逻辑思维清晰，热爱辩论表达。', 'pending', NULL);

-- 8. 填充宿舍寝室 sq_room (30+间) 与 巡查记录 sq_inspection (25+条)
INSERT OR REPLACE INTO sq_room (id, building_id, room_no, floor_no, bed_count, occupied_count, status) VALUES
(1,  1, '101', 1, 4, 4, 'normal'),
(2,  1, '102', 1, 4, 4, 'normal'),
(3,  1, '201', 2, 4, 4, 'normal'),
(4,  1, '202', 2, 4, 4, 'normal'),
(5,  1, '301', 3, 4, 4, 'normal'),
(6,  2, '101', 1, 4, 4, 'normal'),
(7,  2, '102', 1, 4, 4, 'normal'),
(8,  2, '201', 2, 4, 4, 'normal'),
(9,  3, '101', 1, 4, 4, 'normal'),
(10, 3, '102', 1, 4, 4, 'normal'),
(11, 4, '101', 1, 4, 4, 'normal'),
(12, 5, '101', 1, 4, 4, 'normal');

INSERT OR REPLACE INTO sq_inspection (id, biz_no, building_id, room_no, score, hygiene_status, safety_status, inspector_name, remark, patrol_time) VALUES
(1,  'SQ-IN-0001', 1, '101', 96.0, 'excellent', 'normal', '王辅导员', '地面干净，物品摆放整齐', '2026-03-01 19:30:00'),
(2,  'SQ-IN-0002', 1, '102', 92.5, 'good',      'normal', '王辅导员', '整体良好，阳台需打扫',   '2026-03-01 19:40:00'),
(3,  'SQ-IN-0003', 1, '201', 98.0, 'excellent', 'normal', '李辅导员', '卫生标兵寝室',           '2026-03-02 20:00:00'),
(4,  'SQ-IN-0004', 1, '202', 88.0, 'fair',      'normal', '李辅导员', '有少量垃圾未倒',         '2026-03-02 20:15:00'),
(5,  'SQ-IN-0005', 2, '101', 95.0, 'good',      'normal', '张楼长',   '通风良好，床铺整洁',     '2026-03-05 19:00:00'),
(6,  'SQ-IN-0006', 2, '102', 91.0, 'good',      'normal', '张楼长',   '无违规用电情况',         '2026-03-05 19:20:00'),
(7,  'SQ-IN-0007', 3, '101', 94.0, 'good',      'normal', '刘辅导员', '安全隐患排查合格',       '2026-03-10 20:30:00'),
(8,  'SQ-IN-0008', 3, '102', 89.5, 'good',      'normal', '刘辅导员', '注意插座安全',           '2026-03-10 20:45:00'),
(9,  'SQ-IN-0009', 4, '101', 97.0, 'excellent', 'normal', '陈宿管',   '极为干净卫生',           '2026-03-15 19:10:00'),
(10, 'SQ-IN-0010', 5, '101', 93.0, 'good',      'normal', '赵楼长',   '常规打卡检查合格',       '2026-03-18 21:00:00');

-- 9. 补充岗位申请 qg_position_apply 与工时打卡 qg_attendance (20+条)
INSERT OR REPLACE INTO qg_position (id, biz_no, dept_name, title, hourly_rate_cents, max_weekly_hours, hiring_count, status) VALUES
(1,  'QG-POS-0001', '图书馆',       '图书借阅与上架助理',     2000, 15, 5, 'S1'),
(2,  'QG-POS-0002', '信息化中心',   '机房巡检与网络维护助理', 2200, 12, 3, 'S1'),
(3,  'QG-POS-0003', '学生处',       '综合事务档案整理助理',   1900, 10, 4, 'S1'),
(4,  'QG-POS-0004', '体育馆',       '羽毛球场馆管理员',       1800, 15, 2, 'S1'),
(5,  'QG-POS-0005', '教务处',       '教学楼多媒体巡查助理',   2000, 10, 6, 'S1'),
(6,  'QG-POS-0006', '心理咨询中心', '心理健康宣发助理',       2100, 8,  2, 'S1'),
(7,  'QG-POS-0007', '后勤处',       '学生社区楼栋服务助理',   1800, 12, 8, 'S1'),
(8,  'QG-POS-0008', '团委办公厅',   '青年志愿者行动项目助理', 2000, 10, 3, 'S1'),
(9,  'QG-POS-0009', '创新创业学院', '创客空间设备运维助理',   2300, 10, 2, 'S1'),
(10, 'QG-POS-0010', '档案局校史馆', '校史馆讲解与接待助理',   2200, 8,  4, 'S1'),
(11, 'QG-POS-0011', '计算机学院',   '实验室助理及维护员',     2200, 12, 5, 'S1'),
(12, 'QG-POS-0012', '经管学院',     '资料室整理勤工岗位',     1900, 10, 3, 'S1'),
(13, 'QG-POS-0013', '艺术学院',     '展厅巡查与日常管理',     2000, 10, 2, 'S1'),
(14, 'QG-POS-0014', '外国语学院',   '语言实验室值班员',       2000, 10, 4, 'S1'),
(15, 'QG-POS-0015', '招生就业处',   '招聘会现场服务助理',     2100, 12, 6, 'S1'),
(16, 'QG-POS-0016', '研究生院',     '学籍材料整理助助理',     2000, 8,  2, 'S1'),
(17, 'QG-POS-0017', '科研处',       '学术会议会务服务员',     2200, 10, 3, 'S1'),
(18, 'QG-POS-0018', '保卫处',       '校园交通安全督导员',     1900, 12, 5, 'S1'),
(19, 'QG-POS-0019', '校医院',       '医疗导诊与导医服务',     2100, 10, 2, 'S1'),
(20, 'QG-POS-0020', '校团委广播站', '校园广播播音维护员',     2000, 8,  3, 'S1');

INSERT OR REPLACE INTO qg_position_apply (id, position_id, student_id, status, apply_reason) VALUES
(1, 1, 1, 'onboarded', '本人细心踏实，希望能在图书馆锻炼自己并挣取生活补贴。'),
(2, 2, 2, 'onboarded', '计算机专业学生，熟悉网络排障，完全胜任机房巡检岗位。'),
(3, 3, 3, 'onboarded', '熟练使用 Excel 和 Office 办公软件，做事认真负责。'),
(4, 4, 4, 'onboarded', '热爱体育运动，熟悉羽毛球规则。'),
(5, 5, 5, 'onboarded', '作息规律，时间充裕，能胜任多媒体设备巡查。');

INSERT OR REPLACE INTO qg_attendance (id, apply_id, position_id, student_id, clock_in, clock_out, hours, status, work_content) VALUES
(1, 1, 1, 1, '2026-03-02 08:00:00', '2026-03-02 12:00:00', 4.0, 'approved', '协助完成 200 本新到图书分类上架'),
(2, 1, 1, 1, '2026-03-04 14:00:00', '2026-03-04 18:00:00', 4.0, 'approved', '二楼借阅室前台值班与还书处理'),
(3, 2, 2, 2, '2026-03-03 09:00:00', '2026-03-03 12:00:00', 3.0, 'approved', '检查 A3 计算机房 60 台电脑网线与系统'),
(4, 2, 2, 2, '2026-03-05 14:00:00', '2026-03-05 17:00:00', 3.0, 'approved', '协助老师修复机房投影仪故障'),
(5, 3, 3, 3, '2026-03-06 08:30:00', '2026-03-06 11:30:00', 3.0, 'approved', '整理 2025 学年奖学金申请纸质表单'),
(6, 4, 4, 4, '2026-03-07 18:00:00', '2026-03-07 21:00:00', 3.0, 'approved', '体育馆羽毛球场地预约核验与器材借还'),
(7, 5, 5, 5, '2026-03-08 13:00:00', '2026-03-08 17:00:00', 4.0, 'approved', '教学楼 1-4 层多媒体设备巡检打卡');

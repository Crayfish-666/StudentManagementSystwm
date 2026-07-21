-- Flyway V1.2__seed_dashboard_data.sql
-- 为看板和排名页补充演示数据

-- ========== 1. sys_role: 系统角色（DataInitializer 也会创建，此处用 INSERT OR IGNORE 防冲突） ==========
INSERT OR IGNORE INTO sys_role (id, code, name, scope, description) VALUES
(1, 'R-SY-ADMIN', '系统管理员', 'school', '拥有所有模块的全部权限'),
(2, 'R-IDX-COUNSELOR', '辅导员', 'class', '管理所辖班级学生'),
(3, 'R-TY-SECRETARY', '团委书记', 'college', '管理团员发展全流程'),
(4, 'R-ST-TUTOR', '社团指导老师', 'school', '审核社团活动'),
(5, 'R-SQ-TUTOR', '社区指导老师', 'building', '管理宿舍社区'),
(6, 'R-QG-OFFICER', '勤工助学管理员', 'school', '审核岗位与申请'),
(7, 'R-CMP-ADMIN', '综合素质管理员', 'school', '管理量化规则与计算'),
(8, 'R-STUDENT', '学生', 'self', '查看本人信息');

-- ========== 1.5 sys_user: 初始系统用户（明文密码，DataInitializer 启动后会自动升级为 BCrypt） ==========
-- 注意：此处密码为明文，仅用于满足外键约束；DataInitializer 会检测并升级为 BCrypt
INSERT OR IGNORE INTO sys_user (id, username, password_hash, display_name, user_type, status, student_id) VALUES
(1, 'admin', 'admin@123', '系统管理员', 'admin', 'active', NULL),
(2, 'counselor', 'counselor@123', '张辅导员', 'counselor', 'active', NULL);

-- 为前 5 个学生创建用户账号（满足 sq_incident 等表的 reporter_id 外键）
INSERT OR IGNORE INTO sys_user (id, username, password_hash, display_name, user_type, status, student_id) VALUES
(3,  '2023010101', 'student@123', '张伟', 'student', 'active', 1),
(4,  '2023010102', 'student@123', '王芳', 'student', 'active', 2),
(5,  '2023010103', 'student@123', '李娜', 'student', 'active', 3),
(6,  '2023010104', 'student@123', '刘洋', 'student', 'active', 4),
(7,  '2023010105', 'student@123', '陈杰', 'student', 'active', 5),
(8,  '2023010106', 'student@123', '杨光', 'student', 'active', 6),
(9,  '2023010107', 'student@123', '黄磊', 'student', 'active', 7),
(10, '2023010108', 'student@123', '周敏', 'student', 'active', 8),
(11, '2023010109', 'student@123', '吴强', 'student', 'active', 9),
(12, '2023010110', 'student@123', '徐霞', 'student', 'active', 10),
(13, '2023010111', 'student@123', '孙浩', 'student', 'active', 11),
(14, '2023010112', 'student@123', '胡婷', 'student', 'active', 12),
(15, '2023010113', 'student@123', '朱勇', 'student', 'active', 13),
(16, '2023010114', 'student@123', '高丽', 'student', 'active', 14),
(17, '2023010115', 'student@123', '林涛', 'student', 'active', 15),
(18, '2023010116', 'student@123', '何静', 'student', 'active', 16),
(19, '2023010117', 'student@123', '郭平', 'student', 'active', 17),
(20, '2023010118', 'student@123', '马明', 'student', 'active', 18),
(21, '2023010119', 'student@123', '罗军', 'student', 'active', 19),
(22, '2023010120', 'student@123', '梁晨', 'student', 'active', 20);

-- sys_user_role: 用户-角色关联（DataInitializer 也会创建，此处先建立基础关联）
INSERT OR IGNORE INTO sys_user_role (user_id, role_id) VALUES
(1, 1),   -- admin → 系统管理员
(2, 2);   -- counselor → 辅导员

-- ========== 2. sys_dict: 系统字典（前端 dict store 需要） ==========
INSERT OR IGNORE INTO sys_dict (category, code, name_zh, sort, is_active) VALUES
('political_status', 'masses', '群众', 1, 1),
('political_status', 'league_member', '共青团员', 2, 1),
('political_status', 'party_member', '中共党员', 3, 1),
('political_status', 'probationary_member', '中共预备党员', 4, 1),
('political_status', 'activist', '入党积极分子', 5, 1),
('gender', 'M', '男', 1, 1),
('gender', 'F', '女', 2, 1),
('gender', 'U', '未知', 3, 1),
('student_status', 'enrolled', '在读', 1, 1),
('student_status', 'suspended', '休学', 2, 1),
('student_status', 'graduated', '毕业', 3, 1),
('student_status', 'dropped', '退学', 4, 1),
('assoc_status', 'preparing', '筹备中', 1, 1),
('assoc_status', 'active', '活跃', 2, 1),
('assoc_status', 'frozen', '冻结', 3, 1),
('assoc_status', 'disbanded', '解散', 4, 1),
('incident_level', 'L1', '一般事件', 1, 1),
('incident_level', 'L2', '较重事件', 2, 1),
('incident_level', 'L3', '严重事件', 3, 1),
('incident_level', 'L4', '特大事件', 4, 1),
('qg_position_status', 'S0', '草稿', 1, 1),
('qg_position_status', 'S1', '招聘中', 2, 1),
('qg_position_status', 'S2', '已满员', 3, 1),
('qg_position_status', 'S3', '已关闭', 4, 1),
('qg_position_status', 'S4', '已撤销', 5, 1);

-- ========== 3. st_activity: 社团活动（10 条，含不同状态） ==========
INSERT OR IGNORE INTO st_activity (id, biz_no, assoc_id, title, level, budget_cents, start_time, end_time, location, activity_status) VALUES
(1, 'ST-A-0001', 1, '算法竞赛入门讲座', 'D', 50000, '2026-03-15 14:00:00', '2026-03-15 16:00:00', 'A栋101', 'S3'),
(2, 'ST-A-0002', 2, '英语角春季开场活动', 'D', 30000, '2026-03-20 19:00:00', '2026-03-20 21:00:00', '图书馆报告厅', 'S3'),
(3, 'ST-A-0003', 5, '汉服文化展示周', 'C', 200000, '2026-04-01 09:00:00', '2026-04-07 17:00:00', '校园广场', 'S2'),
(4, 'ST-A-0004', 6, 'AI 创新工作坊', 'C', 150000, '2026-04-10 13:00:00', '2026-04-10 18:00:00', '实验楼201', 'S2'),
(5, 'ST-A-0005', 7, '社区志愿服务周', 'C', 80000, '2026-04-15 08:00:00', '2026-04-21 18:00:00', '周边社区', 'S2'),
(6, 'ST-A-0006', 10, '校园辩论赛决赛', 'B', 300000, '2026-05-01 19:00:00', '2026-05-01 22:00:00', '大礼堂', 'S1'),
(7, 'ST-A-0007', 19, '数学建模校赛', 'B', 250000, '2026-05-10 08:00:00', '2026-05-12 18:00:00', '教学楼A区', 'S1'),
(8, 'ST-A-0008', 14, '3D打印创客马拉松', 'C', 180000, '2026-05-20 09:00:00', '2026-05-22 18:00:00', '创客空间', 'S1'),
(9, 'ST-A-0009', 20, '合唱团春季音乐会', 'B', 350000, '2026-06-01 19:30:00', '2026-06-01 21:30:00', '大礼堂', 'S0'),
(10, 'ST-A-0010', 17, '书法作品展', 'D', 40000, '2026-06-10 10:00:00', '2026-06-15 17:00:00', '艺术展厅', 'S0');

-- ========== 4. st_recruit_plan: 社团招新计划（10 条） ==========
INSERT OR IGNORE INTO st_recruit_plan (id, assoc_id, title, target_count, accepted_count, status) VALUES
(1, 1, '2026春季算法社招新', 30, 28, 'S2'),
(2, 2, '2026春季英语角招新', 50, 45, 'S2'),
(3, 5, '2026春季汉服社招新', 40, 40, 'S2'),
(4, 6, '2026春季AI社招新', 25, 22, 'S2'),
(5, 7, '2026春季志愿者招新', 100, 95, 'S2'),
(6, 10, '2026春季辩论社招新', 20, 18, 'S2'),
(7, 19, '2026春季数模协会招新', 35, 30, 'S2'),
(8, 14, '2026春季创客社招新', 30, 25, 'S2'),
(9, 20, '2026春季合唱团招新', 40, 38, 'S2'),
(10, 17, '2026春季书法社招新', 25, 20, 'S1');

-- ========== 5. sq_incident: 社区事件（15 条，覆盖各级别） ==========
INSERT OR IGNORE INTO sq_incident (id, biz_no, building_id, level, incident_type, reporter_id, handler_id, status, description, resolution, closed_at) VALUES
(1,  'SQ-EV-0001', 1,  'L1', 'noise',        1, 2, 'resolved', '夜间宿舍噪音扰民', '已与当事人沟通', '2026-03-05 10:00:00'),
(2,  'SQ-EV-0002', 2,  'L1', 'cleanliness',  2, 2, 'resolved', '宿舍卫生不达标',   '已安排打扫',     '2026-03-08 14:00:00'),
(3,  'SQ-EV-0003', 3,  'L2', 'damage',       3, 2, 'resolved', '损坏公共设施',     '已照价赔偿',     '2026-03-10 09:00:00'),
(4,  'SQ-EV-0004', 1,  'L1', 'noise',        4, 2, 'resolved', '午休时间打球',     '已批评教育',     '2026-03-12 13:00:00'),
(5,  'SQ-EV-0005', 5,  'L2', 'safety',       5, 2, 'resolved', '使用违规电器',     '没收并警告',     '2026-03-15 16:00:00'),
(6,  'SQ-EV-0006', 2,  'L1', 'cleanliness',  6, 2, 'handling', '走廊堆放杂物',     NULL,             NULL),
(7,  'SQ-EV-0007', 7,  'L3', 'safety',       7, 2, 'handling', '私拉电线充电',     NULL,             NULL),
(8,  'SQ-EV-0008', 3,  'L1', 'noise',        8, 2, 'handling', '夜间大声喧哗',     NULL,             NULL),
(9,  'SQ-EV-0009', 8,  'L2', 'damage',       9, 2, 'reported', '玻璃门破损',       NULL,             NULL),
(10, 'SQ-EV-0010', 4,  'L1', 'cleanliness', 10, 2, 'reported', '卫生间脏乱',       NULL,             NULL),
(11, 'SQ-EV-0011', 9,  'L3', 'safety',      11, 2, 'reported', '消防通道堵塞',     NULL,             NULL),
(12, 'SQ-EV-0012', 5,  'L1', 'noise',       12, 2, 'reported', '夜间电视音量大',   NULL,             NULL),
(13, 'SQ-EV-0013', 6,  'L4', 'safety',      13, 2, 'reported', '疑似火情报警',     NULL,             NULL),
(14, 'SQ-EV-0014', 10, 'L2', 'damage',      14, 2, 'reported', '空调损坏',         NULL,             NULL),
(15, 'SQ-EV-0015', 7,  'L1', 'cleanliness', 15, 2, 'reported', '阳台积水',         NULL,             NULL);

-- ========== 6. qg_difficulty_cert: 家庭困难认定（15 条） ==========
INSERT OR IGNORE INTO qg_difficulty_cert (id, biz_no, student_id, academic_year, level, cert_status) VALUES
(1,  'QG-DC-0001', 1,  '2025-2026', 'normal',    'S3'),
(2,  'QG-DC-0002', 3,  '2025-2026', 'difficult', 'S3'),
(3,  'QG-DC-0003', 5,  '2025-2026', 'special',   'S3'),
(4,  'QG-DC-0004', 7,  '2025-2026', 'normal',    'S3'),
(5,  'QG-DC-0005', 9,  '2025-2026', 'difficult', 'S3'),
(6,  'QG-DC-0006', 11, '2025-2026', 'special',   'S3'),
(7,  'QG-DC-0007', 13, '2025-2026', 'normal',    'S3'),
(8,  'QG-DC-0008', 15, '2025-2026', 'difficult', 'S2'),
(9,  'QG-DC-0009', 17, '2025-2026', 'normal',    'S2'),
(10, 'QG-DC-0010', 19, '2025-2026', 'special',   'S2'),
(11, 'QG-DC-0011', 2,  '2025-2026', 'normal',    'S1'),
(12, 'QG-DC-0012', 4,  '2025-2026', 'difficult', 'S1'),
(13, 'QG-DC-0013', 6,  '2025-2026', 'normal',    'S1'),
(14, 'QG-DC-0014', 8,  '2025-2026', 'special',   'S0'),
(15, 'QG-DC-0015', 10, '2025-2026', 'normal',    'S0');

-- ========== 7. qg_position_apply: 勤工助学申请（15 条） ==========
-- 注意：此表在 V1.0 schema 中未定义，这里跳过
-- 实际由 QgModuleController 用 qg_position 表查询

-- ========== 8. file_meta: 文件元数据示例（5 条） ==========
INSERT OR IGNORE INTO file_meta (id, file_key, original_name, file_size, content_type, uploaded_by) VALUES
(1, 'upload/2026/03/uuid-001.pdf', '团员申请表-张伟.pdf',    153600, 'application/pdf', 1),
(2, 'upload/2026/03/uuid-002.pdf', '困难认定证明-李娜.pdf',  204800, 'application/pdf', 1),
(3, 'upload/2026/03/uuid-003.jpg', '活动照片-算法竞赛.jpg', 819200, 'image/jpeg',      1),
(4, 'upload/2026/04/uuid-004.docx','活动策划书-辩论赛.docx', 102400, 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 1),
(5, 'upload/2026/04/uuid-005.xlsx','招新统计表.xlsx',        51200,  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', 1);

-- ========== 9. cmp_ai_evaluation: AI 评语示例（5 条） ==========
INSERT OR IGNORE INTO cmp_ai_evaluation (id, student_id, academic_term, ai_summary, ai_suggestions, final_score, status) VALUES
(1, 1, '2025-2026-2', '该生思想品德优秀，积极参与团组织生活；社团表现突出，担任算法社骨干；学业成绩名列前茅。', '建议保持现有学习节奏，可尝试参与学科竞赛拓宽视野。', 98.0, 'published'),
(2, 5, '2025-2026-2', '该生作为中共党员，政治觉悟高；社团管理能力强；学业稳定。', '可加强社会工作实践，提升综合素质。', 93.5, 'published'),
(3, 7, '2025-2026-2', '该生热心公益，志愿服务时长居班级前列；学业略有波动。', '建议加强专业课学习，平衡社团与学业。', 91.5, 'draft'),
(4, 10, '2025-2026-2', '该生艺术素养突出，参与多次文艺演出；社区表现良好。', '可尝试跨学科发展，丰富知识结构。', 88.0, 'draft'),
(5, 19, '2025-2026-2', '该生数学建模能力突出，曾获校级奖项；劳动教育表现积极。', '建议加强英语能力，为后续竞赛做准备。', 79.5, 'draft');

-- ========== 10. sys_menu: 菜单数据（前端权限需要） ==========
INSERT OR IGNORE INTO sys_menu (code, title, icon, path, component, parent_id, sort, roles, visible) VALUES
('dashboard',   '仪表盘',       'Odometer',      '/dashboard',         'Dashboard',                  NULL, 1,  '["R-SY-ADMIN","R-IDX-COUNSELOR","R-TY-SECRETARY","R-ST-TUTOR","R-SQ-TUTOR","R-QG-OFFICER","R-CMP-ADMIN","R-STUDENT"]', 1),
('idx',         '学生身份库',   'User',          '/idx',               'Layout',                     NULL, 2,  '["R-SY-ADMIN","R-IDX-COUNSELOR"]', 1),
('idx-students','学生管理',     'UserFilled',    '/idx/students',      'idx/Students',               2,    1,  '["R-SY-ADMIN","R-IDX-COUNSELOR"]', 1),
('ty',          '团员发展',     'Star',          '/ty',                'Layout',                     NULL, 3,  '["R-SY-ADMIN","R-TY-SECRETARY"]', 1),
('ty-applications','入团申请',  'Document',      '/ty/applications',   'ty/Applications',            4,    1,  '["R-SY-ADMIN","R-TY-SECRETARY"]', 1),
('st',          '社团活动',     'Football',      '/st',                'Layout',                     NULL, 4,  '["R-SY-ADMIN","R-ST-TUTOR"]', 1),
('st-associations','社团管理',  'Flag',          '/st/associations',   'st/Associations',            6,    1,  '["R-SY-ADMIN","R-ST-TUTOR"]', 1),
('sq',          '学生社区',     'HomeFilled',    '/sq',                'Layout',                     NULL, 5,  '["R-SY-ADMIN","R-SQ-TUTOR"]', 1),
('sq-incidents','事件管理',     'Warning',       '/sq/incidents',      'sq/Incidents',               8,    1,  '["R-SY-ADMIN","R-SQ-TUTOR"]', 1),
('qg',          '勤工助学',     'Wallet',        '/qg',                'Layout',                     NULL, 6,  '["R-SY-ADMIN","R-QG-OFFICER"]', 1),
('qg-positions','岗位管理',     'Briefcase',     '/qg/positions',      'qg/Positions',               10,   1,  '["R-SY-ADMIN","R-QG-OFFICER"]', 1),
('cmp',         '综合素质',     'Trophy',        '/cmp',               'Layout',                     NULL, 7,  '["R-SY-ADMIN","R-CMP-ADMIN","R-STUDENT"]', 1),
('cmp-dashboard','素质看板',    'DataAnalysis',  '/cmp/dashboard',     'cmp/Dashboard',              12,   1,  '["R-SY-ADMIN","R-CMP-ADMIN"]', 1),
('cmp-ranking', '成绩排名',     'Rank',          '/cmp/ranking',       'cmp/ScoreRanking',           12,   2,  '["R-SY-ADMIN","R-CMP-ADMIN"]', 1),
('cmp-myscore', '我的成绩',     'Ticket',        '/cmp/myscore',       'cmp/MyScore',                12,   3,  '["R-SY-ADMIN","R-STUDENT"]', 1),
('sys',         '系统管理',     'Setting',       '/sys',               'Layout',                     NULL, 99, '["R-SY-ADMIN"]', 1),
('sys-users',   '用户管理',     'UserFilled',    '/sys/users',         'sys/Users',                  16,   1,  '["R-SY-ADMIN"]', 1),
('sys-roles',   '角色管理',     'Avatar',        '/sys/roles',         'sys/Roles',                  16,   2,  '["R-SY-ADMIN"]', 1),
('sys-dicts',   '字典管理',     'Collection',    '/sys/dicts',         'sys/Dicts',                  16,   3,  '["R-SY-ADMIN"]', 1);

const pptxgen = require('pptxgenjs');
const path = require('path');

const ppt = new pptxgen();

ppt.layout = 'LAYOUT_16x9';
ppt.title = 'StudentHub 项目汇报 PPT';

// Theme colors (Google Stitch Nexus Campus Palette)
const NAVY = '00236F';
const LIGHT_BLUE = 'D5E3FC';
const SLATE = '515F74';
const LIGHT_BG = 'F7F9FB';
const WHITE = 'FFFFFF';
const DARK_TEXT = '191C1E';
const ACCENT_GREEN = '006C4C';
const BORDER_COLOR = 'E0E3E5';

// Helper: Add header banner to slides
function addSlideHeader(slide, titleText, subtitleText) {
  // Background
  slide.background = { color: LIGHT_BG };
  
  // Header text
  slide.addText(titleText, {
    x: 0.6,
    y: 0.4,
    w: 12,
    h: 0.6,
    fontSize: 24,
    bold: true,
    color: NAVY,
    fontFace: 'Microsoft YaHei'
  });

  if (subtitleText) {
    slide.addText(subtitleText, {
      x: 0.6,
      y: 0.95,
      w: 12,
      h: 0.35,
      fontSize: 13,
      color: SLATE,
      fontFace: 'Microsoft YaHei'
    });
  }

  // Top accent bar
  slide.addShape(ppt.ShapeType.rect, {
    x: 0.6,
    y: 0.3,
    w: 0.15,
    h: 0.7,
    fill: { color: NAVY }
  });
}

// -----------------------------------------------------------------------------
// SLIDE 1: 封面 (Cover)
// -----------------------------------------------------------------------------
const slide1 = ppt.addSlide();
slide1.background = { color: LIGHT_BG };

// Card Background
slide1.addShape(ppt.ShapeType.roundRect, {
  x: 0.8, y: 0.8, w: 11.73, h: 5.9,
  fill: { color: WHITE },
  line: { color: BORDER_COLOR, width: 1.5 }
});

// Top Pill Badge
slide1.addShape(ppt.ShapeType.roundRect, {
  x: 1.3, y: 1.3, w: 3.5, h: 0.45,
  fill: { color: LIGHT_BLUE }
});
slide1.addText('Google Stitch Nexus Campus 设计规范', {
  x: 1.3, y: 1.3, w: 3.5, h: 0.45,
  fontSize: 12, bold: true, color: NAVY, align: 'center', fontFace: 'Microsoft YaHei'
});

// Title & Subtitle
slide1.addText('StudentHub · 学生一站式自主管理过程管理系统', {
  x: 1.3, y: 2.0, w: 10.7, h: 1.0,
  fontSize: 28, bold: true, color: NAVY, fontFace: 'Microsoft YaHei'
});

slide1.addText('围绕“学生主体 + 过程档案”，覆盖 TY / ST / SQ / QG / CMP 五大业务模块的校园中台', {
  x: 1.3, y: 3.0, w: 10.7, h: 0.6,
  fontSize: 15, color: SLATE, fontFace: 'Microsoft YaHei'
});

// Horizontal Divider
slide1.addShape(ppt.ShapeType.line, {
  x: 1.3, y: 3.8, w: 10.7, h: 0,
  line: { color: BORDER_COLOR, width: 1 }
});

// Info Grid
slide1.addText([
  { text: '开发周期：', options: { bold: true, color: NAVY } },
  { text: '2026.07.20 – 2026.07.22 (3天击穿垂直迭代)\n\n' },
  { text: '项目团队：', options: { bold: true, color: NAVY } },
  { text: '陈宇晗 (PM / 架构师 / 前后端负责人 / 组长) | 童子涵 (QA & DevOps 测试运维)\n\n' },
  { text: '汇报场景：', options: { bold: true, color: NAVY } },
  { text: '实训成果汇报 / 课程评测 / 综合事务中台演示' }
], {
  x: 1.3, y: 4.1, w: 10.7, h: 2.2,
  fontSize: 14, color: DARK_TEXT, fontFace: 'Microsoft YaHei', lineSpacing: 22
});


// -----------------------------------------------------------------------------
// SLIDE 2: 需求拆解 (Requirements Breakdown)
// -----------------------------------------------------------------------------
const slide2 = ppt.addSlide();
addSlideHeader(slide2, '需求拆解：原始需求 + 细化迭代需求', '明确划分官方核心必做功能与加分拓展创新功能');

// Core Requirements Box
slide2.addShape(ppt.ShapeType.roundRect, {
  x: 0.6, y: 1.5, w: 5.9, h: 5.2,
  fill: { color: WHITE }, line: { color: BORDER_COLOR, width: 1 }
});
slide2.addText('核心必做功能 (Base Requirements)', {
  x: 0.9, y: 1.7, w: 5.3, h: 0.4,
  fontSize: 16, bold: true, color: NAVY, fontFace: 'Microsoft YaHei'
});
slide2.addText([
  { text: '1. 学生档案 (IDX)：', options: { bold: true } },
  { text: '学籍信息库、学号索引、班级组织关联与档案明细。\n' },
  { text: '2. 团员发展 (TY)：', options: { bold: true } },
  { text: '入团申请、推优大会、培养记录、政审管理、团员花名册。\n' },
  { text: '3. 社团活动 (ST)：', options: { bold: true } },
  { text: '社团立项、招新计划、活动审批与全校活动广场。\n' },
  { text: '4. 学生社区 (SQ)：', options: { bold: true } },
  { text: '楼栋寝室网格、巡查打卡评分、宿舍违规异常处理。\n' },
  { text: '5. 勤工助学 (QG)：', options: { bold: true } },
  { text: '贫困生困难认定、岗位招聘发布、工时考勤打卡。\n' },
  { text: '6. 综合评价 (CMP)：', options: { bold: true } },
  { text: '德智体美劳五维得分量化与全校综合分实时排行。' }
], {
  x: 0.9, y: 2.2, w: 5.3, h: 4.3,
  fontSize: 12, color: DARK_TEXT, fontFace: 'Microsoft YaHei', lineSpacing: 18
});

// Innovative Bonus Requirements Box
slide2.addShape(ppt.ShapeType.roundRect, {
  x: 6.8, y: 1.5, w: 5.9, h: 5.2,
  fill: { color: WHITE }, line: { color: ACCENT_GREEN, width: 1.5 }
});
slide2.addText('★ 加分拓展创新功能 (Innovative Highlights)', {
  x: 7.1, y: 1.7, w: 5.3, h: 0.4,
  fontSize: 16, bold: true, color: ACCENT_GREEN, fontFace: 'Microsoft YaHei'
});
slide2.addText([
  { text: '1. Google Stitch 顶尖视觉重构：', options: { bold: true } },
  { text: '引入 Nexus Campus 规范，海蓝 Pill 标签与纯白石墨极简风 UI。\n' },
  { text: '2. 状态机引擎 (State Machine)：', options: { bold: true } },
  { text: '统一逻辑推进 (S0草稿→S1公示→S2审批→S3通过)，带全流程审计。\n' },
  { text: '3. AI 大模型智慧评语 (DeepSeek)：', options: { bold: true } },
  { text: '自动分析学生五维数据，一键生成多维综合素质评价与成长建议。\n' },
  { text: '4. SQLite Flyway 自动化持久化：', options: { bold: true } },
  { text: '100% 数据库 SQL JOIN 关联查询，拒绝后端假数据硬编码。\n' },
  { text: '5. 一键并发启动脚本 (start.bat)：', options: { bold: true } },
  { text: 'Windows CMD 零乱码并行调起 Spring Boot 与 Vite 双端服务。' }
], {
  x: 7.1, y: 2.2, w: 5.3, h: 4.3,
  fontSize: 12, color: DARK_TEXT, fontFace: 'Microsoft YaHei', lineSpacing: 18
});


// -----------------------------------------------------------------------------
// SLIDE 3: 技术栈清单 (Tech Stack)
// -----------------------------------------------------------------------------
const slide3 = ppt.addSlide();
addSlideHeader(slide3, '技术栈清单 (Full Tech Stack Architecture)', '前后端分离 + 高性能 SQLite 数据库 + AI 大模型中间件');

const techGrid = [
  { title: '前端技术栈 (Frontend)', items: '• Vue 3.5 (<script setup>)\n• Vite 5 构建工具\n• Element Plus 2.8+ UI 库\n• Pinia 3 状态管理\n• ECharts 5 数据可视化\n• Axios 透明刷令牌封包' },
  { title: '后端技术栈 (Backend)', items: '• Java 21 / Spring Boot 3.3\n• MyBatis-Plus 3.5 ORM\n• Flyway 数据库自动迁移\n• Sa-Token 权限鉴权框架\n• Logback 结构化 JSON 日志\n• Spring JdbcTemplate 关联' },
  { title: '数据库与存储 (Storage)', items: '• SQLite3 关系型数据库\n• WAL 模式 (journal_mode=WAL)\n• Foreign Keys 外键硬约束\n• 本地文件/MinIO 对象存储\n• AES-256-GCM 敏感字段加密' },
  { title: 'AI 赋能与 SDK (AI Tools)', items: '• DeepSeek Open AI 大模型 API\n• Google Stitch 设计 Token\n• Vitest / JUnit5 单元测试\n• Docker / Nginx 容器部署\n• start.bat UTF-8一键脚本' }
];

techGrid.forEach((box, idx) => {
  const row = Math.floor(idx / 2);
  const col = idx % 2;
  const x = 0.6 + col * 6.1;
  const y = 1.5 + row * 2.7;

  slide3.addShape(ppt.ShapeType.roundRect, {
    x: x, y: y, w: 5.8, h: 2.4,
    fill: { color: WHITE }, line: { color: BORDER_COLOR, width: 1 }
  });

  slide3.addText(box.title, {
    x: x + 0.3, y: y + 0.2, w: 5.2, h: 0.35,
    fontSize: 15, bold: true, color: NAVY, fontFace: 'Microsoft YaHei'
  });

  slide3.addText(box.items, {
    x: x + 0.3, y: y + 0.6, w: 5.2, h: 1.6,
    fontSize: 12, color: DARK_TEXT, fontFace: 'Microsoft YaHei', lineSpacing: 16
  });
});


// -----------------------------------------------------------------------------
// SLIDE 4: 分工说明 (Team Assignment Matrix)
// -----------------------------------------------------------------------------
const slide4 = ppt.addSlide();
addSlideHeader(slide4, '成员分工说明 (Team Assignment Matrix)', '展示《成员分工记录表》，明确团队角色与核心职责');

// Table Data
const tableRows = [
  [
    { text: '团队成员', options: { bold: true, color: WHITE, fill: { color: NAVY } } },
    { text: '负责角色', options: { bold: true, color: WHITE, fill: { color: NAVY } } },
    { text: '负责模块与具体工作内容', options: { bold: true, color: WHITE, fill: { color: NAVY } } },
    { text: '产出成果', options: { bold: true, color: WHITE, fill: { color: NAVY } } }
  ],
  [
    { text: '陈宇晗', options: { bold: true, color: NAVY } },
    { text: 'PM / 架构师\n后端领头人\n前端领头人\n答辩组长' },
    { text: '• 系统整体单体架构设计 (Modular Monolith) 与 PRD/SRD 编写\n• 数据库 SQLite 模式设计及 Flyway V1.0/V1.1 脚本编写\n• Spring Boot 后端 35+ 子页面 API 控制器与 Service 层开发\n• Vue 3 前端全部视图重构与 Google Stitch Nexus UI 适配\n• DeepSeek AI 大模型综合评价引擎接入与 API 联调' },
    { text: '后端全量代码\n前端全量页面\n数据迁移脚本\nDeepSeek 接入' }
  ],
  [
    { text: '童子涵', options: { bold: true, color: NAVY } },
    { text: 'QA 测试工程师\nDevOps 运维工程师' },
    { text: '• 编写 SQLite 数据库单元与集成测试用例 (StudentHubCrudTest.java)\n• 验证 CRUD 增删改查无报错与自动化断言校验\n• 编写根目录 start.bat / run.bat 一键并发启动批处理脚本\n• 负责系统部署压测、日志排错与终测质量把控' },
    { text: '单元测试用例\nstart.bat 启动脚本\n部署与测试报告' }
  ]
];

slide4.addTable(tableRows, {
  x: 0.6, y: 1.5, w: 12.1, h: 5.2,
  colW: [1.5, 2.2, 6.4, 2.0],
  fontSize: 12,
  fontFace: 'Microsoft YaHei',
  border: { pt: 1, color: BORDER_COLOR },
  align: 'left',
  valign: 'middle'
});


// -----------------------------------------------------------------------------
// SLIDE 5: 迭代过程 (Iteration Process)
// -----------------------------------------------------------------------------
const slide5 = ppt.addSlide();
addSlideHeader(slide5, '迭代过程：需求迭代表 (v1.0 → v1.1 → v2.0)', '遵循 Vibe Coding 快速击穿模式，完成完整迭代链路');

const iterRows = [
  [
    { text: '迭代版本', options: { bold: true, color: WHITE, fill: { color: NAVY } } },
    { text: '目标定位', options: { bold: true, color: WHITE, fill: { color: NAVY } } },
    { text: '核心解决问题与新增功能', options: { bold: true, color: WHITE, fill: { color: NAVY } } },
    { text: '闭环验证指标', options: { bold: true, color: WHITE, fill: { color: NAVY } } }
  ],
  [
    { text: 'v1.0 MVP', options: { bold: true } },
    { text: '基础 MVP 架构搭建' },
    { text: '• 完成基于 Spring Boot 3 与 Vue 3 的前后端骨架搭建\n• 建立 SQLite 数据库 DDL 表结构 (sys_user, ty_application 等)\n• 实现 JWT 登录鉴权与 Sa-Token 菜单权限隔离' },
    { text: '通过用户登录\n与菜单动态加载' }
  ],
  [
    { text: 'v1.1 极简UI', options: { bold: true } },
    { text: '设计重构与体验升级' },
    { text: '• 参考 Google Stitch Nexus Campus 设计重构高颜值 UI\n• Flyway 自动化落盘 20 条标准数据到 SQLite 数据库中\n• 对齐前后端 35+ 页面 JSON 字段名，解决字段显示空白问题' },
    { text: '35 个页面全量\n数据正确显示' }
  ],
  [
    { text: 'v2.0 创新加分', options: { bold: true } },
    { text: 'AI 创新与测试部署闭环' },
    { text: '• 接入 DeepSeek LLM 大模型，实现五维综合素质智能评估\n• 编写 StudentHubCrudTest 自动化单元测试 (100% 通过)\n• 根目录提供 start.bat 脚本，实现无冲突一键并发启动' },
    { text: 'JUnit 全绿\nDeepSeek 智能响应' }
  ]
];

slide5.addTable(iterRows, {
  x: 0.6, y: 1.5, w: 12.1, h: 5.2,
  colW: [1.6, 2.3, 6.2, 2.0],
  fontSize: 12,
  fontFace: 'Microsoft YaHei',
  border: { pt: 1, color: BORDER_COLOR },
  align: 'left',
  valign: 'middle'
});


// -----------------------------------------------------------------------------
// SLIDE 6: AI 赋能落地点 (AI Empowerment Scenarios)
// -----------------------------------------------------------------------------
const slide6 = ppt.addSlide();
addSlideHeader(slide6, 'AI 赋能落地点 (AI-Driven Development Scenarios)', '列举 AI 辅助开发的全部 5 大核心落地场景');

const aiScenarios = [
  { title: '1. AI 编写后端 CRUD & 状态机', desc: '利用 AI 快速生成 Spring Boot REST Controller、Service 逻辑及 Mapper 接口，规范实现 statem.Apply() 状态推进与审计日志，减少 80% 重复工作量。' },
  { title: '2. AI 优化 SQL & SQLite 迁移', desc: 'AI 生成复杂的 LEFT JOIN 多表联查 SQL 语句与 SQLite 窗口函数 (RANK() OVER)，并自动构建 Flyway DML 自动化播种脚本 V1.1__seed_20_samples.sql。' },
  { title: '3. AI 生成前端 Stitch 视觉页面', desc: '根据 Google Stitch Nexus Campus 视觉规范，AI 编写 Element Plus 表单、Pinia 状态管理及 ECharts 可视化图表组件，实现高质量界面展示。' },
  { title: '4. AI 自动编写测试用例与文档', desc: 'AI 生成后端 StudentHubCrudTest.java 单元测试用例，并自动编写 AGENTS.md 规范、README.md、截图指引与答辩 PPT 脚本。' },
  { title: '5. AI 接入 DeepSeek 智能评估', desc: '封装 DeepSeek LLM API 接口，实现基于学生“德智体美劳”五维考研数据的个性化评语生成与智能化提升建议。' }
];

aiScenarios.forEach((sc, idx) => {
  const y = 1.5 + idx * 1.05;

  slide6.addShape(ppt.ShapeType.roundRect, {
    x: 0.6, y: y, w: 12.1, h: 0.9,
    fill: { color: WHITE }, line: { color: BORDER_COLOR, width: 1 }
  });

  slide6.addText(sc.title, {
    x: 0.9, y: y + 0.15, w: 3.5, h: 0.6,
    fontSize: 14, bold: true, color: NAVY, fontFace: 'Microsoft YaHei'
  });

  slide6.addText(sc.desc, {
    x: 4.4, y: y + 0.15, w: 8.0, h: 0.6,
    fontSize: 12, color: DARK_TEXT, fontFace: 'Microsoft YaHei'
  });
});


// -----------------------------------------------------------------------------
// SLIDE 7: 系统演示截图 / 流程图 (System Demonstrations & Architecture)
// -----------------------------------------------------------------------------
const slide7 = ppt.addSlide();
addSlideHeader(slide7, '系统架构、业务流程与界面展示', '核心状态机流转、ECharts 驾驶舱大屏与 AI 智能评估');

const demoBoxes = [
  { title: '1. 状态机审批流 (State Engine)', content: '• 状态推进：S0 草稿 → S1 公示 → S2 审批 → S3 通过 → S4 归档\n• 硬隔离校验：状态变更统一走动作端点 (如 /submit, /approve)\n• 审计追踪：记录操作人、时间戳与变前变后状态' },
  { title: '2. ECharts 管理驾驶舱大屏', content: '• 可视化卡片展示全校团员比例、社团活跃排行榜\n• 社区违规趋势分析与勤工助学岗位核算动态图表\n• 基于 ECharts 5 响应式自适应布局' },
  { title: '3. 数据模型与持久化 (SQLite)', content: '• 多表关联：idx_student 核心表关联 5 大模块业务表\n• WAL 模式：高并发读写性能与日志持久化\n• Flyway 迁移：版本化数据库变更控制' },
  { title: '4. AI 智能综合评估模块', content: '• 采集德智体美劳五维综合素质考评得分\n• 实时调用 DeepSeek 大模型生成智能评语与建议\n• 支持人工覆盖与审核确认机制' }
];

demoBoxes.forEach((box, idx) => {
  const row = Math.floor(idx / 2);
  const col = idx % 2;
  const x = 0.6 + col * 6.1;
  const y = 1.5 + row * 2.7;

  slide7.addShape(ppt.ShapeType.roundRect, {
    x: x, y: y, w: 5.8, h: 2.4,
    fill: { color: WHITE }, line: { color: BORDER_COLOR, width: 1 }
  });

  slide7.addText(box.title, {
    x: x + 0.3, y: y + 0.2, w: 5.2, h: 0.35,
    fontSize: 15, bold: true, color: NAVY, fontFace: 'Microsoft YaHei'
  });

  slide7.addText(box.content, {
    x: x + 0.3, y: y + 0.6, w: 5.2, h: 1.6,
    fontSize: 12, color: DARK_TEXT, fontFace: 'Microsoft YaHei', lineSpacing: 16
  });
});


// -----------------------------------------------------------------------------
// SLIDE 8: 项目亮点 (Project Highlights)
// -----------------------------------------------------------------------------
const slide8 = ppt.addSlide();
addSlideHeader(slide8, '项目亮点与加分创新点', '业务价值击穿 + Google Stitch 极简设计 + Vibe Coding 高效范式');

const highlightGrid = [
  { title: '业务价值击穿 (Business Value)', text: '围绕“学生主体 + 过程档案”，彻底改变传统“被动管理”模式。全方位覆盖团员发展、社团活动、社区网格、勤工助学与综合评价五大场景。' },
  { title: 'Google Stitch 视觉交互', text: '遵循 Google Stitch Nexus Campus 设计规范，采用高颜值海蓝 Active Pill 标签、纯白侧边栏与石墨极简卡片，极大地提升用户视觉体验。' },
  { title: '状态机与全流程审计引擎', text: '封装统一的 State Machine 引擎，严格卡控业务状态变更路径，任何关键操作实时记录审计日志，保证校园管理数据的严肃性与追溯性。' },
  { title: 'Vibe Coding 高效敏捷范式', text: '借助 AI 辅助开发，在 3 天内快速击穿 35 个子页面、全模块 20 条标准数据 SQL 关联查询、100% 单元测试全绿及一键并发启动脚本。' }
];

highlightGrid.forEach((hl, idx) => {
  const row = Math.floor(idx / 2);
  const col = idx % 2;
  const x = 0.6 + col * 6.1;
  const y = 1.5 + row * 2.7;

  slide8.addShape(ppt.ShapeType.roundRect, {
    x: x, y: y, w: 5.8, h: 2.4,
    fill: { color: WHITE }, line: { color: ACCENT_GREEN, width: 1.2 }
  });

  slide8.addText(hl.title, {
    x: x + 0.3, y: y + 0.2, w: 5.2, h: 0.35,
    fontSize: 15, bold: true, color: ACCENT_GREEN, fontFace: 'Microsoft YaHei'
  });

  slide8.addText(hl.text, {
    x: x + 0.3, y: y + 0.6, w: 5.2, h: 1.6,
    fontSize: 12, color: DARK_TEXT, fontFace: 'Microsoft YaHei', lineSpacing: 16
  });
});


// -----------------------------------------------------------------------------
// SLIDE 9: 总结与待优化方向 (Summary & Roadmap)
// -----------------------------------------------------------------------------
const slide9 = ppt.addSlide();
addSlideHeader(slide9, '总结与未来待优化方向 (Summary & Roadmap)', '收获与实践总结 + 未来架构演进计划');

// Harvest Box
slide9.addShape(ppt.ShapeType.roundRect, {
  x: 0.6, y: 1.5, w: 5.9, h: 5.2,
  fill: { color: WHITE }, line: { color: BORDER_COLOR, width: 1 }
});
slide9.addText('迭代收获 (Harvest & Takeaways)', {
  x: 0.9, y: 1.7, w: 5.3, h: 0.4,
  fontSize: 16, bold: true, color: NAVY, fontFace: 'Microsoft YaHei'
});
slide9.addText([
  { text: '1. 架构能力提升：', options: { bold: true } },
  { text: '掌握了 Modular Monolith 单体模块化设计与分层架构规范。\n' },
  { text: '2. 数据库高可用：', options: { bold: true } },
  { text: '实践了 SQLite WAL 模式、Flyway DML 自动化迁移与 SQL JOIN 优化。\n' },
  { text: '3. 极简 UI 设计：', options: { bold: true } },
  { text: '深入理解 Google Stitch 设计语言，成功打造现代化 Campus 界面。\n' },
  { text: '4. Vibe Coding 实践：', options: { bold: true } },
  { text: '验证了 AI 在需求拆解、全栈代码编写、测试断言及部署脚本中的巨大赋能价值。' }
], {
  x: 0.9, y: 2.2, w: 5.3, h: 4.3,
  fontSize: 12, color: DARK_TEXT, fontFace: 'Microsoft YaHei', lineSpacing: 18
});

// Roadmap Box
slide9.addShape(ppt.ShapeType.roundRect, {
  x: 6.8, y: 1.5, w: 5.9, h: 5.2,
  fill: { color: WHITE }, line: { color: BORDER_COLOR, width: 1 }
});
slide9.addText('未来待优化方向 (Future Roadmap)', {
  x: 7.1, y: 1.7, w: 5.3, h: 0.4,
  fontSize: 16, bold: true, color: NAVY, fontFace: 'Microsoft YaHei'
});
slide9.addText([
  { text: '1. 微服务架构演进 (v3.0)：', options: { bold: true } },
  { text: '引入 Nacos / Spring Cloud，将 5 大业务模块平滑拆分为微服务。\n' },
  { text: '2. 分布式缓存与 Redis：', options: { bold: true } },
  { text: '引入 Redis 实现分布式 Session 共享与热点综合分排行榜缓存。\n' },
  { text: '3. 多模态 AI 能力扩充：', options: { bold: true } },
  { text: '对接 OCR 识别团课证书、摄像头人脸识别实现活动与打卡考勤签到。\n' },
  { text: '4. 移动端小程序适配：', options: { bold: true } },
  { text: '基于 UniApp 拓展微信小程序端，方便学生随时随地打卡与查分。' }
], {
  x: 7.1, y: 2.2, w: 5.3, h: 4.3,
  fontSize: 12, color: DARK_TEXT, fontFace: 'Microsoft YaHei', lineSpacing: 18
});

// Write Presentation File to Workspace Root
const outputPath = path.join(__dirname, '../StudentHub_Project_Presentation.pptx');
ppt.writeFile({ fileName: outputPath }).then(fileName => {
  console.log(`PPTX successfully generated at: ${fileName}`);
}).catch(err => {
  console.error('Error generating PPTX:', err);
});

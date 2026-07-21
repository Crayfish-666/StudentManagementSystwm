# StudentHub 页面原型与 Figma 全量页面 UI/UX 设计提示词 (Figma All-Pages Prompt)

> **使用说明**：本提示词用于提交给 Figma AI、Galileo AI、Relume 或 UI/UX 设计师。
> **设计原则**：**本提示词不对任何视觉风格、色彩搭配或字体样式做干涉**。本提示词将导航树中所有 35 个子视图页面**一个不漏地逐一画出**，精确标注每一个页面的 **HTML/Div 结构划分、区域布局、包含的信息字段、使用的 UI 组件、指定的图标名称 (Icon Spec) 以及特定的图表类型 (Chart Spec)**！

---

## 目录与 35 个全量页面清单

* [0. 全局通用框架结构 (`DefaultLayout`)](#0-全局通用框架结构-defaultlayout)
* [1. 登录页面 (`/login`)](#1-登录页面-login)
* [2. 工作台 (dashboard)](#2-工作台-dashboard)
  * [2.1 管理驾驶舱 (`/cmp/dashboard`)](#21-管理驾驶舱-cmpdashboard)
  * [2.2 综合分排行 (`/cmp/ranking`)](#22-综合分排行-cmpranking)
* [3. 团员发展 (ty)](#3-团员发展-ty)
  * [3.1 入团申请 (`/ty/application`)](#31-入团申请-tyapplication)
  * [3.2 审批中心 (`/ty/approval`)](#32-审批中心-tyapproval)
  * [3.3 支部推优大会 (`/ty/recommendation-meeting`)](#33-支部推优大会-tyrecommendation-meeting)
  * [3.4 培养记录管理 (`/ty/cultivation`)](#34-培养记录管理-tycultivation)
  * [3.5 发展对象管理 (`/ty/development-object`)](#35-发展对象管理-tydevelopment-object)
  * [3.6 政治审查管理 (`/ty/political-review`)](#36-政治审查管理-typolitical-review)
  * [3.7 接收发展大会 (`/ty/development-meeting`)](#37-接收发展大会-tydevelopment-meeting)
  * [3.8 预备团员转正 (`/ty/probationary`)](#38-预备团员转正-typrobationary)
  * [3.9 团员花名册 (`/ty/member-roster`)](#39-团员花名册-tymember-roster)
* [4. 社团活动 (st)](#4-社团活动-st)
  * [4.1 社团管理 (`/st/association`)](#41-社团管理-stassociation)
  * [4.2 招新计划管理 (`/st/recruit-plan`)](#42-招新计划管理-strecruit-plan)
  * [4.3 招新申请广场 (`/st/recruit-apply`)](#43-招新申请广场-strecruit-apply)
  * [4.4 活动管理与审批 (`/st/activity`)](#44-活动管理与审批-stactivity)
* [5. 学生社区 (sq)](#5-学生社区-sq)
  * [5.1 楼栋与寝室网格 (`/sq/building`)](#51-楼栋与寝室网格-sqbuilding)
  * [5.2 巡查记录大厅 (`/sq/inspection`)](#52-巡查记录大厅-sqinspection)
  * [5.3 异常事件处置 (`/sq/incident`)](#53-异常事件处置-sqincident)
* [6. 勤工助学 (qg)](#6-勤工助学-qg)
  * [6.1 困难认定库 (`/qg/difficulty`)](#61-困难认定库-qgdifficulty)
  * [6.2 岗位管理 (`/qg/position`)](#62-岗位管理-qgposition)
  * [6.3 工时打卡与考勤 (`/qg/attendance`)](#63-工时打卡与考勤-qgattendance)
* [7. 我的申请/个人中心 (mine)](#7-我的申请个人中心-mine)
  * [7.1 我的团员发展 (`/mine/ty-development`)](#71-我的团员发展-minety-development)
  * [7.2 我的入团申请 (`/mine/ty-application`)](#72-我的入团申请-minety-application)
  * [7.3 我的思想汇报 (`/mine/thought-report`)](#73-我的思想汇报-minethought-report)
  * [7.4 我的社团履历 (`/mine/activity`)](#74-我的社团履历-mineactivity)
  * [7.5 我的勤工记录 (`/mine/work`)](#75-我的勤工记录-minework)
  * [7.6 我的综合分 (`/mine/score`)](#76-我的综合分-minescore)
  * [7.7 我的学籍档案 (`/mine/profile`)](#77-我的学籍档案-mineprofile)
* [8. 学生管理 (idx)](#8-学生管理-idx)
  * [8.1 学生列表与履历 (`/idx/student`)](#81-学生列表与履历-idxstudent)
  * [8.2 学生批量导入 (`/idx/import`)](#82-学生批量导入-idximport)
* [9. 系统管理 (sys)](#9-系统管理-sys)
  * [9.1 字典管理 (`/sys/dict`)](#91-字典管理-sysdict)
  * [9.2 用户账号管理 (`/sys/user`)](#92-用户账号管理-sysuser)
  * [9.3 组织机构树 (`/sys/org`)](#93-组织机构树-sysorg)
  * [9.4 定时任务监控 (`/sys/job`)](#94-定时任务监控-sysjob)
* [10. 消息中心 (`/notifications`)](#10-消息中心-notifications)

---

## 0. 全局通用框架结构 (`DefaultLayout`)

所有后台功能视图均嵌套在 `DefaultLayout` 整体容器内，结构划分为三大固定 Div 区域：

* **Div 0.1 Header 顶部栏 (`div.header-top-bar`)**：
  * `div.brand-logo`：图标 `School` / `GraduationCap` + 标题 `StudentHub 学生一站式自主管理系统`。
  * `div.role-switcher`：图标 `SwitchUser` + 下拉选择 `ElSelect`（角色：`普通学生`、`团支部书记`、`社长`、`楼层长`、`辅导员`、`院系团委`、`校团委`、`管理员`）。
  * `div.notification-bell`：图标 `Bell` + `ElBadge` 消息数 + 下拉通知卡片。
  * `div.user-profile`：头像 `ElAvatar` + 用户名与学号 + 下拉菜单（`我的档案 User`、`修改密码 Key`、`退出登录 LogOut`）。
* **Div 0.2 Left Sidebar 左侧导航 (`div.left-sidebar-menu`)**：
  * 级联菜单 `ElMenu`，包含 8 大一级图标（`Odometer`, `Flag`, `Trophy`, `House`, `Briefcase`, `Document`, `User`, `Setting`）与子菜单。
* **Div 0.3 Main Container 主内容区 (`div.main-container`)**：
  * 顶部：面包屑 `ElBreadcrumb` + 页面标题 `div.page-header` + 右侧控制按钮组。
  * 中间：筛选条件区 `div.filter-box` + 数据展示区 `div.data-box`。
  * 底部：分页控制条 `ElPagination`。

---

## 1. 登录页面 (`/login`)

```text
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ div.login-page-wrapper                                                                       │
│ ┌──────────────────────────────────────┬───────────────────────────────────────────────────┐ │
│ │ div.login-left-banner                │ div.login-right-card                              │ │
│ │ ├── 图标插画: GraduationIllustration  │ ├── div.header (标题: 用户登录)                   │ │
│ │ ├── 标题: StudentHub 智慧校园        │ ├── div.input-user (图标: User, 学号/工号)        │ │
│ │ └── Slogan: 过程留痕·规则卡控·综合量化│ ├── div.input-pass (图标: Lock, 密码, 显隐切换)   │ │
│ │                                      │ ├── div.options (记住密码 Checkbox + 忘记密码)    │ │
│ │                                      │ ├── div.btn-submit (登录大按钮, 图标: ArrowRight) │ │
│ │                                      │ └── div.footer (版本与版权信息)                   │ │
│ └──────────────────────────────────────┴───────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. 工作台 (dashboard)

### 2.1 管理驾驶舱 (`/cmp/dashboard`)
* **Div 2.1.1 `div.welcome-card`**：用户头像、问候语 `欢迎回来，张三老师！`、快捷按钮组（`新增申请 Plus`、`发起立项 Trophy`、`巡查打卡 House`）。
* **Div 2.1.2 `div.kpi-cards-grid` (4 大 KPI 卡片)**：
  1. `在读学生总数` (图标 `User`, 数值 `3,421`, 同比 `↑ 5.2%`)
  2. `待我审批事项` (图标 `DocumentChecked`, 红色角标 `12`)
  3. `活跃社团数` (图标 `Trophy`, 数值 `48`, 星级 `4.2 ★`)
  4. `勤工在岗学生` (图标 `Briefcase`, 数值 `256`, 本月总工时 `8,420h`)
* **Div 2.1.3 `div.spring-ai-eval-box` (Spring AI 大模型综测助手)**：
  * 控件：学生选择器 `ElSelect` (搜姓名/学号)、学期选择器。
  * 按钮：`ElButton` (type=success, icon=`Sparkles`/`Cpu`) `一键生成 AI 综测评语初稿`。
  * 输出区：AI 评语文本域 `ElInput` (textarea)、改进建议 `ElTag` 列表、`人工覆写评语` 文本域及 `保存` 按钮 (图标 `Check`)。
* **Div 2.1.4 `div.chart-row-1`**：
  * 左：ECharts **综合素质 5 维雷达图 (Radar Chart)**（维度：团内、社团、社区、勤工、学业）。
  * 右：ECharts **月度参与趋势双折线图 (Line Chart)**（X轴：9-6月，Y轴：人次；图例：活动签到数、勤工打卡数）。
* **Div 2.1.5 `div.chart-row-2`**：
  * 左：ECharts **社团分类占比饼图 (Pie Chart)**（思想政治、学术科技、文化体育、志愿公益）。
  * 右：ECharts **楼栋隐患与晚归柱状图 (Bar Chart)**（X轴：1-10号楼，Y轴：件数；颜色区分 L1-L4）。

### 2.2 综合分排行 (`/cmp/ranking`)
* **Div 2.2.1 `div.filter-bar`**：学期下拉框、院系下拉框、专业下拉框、搜索框 (图标 `Search`)、`查询`、`重置`、`导出 Excel (Download)` 按钮。
* **Div 2.2.2 `div.podium-cards`**：前三名金银铜卡片 (皇冠/奖牌图标 `Medal`)，展示头像、姓名、专业、总分大字。
* **Div 2.2.3 `div.ranking-table`**：`ElTable` 包含排名 (前3高亮)、学生姓名/学号、院系专业、5 维细分得分、综合总分 (`ElProgress` 进度条)、`查看履历全景` 按钮。

---

## 3. 团员发展 (ty)

### 3.1 入团申请 (`/ty/application`)
* **Div 3.1.1 `div.filter-bar`**：状态筛选 (S0-S4)、申请人搜索、`提交新申请 (Plus)` 按钮。
* **Div 3.1.2 `div.application-table`**：列：申请单号、姓名、学号、团支部、申请日期、当前节点 Tag、状态 Tag (草稿/待审/通过/驳回)、操作（查看/编辑/撤回）。
* **Div 3.1.3 `div.application-form-modal` (新增/编辑表单)**：
  * 年龄校验 Card：年龄 14-28 岁 Green Tag / 超龄 Red Tag。
  * 政治思想自述：`ElInput` (textarea, 强制 ≥500字，带有字数统计)。
  * MinIO 附件：手写申请书照片上传组件 (图标 `Upload`，含 `眼睛 View` 预签名在线预览)。

### 3.2 审批中心 (`/ty/approval`)
* **Div 3.2.1 `div.approval-tabs`**：`待我审批 (Badge 数量)` vs `我已审批` 选项卡。
* **Div 3.2.2 `div.approval-cards-grid`**：待办申请卡片列表，展示学生头像、申请编号、自述摘要、当前阶段。
* **Div 3.2.3 `div.approval-dialog` (审批弹窗)**：展现申请全景，选择 `同意 (Pass)` / `驳回 (Reject)`，驳回时强制输入意见（≥30字）。

### 3.3 支部推优大会 (`/ty/recommendation-meeting`)
* **Div 3.3.1 `div.meeting-config`**：支部选择、会议时间选择、会议地点。
* **Div 3.3.2 `div.attendance-hard-control` (到会率硬卡控)**：
  * 输入：应到团员数、实到团员数。
  * 硬卡控 Tag：到会率 `83.3%`（`≥ 66.7%` 显示绿 Tag `满足推优条件`；`< 66.7%` 显示红 Alert `到会率不足，禁止提交`）。
* **Div 3.3.3 `div.candidate-vote-table`**：候选人列表、赞成票 (ElInputNumber)、反对票、弃权票、过半通过 Tag (`赞成票 ≥ 实到 1/2`)。
* **Div 3.3.4 `div.photo-wall`**：MinIO 会场照片上传墙（强制上传 2 张：会场全景 + 投票特写）。

### 3.4 培养记录管理 (`/ty/cultivation`)
* **Div 3.4.1 `div.tutor-binding-card`**：绑定 2 名培养联系人信息（姓名、联系方式、党团身份）。
* **Div 3.4.2 `div.cultivation-log-table`**：按月显示考察得分 (0-100)、评语、联系人签字。
* **Div 3.4.3 `div.thought-report-box`**：思想汇报提交列表、季度选择、正文预览、AI 查重率 Badge（如 `12% 合规`，`>30% 自动打回`）。

### 3.5 发展对象管理 (`/ty/development-object`)
* **Div 3.5.1 `div.qualification-check`**：团课结业证书编号输入、结业成绩 (≥80分卡控)、志愿服务时长 (≥20h)。
* **Div 3.5.2 `div.opinion-collection`**：培养联系人意见文本域、辅导员意见文本域、群众座谈记录 (≥10人参与)。

### 3.6 政治审查管理 (`/ty/political-review`)
* **Div 3.6.1 `div.review-scope-card`**：政审范围选择（本人、父母、配偶）。
* **Div 3.6.2 `div.review-document-box`**：函调盖章公文 MinIO 上传组件、预签名 PDF 在线预览框。
* **Div 3.6.3 `div.conclusion-selector`**：政审结论 Radio（合格 / 基本合格 / 不合格）。

### 3.7 接收发展大会 (`/ty/development-meeting`)
* **Div 3.7.1 `div.prerequisite-check`**：公示 5 个工作日检查 Tag、个人自传 (≥2000字) 上传标记。
* **Div 3.7.2 `div.meeting-vote-box`**：发展大会表决票数录入、表决结果判定、《入团志愿书》生成。

### 3.8 预备团员转正 (`/ty/probationary`)
* **Div 3.8.1 `div.probationary-timer`**：预备期倒计时进度条 (满 1 年解锁转正按钮)。
* **Div 3.8.2 `div.quarterly-inspection-table`**：4 个季度考察表提交状态。
* **Div 3.8.3 `div.transfer-apply-box`**：转正申请书提交与校团委终审盖章按钮。

### 3.9 团员花名册 (`/ty/member-roster`)
* **Div 3.9.1 `div.roster-filter`**：全校团员搜索、团员证号检索、转出/离团筛选。
* **Div 3.9.2 `div.roster-table`**：学号、姓名、所在支部、全国统一团员证号、入团时间、转正时间、团籍状态。

---

## 4. 社团活动 (st)

### 4.1 社团管理 (`/st/association`)
* **Div 4.1.1 `div.assoc-card-grid`**：社团卡片网格，展示 Logo、名称、指导教师、会长、星级 `ElRate` (1-5星)、状态 Tag (筹备/试运行/注册/整顿/注销)。
* **Div 4.1.2 `div.assoc-detail-drawer`**：社团章程预览、历任干部列表、申请换届按钮、申请注销按钮。

### 4.2 招新计划管理 (`/st/recruit-plan`)
* **Div 4.2.1 `div.plan-table`**：招新标题、目标人数、已录取 `ElProgress` 进度条、状态 Tag。
* **Div 4.2.2 `div.finish-action-box`**：**`提前结束招新 (SwitchButton)`** 按钮。
* **Div 4.2.3 `div.finish-modal` (提前结束招新二次确认弹窗)**：
  * 图标：`Warning` 警告。
  * 提示：`提前结束招新操作不可逆！结束之后学生将无法在招新广场投递本计划。`
  * 输入：`结束原因` (textarea, 必填)、`确认结束` 按钮。

### 4.3 招新申请广场 (`/st/recruit-apply`)
* **Div 4.3.1 `div.recruit-square-grid`**：招新海报卡片流（社团简介、需求岗位、面试时间/地点、“立即投递简历”按钮）。
* **Div 4.3.2 `div.apply-modal`**：学生简历预览、个人优势说明、提交投递。

### 4.4 活动管理与审批 (`/st/activity`)
* **Div 4.4.1 `div.activity-form` (立项表单)**：
  * 活动名称、预算金额 (元)、预计人数、活动范围 (院系内/跨院系/跨校)。
  * **动态级别计算 Box**：自动匹配并渲染 **A 级 (红 Badge)** / **B 级 (橙 Badge)** / **C 级 (蓝 Badge)** / **D 级 (绿 Badge)**。
  * **MinIO 上传区**：A/B 级动态提示上传 `* 应急预案 PDF` 与 `* 安全承诺书 PDF`。
* **Div 4.4.2 `div.activity-approval-flow`**：流水线审批链预览（指导教师 ➔ 院系 ➔ 校社联 ➔ 校团委 ➔ 校领导）。
* **Div 4.4.3 `div.activity-checkin-box`**：签到二维码动态刷新框、GPS 定位地图、实时已签到列表（迟到 >15min 标记）。
* **Div 4.4.4 `div.activity-summary-box`**：活动总结文本、照片上传墙 (≥3张)、发票报销表。

---

## 5. 学生社区 (sq)

### 5.1 楼栋与寝室网格 (`/sq/building`)
* **Div 5.1.1 `div.building-left-menu`**：楼栋列表（1号楼-10号楼、男舍/女舍、楼管会指导教师）。
* **Div 5.1.2 `div.floor-tabs`**：楼层选择单选框 (1楼-6楼)。
* **Div 5.1.3 `div.room-grid`**：寝室卡片网格。
  * 元素：寝室号 (如 `302`)、床位进度条 (`4/4`)、寝室长姓名头像；连续 3 次卫生不达标卡片右上角显示红闪 Tag `重点关注寝室`。点击卡片弹出人员与床位调整 Modal。

### 5.2 巡查记录大厅 (`/sq/inspection`)
* **Div 5.2.1 `div.type-tabs`**：巡查类型切换（卫生巡查 / 晚归检查 / 违规电器 / 安全隐患 / 消防通道）。
* **Div 5.2.2 `div.inspection-form`**：寝室选择、卫生评分 Slider (0-100)、扣分项 Checkbox、违规电器拍照取证上传 (MinIO)、晚归人员登记。

### 5.3 异常事件处置 (`/sq/incident`)
* **Div 5.3.1 `div.kanban-board` (L1-L4 看板)**：
  * L1 (常规报修) / L2 (违规/矛盾) / L3 (严重隐患) / **L4 (火警/突发事件)** 4 列。
  * **L4 列特效**：带红框呼吸闪烁与 10min 倒计时，卡片提供指导教师 `一键确认结案` 按钮。

---

## 6. 勤工助学 (qg)

### 6.1 困难认定库 (`/qg/difficulty`)
* **Div 6.1.1 `div.year-select`**：学年选择器、认定等级 Tag (特别困难-紫 / 困难-红 / 一般困难-黄 / 不困难-灰)。
* **Div 6.1.2 `div.cert-table`**：困难学生列表、证明材料 MinIO 预览按钮、班级评议得分。
* **门禁提示 Modal**：非困难生投递岗位时弹出红字 Alert `须先通过家庭经济困难认定方可申请勤工岗位！`。

### 6.2 岗位管理 (`/qg/position`)
* **Div 6.2.1 `div.position-grid`**：岗位卡片（岗位名称、用人部门、时薪、每周工时上限 `≤20h` 卡控 Tag、需求/已录用人数）。
* **Div 6.2.2 `div.apply-btn`**：`申请上岗` 按钮（自动校验同时在岗数 ≤1）。

### 6.3 工时打卡与考勤 (`/qg/attendance`)
* **Div 6.3.1 `div.clock-box`**：中央数字时钟 `14:30:25`、GPS 定位 Tag、`上班打卡 (Pointer)` / `下班打卡` 巨型按钮。
* **Div 6.3.2 `div.monthly-limit-box` (月 40h 硬卡控)**：
  * 进度条 `38.5 / 40h`。
  * **当满 `40.0h` 时，进度条变红，打卡大按钮强制置灰禁用 (`disabled`)，下方显示警告 `根据规定，单月工时上限为 40 小时，已达封顶，不可继续打卡！`**。
* **Div 6.3.3 `div.makeup-modal`**：双签补卡申请弹窗（辅导员 + 用人部门双签状态）。

---

## 7. 我的申请/个人中心 (mine)

### 7.1 我的团员发展 (`/mine/ty-development`)
* 展示学生本人入团发展 7 步轨迹时间轴、当前节点进度、下阶段任务提示。

### 7.2 我的入团申请 (`/mine/ty-application`)
* 查看/修改本人入团申请书、查看审批进度与各级审批意见。

### 7.3 我的思想汇报 (`/mine/thought-report`)
* 提交本季度思想汇报正文，实时查看 AI 查重率结果与联系人评语。

### 7.4 我的社团履历 (`/mine/activity`)
* 展示已加入社团列表、担任职务、参与的活动签到历史与获奖记录。

### 7.5 我的勤工记录 (`/mine/work`)
* 查看当前勤工岗位、当月累计工时进度条、打卡历史与历月薪酬明细。

### 7.6 我的综合分 (`/mine/score`)
* 个人综合素质 5 维雷达图、全校/院系排名、各维度得分明细、**查看 Spring AI 生成的个人综测评语**。

### 7.7 我的学籍档案 (`/mine/profile`)
* 展示个人基础学籍信息、政治面貌、脱敏身份证/手机号、紧急联系人与宿舍床位。

---

## 8. 学生管理 (idx)

### 8.1 学生列表与履历 (`/idx/student`)
* **筛选栏**：院系/专业/班级下拉框、姓名学号搜索框。
* **表格**：学号、姓名、脱敏身份证 (`110101********0012`)、脱敏手机号 (`138****5678`)、政治面貌 Tag、困难等级 Tag；操作栏包含 `查看全景履历` 弹窗按钮（触发敏数据解密审计）。

### 8.2 学生批量导入 (`/idx/import`)
* **拖拽上传区**：MinIO/Excel 文件拖拽框 `ElUpload`。
* **解析预览表**：导入数据实时校验列表（重复学号标红高亮），提供 `下载导入异常报告` 按钮。

---

## 9. 系统管理 (sys)

### 9.1 字典管理 (`/sys/dict`)
* 左侧：字典分类树（性别、政治面貌、活动级别、困难等级、巡查类型等）。
* 右侧：字典项 Key-Value 配置表格、排序、启用/禁用开关。

### 9.2 用户账号管理 (`/sys/user`)
* 用户账号表格、重置密码弹窗、绑定学生/教师 ID、Sa-Token 角色分配穿梭框 (`ElTransfer`)。

### 9.3 组织机构树 (`/sys/org`)
* 树形面板编辑院系 (`sys_college`)、专业 (`sys_major`)、班级 (`idx_class`) 及分配辅导员。

### 9.4 定时任务监控 (`/sys/job`)
* Cron 任务列表（工时自动核算、评优计算、日志归档）、执行状态 Tag、`立即执行一次` 按钮、运行日志抽屉 (`ElDrawer`)。

---

## 10. 消息中心 (`/notifications`)

### 10.1 消息中心主页 (`/notifications`)
* **分类 Tab**：`全部` / `审批提醒` / `告警通知` / `考勤通知` / `系统消息`。
* **消息列表**：图标（审批 `DocumentChecked` / 告警 `Warning` / 考勤 `Clock`）、标题、发送时间、摘要内容、`一键已读` 按钮、`一键跳转关联单据` 按钮。

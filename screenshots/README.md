# StudentHub 产品演示截图

> 自动生成时间：2026/6/25 01:41:57
> 截图方式：Playwright + Chromium 真实运行前端 Vite (`:5173`) 与后端 (`:8080`)
> 视口规格：1440 × 900 @ 1x
> 角色账号：`admin / admin@123`（管理员）、`20231001 / student@123`（学生）

## 目录结构

```
screenshots/
├── admin-*.png            # 管理员视角（29 张）
├── student-*.png          # 学生视角（9 张）
├── README.md              # 本文件（总索引）
├── README-admin.md        # 管理员分索引
└── README-student.md      # 学生分索引
```

## 一、管理员视角（29 张）

| # | 文件 | 路径 | 描述 |
| - | ---- | ---- | ---- |
| 01 | `admin-00-login.png` | `/login` | 登录页 |
| 02 | `admin-01-dashboard.png` | `/dashboard` | 管理驾驶舱 Dashboard |
| 03 | `admin-02-ty-application-list.png` | `/ty/application` | 入团申请列表 |
| 04 | `admin-03-ty-approval-center.png` | `/ty/approval` | 审批中心（待我审批） |
| 05 | `admin-04-ty-recommendation-meeting.png` | `/ty/recommendation-meeting` | 推优大会 |
| 06 | `admin-05-ty-cultivation.png` | `/ty/cultivation` | 培养记录管理 |
| 07 | `admin-06-ty-development-object.png` | `/ty/development-object` | 发展对象管理 |
| 08 | `admin-07-ty-political-review.png` | `/ty/political-review` | 政审管理 |
| 09 | `admin-08-ty-development-meeting.png` | `/ty/development-meeting` | 发展大会 |
| 10 | `admin-09-ty-probationary.png` | `/ty/probationary` | 转正流程 |
| 11 | `admin-10-ty-member-roster.png` | `/ty/member-roster` | 团员花名册 |
| 12 | `admin-11-st-association.png` | `/st/association` | 社团管理 |
| 13 | `admin-12-st-recruit-plan.png` | `/st/recruit-plan` | 招新计划 |
| 14 | `admin-13-st-activity.png` | `/st/activity` | 活动管理 |
| 15 | `admin-14-sq-building.png` | `/sq/building` | 楼栋管理 |
| 16 | `admin-15-sq-inspection.png` | `/sq/inspection` | 巡查记录 |
| 17 | `admin-16-sq-incident.png` | `/sq/incident` | 异常事件 |
| 18 | `admin-17-qg-difficulty.png` | `/qg/difficulty` | 困难认定 |
| 19 | `admin-18-qg-position.png` | `/qg/position` | 岗位管理 |
| 20 | `admin-19-qg-attendance.png` | `/qg/attendance` | 工时打卡 |
| 21 | `admin-20-cmp-dashboard.png` | `/cmp/dashboard` | 综合看板（驾驶舱） |
| 22 | `admin-21-cmp-ranking.png` | `/cmp/ranking` | 综合分排行 |
| 23 | `admin-22-idx-student.png` | `/idx/student` | 学生列表 |
| 24 | `admin-23-idx-import.png` | `/idx/import` | 学生导入 |
| 25 | `admin-24-sys-dict.png` | `/sys/dict` | 字典管理 |
| 26 | `admin-25-sys-user.png` | `/sys/user` | 用户管理 |
| 27 | `admin-26-sys-org.png` | `/sys/org` | 组织管理 |
| 28 | `admin-27-sys-job.png` | `/sys/job` | 任务监控 |
| 29 | `admin-28-notifications.png` | `/notifications` | 通知中心 |

## 二、学生视角（9 张）

| # | 文件 | 路径 | 描述 |
| - | ---- | ---- | ---- |
| 01 | `student-00-login.png` | `/login` | 登录页 |
| 02 | `student-01-dashboard.png` | `/dashboard` | 学生 Dashboard |
| 03 | `student-02-mine-ty-development.png` | `/mine/ty-development` | 我的团员发展 |
| 04 | `student-03-mine-ty-application.png` | `/mine/ty-application` | 我的入团申请 |
| 05 | `student-04-mine-thought-report.png` | `/mine/thought-report` | 我的思想汇报 |
| 06 | `student-05-mine-activity.png` | `/mine/activity` | 我的社团 |
| 07 | `student-06-mine-work.png` | `/mine/work` | 我的勤工 |
| 08 | `student-07-mine-score.png` | `/mine/score` | 我的综合分（雷达图） |
| 09 | `student-08-mine-profile.png` | `/mine/profile` | 我的档案 |

## 三、模块覆盖一览

| 模块 | 管理员截图 | 学生截图 | 合计 |
| ---- | ---- | ---- | ---- |
| 登录 + Dashboard | 2 | 2 | 4 |
| TY 团员发展 | 9 | 1（我的团员发展） | 10 |
| ST 社团活动 | 3 | 1（我的社团） | 4 |
| SQ 学生社区 | 3 | 0 | 3 |
| QG 勤工助学 | 3 | 1（我的勤工） | 4 |
| CMP 综合看板 | 2 | 1（我的综合分） | 3 |
| IDX 学生管理 | 2 | 0 | 2 |
| SYS 系统管理 | 4 | 0 | 4 |
| 通知中心 | 1 | 0 | 1 |
| **合计** | **29** | **9** | **38** |

## 四、复现方式

```bash
# 1. 启动后端（默认 :8080）
cd backend && go run ./cmd/server

# 2. 启动前端（默认 :5173）
cd frontend && pnpm dev

# 3. 截图（依赖：screenshot-tool/playwright）
cd screenshot-tool
node capture.mjs admin     # 管理员视角
node capture.mjs student   # 学生视角
```
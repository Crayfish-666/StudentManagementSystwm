# StudentHub · 学生“一站式”自主管理过程管理系统 (Spring Boot 重构版)

> 一个面向高校“第二课堂”与学生事务管理的统一管理平台，围绕 **学生主体 + 过程档案 + 时间戳** 沉淀数据，覆盖 **团员发展、社团活动、学生社区与自治队伍、勤工助学** 四大核心模块，最终结合大模型 (LLM) 形成可量化的综合素质档案与 AI 综测初稿。

![backend](https://img.shields.io/badge/backend-Spring%20Boot%203.2-green)
![frontend](https://img.shields.io/badge/frontend-Vue%203.5%20%2B%20Vite5-42b883)
![db](https://img.shields.io/badge/db-SQLite3%20WAL-003B57)
![auth](https://img.shields.io/badge/auth-Sa--Token%20JWT-red)
![storage](https://img.shields.io/badge/storage-MinIO-orange)
![ai](https://img.shields.io/badge/ai-Spring%20AI%20%2F%20DeepSeek-blue)

---

## 核心设计与规范文档导航（“宪法级” SSOT）

所有架构设计、设计提示词与规范文档均位于项目根目录：

| 文档名称 | 路径 | 核心内容说明 |
| :--- | :--- | :--- |
| **PRD 产品需求文档** | [PRD.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/PRD.md) | 包含 US-001 ~ US-017 用户故事、硬卡控规则、五态状态机、MinIO 分片上传与 LLM AI 综测要求 |
| **ADR 架构决策文档** | [ADR.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/ADR.md) | 包含分层架构图、ADR-001 ~ ADR-009 架构决策（SpringBoot3/SQLite WAL/Sa-Token/Flyway/MinIO/Spring AI） |
| **数据库设计规范** | [database_design_spec.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/database_design_spec.md) | 包含 Mermaid ER 图、SQLite 全量 DDL 脚本、唯一索引、CHECK 约束与 Flyway 初始化迁移脚本 |
| **API 规范文档 (SRD)** | [SRD_api_specifications.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/SRD_api_specifications.md) | RESTful API 契约、Sa-Token 鉴权、统一 Response 封包、错误码表、MinIO & AI 端点与 JSON 样例 |
| **Figma UI 设计提示词** | [figma_prompt.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/figma_prompt.md) | **Figma UI/UX 原型设计提示词**（包含全页面清单、布局结构、元素及跳转关系，无样式干涉） |
| **业务分析报告** | [Analyze.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/Analyze.md) | 10 大维度纯业务分析报告（项目定位、角色、流程、数据模型、页面数据流等） |

---

## 答辩与答辩材料清单

| 材料名称 | 路径 | 适用答辩环节 |
| :--- | :--- | :--- |
| **《成员分工记录表》** | [member_work_division.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/member_work_division.md) | **陈宇晗**（产品/架构/前后端/组长），**童子涵**（测试运维） |
| **《需求迭代表 (v1.0/v1.1/v2.0)》** | [iteration_records.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/iteration_records.md) | 环节 2：三轮快速迭代脉络说明 |
| **《AI 编程赋能落地说明表》** | [ai_empowerment_log.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/ai_empowerment_log.md) | 环节 3：5 大 AI 辅助开发落地场景 |
| **《答辩演示与 PPT 9 页指南》** | [presentation_guide.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/presentation_guide.md) | 全程演练、PPT 制作与 5 分钟实操演练话术 |

---

## 导航栏分支结构与功能全景

结合样例项目（[Neurbyte-ZQJ/StudentHub](https://github.com/Neurbyte-ZQJ/StudentHub) / 线上部署 `http://47.99.245.109/`）对齐的 **9 大顶级导航 + 35 个子页面** 结构：

```text
StudentHub 导航树
├── 📊 1. 工作台 (dashboard) ➔ 管理驾驶舱 / 综合分排行
├── 🚩 2. 团员发展 (ty) ➔ 入团申请 / 审批中心 / 推优大会 / 培养记录 / 发展对象 / 政审 / 接收大会 / 转正 / 团员花名册
├── 🎪 3. 社团活动 (st) ➔ 社团管理 / 招新计划 / 招新广场 / 活动管理与分级审批
├── 🏢 4. 学生社区 (sq) ➔ 楼栋寝室网格 / 巡查记录 / 异常事件处置
├── 🛠️ 5. 勤工助学 (qg) ➔ 困难认定库 / 岗位管理 / 工时打卡与考勤
├── 👤 6. 我的申请/个人中心 (mine) ➔ 我的团员发展 / 我的申请 / 思想汇报 / 社团履历 / 勤工记录 / 综合分 / 学籍档案
├── 🎓 7. 学生管理 (idx) ➔ 学生列表与履历 / 学生批量导入
├── ⚙️ 8. 系统管理 (sys) ➔ 字典管理 / 用户账号 / 组织机构树 / 任务监控
└── 🔔 9. 消息中心 (noti) ➔ 通知列表与消息推送
```

### 样例项目功能 vs 重构版增强功能对比

| 导航模块 | 样例项目实现的功能 (Go/Gin) | 我们的重构项目规划的功能 (Spring Boot 3 + Vue 3 + MinIO + Spring AI) |
| :--- | :--- | :--- |
| **工作台** | 快捷菜单、待办统计、基础 ECharts 统计图表、综合分排行。 | **智能驾驶舱**：集成 **Spring AI / DeepSeek API** 自动生成学生综测评语初稿，支持辅导员人工覆写保存。 |
| **团员发展** | 7 步发展流程、思想汇报、推优表决、政审与团员花名册。 | **刚性卡控与查重**：推优“实到≥2/3、赞成≥1/2”系统卡控；思想汇报文本查重（>30%退回）；政审材料 MinIO 归档。 |
| **社团活动** | 社团生命周期、招新计划、提前结束招新、A/B/C/D 分级审批。 | **对象存储与签到**：招新提前结束动作（`:finish`）；A/B 级活动预案 MinIO 分片上传与预签名 URL 在线预览。 |
| **学生社区** | 楼栋网格图、宿舍巡查、晚归/违规电器登记、L1~L4 异常处置。 | **网格管理与分级响应**：图形化床位映射，L4 级（火警/突发）10min 启动应急强提醒，指导教师线上结案闭环。 |
| **勤工助学** | 困难认定、岗位发布、考勤打卡、月度算薪。 | **工时刚性卡控**：非困难生投递拦截；**每月累计满 40h 刚性卡控阻断打卡**；自动算薪导出脱敏表格。 |
| **文件与系统** | 本地文件存储、基础字典、用户授权与角色。 | **MinIO & Sa-Token 中台**：S3 协议分片上传与时效预览；Sa-Token 双令牌无感刷新与 ABAC 数据隔离。 |

---

## 技术栈与 AI 配置说明

* **后端框架**：Java 17 + Spring Boot 3.2.x + MyBatis-Plus 3.5.x + Sa-Token 1.37+ (JWT 模式) + Flyway 10.x
* **数据库引擎**：SQLite 3 (开启 WAL 写前日志、强外键约束、5000ms 锁超时)
* **前端框架**：Vue 3.5 (Composition API `<script setup>`) + Vite 5 + Element Plus 2.8+ + Pinia 3 + Axios + ECharts 5
* **对象存储**：MinIO Java SDK 8.5.x（分片上传、预签名临时 URL 在线预览、ZIP 打包下载）
* **AI 大模型**：Spring AI + DeepSeek API（支持通过环境变量 `DEEPSEEK_API_KEY` 注入）

---

## 快速开始

### 1. 后端配置与启动 (Spring Boot)
在 `src/main/resources/application.yml` 中配置 DeepSeek API Key 与 MinIO 参数：
```yaml
spring:
  ai:
    openai:
      api-key: ${DEEPSEEK_API_KEY:your_deepseek_api_key_here}
      base-url: https://api.deepseek.com
```

启动命令：
```bash
# 源码运行 (默认监听 8080)
mvn spring-boot:run

# 打包运行
mvn clean package -DskipTests
java -jar target/studenthub-backend-3.2.0.jar
```

### 2. 前端启动 (Vue 3 + Vite)
```bash
cd frontend
pnpm install
pnpm dev
```
访问地址：`http://127.0.0.1:5173`。

---

## Git 仓库同步

本项目已全量提交并同步至 GitHub 仓库：
`https://github.com/Crayfish-666/StudentManagementSystwm.git`

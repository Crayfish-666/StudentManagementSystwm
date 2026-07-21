# StudentHub 项目组 · 成员分工记录表

| 开发周期 | 2026.7.20 – 2026.7.22 | 团队名称 | Vibe Coding AI 重构组 |
| :--- | :--- | :--- | :--- |

---

## 1. 团队角色与模块分工清单

| 角色 | 负责人 | 核心职责描述 | 负责的业务模块 / 交付物 |
| :--- | :--- | :--- | :--- |
| **产品经理 (PM) & 组长** | **陈宇晗** | 原始需求拆解、PRD 文档撰写、迭代路线图规划、答辩 PPT 制作与演练话术准备 | 需求分析、PRD.md、PPT 制作、环节 1 演示 |
| **后端架构师 (BE Lead)** | **陈宇晗** | Spring Boot 3 骨架搭建、Sa-Token 鉴权、SQLite WAL 配置、Flyway 迁移、MinIO & Spring AI 集成 | 后端主工程、ADR.md、SRD_api_specifications.md、MinIO & LLM API |
| **后端开发工程师 (BE)** | **陈宇晗** | 业务逻辑编排、MyBatis-Plus CRUD 生成、刚性卡控规则实现、定时任务与状态机 | TY/ST/SQ/QG 模块 Service & Mapper 实现、后端控制层 |
| **前端开发工程师 (FE)** | **陈宇晗** | Vue 3.5 + Vite 5 + Element Plus 脚手架搭建、Pinia Store 管理、Axios 拦截与无感刷新、ECharts 看板 | 前端 Views/Components、导航栏菜单、大屏可视化、MinIO 切片上传 UI |
| **测试与运维 (QA & DevOps)** | **童子涵** | 测试用例编写、Postman 接口测试、数据库自检、Docker 容器化部署与打包 | database_design_spec.md、测试 Bug 记录表、Dockerfile / Compose |

---

## 2. 协作规范与提交记录
* 成员：**陈宇晗**（产品 / 架构 / 后端 / 前端 / 组长），**童子涵**（测试 / 运维）。
* 采用 Conventional Commits 提交规范（如 `feat(...)`, `fix(...)`, `docs(...)`）。
* 代码与文档统一托管于 GitHub 仓库：`https://github.com/Crayfish-666/StudentManagementSystwm.git`。

# StudentHub · AI 助手协作指南 (AGENTS.md)

> 本文件是给所有 AI 编码助手（Trae / Cursor / Claude Code / Antigravity / Aider / Continue 等）的**统一入口配置**。
> 它的目的不是重复项目文档，而是**告诉 AI 助手：项目是什么、文档在哪里、要遵守什么规矩、禁止做什么**。
> 具体细节以 `docs/` 下对应的"宪法级"文档为唯一事实来源（SSOT）。

---

## 0. 项目一句话

**学生"一站式"自主管理过程管理系统 (Spring Boot 3 重构版)** —— 围绕"学生主体 + 过程档案"，覆盖 **团员发展 (TY) / 社团活动 (ST) / 学生社区 (SQ) / 勤工助学 (QG) / 综合素质量化 (CMP)** 五大业务模块的校园管理后台。

---

## 1. 必读文档清单（SSOT，缺一不可）

接到任何任务前，**先定位到对应章节**再动手；禁止凭直觉编写。

| 文档 | 路径 | 何时读 |
| ---- | ---- | ------ |
| **PRD** 产品需求 | [`docs/01_PRD.md`](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/docs/01_PRD.md) | 涉及业务规则、状态流转、角色权限时 |
| **ADR** 架构决策 | [`docs/02_ADR.md`](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/docs/02_ADR.md) | 涉及技术选型、目录结构、API 风格、错误处理时 |
| **数据库设计** | [`docs/03_database_design_spec.md`](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/docs/03_database_design_spec.md) | 写 DDL / MyBatis-Plus Mapper / Flyway 迁移 / 索引时 |
| **API 规范 (SRD)** | [`docs/04_SRD_api_specifications.md`](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/docs/04_SRD_api_specifications.md) | 写后端 Controller / 前端 api/* / 联调时 |
| **迭代路线图** | [`docs/05_superpowers_iteration_plan.md`](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/docs/05_superpowers_iteration_plan.md) | 决定"现在做哪个切片"时 |

---

## 2. 技术栈速查

### 2.1 后端
- **语言 / SDK**：Java 21 / OpenJDK 21
- **核心框架**：Spring Boot 3.3.14
- **ORM / 持久层**：MyBatis-Plus 3.5.7 + Spring JDBC (`JdbcTemplate`)
- **数据库**：SQLite3（**WAL 模式**：`PRAGMA journal_mode=WAL; foreign_keys=ON; busy_timeout=5000;`）
- **数据库迁移**：Flyway 10.x（`V1.0` ~ `V1.4` 自动化脚本管控）
- **鉴权 & 权限**：Sa-Token 1.37+ (JWT 模式，`Authorization: Bearer {token}`)
- **密码加密**：BCrypt 算法 (Sa-Token 自带 / Spring Security Crypto)
- **字段加密**：AES-256-GCM (字段级敏感信息加密，密钥 `APP_DATA_KEY`)
- **对象存储**：MinIO Java SDK 8.5.x (S3 协议分片上传与预览)
- **AI 智能评语**：Spring AI + DeepSeek API（自动生成综合素质评价）

### 2.2 前端
- **框架**：Vue 3.5 + Composition API (`<script setup>`) + Vite 5
- **视觉设计系统**：Google Stitch Tokens + Nexus Campus Design System
- **状态管理**：Pinia 3
- **路由管理**：Vue Router 4 (支持静态与根据 Sa-Token 动态鉴权)
- **UI 组件库**：Element Plus 2.8+
- **HTTP 客户端**：Axios (统一封装在 `frontend/src/api/http.js`)
- **图表展示**：ECharts 5 (带 resize 监听与销毁回收)
- **时间处理**：Day.js
- **构建工具**：Vite 5 (前端单页打包至 dist)

### 2.3 运行与部署
- 单体架构 (Modular Monolith) + 单文件 SQLite3 数据库
- 一键批处理脚本：根目录 `start.bat` 并行启动后端 8088 与前端 5173

---

## 3. 仓库目录结构

```
StudentManagementSystemVD/
├── start.bat                   # 一键启动脚本（并行拉起后端:8088与前端:5173）
├── StudentHub_Project_Presentation.pptx # 标准 9 页答辩演示文稿
├── AUDIT_FIX_PLAN.md           # 代码审查与 P0 修复记录
├── backend/                    # Java Backend (Maven 工程)
│   ├── pom.xml                 # Maven 依赖配置文件
│   ├── data/                   # SQLite 数据库目录 (studenthub.db)
│   └── src/
│       ├── main/
│       │   ├── java/com/studenthub/
│       │   │   ├── StudentHubApplication.java # 主程序入口
│       │   │   ├── common/     # 统一响应封装 (R.java)
│       │   │   ├── config/     # Spring 配置 (CorsConfig, FlywayConfig, WebMvcConfig)
│       │   │   └── modules/    # ★ 业务模块 (API / Entity / Service / Mapper)
│       │   │       ├── auth/   # 认证与 Token 刷新
│       │   │       ├── idx/    # 学生身份库
│       │   │       ├── ty/     # 团员发展 (全流程)
│       │   │       ├── st/     # 社团活动 & 招新广场
│       │   │       ├── sq/     # 学生社区 & 楼栋网络
│       │   │       ├── qg/     # 勤工助学 & 工时打卡
│       │   │       ├── cmp/    # 综合素质量化 & DeepSeek AI
│       │   │       └── sys/    # 系统管理 (用户/角色/字典)
│       │   └── resources/
│       │       ├── application.yml
│       │       └── db/migration/ # Flyway 迁移脚本 (V1.0~V1.4)
│       └── test/java/com/studenthub/
│           └── StudentHubCrudTest.java # 7 大模块 CRUD 集成单元测试
├── frontend/                   # Vue3 前端
│   ├── src/
│   │   ├── api/                # 模块 API (auth.js, ty.js, st.js, sq.js, qg.js, idx.js, sys.js)
│   │   ├── components/         # 基础组件
│   │   ├── layouts/            # DefaultLayout.vue (含 Keep-Alive 视图容器)
│   │   ├── router/             # 路由配置 (含 RBAC 权限守卫与 Meta 字段)
│   │   ├── stores/             # Pinia store (auth.js, menu.js, dict.js)
│   │   ├── utils/              # 通用工具
│   │   └── views/              # 业务页面 (按模块子目录组织)
│   ├── vite.config.js
│   └── package.json
└── docs/                       # SSOT 架构与需求规范文档
```

---

## 4. API 约定（速查）

### 4.1 URL 契约
- **基础地址**：`/api/v1`
- **模块前缀**：`/auth /idx /ty /st /sq /qg /cmp /sys`
- **响应格式**：
```json
{
  "code": 0,
  "message": "ok",
  "data": { "items": [...], "total": 42, "page": 1, "page_size": 20 }
}
```
- **鉴权 Header**：`Authorization: Bearer {token}`

---

## 5. 编码与安全硬规则（必须遵守）

1. ❌ **禁止**硬编码生产密钥，所有敏感配置（`APP_DATA_KEY`, `MINIO_ACCESS_KEY` 等）通过环境变量注入。
2. ❌ **禁止**直接 `v-html` 渲染未经过滤的 HTML 内容。
3. ❌ **禁止**在前端静态代码中暴露真实账号密码（仅允许在开发环境 `import.meta.env.DEV` 下展示辅助按纽）。
4. ❌ **禁止**绕过权限控制（导员及系统管理员角色需按规则在路由守卫中匹配 `permission`）。
5. ✅ 修改后端或数据库后，必须通过 `mvn test` 验证单元测试通过！
6. ✅ 修改前端代码后，必须通过 `npm run build` 验证打包通过！

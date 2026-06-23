# StudentHub · AI 助手协作指南 (AGENTS.md)

> 本文件是给所有 AI 编码助手（Trae / Cursor / Claude Code / Aider / Continue 等）的**统一入口配置**。
> 它的目的不是重复项目文档，而是**告诉 AI 助手：项目是什么、文档在哪里、要遵守什么规矩、禁止做什么**。
> 具体细节以 `docs/` 下对应的"宪法级"文档为唯一事实来源（SSOT）。

---

## 0. 项目一句话

**学生"一站式"自主管理过程管理系统** —— 围绕"学生主体 + 过程档案"，覆盖 **团员发展 (TY) / 社团活动 (ST) / 学生社区 (SQ) / 勤工助学 (QG) / 综合素质量化 (CMP)** 五大业务模块的校园管理后台。

---

## 1. 必读文档清单（SSOT，缺一不可）

接到任何任务前，**先定位到对应章节**再动手；禁止凭直觉编写。

| 文档 | 路径 | 何时读 |
| ---- | ---- | ------ |
| **PRD** 产品需求 | [`docs/01_PRD.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/01_PRD.md) | 涉及业务规则、状态流转、角色权限时 |
| **ADR** 架构决策 | [`docs/02_ADR.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/02_ADR.md) | 涉及技术选型、目录结构、API 风格、错误处理时 |
| **数据库设计** | [`docs/03_database_design_spec.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/03_database_design_spec.md) | 写 GORM Model / 迁移 / 索引时 |
| **API 规范 (SRD)** | [`docs/04_SRD_api_specifications.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/04_SRD_api_specifications.md) | 写后端 Handler / 前端 api/* / 联调时 |
| **迭代路线图** | [`docs/05_superpowers_iteration_plan.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/05_superpowers_iteration_plan.md) | 决定"现在做哪个切片"时（**S01~S12 步长**） |
| **项目铁律** | [`.trae/rules/project_rules.md`](file:///d:/Teach/AI_Coding/StudentHub/.trae/rules/project_rules.md) | 始终遵循，与本文件同效 |

---

## 2. 技术栈速查

### 2.1 后端
- **语言 / 版本**：Go 1.25+（[`backend/go.mod`](file:///d:/Teach/AI_Coding/StudentHub/backend/go.mod)）
- **HTTP 框架**：Gin v1.10
- **ORM**：GORM v1.25 + `gorm.io/driver/sqlite`
- **数据库**：SQLite3（**WAL 模式**：`PRAGMA journal_mode=WAL; foreign_keys=ON;`）
- **日志**：`go.uber.org/zap`（结构化 JSON）
- **配置**：`spf13/viper`（env 优先于 yaml）
- **鉴权**：`golang-jwt/jwt/v5`，HS256，**Access 15min + Refresh 7d (HttpOnly Cookie)**
- **密码**：`bcrypt cost=12`
- **调度**：`robfig/cron/v3`
- **缓存**：`hashicorp/golang-lru/v2`（进程内，V1 不上 Redis）
- **加密**：`AES-256-GCM`（字段级，密钥 `APP_DATA_KEY` 经环境变量注入）

### 2.2 前端
- **框架**：Vue 3.5 + `<script setup>` + Vite 5
- **状态**：Pinia 3（**禁用 Vuex**）
- **路由**：vue-router 4
- **UI 库**：Element Plus 2.8+（按需自动引入）
- **HTTP**：axios（统一封装在 `frontend/src/api/http.js`）
- **图表**：echarts 5
- **时间**：dayjs
- **测试**：Vitest + @vue/test-utils
- **质量**：ESLint + Prettier

### 2.3 部署
- 单体应用（Modular Monolith） + 单 SQLite 文件
- Docker 化（`Dockerfile` + `docker-compose.yml`）
- 反代：Nginx

---

## 3. 仓库目录结构

```
studenthub/
├── backend/                    # Go 后端（module: student-system）
│   ├── cmd/
│   │   ├── server/             # 主入口
│   │   ├── dbcheck/            # 数据库自检工具
│   │   ├── seedstudent/        # 学生数据灌入
│   │   └── ...                 # 其他运维子命令
│   ├── configs/                # config.yaml
│   ├── data/                   # SQLite 文件 + 本地文件存储
│   ├── internal/
│   │   ├── boot/               # 启动装配、迁移、Seed
│   │   ├── middleware/         # auth / rbac 中间件
│   │   ├── models/             # GORM 实体（按模块拆文件）
│   │   ├── modules/            # ★ 业务模块（API/Service/Repository 分层）
│   │   │   ├── auth/           # 认证
│   │   │   ├── idx/            # 学生身份库
│   │   │   ├── ty/             # 团员发展
│   │   │   ├── st/             # 社团活动
│   │   │   ├── sq/             # 学生社区
│   │   │   ├── qg/             # 勤工助学
│   │   │   ├── cmp/            # 综合素测量化
│   │   │   ├── noti/           # 通知中心
│   │   │   ├── file/           # 文件服务
│   │   │   ├── sys/            # 系统（用户/角色/字典）
│   │   │   └── dashboard/      # 聚合 Dashboard
│   │   ├── eventx/             # 领域事件总线
│   │   ├── scheduler/          # 定时任务
│   │   ├── statem/             # 状态机引擎
│   │   ├── notifyx/            # 通知通道
│   │   └── idgen/              # 业务编号生成
│   ├── pkg/                    # 通用工具（logger / cryptox / cachex / response）
│   ├── storage/                # 上传文件本地存储
│   └── go.mod
├── frontend/                   # Vue3 前端
│   ├── src/
│   │   ├── api/                # 按后端模块分子文件（auth.js / ty.js / st.js ...）
│   │   ├── components/         # 通用组件
│   │   ├── layouts/            # DefaultLayout
│   │   ├── router/             # 路由 + meta 权限
│   │   ├── stores/             # Pinia（auth / menu / dict）
│   │   ├── utils/              # datetime / echarts
│   │   └── views/              # 页面（按模块目录：cmp/ idx/ qg/ sq/ st/ ty/ sys/）
│   ├── vite.config.js
│   └── package.json
├── docs/                       # ★ 必读 SSOT 文档
├── deploy/                     # docker-compose / env 模板
├── .trae/                      # Trae IDE 规则与技能
└── README.md
```

---

## 4. API 约定（速查）

### 4.1 URL
- **基础地址**：`/api/v1`
- **模块前缀**：`/auth /idx /ty /st /sq /qg /cmp /noti /file /sys`
- **资源复数**：`/ty/applications`、`/st/activities`
- **状态推进动宾**：`POST /ty/applications/{id}/submit`、`POST /ty/applications/{id}/approve` —— **状态机推进必须走动作端点，禁止 PATCH status 字段**
- **查询**：`?page=1&page_size=20&sort=created_at:desc&q[status]=S1`
- **嵌套最多 2 层**：`/classes/{id}/students`

### 4.2 统一响应封包
```json
{
  "code": 0,
  "message": "ok",
  "data": { },
  "request_id": "01J0X1V8P9KQYWZS2H3FYZRN1A"
}
```
列表分页：
```json
{ "code": 0, "data": { "items": [...], "page": 1, "page_size": 20, "total": 3421 } }
```

### 4.3 错误码段
| 段 | 模块 |
| -- | ---- |
| 0 | 成功 |
| 1000–1099 | 通用（参数 / 权限 / 系统） |
| 1100–1199 | IDX |
| 2000–2099 | TY 团员发展 |
| 3000–3099 | ST 社团活动 |
| 4000–4099 | SQ 学生社区 |
| 5000–5099 | QG 勤工助学 |
| 6000–6099 | CMP |
| 7000–7099 | NOTI |
| 8000–8099 | FILE |
| 9000–9099 | SYS |

错误响应：`{ "code": 2401, "message": "申请人年龄超出 14–28 周岁", "biz_code": "TY.APPLICATION.AGE_OUT_OF_RANGE", "request_id": "..." }`

### 4.4 时区 / 时间
- 后端 → 前端：RFC3339 + `+08:00`
- 前端**不做时区转换**，原样展示

### 4.5 鉴权
- `Authorization: Bearer {access_token}`
- 401 → 透明刷新（Refresh Token in HttpOnly Cookie）
- 403 → 跳转无权限页
- ≥500 → 全局错误页 + 暴露 `request_id`

---

## 5. 数据库约定（速查）

- **命名**：表 `{module}_{entity}`（`ty_application`、`st_activity`），字段 `snake_case`
- **通用字段**（每张业务表必备）：
  ```
  id          INTEGER PRIMARY KEY AUTOINCREMENT
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  created_by  INTEGER
  updated_by  INTEGER
  is_deleted  INTEGER NOT NULL DEFAULT 0   -- 软删
  ```
- **业务编号** `biz_no`（unique），格式 `<MODULE>-<YYYY>-<4位流水>`（如 `TY-2026-0001`）
- **外键**：`{ref_table_singular}_id`（如 `student_id`）
- **索引**：`idx_{table}_{col1}_{col2}`
- **软删**：业务查询统一通过 `repository` 过滤 `is_deleted=0`；物理删除仅允许在归档任务中
- **GORM Model**：字段名、类型、长度、默认值、`gorm` / `json` tag 必须与 [`docs/03_database_design_spec.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/03_database_design_spec.md) **完全一致**

---

## 6. 状态机与审计（必须理解）

- 任何状态变更**必须**走 `statem.Apply()`，**禁止**业务代码直接 `UPDATE status=...`
- 状态转移 = `(ObjectType, FromState, ToState, Action, Guard, Effect)`
- 每一笔状态变更写 `event_log`（append-only，禁止物理删除）
- 任何 S0→S1、S2→S3 这类关键变更 API 必须经过 `audit.Middleware`
- 审计保留 ≥ 5 年，应用日志 ≥ 180 天

---

## 7. 模块化与跨模块规则

- **模块边界**：每个模块**只能**通过自己的 `service` 暴露能力
- **跨模块调用**：通过 `boot` 注入的接口引用；**禁止**直接访问他人 `repository` / DB session
- **事件总线**：`eventx.Publish(ctx, &event.TyApplicationSubmitted{...})`；订阅在模块启动时 `eventx.Subscribe(...)`
- 出现"跨模块事务"时，优先 **Saga 编排式事务**（事件补偿），**禁止跨模块大事务**

---

## 8. 编码规范（强制）

### 8.1 Go
- `gofmt` / `goimports` 强制；提交前必须通过
- `golangci-lint` 必跑（`go vet` / `staticcheck` / `errcheck` / `ineffassign` / `revive` / `gocyclo ≤ 20`）
- 函数圈复杂度 ≤ 20
- **业务代码禁止 `panic`**（仅 `cmd/boot` 启动校验可 fatal）
- **禁止 `_ = err` 静默吞错**；`error` 必须 `fmt.Errorf("...: %w", err)` wrap
- 单测覆盖：模块核心 `service` ≥ 70%
- 文件命名 `snake_case.go`；结构体 `PascalCase`；接口 `XxxRepository`
- 局部变量 `camelCase`；常量 `PascalCase` 或全大写

### 8.2 Vue / TS
- **必须** `<script setup>` + `defineProps` 泛型化
- **禁止 `any`**（`unknown` + 类型守卫代替）
- 组件名 `PascalCase.vue`
- 路由 `meta` 必须含：`title / icon / permission / keepAlive / module`
- 表单统一 `el-form` 校验
- 异步动作封装在 Pinia store 的 `action` 中；**组件不直接调 axios**，统一走 `src/api/*`

### 8.3 提交规范
- Conventional Commits：`<type>(<scope>): <subject>`
- type：`feat / fix / refactor / perf / docs / test / chore / build / ci`
- 例：`feat(ty): 支持推优大会到会率硬卡控`
- 单 PR ≤ 600 行净增，至少 1 名同模块同事 + 1 名 TL 评审

### 8.4 命名
| 类型 | 规则 | 示例 |
| ---- | ---- | ---- |
| 包名 | 小写单词 | `modules/ty/service` |
| DB 表 | 模块前缀蛇形 | `ty_application` |
| 事件名 | PascalCase 过去式 | `TyApplicationSubmitted` |
| 错误码 | 4 位数字 | `2401` |
| 前端 store | `use` + PascalCase | `useUserStore` |
| 前端 API 文件 | 模块名小写 | `src/api/ty.js` |

---

## 9. 安全 & 敏感数据

- 静态敏感字段（身份证 / 银行卡 / 家庭经济信息）走 **AES-256-GCM**，统一 `pkg/cryptox`
- 列表 / 导出按角色脱敏（如身份证 `110***********0023`）
- **禁止日志打印**：密码、token、身份证、银行卡、家庭经济情况
- 配置文件 `.env*` 不入仓；密钥**只**经环境变量
- 文件上传：MIME 白名单 + 后缀白名单 + 大小 ≤ 50MB + 重命名

---

## 10. 迭代节奏（强约束）

- 严格按 [`docs/05_superpowers_iteration_plan.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/05_superpowers_iteration_plan.md) 的 **S01~S12 步长**逐个垂直击穿
- 每个切片闭环：`DB Migration → Model → Repo → Service → API → 前端联调 → 测试`
- **未到达的切片不提前实现**；**已到达的切片不省略环节**
- 完成后必须通过：
  - 后端：`go build ./... && go test ./...`
  - 前端：`pnpm lint && pnpm build`
  - 关键切片：补充 Postman / curl 烟测脚本

---

## 11. AI 助手必须遵守的"硬禁忌"

1. ❌ **禁止**凭直觉新增 / 删除 / 变更状态机节点 → 必须核对 PRD
2. ❌ **禁止**直接 `UPDATE status=...` 绕过状态机 → 必须走 `statem.Apply()`
3. ❌ **禁止**直接访问其他模块的 `repository` / 表结构 → 走接口注入
4. ❌ **禁止**引入未经 ADR 批准的框架、库、中间件
5. ❌ **禁止**修改 GORM tag / 字段类型 / 索引而不同步 `docs/03_database_design_spec.md`
6. ❌ **禁止**在前端用 `v-html`（除非白名单审计）
7. ❌ **禁止**业务代码 `panic`、静默吞错 ` _ = err`
8. ❌ **禁止**日志打印身份证 / 密码 / token / 银行卡
9. ❌ **禁止**改 `.env*`、提交密钥；CI 注入
10. ❌ **禁止**跳过"垂直击穿"步骤，提前实现未到切片

---

## 12. 接到任务时的标准动作

1. **先定位文档**：本任务属于哪个切片（S0X）？读 `docs/05` 定位 → 读对应 `docs/01-04`
2. **再读现有代码**：`Grep` / `Glob` 找相似模块的实现（如新写 TY 的接口，先看 `internal/modules/ty/` 已有代码）
3. **再读 .trae 规则**：永远把 [`.trae/rules/project_rules.md`](file:///d:/Teach/AI_Coding/StudentHub/.trae/rules/project_rules.md) 当作硬约束
4. **动手前对齐**：任何与 SSOT 文档不一致的需求 → **先向用户澄清**，禁止擅自决断
5. **改动后验证**：`go build` / `go test` / `pnpm lint` / `pnpm build` 全绿后才算完成
6. **回复与注释统一使用中文**

---

## 13. 常用命令速查

```bash
# 后端
cd backend
go run cmd/server/main.go              # 启动服务（默认 :8088）
go build ./...                          # 编译
go test ./...                           # 测试
go vet ./...                            # 静态检查
golangci-lint run                       # 完整 lint（如有）

# 前端
cd frontend
pnpm install                            # 安装依赖
pnpm dev                                # 启动 Vite（默认 :5173）
pnpm build                              # 打包到 dist/
pnpm lint                               # ESLint
pnpm test                               # Vitest

# 一键烟测（参考 backend/data 下的 verify_*.ps1）
```

---

**维护说明**：本文件是 AI 助手的"快速入门 + 速查手册"。任何对技术栈、目录、API 风格、编码规范的变更，**必须同步更新** `docs/02_ADR.md` 与本文件。**禁止二者在事实上不一致。**

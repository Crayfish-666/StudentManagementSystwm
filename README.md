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

所有架构设计与规范文档均位于项目根目录：

| 文档名称 | 路径 | 核心内容说明 |
| :--- | :--- | :--- |
| **PRD 产品需求文档** | [PRD.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/PRD.md) | 包含 US-001 ~ US-017 用户故事、硬卡控规则、五态状态机、MinIO 分片上传与 LLM AI 综测要求 |
| **ADR 架构决策文档** | [ADR.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/ADR.md) | 包含分层架构图、ADR-001 ~ ADR-009 架构决策（SpringBoot3/SQLite WAL/Sa-Token/Flyway/MinIO/Spring AI） |
| **数据库设计规范** | [database_design_spec.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/database_design_spec.md) | 包含 Mermaid ER 图、SQLite 全量 DDL 脚本、唯一索引、CHECK 约束与 Flyway 初始化迁移脚本 |
| **API 规范文档 (SRD)** | [SRD_api_specifications.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/SRD_api_specifications.md) | RESTful API 契约、Sa-Token 鉴权、统一 Response 封包、错误码表、MinIO & AI 端点与 JSON 样例 |
| **业务分析报告** | [Analyze.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/Analyze.md) | 10 大维度纯业务分析报告（项目定位、角色、流程、数据模型、页面数据流等） |

---

## 答辩与答辩材料清单

| 材料名称 | 路径 | 适用答辩环节 |
| :--- | :--- | :--- |
| **《成员分工记录表》** | [member_work_division.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/member_work_division.md) | 环节 2：开发分工与角色展示 |
| **《需求迭代表 (v1.0/v1.1/v2.0)》** | [iteration_records.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/iteration_records.md) | 环节 2：三轮快速迭代脉络说明 |
| **《AI 编程赋能落地说明表》** | [ai_empowerment_log.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/ai_empowerment_log.md) | 环节 3：5 大 AI 辅助开发落地场景 |
| **《答辩演示与 PPT 9 页指南》** | [presentation_guide.md](file:///d:/Developing/WorkSpace/StudentManagementSystemVD/presentation_guide.md) | 全程演练、PPT 制作与 5 分钟实操演练话术 |

---

## 技术栈选型

* **后端**：Java 17 + Spring Boot 3.2.x + MyBatis-Plus 3.5.x + Sa-Token 1.37+ (JWT 模式) + Flyway 10.x
* **数据库**：SQLite 3 (开启 WAL 写前日志、外键约束、5000ms 锁超时)
* **前端**：Vue 3.5 (Composition API `<script setup>`) + Vite 5 + Element Plus 2.8+ + Pinia 3 + Axios + ECharts 5
* **中间件**：MinIO 对象存储（支持分片上传、预签名链接在线预览、打包下载）
* **AI 大模型**：Spring AI / DeepSeek API（AI 综测初稿生成与人工复核覆写）

---

## 快速开始

### 1. 后端启动 (Spring Boot)
```bash
# 源码运行 (默认端口 8080)
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

## Git 仓库提交说明

本项目已同步提交并推送至目标 GitHub 仓库：
`https://github.com/Crayfish-666/StudentManagementSystwm.git`

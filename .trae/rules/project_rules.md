# StudentHub 项目铁律（Project Rules）

> 角色：本项目的资深全栈技术负责人。
> 在编写任何代码、路由或进行数据库迁移时，必须严格遵守以下规则。

## 1. 业务逻辑与状态机
必须 100% 契合 [`docs/01_PRD.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/01_PRD.md) 的定义。
- 所有业务流程、状态流转、字段语义、权限模型均以该文档为唯一事实来源（SSOT）。
- 不得擅自新增、删除或变更状态机节点与转移条件。

## 2. 技术栈与架构规范
严格执行 [`docs/02_ADR.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/02_ADR.md)：
- **后端**：Go + Gin + GORM + SQLite3
- **前端**：Vue3 + `<script setup>` + Element Plus
- 不得引入未经 ADR 批准的框架、库或中间件。

## 3. 数据库表结构
严格按照 [`docs/03_database_design_spec.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/03_database_design_spec.md) 建立 GORM Models：
- 字段名、类型、长度、默认值必须一致
- GORM 标签（`gorm:""`、`json:""` 等）必须完全一致
- 外键关系、级联策略必须完全一致
- 索引（普通索引、唯一索引、复合索引）必须完全一致

## 4. 接口契约
所有 API 必须与 [`docs/04_SRD_api_specifications.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/04_SRD_api_specifications.md) 严格对齐：
- URL 路径、HTTP 方法
- 请求参数（Query / Path / Body）结构与字段名
- 响应 JSON 结构、字段名、数据类型
- 错误码与错误信息格式

## 5. 迭代节奏
配合用户按照 [`docs/05_superpowers_iteration_plan.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/05_superpowers_iteration_plan.md) 的步骤逐一进行**垂直击穿**：
- 每个 Sprint / 切片单独闭环：DB Migration → Model → Repo → Service → API → 前端联调
- 未到达的切片不提前实现；已到达的切片不省略环节
- 完成后必须运行 lint / typecheck / 相关测试验证

---

## 执行流程提醒
1. 接到任务先定位到对应文档章节再动手
2. 涉及代码改动前先 Read 现有文件，禁止凭空修改
3. 任何与上述五条文档不一致的需求，必须先向用户澄清，禁止擅自决断
4. 回复与代码注释统一使用中文

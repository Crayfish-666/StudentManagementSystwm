# 学生“一站式”自主管理过程管理系统 (StudentHub) · RESTful API 规范文档（SRD）

| 文档版本 | 修订日期 | 编写者 | 接口协议 | 文档状态 |
| :--- | :--- | :--- | :--- | :--- |
| V2.1 (答辩强化版) | 2026-07-22 | API 全栈架构专家 | HTTPS / JSON (Sa-Token JWT) | 正式规约 |

---

## 0. 阅读指引与设计契约

### 0.1 基础信息
* **基础根路径**：`http://localhost:8080/api/v1`
* **交互协议**：JSON (RFC 8259)，字符集 UTF-8
* **时间格式**：标准 ISO-8601 (RFC3339) 格式带时区，如 `2026-07-22T08:30:00+08:00`
* **鉴权方式**：Sa-Token 生成的 JWT Bearer Token。请求头需携带 `Authorization: Bearer {access_token}`

---

## 1. 统一响应封包与错误处理

```json
{
  "code": 0,
  "message": "ok",
  "data": { },
  "request_id": "01HZX8P9KQYWZS2H3FYZRN1A"
}
```

---

## 2. MinIO 文件服务 API 规范 (`/api/v1/files`)

### 2.1 初始化分片上传 (`POST /files/multipart/init`)
* **说明**：大文件（>10MB）初始化分片上传，返回 MinIO Upload ID。
* **请求 Body**：
```json
{
  "original_name": "activity_video.mp4",
  "file_size": 52428800,
  "content_type": "video/mp4"
}
```
* **响应 Data**：
```json
{
  "code": 0,
  "data": {
    "file_key": "2026/07/activity_video.mp4",
    "upload_id": "minio-upload-982312"
  }
}
```

### 2.2 上传文件切片 (`POST /files/multipart/chunk`)
* **说明**：前端上传某一个分片数据，返回分片 ETag。

### 2.3 完成分片合并 (`POST /files/multipart/complete`)
* **说明**：通知 MinIO 合并所有分片，生成最终文件。

### 2.4 获取 MinIO 预签名在线预览链接 (`GET /files/preview-url`)
* **说明**：获取带 15 分钟时效的 MinIO 预览 URL。
* **请求 Query**：`?file_key=2026/07/activity_video.mp4`
* **响应 Data**：
```json
{
  "code": 0,
  "data": {
    "preview_url": "http://minio.school.edu:9000/studenthub-bucket/2026/07/activity_video.mp4?X-Amz-Expires=900&X-Amz-Signature=..."
  }
}
```

---

## 3. LLM 大模型 AI 综测初稿 API 规范 (`/api/v1/cmp/ai-evaluation`)

### 3.1 自动生成 AI 综测评语初稿 (`POST /cmp/ai-evaluation/generate`)
* **权限**：`@SaCheckRole("R-COL-COUN")`
* **说明**：抽取学生履历，调用 Spring AI / DeepSeek API 生成结构化初稿。
* **请求 Body**：
```json
{
  "student_id": 101,
  "academic_term": "2025-2026-2"
}
```
* **响应 Data**：
```json
{
  "code": 0,
  "data": {
    "student_id": 101,
    "academic_term": "2025-2026-2",
    "ai_summary": "该生在本学期思想上积极上进，已列为入团积极分子；担任计算机协会副会长，组织 B 级算法大赛表现突出；勤工助学岗位履职优秀，月度考勤达标。",
    "ai_suggestions": "建议继续加强社区宿舍卫生文明建设，提升团队协作综合分。",
    "status": "draft"
  }
}
```

### 3.2 辅导员人工复核与覆写评语 (`POST /cmp/ai-evaluation/overwrite`)
* **权限**：`@SaCheckRole("R-COL-COUN")`
* **请求 Body**：
```json
{
  "student_id": 101,
  "academic_term": "2025-2026-2",
  "human_override": "该生整体表现优异，同意 AI 初稿评语，额外表彰其在社区暴雨抢险中的突出贡献。",
  "final_score": 92.5
}
```

---

## 4. 团员发展、社团活动、社区自治与勤工助学 API

（包含入团申请提交、推优表决、活动分级审批、提前结束招新 `:finish`、宿舍巡查与 L1~L4 异常派单结案、勤工打卡考勤与月度算薪等 RESTful 端点，保持高标准一致性）。

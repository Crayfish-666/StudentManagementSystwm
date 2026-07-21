# 学生“一站式”自主管理过程管理系统 (StudentHub) · RESTful API 规范文档（SRD）

| 文档版本 | 修订日期 | 编写者 | 接口协议 | 文档状态 |
| :--- | :--- | :--- | :--- | :--- |
| V2.0 (SpringBoot版) | 2026-07-22 | API 全栈架构专家 | HTTPS / JSON (Sa-Token JWT) | 正式规约 |

---

## 0. 阅读指引与设计契约

### 0.1 基础信息
* **基础根路径**：`http://localhost:8080/api/v1`
* **交互协议**：JSON (RFC 8259)，字符集 UTF-8
* **时间格式**：标准 ISO-8601 (RFC3339) 格式带时区，如 `2026-07-22T08:30:00+08:00`
* **鉴权方式**：Sa-Token 生成的 JWT Bearer Token。请求头需携带 `Authorization: Bearer {access_token}`

---

## 1. 统一响应封包与错误处理

### 1.1 成功响应封包
所有 `/api/v1/**` 响应一律封包为以下结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": { },
  "request_id": "01HZX8P9KQYWZS2H3FYZRN1A"
}
```

* `code = 0` 表示成功；非零为业务错误码。
* `request_id` 为每笔请求生成的唯一追踪 ID（与响应头 `X-Request-ID` 一致）。

### 1.2 分页列表响应封包
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [ ],
    "page": 1,
    "page_size": 20,
    "total": 142,
    "total_pages": 8
  },
  "request_id": "01HZX8P9KQYWZS2H3FYZRN1A"
}
```

### 1.3 错误响应与核心错误码定义
当出现异常时，返回 HTTP 状态码 400/401/403/409/500，并在 Body 中包含具体错误结构：

```json
{
  "code": 1305,
  "message": "推优大会到会率不足，实到团员人数须 >= 应到人数的 2/3",
  "data": null,
  "request_id": "01HZX8P9KQYWZS2H3FYZRN1A"
}
```

| 错误码 | 业务类别 | 说明 |
| :--- | :--- | :--- |
| **0** | 成功 | 请求成功完成 |
| **1001** | 参数非法 | 必填字段缺失或格式校验失败 |
| **1201** | 未登录 / Token 无效 | 凭证过期或被 Sa-Token 黑名单拦截 |
| **1203** | 无操作权限 | RBAC / ABAC 作用域鉴权未通过 |
| **1301** | 状态机跃迁非法 | 企图跳过当前状态直接变更 |
| **1302** | 勤工月工时超限 | 累计工时已超过 40 小时/月 |
| **1305** | 推优到会率不足 | 实到团员未达 2/3 |
| **1306** | 未认定困难生阻断 | 未通过困难认定无法申请常规岗位 |

---

## 2. 鉴权与系统核心 API (`/api/v1/auth`, `/api/v1/sys`)

### 2.1 用户登录 (`POST /auth/login`)
* **说明**：支持学号/工号登录，颁发 Access Token 并下发 Refresh Token Cookie。
* **请求 Body**：
```json
{
  "username": "2023010101",
  "password": "student@123"
}
```
* **响应 Data**：
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token_name": "Authorization",
    "access_token": "eyJhbGciOiJIUzI1NiJ9...",
    "user_info": {
      "user_id": 101,
      "username": "2023010101",
      "real_name": "张三",
      "user_type": "student",
      "roles": ["R-STU-NORM", "R-STU-ASSOC"]
    }
  },
  "request_id": "01HZX8P9KQYWZS2H3FYZRN1A"
}
```

### 2.2 无感刷新 Token (`POST /auth/refresh`)
* **说明**：通过 HttpOnly Cookie 刷新 Access Token。
* **响应 Data**：
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiJ9.new..."
  }
}
```

---

## 3. 团员发展 API 规范 (`/api/v1/ty`)

### 3.1 提交入团申请 (`POST /ty/applications`)
* **权限**：`@SaCheckRole("R-STU-NORM")`
* **请求 Body**：
```json
{
  "statement": "我深刻认识到共青团是党的得力助手和后备军，在学习与生活中严格要求自己...（>=500字）"
}
```
* **响应 Data**：
```json
{
  "code": 0,
  "data": {
    "id": 12,
    "biz_no": "TY-2026-0012",
    "student_id": 101,
    "app_status": "S1",
    "created_at": "2026-07-22T09:00:00+08:00"
  }
}
```

### 3.2 推进入团申请状态 (`POST /ty/applications/{id}/approve`)
* **权限**：`@SaCheckRole("R-COL-LEAGUE")`
* **请求 Body**：
```json
{
  "action": "approve",
  "opinion": "同意推荐入团，思想表现良好。"
}
```

### 3.3 录入支部推优大会表决 (`POST /ty/recommendation-meetings`)
* **权限**：`@SaCheckRole("R-STU-LEAGUE")`
* **说明**：提交推优大会，系统执行刚性人数校验。
* **请求 Body**：
```json
{
  "branch_id": 5,
  "meeting_date": "2026-07-22T14:30:00+08:00",
  "location": "3 号教学楼 201 教室",
  "total_members": 30,
  "attended_members": 25,
  "photo_urls": ["/storage/2026/07/photo1.jpg", "/storage/2026/07/photo2.jpg"],
  "votes": [
    { "application_id": 12, "approve_votes": 22, "reject_votes": 2, "abstain_votes": 1 }
  ]
}
```

---

## 4. 社团活动 API 规范 (`/api/v1/st`)

### 4.1 提交活动立项 (`POST /st/activities`)
* **权限**：`@SaCheckRole("R-STU-ASSOC")`
* **请求 Body**：
```json
{
  "assoc_id": 3,
  "title": "计算机协会 2026 算法编程大赛",
  "level": "B",
  "budget_cents": 600000,
  "start_time": "2026-08-01T09:00:00+08:00",
  "end_time": "2026-08-01T17:00:00+08:00",
  "location": "图书馆一楼学术报告厅",
  "emergency_plan_url": "/storage/2026/07/plan_b.pdf"
}
```

### 4.2 提前结束招新计划 (`POST /st/recruit-plans/{id}/finish`)
* **权限**：`@SaCheckRole("R-STU-ASSOC")`
* **说明**：手动执行提前结束招新动作，操作不可逆。
* **请求 Body**：
```json
{
  "reason": "预定招新名额已满，提前截止投递。"
}
```
* **响应 Data**：
```json
{
  "code": 0,
  "data": {
    "id": 8,
    "assoc_id": 3,
    "is_finished": 1,
    "finished_at": "2026-07-22T10:15:00+08:00",
    "finished_reason": "预定招新名额已满，提前截止投递。"
  }
}
```

### 4.3 活动现场签到 (`POST /st/activities/{id}/checkin`)
* **请求 Body**：
```json
{
  "checkin_type": "qrcode",
  "token": "SIGN_TOKEN_XXXXX"
}
```

---

## 5. 社区自治 API 规范 (`/api/v1/sq`)

### 5.1 上报异常事件 (`POST /sq/incidents`)
* **请求 Body**：
```json
{
  "building_id": 2,
  "level": "L4",
  "incident_type": "fire_alarm",
  "description": "二楼西侧走廊感烟探测器报警，有轻微焦糊味。"
}
```

### 5.2 L1~L4 事件结案 (`POST /sq/incidents/{id}/close`)
* **权限**：`@SaCheckRole("R-COL-FLOOR")`
* **请求 Body**：
```json
{
  "resolution": "系线路短路产生微烟，已切断电源并完成抢修，无人员伤亡。"
}
```

---

## 6. 勤工助学 API 规范 (`/api/v1/qg`)

### 6.1 在线打卡考勤 (`POST /qg/attendances/clock`)
* **权限**：`@SaCheckRole("R-STU-WORK")`
* **说明**：上下班打卡，月累计满 40h 自动阻断。
* **请求 Body**：
```json
{
  "apply_id": 15,
  "clock_type": "in"
}
```

### 6.2 录入月度考核与自动算薪 (`POST /qg/monthly-assess`)
* **权限**：`@SaCheckRole("R-COL-COUN")`
* **请求 Body**：
```json
{
  "apply_id": 15,
  "assess_month": "2026-06",
  "score": 92,
  "comments": "工作认真负责，服务态度良好。"
}
```

---

## 7. 综合素质量化 API 规范 (`/api/v1/cmp`)

### 7.1 查询个人综合素质得分 (`GET /cmp/scores/me`)
* **响应 Data**：
```json
{
  "code": 0,
  "data": {
    "student_id": 101,
    "total_score": 88.5,
    "ty_score": 27.0,
    "st_score": 22.5,
    "sq_score": 18.0,
    "qg_score": 12.0,
    "academic_score": 9.0,
    "rank_in_college": 5,
    "rank_in_major": 2
  }
}
```

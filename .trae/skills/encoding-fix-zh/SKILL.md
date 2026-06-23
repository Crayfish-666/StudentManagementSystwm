---
name: "encoding-fix-zh"
description: "排查并修复中文/全角标点在页面/数据库中显示为 '?' 的乱码问题。在用户描述 '中文乱码 / 显示问号 / ??? / 标点变成问号 / Excel 导入后变 ?' 时立即触发。"
---

# 中文 / 全角标点显示为 "?" 排查修复指南

> 适用项目：StudentHub（Go + Gin + GORM + SQLite + Vue3 + Element Plus）
> 适用现象：页面、列表、详情、SQL 查询结果中，本应是中文或全角标点的位置出现 `?` 或 `??`，且**字符数固定为 1 个 `?` 替换一个汉字**。

---

## 1. 决策链：先定位「字节」在哪一层丢失

中文丢失为 `?` 的本质是：**某一层把不能识别的字节用 ASCII `0x3F`（'?'）做了替换**。
按数据流向逐层定位（命中即停）：

```
CSV/Excel 文件 → HTTP Body → Go 后端 string → GORM → SQLite TEXT → API JSON → 浏览器
        ①            ②            ③          ④         ⑤          ⑥        ⑦
```

| 层 | 典型症状 | 定位手段 |
|----|----------|---------|
| ① 源文件本身就是乱码 | 用记事本/VSCode 切换为 GBK 后能正常显示 | 用 VSCode 右下角切换编码查看；或 `file <name>.csv` |
| ② HTTP 上传时被错误解码 | 后端 log 中入参就已为 `?` | 在 Service 入口 `zlog.Info` 打印原始字节 `%x` |
| ③ Go 字符串处理 | 几乎不会发生（Go string 是字节串）| —— |
| ④ GORM/驱动 | 几乎不会发生（mattn/go-sqlite3 走 UTF-8）| —— |
| ⑤ **SQLite 存储为 `0x3F`**（最常见根因）| `SELECT hex(name)` 返回 `3F3F` 而不是 `E5BCA0E4B889` | `sqlite3 db "SELECT id, name, hex(name) FROM <table>;"` |
| ⑥ API JSON 序列化 | 浏览器 Network 面板里 JSON 已是 `?` | 浏览器 DevTools → Network → Response |
| ⑦ 浏览器渲染 | Network 里 JSON 是中文，但页面变 `?` | 检查 `<meta charset>` / Content-Type |

> **黄金 1 分钟定位法**：
> ```bash
> sqlite3 data/studenthub.db "SELECT id, <字段>, hex(<字段>) FROM <表> LIMIT 20;"
> ```
> - `hex()` 全是 `3F` → ⑤ 已落库脏数据，必须从源头修；
> - `hex()` 是合法 UTF-8（汉字三字节 `E?????`）→ 问题在 ⑥ 或 ⑦。

---

## 2. 根因清单（StudentHub 已知 / 高频）

### R1. CSV 上传未做编码识别（已修复）
**触发路径**：用户在 Windows 用 Excel "另存为 CSV"，默认 GBK；后端 `csv.NewReader(reader)` 直接当 UTF-8 解析 → 无效字节落库为 `?`。
**修复点**：[student_service.go](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/modules/idx/service/student_service.go) 中 `decodeCSVReader` 函数：自动识别 UTF-8 BOM / 合法 UTF-8 / GB18030，统一转 UTF-8 后再交给 `csv.Reader`。
**关键依赖**：`golang.org/x/text/encoding/simplifiedchinese` + `golang.org/x/text/transform`。

### R1b. PowerShell 5 / curl 发 JSON 时按 GBK 编码 body（已加全局守护）
**触发路径**：在 PowerShell 5 中 `Invoke-WebRequest -Body '{"name":"中文"}' -ContentType 'application/json'`，PowerShell 5 默认把字符串按系统 ANSI（中文 Windows = GBK / CP936）写入 body；后端 `c.ShouldBindJSON` 按 UTF-8 解码失败位置 → 字节落入 string → 写 SQLite TEXT 时被替换为 `0x3F`。
**全局防御**：[middleware.go](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/boot/middleware.go) 中 `utf8GuardMiddleware`，对所有 POST/PUT/PATCH 且 Content-Type 含 json 的请求：
1. 合法 UTF-8 → 直接放行；
2. 非合法 UTF-8 但能用 GB18030 解码 → 重写 body 为 UTF-8（兼容老脚本）；
3. 都失败 → 返回 `41000` 提示客户端切 UTF-8。
**前提**：`boot.go` 必须 `r.Use(utf8GuardMiddleware())`。
**调用方推荐姿势**：
```powershell
# PowerShell 5 正确发送 UTF-8 JSON：
$body = [System.Text.Encoding]::UTF8.GetBytes('{"name":"测试社团"}')
Invoke-WebRequest -Uri http://localhost:8081/api/v1/st/associations `
  -Method POST -Body $body `
  -ContentType 'application/json; charset=utf-8' `
  -Headers @{ Authorization = "Bearer $token" }
```
或者改用 PowerShell 7 / Postman / Apifox / `curl.exe`（注意是 `curl.exe` 不是 PS 别名）。

### R2. 后端响应未声明 UTF-8
Gin 的 `c.JSON` 默认 `application/json; charset=utf-8`，**通常不会出问题**。如果自定义中间件覆写了 Content-Type，需要确保保留 `charset=utf-8`。

### R3. 前端 HTML 缺少 charset
检查：[index.html](file:///d:/Teach/AI_Coding/StudentHub/frontend/index.html) 必须含 `<meta charset="UTF-8" />`。✅ 当前项目正常。

### R4. 终端/控制台编码（伪问题）
PowerShell 默认输出编码不是 UTF-8，导致 `sqlite3` / `go run` 终端打印的中文变 `?`，但**数据库本身正常**。
排除方法：执行 `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8` 后再观察。

### R5. 源代码文件用 GBK 保存
所有 Go / Vue 源文件必须以 UTF-8 (无 BOM) 保存。如怀疑：用 VSCode 打开右下角看编码。

---

## 3. 修复 Workflow（强制按序）

### Step 1：固定快照
```powershell
# 备份数据库（StudentHub 约定 data/ 目录）
Copy-Item backend\data\studenthub.db backend\data\studenthub.db.bak.$((Get-Date).ToString('yyyyMMdd_HHmmss'))
```

### Step 2：定位根因层
```powershell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
sqlite3 backend\data\studenthub.db "SELECT id, <字段>, hex(<字段>) FROM <表> WHERE hex(<字段>) LIKE '%3F%';"
```

### Step 3：修源头
- 命中 R1：确认 `decodeCSVReader` 已生效；
- 命中 R5：把源文件转 UTF-8 重新保存。

### Step 4：清理脏数据
**⚠️ StudentHub 项目铁律**：禁止擅自变更业务数据，需先与用户对齐方案，三选一：

1. **软删除并重导**（推荐，数据可重新上传）：
   ```sql
   UPDATE idx_student SET is_deleted = 1, updated_at = CURRENT_TIMESTAMP
   WHERE hex(name) LIKE '%3F3F%' AND is_deleted = 0;
   ```
2. **按学号映射回填**（用户提供 `student_no → 真实姓名` 映射表）：
   ```sql
   UPDATE idx_student SET name = '张三' WHERE student_no = '2023003';
   ```
3. **物理删除**（仅当确无业务关联时）：
   ```sql
   DELETE FROM idx_student WHERE hex(name) LIKE '%3F3F%';
   ```

### Step 5：回归验证
```powershell
# 1. 后端编译
cd backend; go build ./...
# 2. 重启后端
go run ./cmd/server
# 3. 重新上传一份 GBK 编码 CSV，验证 hex(name) 为合法 UTF-8 字节
sqlite3 data\studenthub.db "SELECT id, name, hex(name) FROM idx_student ORDER BY id DESC LIMIT 5;"
```

---

## 4. 预防 Checklist（提交代码前自检）

- [ ] 任何接收用户上传文本（CSV / TXT / JSON）的接口，都必须显式做编码探测（UTF-8 BOM / 合法 UTF-8 / GB18030）；
- [ ] 涉及中文的单元测试至少包含一个 GBK 字节序列样本（`"\xd5\xc5\xc8\xfd"` = "张三"）；
- [ ] 数据库迁移 / 种子脚本中的中文字面量必须由 UTF-8 源文件直接写入，禁止 `cmd > file.sql` 这种被 GBK 重定向的写法；
- [ ] 前端模板文件 `<meta charset="UTF-8" />` 不可缺失；
- [ ] 自定义 Gin 中间件改写响应头时保留 `charset=utf-8`。

---

## 5. 常用排查命令清单

```powershell
# 强制 PowerShell 输出 UTF-8（每次新开终端都要执行一次）
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 看某字段所有「问号脏数据」
sqlite3 backend\data\studenthub.db "SELECT id, name, hex(name) FROM idx_student WHERE hex(name) LIKE '%3F%';"

# 看某字段编码分布（合法 UTF-8 汉字以 E?/F? 开头）
sqlite3 backend\data\studenthub.db "SELECT substr(hex(name),1,2) AS prefix, COUNT(*) FROM idx_student GROUP BY prefix;"

# 文件编码探测（PowerShell）
Get-Content -Path .\students.csv -Encoding Byte -TotalCount 16 | ForEach-Object { '{0:X2}' -f $_ }
```

---

## 6. 触发本 Skill 的关键词

中文乱码 / 显示成问号 / 显示 ? / 显示 ?? / 显示 ??? / 字符变 ? / 标点符号变 ? /
encoding 问号 / GBK 乱码 / Excel CSV 中文 / 导入后中文丢失 / SQLite 中文 ? / 全角标点 ? /
"页面显示" + "?"

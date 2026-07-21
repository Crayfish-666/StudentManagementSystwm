# P0 审计修复任务清单

## 第一组：删除 Java 残留 + 基础修正
- [ ] 1.1 删除 backend/src/ 目录
- [ ] 1.2 删除 backend/pom.xml
- [ ] 1.3 清理 backend/.gitignore
- [ ] 1.4 修正 vite.config.js

## 第二组：后端安全加固
- [ ] 2.1 AES 密钥加固
- [ ] 2.2 JWT 密钥加固
- [ ] 2.3 CORS 白名单
- [ ] 2.4 Cookie Secure 动态化
- [ ] 2.5 MIME 嗅探加固
- [ ] 2.6 错误吞没修复

## 第三组：状态机绕过修复
- [ ] 3.1-3.5 QG/SQ/Auth 状态机

## 第四组：数据库约束
- [ ] 4.1-4.3 模型修复 + 约束补建

## 第五组：前端修复
- [ ] 5.1-5.6 路由/http.js/Login/Dashboard

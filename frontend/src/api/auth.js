import http from './http'

// 认证 API：与 04_SRD 对齐
export const authApi = {
  // 登录 POST /api/v1/auth/login
  login: (data) => http.post('/auth/login', data),

  // 刷新 Token POST /api/v1/auth/refresh
  refresh: () => http.post('/auth/refresh'),

  // 登出 POST /api/v1/auth/logout
  logout: () => http.post('/auth/logout'),

  // 获取当前用户 GET /api/v1/auth/me
  me: () => http.get('/auth/me'),

  // 修改密码 POST /api/v1/auth/password
  changePassword: (data) => http.post('/auth/password', data)
}

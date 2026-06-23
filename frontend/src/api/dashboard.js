import http from './http'

// 工作台 API
export const dashboardApi = {
  // 工作台概览 GET /api/v1/dashboard/overview
  overview() {
    return http.get('/dashboard/overview')
  }
}

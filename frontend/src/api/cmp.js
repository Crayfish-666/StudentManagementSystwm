import http from './http'

// 综合素质量化 API
export const cmpScoreApi = {
  // 学生本人综合分
  myScore(term) {
    return http.get('/cmp/scores/me', { params: { term } })
  },
  // 学生本人历史综合分
  myHistory() {
    return http.get('/cmp/scores/me/history')
  },
  // 综合分列表
  list(params) {
    return http.get('/cmp/scores', { params })
  },
  // 综合分详情
  get(studentId, term) {
    return http.get(`/cmp/scores/${studentId}`, { params: { term } })
  },
  // 手动重算单学生
  recompute(studentId, term) {
    return http.post(`/cmp/scores/${studentId}/recompute`, null, { params: { term } })
  },
  // 批量重算
  batchCompute(data) {
    return http.post('/cmp/scores/compute', data)
  }
}

// 综合素质看板 API
export const cmpDashboardApi = {
  // 关键 KPI
  kpi(term) {
    return http.get('/cmp/dashboard/kpi', { params: { term } })
  },
  // 趋势
  trends(metric, range) {
    return http.get('/cmp/dashboard/trends', { params: { metric, range } })
  },
  // 分布
  distribution(dim, term) {
    return http.get('/cmp/dashboard/distribution', { params: { dim, term } })
  },
  // 活跃社团按院系
  activeAssocByCollege() {
    return http.get('/cmp/dashboard/active-assoc-by-college')
  },
  // 事件等级分布
  incidentLevel() {
    return http.get('/cmp/dashboard/incident-level')
  }
}

// 规则版本 API
export const cmpRuleApi = {
  list() {
    return http.get('/cmp/rule-versions')
  },
  create(data) {
    return http.post('/cmp/rule-versions', data)
  },
  activate(id) {
    return http.post(`/cmp/rule-versions/${id}/activate`)
  }
}

import http from './http'

// 入团申请 API
export const tyApplicationApi = {
  // 分页查询入团申请列表
  list(params) {
    return http.get('/ty/applications', { params })
  },
  // 获取入团申请详情
  get(id) {
    return http.get(`/ty/applications/${id}`)
  },
  // 创建入团申请（保存为草稿）
  create(data) {
    return http.post('/ty/applications', data)
  },
  // 更新入团申请（仅草稿状态可改）
  update(id, data) {
    return http.put(`/ty/applications/${id}`, data)
  },
  // 提交入团申请（S0 → S1）
  submit(id) {
    return http.post(`/ty/applications/${id}/submit`)
  },
  // 撤回入团申请（S1 → S0）
  withdraw(id, reason) {
    return http.post(`/ty/applications/${id}/withdraw`, { reason })
  },
  // 删除入团申请（仅 S0/S4）
  delete(id) {
    return http.delete(`/ty/applications/${id}`)
  },
  // 三级审批 通过/驳回（S06）
  approve(id, data) {
    return http.post(`/ty/applications/${id}/approve`, data)
  },
  // 审批记录列表
  listApprovals(id) {
    return http.get(`/ty/applications/${id}/approvals`)
  },
  // 事件流时间线
  timeline(id) {
    return http.get(`/ty/applications/${id}/timeline`)
  },
  // 我的待办审批
  listPending(params) {
    return http.get('/ty/approvals/pending', { params })
  },
  // 团员发展轨迹（静默模式：后端未重启时不弹错误提示）
  developmentTrack(studentId) {
    return http.get(`/ty/students/${studentId}/development-track`, { silent: true })
  }
}

// 团支部 API
export const tyBranchApi = {
  // 获取团支部列表（下拉选择用）
  list(collegeId) {
    const params = {}
    if (collegeId) params.college_id = collegeId
    return http.get('/ty/branches', { params })
  }
}

// 推优大会 API
export const tyRecommendationMeetingApi = {
  list(params) { return http.get('/ty/recommendation-meetings', { params }) },
  get(id) { return http.get(`/ty/recommendation-meetings/${id}`) },
  getByApplication(applicationId) { return http.get(`/ty/recommendation-meetings/application/${applicationId}`) },
  create(data) { return http.post('/ty/recommendation-meetings', data) }
}

// 培养联系人 API
export const tyCultivationLinkApi = {
  list(params) { return http.get('/ty/cultivation-links', { params }) },
  create(data) { return http.post('/ty/cultivation-links', data) },
  end(id) { return http.post(`/ty/cultivation-links/${id}/end`) }
}

// 培养记录 API
export const tyCultivationRecordApi = {
  list(params) { return http.get('/ty/cultivation-records', { params }) },
  create(data) { return http.post('/ty/cultivation-records', data) }
}

// 团课记录 API
export const tyCourseRecordApi = {
  list(params) { return http.get('/ty/course-records', { params }) },
  create(data) { return http.post('/ty/course-records', data) },
  updatePassStatus(id) { return http.put(`/ty/course-records/${id}/pass`) }
}

// 思想汇报 API
export const tyThoughtReportApi = {
  list(params) { return http.get('/ty/thought-reports', { params }) },
  get(id) { return http.get(`/ty/thought-reports/${id}`) },
  create(data) { return http.post('/ty/thought-reports', data) }
}

// 发展对象 API
export const tyDevelopmentObjectApi = {
  list(params) { return http.get('/ty/development-objects', { params }) },
  get(id) { return http.get(`/ty/development-objects/${id}`) },
  create(data) { return http.post('/ty/development-objects', data) },
  publicize(id) { return http.post(`/ty/development-objects/${id}/publicize`) },
  approve(id, data) { return http.post(`/ty/development-objects/${id}/approve`, data) }
}

// 政审记录 API
export const tyPoliticalReviewApi = {
  list(params) { return http.get('/ty/political-reviews', { params }) },
  create(data) { return http.post('/ty/political-reviews', data) }
}

// 发展大会 API
export const tyDevelopmentMeetingApi = {
  list(params) { return http.get('/ty/development-meetings', { params }) },
  create(data) { return http.post('/ty/development-meetings', data) }
}

// 预备期考察 API
export const tyProbationaryRecordApi = {
  list(params) { return http.get('/ty/probationary-records', { params }) },
  get(id) { return http.get(`/ty/probationary-records/${id}`) },
  create(data) { return http.post('/ty/probationary-records', data) }
}

// 转正大会 API
export const tyProbationaryMeetingApi = {
  list(params) { return http.get('/ty/probationary-meetings', { params }) },
  get(id) { return http.get(`/ty/probationary-meetings/${id}`) },
  create(data) { return http.post('/ty/probationary-meetings', data) }
}

// 团员花名册 API
export const tyMemberRosterApi = {
  list(params) { return http.get('/ty/members', { params }) },
  get(id) { return http.get(`/ty/members/${id}`) },
  update(id, data) { return http.patch(`/ty/members/${id}`, data) },
  transferOut(id) { return http.post(`/ty/members/${id}/transfer-out`) },
  overtime(id) { return http.post(`/ty/members/${id}/overtime`) },
  archive(id) { return http.post(`/ty/members/${id}/archive`) }
}

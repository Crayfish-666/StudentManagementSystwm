import http from './http'

// 社团 API
export const stAssociationApi = {
  // 分页查询社团列表
  list(params) {
    return http.get('/st/associations', { params })
  },
  // 获取社团详情
  get(id) {
    return http.get(`/st/associations/${id}`)
  },
  // 创建社团
  create(data) {
    return http.post('/st/associations', data)
  },
  // 更新社团
  update(id, data) {
    return http.put(`/st/associations/${id}`, data)
  },
  // 删除社团
  delete(id) {
    return http.delete(`/st/associations/${id}`)
  },
  // 查询社团发起人
  listFounders(id) {
    return http.get(`/st/associations/${id}/founders`)
  },
  // 查询社团成员
  listMembers(id) {
    return http.get(`/st/associations/${id}/members`)
  },
  // 查询用户列表(指导教师下拉用,仅教职工)
  listUsers() {
    return http.get('/st/users')
  },
  // 查询学生列表(社长下拉用)
  listStudents() {
    return http.get('/st/students')
  }
}

// 活动 API
export const stActivityApi = {
  // 分页查询活动列表
  list(params) {
    return http.get('/st/activities', { params })
  },
  // 获取活动详情
  get(id) {
    return http.get(`/st/activities/${id}`)
  },
  // 创建活动
  create(data) {
    return http.post('/st/activities', data)
  },
  // 更新活动
  update(id, data) {
    return http.put(`/st/activities/${id}`, data)
  },
  // 提交活动（S0 → S1）
  submit(id) {
    return http.post(`/st/activities/${id}/submit`)
  },
  // 删除活动
  delete(id) {
    return http.delete(`/st/activities/${id}`)
  },
  // 审批活动
  approve(id, data) {
    return http.post(`/st/activities/${id}/approve`, data)
  },
  // 撤回活动
  withdraw(id) {
    return http.post(`/st/activities/${id}/withdraw`)
  },
  // 查询审批记录
  listApprovals(id) {
    return http.get(`/st/activities/${id}/approvals`)
  },
  // 查询审批时间线
  timeline(id) {
    return http.get(`/st/activities/${id}/timeline`)
  },
  // 签到
  checkin(id, data) {
    return http.post(`/st/activities/${id}/checkin`, data)
  },
  // 查询签到记录
  listCheckins(id, params) {
    return http.get(`/st/activities/${id}/checkins`, { params })
  },
  // 提交活动总结
  submitSummary(id, data) {
    return http.post(`/st/activities/${id}/summary`, data)
  },
  // 获取活动总结
  getSummary(id) {
    return http.get(`/st/activities/${id}/summary`)
  }
}

// 招新计划 API
export const stRecruitPlanApi = {
  // 分页查询招新计划
  list(params) {
    return http.get('/st/recruit-plans', { params })
  },
  // 获取计划详情
  get(id) {
    return http.get(`/st/recruit-plans/${id}`)
  },
  // 创建招新计划
  create(data) {
    return http.post('/st/recruit-plans', data)
  },
  // 更新招新计划
  update(id, data) {
    return http.put(`/st/recruit-plans/${id}`, data)
  },
  // 提交计划（S0 → S1）
  submit(id) {
    return http.post(`/st/recruit-plans/${id}/submit`)
  },
  // 撤回计划（S1 → S0）
  withdraw(id) {
    return http.post(`/st/recruit-plans/${id}/withdraw`)
  },
  // 审批通过（S1 → S3）
  approve(id) {
    return http.post(`/st/recruit-plans/${id}/approve`)
  },
  // 驳回（S1 → S4）
  reject(id, data) {
    return http.post(`/st/recruit-plans/${id}/reject`, data)
  },
  // 发布
  publish(id) {
    return http.post(`/st/recruit-plans/${id}/publish`)
  },
  // 提前结束招新（仅 S3 + 未结束可用）
  finish(id, data) {
    return http.post(`/st/recruit-plans/${id}/finish`, data || {})
  }
}

// 招新申请 API
export const stRecruitApplyApi = {
  // 分页查询申请
  list(params) {
    return http.get('/st/recruit-applies', { params })
  },
  // 学生投递
  create(data) {
    return http.post('/st/recruit-applies', data)
  },
  // 录入面试结果
  submitResult(id, data) {
    return http.post(`/st/recruit-applies/${id}/result`, data)
  }
}

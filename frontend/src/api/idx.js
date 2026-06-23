import http from './http'

// 学生管理 API
export const studentApi = {
  // 分页查询学生列表
  list(params) {
    return http.get('/idx/students', { params })
  },
  // 获取学生详情
  get(id) {
    return http.get(`/idx/students/${id}`)
  },
  // 创建学生
  create(data) {
    return http.post('/idx/students', data)
  },
  // 更新学生
  update(id, data) {
    return http.put(`/idx/students/${id}`, data)
  },
  // 删除学生
  delete(id) {
    return http.delete(`/idx/students/${id}`)
  },
  // 批量导入学生（CSV）
  importCSV(file) {
    const formData = new FormData()
    formData.append('file', file)
    return http.post('/idx/students/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  // 获取组织树
  getOrgTree() {
    return http.get('/idx/org-tree')
  },
  // 获取我的画像
  getMyProfile() {
    return http.get('/idx/profile/me')
  }
}

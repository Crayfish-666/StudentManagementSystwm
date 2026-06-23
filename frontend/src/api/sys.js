import http from './http'

// 字典 API
export const dictApi = {
  // 按分类查询字典项
  getItems(category) {
    return http.get(`/sys/dicts/${category}/items`)
  },
  // 列出所有字典分类（管理）
  listCategories() {
    return http.get('/sys/dicts')
  },
  // 新增字典项
  createItem(data) {
    return http.post('/sys/dicts/items', data)
  },
  // 修改字典项
  updateItem(id, data) {
    return http.put(`/sys/dicts/items/${id}`, data)
  },
  // 删除字典项
  deleteItem(id) {
    return http.delete(`/sys/dicts/items/${id}`)
  }
}

// 菜单 API
export const menuApi = {
  // 获取当前用户可见菜单
  getMyMenus() {
    return http.get('/sys/menus/mine')
  }
}

// 用户管理 API（系统管理员）
export const userApi = {
  // 分页列表
  list(params) {
    return http.get('/sys/users', { params })
  },
  // 详情
  get(id) {
    return http.get(`/sys/users/${id}`)
  },
  // 新建
  create(data) {
    return http.post('/sys/users', data)
  },
  // 更新基本信息
  update(id, data) {
    return http.put(`/sys/users/${id}`, data)
  },
  // 软删
  remove(id) {
    return http.delete(`/sys/users/${id}`)
  },
  // 重置密码
  resetPassword(id, newPassword) {
    return http.post(`/sys/users/${id}/reset-password`, { new_password: newPassword })
  },
  // 锁定
  lock(id) {
    return http.post(`/sys/users/${id}/lock`)
  },
  // 解锁
  unlock(id) {
    return http.post(`/sys/users/${id}/unlock`)
  },
  // 启用
  enable(id) {
    return http.post(`/sys/users/${id}/enable`)
  },
  // 禁用
  disable(id) {
    return http.post(`/sys/users/${id}/disable`)
  },
  // 分配角色（覆盖式）
  assignRoles(id, roleIds) {
    return http.post(`/sys/users/${id}/roles`, { role_ids: roleIds })
  },
  // 撤销单个角色
  revokeRole(id, rid) {
    return http.delete(`/sys/users/${id}/roles/${rid}`)
  }
}

// 角色 API
export const roleApi = {
  // 列出所有角色
  list() {
    return http.get('/sys/roles')
  }
}

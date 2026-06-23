import http from './http'

// 楼栋 API
export const sqBuildingApi = {
  // 获取楼栋树形结构
  tree() {
    return http.get('/sq/buildings/tree')
  },
  // 分页查询楼栋列表
  list(params) {
    return http.get('/sq/buildings', { params })
  },
  // 获取楼栋详情
  get(id) {
    return http.get(`/sq/buildings/${id}`)
  },
  // 创建楼栋
  create(data) {
    return http.post('/sq/buildings', data)
  },
  // 更新楼栋
  update(id, data) {
    return http.put(`/sq/buildings/${id}`, data)
  },
  // 删除楼栋
  delete(id) {
    return http.delete(`/sq/buildings/${id}`)
  },
  // 查询楼栋下的楼层
  listFloors(buildingId) {
    return http.get(`/sq/buildings/${buildingId}/floors`)
  },
  // 创建楼层
  createFloor(data) {
    return http.post('/sq/floors', data)
  },
  // 删除楼层
  deleteFloor(id) {
    return http.delete(`/sq/floors/${id}`)
  },
  // 查询楼栋下的寝室
  listRooms(buildingId, params) {
    return http.get(`/sq/buildings/${buildingId}/rooms`, { params })
  },
  // 创建寝室
  createRoom(data) {
    return http.post('/sq/rooms', data)
  },
  // 删除寝室
  deleteRoom(id) {
    return http.delete(`/sq/rooms/${id}`)
  },
  // 查询寝室成员
  getRoomMembers(roomId) {
    return http.get(`/sq/rooms/${roomId}/members`)
  }
}

// 巡查 API
export const sqInspectionApi = {
  // 分页查询巡查列表
  list(params) {
    return http.get('/sq/inspections', { params })
  },
  // 获取巡查详情
  get(id) {
    return http.get(`/sq/inspections/${id}`)
  },
  // 创建巡查记录
  create(data) {
    return http.post('/sq/inspections', data)
  },
  // 删除巡查记录
  delete(id) {
    return http.delete(`/sq/inspections/${id}`)
  }
}

// 事件 API
export const sqIncidentApi = {
  // 分页查询事件列表
  list(params) {
    return http.get('/sq/incidents', { params })
  },
  // 获取事件详情
  get(id) {
    return http.get(`/sq/incidents/${id}`)
  },
  // 创建事件
  create(data) {
    return http.post('/sq/incidents', data)
  },
  // 处置事件
  handle(id, data) {
    return http.post(`/sq/incidents/${id}/handle`, data)
  },
  // 结案事件
  close(id, data) {
    return http.post(`/sq/incidents/${id}/close`, data)
  },
  // 删除事件
  delete(id) {
    return http.delete(`/sq/incidents/${id}`)
  }
}

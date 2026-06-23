import http from './http'

// 通知 API
export const notificationApi = {
  // 查询我的通知列表
  listMine(params) {
    return http.get('/notifications/mine', { params })
  },
  // 获取未读数
  getUnreadCount() {
    return http.get('/notifications/unread-count')
  },
  // 标记已读
  markRead(id) {
    return http.post(`/notifications/${id}/read`)
  },
  // 全部已读
  markAllRead() {
    return http.post('/notifications/read-all')
  }
}

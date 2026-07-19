import http from './http'

// 文件 API
export const fileApi = {
  // 上传文件（multipart/form-data）
  upload(formData) {
    return http.post('/files/upload', formData)
  },
  // 下载文件
  download(key) {
    return http.get(`/files/${key}`, { responseType: 'blob' })
  },
  // 获取文件元数据
  getMeta(key) {
    return http.get(`/files/${key}/meta`)
  },
  // 删除文件
  delete(key) {
    return http.delete(`/files/${key}`)
  }
}

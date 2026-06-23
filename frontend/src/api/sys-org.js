import http from './http'

// 院系 API
export const collegeApi = {
  list() {
    return http.get('/sys/colleges')
  },
  create(data) {
    return http.post('/sys/colleges', data)
  },
  update(id, data) {
    return http.put(`/sys/colleges/${id}`, data)
  },
  delete(id) {
    return http.delete(`/sys/colleges/${id}`)
  }
}

// 专业 API
export const majorApi = {
  list(params) {
    return http.get('/sys/majors', { params })
  },
  create(data) {
    return http.post('/sys/majors', data)
  },
  update(id, data) {
    return http.put(`/sys/majors/${id}`, data)
  },
  delete(id) {
    return http.delete(`/sys/majors/${id}`)
  }
}

// 班级 API
export const classApi = {
  list(params) {
    return http.get('/sys/classes', { params })
  },
  create(data) {
    return http.post('/sys/classes', data)
  },
  update(id, data) {
    return http.put(`/sys/classes/${id}`, data)
  },
  delete(id) {
    return http.delete(`/sys/classes/${id}`)
  }
}

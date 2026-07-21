import axios from 'axios'
import { ElMessage } from 'element-plus'

const http = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  withCredentials: true
})

let isRefreshing = false
let pendingRequests = []

http.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

http.interceptors.response.use(
  (response) => {
    const body = response.data
    if (body == null || typeof body !== 'object' || !('code' in body)) {
      return body
    }
    if (body.code === 0) {
      return body.data
    }
    // 不对 404/资源未找到弹错误 Toast，避免死循环报屏
    if (!body.message?.includes('No static resource')) {
      ElMessage.error(body.message || `请求失败(code=${body.code})`)
    }
    return Promise.reject(Object.assign(new Error(body.message || 'biz error'), {
      code: body.code,
      requestId: body.request_id
    }))
  },
  async (error) => {
    const originalRequest = error.config
    const status = error?.response?.status
    const bizCode = error?.response?.data?.code

    if (status === 401 && bizCode === 40103) {
      pendingRequests.forEach(({ reject }) => reject(error))
      pendingRequests = []
      localStorage.removeItem('access_token')
      import('@/router').then(({ default: router }) => {
        router.push('/login')
      })
      ElMessage.error('登录状态已失效，请重新登录')
      return Promise.reject(error)
    }

    if (status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingRequests.push({ resolve, reject, config: originalRequest })
        })
      }

      isRefreshing = true
      try {
        const { data } = await axios.post('/api/v1/auth/refresh', null, { withCredentials: true })
        const newToken = data?.data?.access_token
        if (newToken) {
          localStorage.setItem('access_token', newToken)
          pendingRequests.forEach(({ resolve, reject, config }) => {
            config.headers.Authorization = `Bearer ${newToken}`
            http(config).then(resolve).catch(reject)
          })
          pendingRequests = []
          originalRequest.headers.Authorization = `Bearer ${newToken}`
          return http(originalRequest)
        }
      } catch (refreshErr) {
        pendingRequests.forEach(({ reject }) => reject(refreshErr))
        pendingRequests = []
        localStorage.removeItem('access_token')
        import('@/router').then(({ default: router }) => {
          router.push('/login')
        })
        return Promise.reject(refreshErr)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

export default http

import axios from 'axios'
import { ElMessage } from 'element-plus'

// 全局 Axios 实例：与 ADR-009 / 010 对齐
const http = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  withCredentials: true
})

// 是否正在刷新 Token
let isRefreshing = false
// 刷新期间排队的请求
let pendingRequests = []

// 请求拦截：注入 Authorization
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

// 响应拦截：统一解包 + 401 刷新重试
http.interceptors.response.use(
  (response) => {
    const body = response.data
    // 兼容文件下载等非标准 JSON 响应
    if (body == null || typeof body !== 'object' || !('code' in body)) {
      return body
    }
    if (body.code === 0) {
      return body.data
    }
    ElMessage.error(body.message || `请求失败(code=${body.code})`)
    return Promise.reject(Object.assign(new Error(body.message || 'biz error'), {
      code: body.code,
      requestId: body.request_id
    }))
  },
  async (error) => {
    const originalRequest = error.config
    const status = error?.response?.status
    const bizCode = error?.response?.data?.code

    // 40103 = RT/Token 已被吊销（改密/禁用），**不**再尝试 refresh，直接登出
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

    // 401 且未重试过 → 尝试刷新 Token
    if (status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      if (isRefreshing) {
        // 已有刷新请求进行中，排队等待
        return new Promise((resolve, reject) => {
          pendingRequests.push({ resolve, reject, config: originalRequest })
        })
      }

      isRefreshing = true
      try {
        // 用 refresh_token（HttpOnly Cookie）刷新
        const { data } = await axios.post('/api/v1/auth/refresh', null, { withCredentials: true })
        const newToken = data?.data?.access_token
        if (newToken) {
          localStorage.setItem('access_token', newToken)
          // 重试排队的请求
          pendingRequests.forEach(({ resolve, reject, config }) => {
            config.headers.Authorization = `Bearer ${newToken}`
            http(config).then(resolve).catch(reject)
          })
          pendingRequests = []
          // 重试原始请求
          originalRequest.headers.Authorization = `Bearer ${newToken}`
          return http(originalRequest)
        }
      } catch (refreshErr) {
        // 刷新失败，清除登录态，跳转登录页
        pendingRequests.forEach(({ reject }) => reject(refreshErr))
        pendingRequests = []
        localStorage.removeItem('access_token')
        // 延迟导入避免循环依赖
        import('@/router').then(({ default: router }) => {
          router.push('/login')
        })
        // 区分 40103（吊销）与其他失效原因
        const refreshBizCode = refreshErr?.response?.data?.code
        ElMessage.error(refreshBizCode === 40103 ? '登录状态已失效，请重新登录' : '登录已过期，请重新登录')
        return Promise.reject(refreshErr)
      } finally {
        isRefreshing = false
      }
    }

    if (status === 403) {
      ElMessage.error('无权限访问')
    } else if (status >= 500) {
      ElMessage.error('服务器异常，请稍后重试')
    } else if (status !== 401 && !originalRequest.silent) {
      ElMessage.error(error.message || '网络异常')
    }

    return Promise.reject(error)
  }
)

export default http

import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

/**
 * HTTP 客户端（适配 Java Spring Boot + Sa-Token）。
 *
 * 关键说明：Sa-Token 鉴权失败/权限不足时返回的是 HTTP 200 + body.code=10401/10403，
 * 不是 HTTP 401/403。因此鉴权相关的错误处理必须放在"成功拦截器"里判断 body.code。
 *
 * 错误码（与后端 GlobalExceptionHandler 对齐）：
 *   0      - 成功
 *   1001   - 参数/业务异常
 *   10401  - 未登录 / 登录态失效
 *   10403  - 权限不足
 *   10404  - 资源不存在
 *   10422  - 参数校验失败
 *   1500   - 系统内部异常
 */
const http = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  withCredentials: true
})

// ---------- 请求拦截器：注入 Bearer Token ----------
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

// ---------- 响应拦截器：统一处理成功/失败 ----------
http.interceptors.response.use(
  // --- 成功响应（HTTP 2xx）---
  (response) => {
    const body = response.data

    // 如果不是标准封包（比如文件下载），直接返回
    if (body == null || typeof body !== 'object' || !('code' in body)) {
      return body
    }

    // 成功
    if (body.code === 0) {
      return body.data
    }

    // ===== 业务错误码（HTTP 200 + body.code !== 0） =====

    // 10401 未登录 / token 失效 → 清登录态跳登录
    if (body.code === 10401) {
      handleAuthExpired()
      return Promise.reject(createBizError(body))
    }

    // 10403 权限不足 → 跳 403 页
    if (body.code === 10403) {
      handleForbidden()
      return Promise.reject(createBizError(body))
    }

    // 1500 系统内部异常 → 弹错误，暴露 request_id
    if (body.code === 1500) {
      const rid = body.request_id || ''
      ElMessage.error({
        message: `服务器内部错误${rid ? '，请求ID: ' + rid : ''}`,
        duration: 5000
      })
      return Promise.reject(createBizError(body))
    }

    // 其他业务错误 → 弹消息
    if (body.message) {
      ElMessage.error(body.message)
    }
    return Promise.reject(createBizError(body))
  },

  // --- HTTP 错误（4xx/5xx）---
  (error) => {
    const status = error?.response?.status

    // HTTP 401（极少触发，Sa-Token 用 body.code）
    if (status === 401) {
      handleAuthExpired()
      return Promise.reject(error)
    }

    // HTTP 403
    if (status === 403) {
      handleForbidden()
      return Promise.reject(error)
    }

    // HTTP >= 500 网络/服务端错误
    if (status >= 500) {
      const rid = error.response?.headers?.['x-request-id'] || ''
      ElMessage.error({
        message: `服务器错误 (${status})${rid ? '，请求ID: ' + rid : ''}`,
        duration: 5000
      })
      return Promise.reject(error)
    }

    // 网络错误 / 超时
    if (!error.response || error.code === 'ECONNABORTED') {
      ElMessage.error('网络错误，请检查网络连接')
      return Promise.reject(error)
    }

    return Promise.reject(error)
  }
)

// ---------- 工具函数 ----------

function handleAuthExpired() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('user_info')
  if (router.currentRoute.value.path !== '/login') {
    ElMessage.warning('登录状态已失效，请重新登录')
    router.push('/login')
  }
}

function handleForbidden() {
  if (router.currentRoute.value.path !== '/403') {
    router.push('/403')
  }
}

function createBizError(body) {
  const err = new Error(body.message || 'biz error')
  err.code = body.code
  err.bizCode = body.biz_code
  err.requestId = body.request_id
  return err
}

export default http

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { useMenuStore } from '@/stores/menu'

// Auth Store：管理登录态、Token、用户信息。
export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(localStorage.getItem('access_token') || '')
  const user = ref(null)

  const isLoggedIn = computed(() => !!accessToken.value)
  const roles = computed(() => user.value?.roles?.map(r => r.code) || [])
  const displayName = computed(() => user.value?.display_name || '')

  // 登录
  async function login(username, password) {
    const data = await authApi.login({ username, password })
    accessToken.value = data.access_token
    localStorage.setItem('access_token', data.access_token)
    user.value = data.user
    return data
  }

  // 刷新 Token
  async function refresh() {
    const data = await authApi.refresh()
    accessToken.value = data.access_token
    localStorage.setItem('access_token', data.access_token)
    return data
  }

  // 获取当前用户
  async function fetchUser() {
    const data = await authApi.me()
    user.value = data
    return data
  }

  // 登出
  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // 忽略登出请求失败
    }
    accessToken.value = ''
    user.value = null
    localStorage.removeItem('access_token')
    // 重置菜单状态
    const menuStore = useMenuStore()
    menuStore.resetMenus()
  }

  // 修改密码：成功后服务端会吊销当前 RT，本地需清态并跳登录页
  async function changePassword(oldPassword, newPassword) {
    await authApi.changePassword({ old_password: oldPassword, new_password: newPassword })
    // 本地清态（由调用方负责跳转登录页）
    accessToken.value = ''
    user.value = null
    localStorage.removeItem('access_token')
    const menuStore = useMenuStore()
    menuStore.resetMenus()
  }

  // 清除本地状态（Token 过期时调用）
  function clearAuth() {
    accessToken.value = ''
    user.value = null
    localStorage.removeItem('access_token')
  }

  return {
    accessToken,
    user,
    isLoggedIn,
    roles,
    displayName,
    login,
    refresh,
    fetchUser,
    logout,
    clearAuth
  }
})

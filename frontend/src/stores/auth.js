import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { useMenuStore } from '@/stores/menu'

// Auth Store：管理登录态、Token、用户信息。
export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(localStorage.getItem('access_token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user_info') || 'null'))

  const isLoggedIn = computed(() => !!accessToken.value)
  const roles = computed(() => user.value?.roles || [])
  const displayName = computed(() => user.value?.displayName || user.value?.username || '系统用户')

  // 登录
  async function login(username, password) {
    const data = await authApi.login({ username, password })
    const token = data.token || data.access_token
    accessToken.value = token
    localStorage.setItem('access_token', token)
    
    const userInfo = {
      userId: data.userId,
      username: data.username,
      displayName: data.displayName,
      userType: data.userType,
      studentId: data.studentId,
      roles: data.roles || []
    }
    user.value = userInfo
    localStorage.setItem('user_info', JSON.stringify(userInfo))
    return data
  }

  // 获取当前用户
  async function fetchUser() {
    if (!accessToken.value) return null
    try {
      const data = await authApi.me()
      if (data) {
        user.value = data
        localStorage.setItem('user_info', JSON.stringify(data))
      }
    } catch {
      // 忽略查询用户信息失败
    }
    return user.value
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
    localStorage.removeItem('user_info')
    const menuStore = useMenuStore()
    menuStore.resetMenus()
  }

  // 清除本地状态
  function clearAuth() {
    accessToken.value = ''
    user.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('user_info')
  }

  return {
    accessToken,
    user,
    isLoggedIn,
    roles,
    displayName,
    login,
    fetchUser,
    logout,
    clearAuth
  }
})

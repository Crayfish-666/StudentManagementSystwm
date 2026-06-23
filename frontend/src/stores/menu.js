import { defineStore } from 'pinia'
import { ref } from 'vue'
import { menuApi } from '@/api/sys'

// 使用 import.meta.glob 预收集所有视图组件，Vite 推荐的动态导入方式
const viewModules = import.meta.glob('../views/**/*.vue')

// Menu Store：管理动态菜单 + 路由注册。
export const useMenuStore = defineStore('menu', () => {
  const menuList = ref([])
  const isLoaded = ref(false)

  // 获取菜单并动态注册路由
  async function fetchMenus(router) {
    if (isLoaded.value) return

    try {
      const data = await menuApi.getMyMenus()
      menuList.value = data.menus || []
      isLoaded.value = true

      // 动态注册路由
      if (router) {
        registerDynamicRoutes(router)
      }
    } catch (err) {
      console.error('获取菜单失败:', err)
    }
  }

  // 动态注册路由
  function registerDynamicRoutes(router) {
    const layoutRoute = router.getRoutes().find(r => r.name === 'Layout')
    if (!layoutRoute) return

    const dynamicRoutes = buildRoutes(menuList.value)
    dynamicRoutes.forEach(route => {
      // 避免重复注册
      if (!router.hasRoute(route.name)) {
        router.addRoute('Layout', route)
      }
    })
  }

  // 从菜单树构建路由配置
  function buildRoutes(menus) {
    const routes = []
    for (const menu of menus) {
      if (menu.children && menu.children.length) {
        routes.push(...buildRoutes(menu.children))
      } else if (menu.component) {
        // menu.component 格式如 "views/Dashboard.vue"
        // viewModules 的 key 格式如 "../views/Dashboard.vue"
        const moduleKey = `../${menu.component}`
        const componentLoader = viewModules[moduleKey]
        if (componentLoader) {
          routes.push({
            name: menu.code,
            path: menu.path,
            component: componentLoader,
            meta: { title: menu.title }
          })
        } else {
          console.warn(`菜单组件未找到: ${moduleKey}`)
        }
      }
    }
    return routes
  }

  // 重置菜单（登出时调用）
  function resetMenus() {
    menuList.value = []
    isLoaded.value = false
  }

  return {
    menuList,
    isLoaded,
    fetchMenus,
    resetMenus
  }
})

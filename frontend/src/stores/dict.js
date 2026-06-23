import { defineStore } from 'pinia'
import { ref } from 'vue'
import { dictApi } from '@/api/sys'

// Dict Store：缓存常用字典数据。
export const useDictStore = defineStore('dict', () => {
  const dictMap = ref({})

  // 获取字典项（带缓存）
  async function getDictItems(category) {
    if (dictMap.value[category]) {
      return dictMap.value[category]
    }
    const data = await dictApi.getItems(category)
    const items = data.items || []
    dictMap.value[category] = items
    return items
  }

  // 清除缓存（字典变更时调用）
  function clearCache(category) {
    if (category) {
      delete dictMap.value[category]
    } else {
      dictMap.value = {}
    }
  }

  return {
    dictMap,
    getDictItems,
    clearCache
  }
})

// useDict 组合式函数：返回响应式字典数据。
export function useDict(category) {
  const items = ref([])
  const loading = ref(false)

  const dictStore = useDictStore()

  async function loadDict() {
    loading.value = true
    try {
      items.value = await dictStore.getDictItems(category)
    } catch (err) {
      console.error(`加载字典 ${category} 失败:`, err)
    } finally {
      loading.value = false
    }
  }

  loadDict()

  return { items, loading, refresh: loadDict }
}

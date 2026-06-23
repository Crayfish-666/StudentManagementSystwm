<template>
  <el-popover
    placement="bottom-end"
    :width="360"
    trigger="click"
    @show="fetchList"
  >
    <template #reference>
      <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99" class="bell-badge">
        <el-icon :size="20" class="bell-icon"><Bell /></el-icon>
      </el-badge>
    </template>

    <div class="noti-popover">
      <div class="noti-header">
        <span class="noti-title">通知中心</span>
        <el-button link type="primary" size="small" @click="handleReadAll" :disabled="unreadCount === 0">全部已读</el-button>
      </div>

      <div class="noti-list" v-loading="loading">
        <div v-if="list.length === 0" class="noti-empty">暂无通知</div>
        <div
          v-for="item in list"
          :key="item.id"
          class="noti-item"
          :class="{ unread: item.is_read === 0 }"
          @click="handleRead(item)"
        >
          <div class="noti-item-title">
            <el-tag v-if="item.is_read === 0" type="danger" size="small" effect="dark" class="dot-tag">未读</el-tag>
            <span>{{ item.title }}</span>
          </div>
          <div class="noti-item-body">{{ item.body }}</div>
          <div class="noti-item-time">{{ formatTime(item.created_at) }}</div>
        </div>
      </div>

      <div class="noti-footer">
        <el-button link type="primary" @click="goNotificationCenter">查看全部通知</el-button>
      </div>
    </div>
  </el-popover>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell } from '@element-plus/icons-vue'
import { notificationApi } from '@/api/notification'
import { formatDateTime as formatTime } from '@/utils/datetime'

const router = useRouter()

const unreadCount = ref(0)
const list = ref([])
const loading = ref(false)

let timer = null

const fetchUnreadCount = async () => {
  try {
    const data = await notificationApi.getUnreadCount()
    // http 拦截器已解包，data 就是 { unread_count: N }
    unreadCount.value = data?.unread_count ?? 0
  } catch {
    // 静默失败
  }
}

const fetchList = async () => {
  loading.value = true
  try {
    const data = await notificationApi.listMine({ page: 1, page_size: 10 })
    list.value = data?.items || []
  } catch {
    // 静默失败
  } finally {
    loading.value = false
  }
}

const handleRead = async (item) => {
  if (item.is_read === 0) {
    try {
      await notificationApi.markRead(item.id)
      item.is_read = 1
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    } catch {
      // 静默失败
    }
  }
}

const handleReadAll = async () => {
  try {
    await notificationApi.markAllRead()
    list.value.forEach(item => { item.is_read = 1 })
    unreadCount.value = 0
  } catch {
    // 静默失败
  }
}

const goNotificationCenter = () => {
  router.push('/notifications')
}

onMounted(() => {
  fetchUnreadCount()
  // 每 60 秒轮询一次未读数
  timer = setInterval(fetchUnreadCount, 60000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})
</script>

<style scoped>
.bell-badge {
  cursor: pointer;
}
.bell-icon {
  color: var(--sh-text-regular);
  transition: color var(--sh-duration-fast) var(--sh-ease-out);
}
.bell-icon:hover {
  color: var(--sh-primary);
}
.noti-popover {
  margin: calc(var(--sh-space-md) * -1);
}
.noti-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--sh-space-md) var(--sh-space-md);
  border-bottom: 1px solid var(--sh-border-light);
}
.noti-title {
  font-size: var(--sh-text-lg);
  font-weight: 600;
  color: var(--sh-text-primary);
}
.noti-list {
  max-height: 400px;
  overflow-y: auto;
}
.noti-empty {
  padding: var(--sh-space-2xl) 0;
  text-align: center;
  color: var(--sh-text-secondary);
  font-size: var(--sh-text-base);
}
.noti-item {
  padding: var(--sh-space-md);
  border-bottom: 1px solid var(--sh-border-light);
  cursor: pointer;
  transition: background var(--sh-duration-fast) var(--sh-ease-out);
}
.noti-item:hover {
  background: var(--sh-bg-elevated);
}
.noti-item.unread {
  background: var(--sh-primary-lighter);
}
.noti-item.unread:hover {
  background: var(--sh-info-light);
}
.noti-item-title {
  display: flex;
  align-items: center;
  gap: var(--sh-space-sm);
  font-size: var(--sh-text-base);
  font-weight: 500;
  color: var(--sh-text-primary);
  margin-bottom: var(--sh-space-xs);
}
.dot-tag {
  font-size: 10px;
  padding: 0 var(--sh-space-xs);
  height: 18px;
  line-height: 18px;
}
.noti-item-body {
  font-size: var(--sh-text-sm);
  color: var(--sh-text-regular);
  line-height: var(--sh-leading-normal);
  margin-bottom: var(--sh-space-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.noti-item-time {
  font-size: var(--sh-text-xs);
  color: var(--sh-text-secondary);
}
.noti-footer {
  padding: var(--sh-space-sm) var(--sh-space-md);
  text-align: center;
  border-top: 1px solid var(--sh-border-light);
}
</style>

<template>
  <div class="notification-center">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>通知中心</span>
          <el-button type="primary" size="small" @click="handleReadAll" :disabled="unreadCount === 0">全部已读</el-button>
        </div>
      </template>

      <!-- 筛选 -->
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="状态">
          <el-select v-model="filters.is_read" placeholder="全部" clearable style="width: 120px">
            <el-option label="未读" :value="0" />
            <el-option label="已读" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 列表 -->
      <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_read === 0 ? 'danger' : 'info'" size="small">
              {{ row.is_read === 0 ? '未读' : '已读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column prop="body" label="内容" min-width="300" show-overflow-tooltip />
        <el-table-column prop="channel" label="通道" width="100" />
        <el-table-column label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.is_read === 0" link type="primary" @click="handleRead(row)">标记已读</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { notificationApi } from '@/api/notification'
import { formatDateTime as formatTime } from '@/utils/datetime'

const filters = reactive({ is_read: '' })
const tableData = ref([])
const loading = ref(false)
const unreadCount = ref(0)

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const fetchUnreadCount = async () => {
  try {
    const data = await notificationApi.getUnreadCount()
    unreadCount.value = data?.unread_count ?? 0
  } catch {
    // 静默
  }
}

const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.is_read !== '' && filters.is_read !== null) {
      params.is_read = filters.is_read
    }
    const data = await notificationApi.listMine(params)
    tableData.value = data?.items || []
    pagination.total = data?.total || 0
  } catch {
    // 静默
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  filters.is_read = ''
  pagination.page = 1
  fetchList()
}

const handleRead = async (row) => {
  try {
    await notificationApi.markRead(row.id)
    row.is_read = 1
    unreadCount.value = Math.max(0, unreadCount.value - 1)
    ElMessage.success('已标记为已读')
  } catch {
    ElMessage.error('标记已读失败')
  }
}

const handleReadAll = async () => {
  try {
    await notificationApi.markAllRead()
    tableData.value.forEach(item => { item.is_read = 1 })
    unreadCount.value = 0
    ElMessage.success('已全部标记为已读')
  } catch {
    ElMessage.error('全部已读失败')
  }
}

onMounted(() => {
  fetchUnreadCount()
  fetchList()
})
</script>

<style scoped>
.notification-center {
  padding: var(--sh-space-lg);
}
/* .card-header 已在 App.vue 全局定义 */
.filter-form {
  margin-bottom: var(--sh-space-md);
}
/* .pagination-wrap 已在 App.vue 全局定义 */
</style>

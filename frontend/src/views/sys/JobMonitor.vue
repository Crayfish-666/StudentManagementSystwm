<template>
  <div class="job-monitor">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>定时任务监控</span>
        </div>
      </template>

      <!-- 任务列表 -->
      <el-table :data="jobs" stripe border style="width: 100%; margin-bottom: 24px">
        <el-table-column prop="name" label="任务名称" min-width="180" />
        <el-table-column prop="cron" label="Cron 表达式" width="140" />
        <el-table-column prop="description" label="说明" min-width="200" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleRun(row)" :loading="row.running">手动执行</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 执行记录筛选 -->
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="任务名称">
          <el-select v-model="filters.job_name" placeholder="全部" clearable style="width: 200px">
            <el-option
              v-for="j in jobs"
              :key="j.name"
              :label="j.name"
              :value="j.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 执行记录 -->
      <el-table :data="runs" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="job_name" label="任务名称" min-width="180" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="runStatusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration_ms" label="耗时(ms)" width="120" />
        <el-table-column prop="error_message" label="错误信息" min-width="200" show-overflow-tooltip />
        <el-table-column label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.started_at) }}
          </template>
        </el-table-column>
        <el-table-column label="完成时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.finished_at) }}
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
          @size-change="fetchRuns"
          @current-change="fetchRuns"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { jobApi } from '@/api/job'
import { formatDateTime as formatTime } from '@/utils/datetime'

// 4 个定时任务定义
const jobs = ref([
  { name: 'ty_overdue_warn', cron: '0 9 * * *', description: '团员培养超期预警', running: false },
  { name: 'qg_payroll_gen', cron: '0 2 1 * *', description: '勤工助学工资单生成', running: false },
  { name: 'sq_late_alert', cron: '30 22 * * *', description: '晚归告警', running: false },
  { name: 'cmp_recompute', cron: '0 2 * * *', description: '综合素质量化重算', running: false }
])

const filters = reactive({ job_name: '' })
const runs = ref([])
const loading = ref(false)

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const runStatusType = (status) => {
  const map = { running: 'warning', success: 'success', failed: 'danger' }
  return map[status] || 'info'
}

const fetchRuns = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.job_name) params.job_name = filters.job_name
    const data = await jobApi.listRuns(params)
    runs.value = data?.items || []
    pagination.total = data?.total || 0
  } catch {
    // 静默
  } finally {
    loading.value = false
  }
}

const handleRun = async (job) => {
  try {
    await ElMessageBox.confirm(`确定手动执行「${job.description}」？`, '确认', {
      confirmButtonText: '执行',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }

  job.running = true
  try {
    const data = await jobApi.runJob(job.name)
    ElMessage.success(`任务已触发，执行ID: ${data?.job_run_id || '-'}`)
    fetchRuns()
  } catch {
    ElMessage.error('任务执行失败')
  } finally {
    job.running = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchRuns()
}

const handleReset = () => {
  filters.job_name = ''
  pagination.page = 1
  fetchRuns()
}

onMounted(() => {
  fetchRuns()
})
</script>

<style scoped>
.job-monitor {
  padding: var(--sh-space-lg);
}
/* .card-header 已在 App.vue 全局定义 */
.filter-form {
  margin-bottom: var(--sh-space-md);
}
/* .pagination-wrap 已在 App.vue 全局定义 */
</style>

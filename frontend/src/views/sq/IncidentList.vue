<template>
  <div class="incident-list">
    <!-- 顶部筛选栏 -->
    <el-card class="filter-card" shadow="never">
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="事件等级">
          <el-select v-model="filters.incident_level" placeholder="全部" clearable style="width: 120px">
            <el-option label="L1" value="L1" />
            <el-option label="L2" value="L2" />
            <el-option label="L3" value="L3" />
            <el-option label="L4" value="L4" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="待处理" value="open" />
            <el-option label="处理中" value="processing" />
            <el-option label="已结案" value="closed" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
        </el-form-item>
        <el-form-item label="楼栋">
          <el-select v-model="filters.building_id" placeholder="全部" clearable style="width: 160px">
            <el-option
              v-for="b in buildings"
              :key="b.id"
              :label="b.name"
              :value="b.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 操作栏 -->
    <div class="action-bar">
      <el-button type="primary" @click="router.push('/sq/incident/new')">新增事件</el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" stripe border style="width: 100%">
      <el-table-column prop="biz_no" label="业务编号" min-width="150" />
      <el-table-column label="事件等级" min-width="100">
        <template #default="{ row }">
          <LevelBadge :level="row.incident_level" size="small" />
        </template>
      </el-table-column>
      <el-table-column prop="incident_type" label="事件类型" min-width="120" />
      <el-table-column prop="building_name" label="楼栋" min-width="120" />
      <el-table-column label="发生时间" min-width="170">
        <template #default="{ row }">
          {{ formatTime(row.occurred_at) }}
        </template>
      </el-table-column>
      <el-table-column prop="reporter_name" label="上报人" min-width="100" />
      <el-table-column label="状态" min-width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ statusMap[row.status] }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-width="240" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/sq/incident/${row.id}`)">查看详情</el-button>
          <el-button v-if="row.status === 'open' || row.status === 'processing'" link type="warning" @click="openHandleDialog(row)">处置</el-button>
          <el-button v-if="row.status === 'open' || row.status === 'processing'" link type="success" @click="openCloseDialog(row)">结案</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <!-- 处置弹窗 -->
    <el-dialog v-model="handleDialogVisible" title="事件处置" width="500px">
      <el-form :model="handleForm">
        <el-form-item label="处置说明">
          <el-input v-model="handleForm.action_text" type="textarea" :rows="4" placeholder="请输入处置说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitHandle">确认</el-button>
      </template>
    </el-dialog>

    <!-- 结案弹窗 -->
    <el-dialog v-model="closeDialogVisible" title="事件结案" width="500px">
      <el-form :model="closeForm">
        <el-form-item label="结案说明">
          <el-input v-model="closeForm.final_action" type="textarea" :rows="4" placeholder="请输入结案说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitClose">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { sqIncidentApi, sqBuildingApi } from '@/api/sq'
import LevelBadge from '@/components/LevelBadge.vue'
import { formatDateTime as formatTime } from '@/utils/datetime'

const router = useRouter()

// 状态字典
const statusMap = { open: '待处理', processing: '处理中', closed: '已结案', cancelled: '已取消' }

const statusTagType = (status) => {
  const map = { open: 'danger', processing: 'warning', closed: 'success', cancelled: 'info' }
  return map[status] || 'info'
}



// 筛选
const filters = reactive({
  incident_level: '',
  status: '',
  building_id: ''
})

// 楼栋列表
const buildings = ref([])

// 表格
const tableData = ref([])
const loading = ref(false)

// 分页
const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

// 处置弹窗
const handleDialogVisible = ref(false)
const handleForm = reactive({ action_text: '' })
const currentRow = ref(null)

// 结案弹窗
const closeDialogVisible = ref(false)
const closeForm = reactive({ final_action: '' })

const submitting = ref(false)

// 获取楼栋列表
const fetchBuildings = async () => {
  try {
    const data = await sqBuildingApi.list()
    buildings.value = data?.items || data || []
  } catch (e) {
    console.error('获取楼栋列表失败', e)
  }
}

// 获取列表
const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.incident_level) params.incident_level = filters.incident_level
    if (filters.status) params.status = filters.status
    if (filters.building_id) params.building_id = filters.building_id

    const data = await sqIncidentApi.list(params)
    tableData.value = data?.items || []
    pagination.total = data?.total || 0
  } catch (e) {
    console.error('获取事件列表失败', e)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  filters.incident_level = ''
  filters.status = ''
  filters.building_id = ''
  pagination.page = 1
  fetchList()
}

// 处置
const openHandleDialog = (row) => {
  currentRow.value = row
  handleForm.action_text = ''
  handleDialogVisible.value = true
}

const submitHandle = async () => {
  if (!handleForm.action_text.trim()) {
    ElMessage.warning('请输入处置说明')
    return
  }
  submitting.value = true
  try {
    await sqIncidentApi.handle(currentRow.value.id, { action_text: handleForm.action_text })
    ElMessage.success('处置成功')
    handleDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error('处置失败')
  } finally {
    submitting.value = false
  }
}

// 结案
const openCloseDialog = (row) => {
  currentRow.value = row
  closeForm.final_action = ''
  closeDialogVisible.value = true
}

const submitClose = async () => {
  if (!closeForm.final_action.trim()) {
    ElMessage.warning('请输入结案说明')
    return
  }
  submitting.value = true
  try {
    await sqIncidentApi.close(currentRow.value.id, { final_action: closeForm.final_action })
    ElMessage.success('结案成功')
    closeDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error('结案失败')
  } finally {
    submitting.value = false
  }
}

// 删除
const handleDelete = (row) => {
  ElMessageBox.confirm('确定删除该事件记录？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await sqIncidentApi.delete(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (e) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

onMounted(() => {
  fetchBuildings()
  fetchList()
})
</script>

<style scoped>
.incident-list {
  padding: var(--sh-space-lg);
}
.filter-card {
  margin-bottom: var(--sh-space-md);
}
/* .action-bar, .pagination-wrap 已在 App.vue 全局定义 */
</style>

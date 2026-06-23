<template>
  <div class="incident-detail">
    <el-page-header @back="router.back()" title="返回" content="事件详情" />

    <div v-loading="loading" style="margin-top: 20px">
      <!-- 基本信息卡片 -->
      <el-card shadow="never" class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件基本信息</span>
            <div class="action-buttons">
              <el-button
                v-if="detail.status === 'open' || detail.status === 'processing'"
                type="warning"
                @click="openHandleDialog"
              >处置</el-button>
              <el-button
                v-if="detail.status === 'open' || detail.status === 'processing'"
                type="success"
                @click="openCloseDialog"
              >结案</el-button>
            </div>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="业务编号">{{ detail.biz_no }}</el-descriptions-item>
          <el-descriptions-item label="事件等级">
            <LevelBadge :level="detail.incident_level" size="small" />
          </el-descriptions-item>
          <el-descriptions-item label="事件类型">{{ detail.incident_type }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(detail.status)">{{ statusMap[detail.status] }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="楼栋">{{ detail.building_name }}</el-descriptions-item>
          <el-descriptions-item label="发生时间">{{ formatTime(detail.occurred_at) }}</el-descriptions-item>
          <el-descriptions-item label="上报人">{{ detail.reporter_name }}</el-descriptions-item>
          <el-descriptions-item label="地点详情">{{ detail.location_detail || '-' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 关联学生 -->
      <el-card shadow="never" class="info-card">
        <template #header>关联学生</template>
        <el-table :data="detail.involved_students || []" stripe border style="width: 100%">
          <el-table-column prop="student_no" label="学号" />
          <el-table-column prop="name" label="姓名" />
        </el-table>
      </el-card>

      <!-- 处置记录时间线 -->
      <el-card shadow="never" class="info-card">
        <template #header>处置记录</template>
        <el-timeline v-if="detail.actions && detail.actions.length">
          <el-timeline-item
            v-for="(action, index) in detail.actions"
            :key="index"
            :timestamp="formatTime(action.action_at)"
            placement="top"
            :type="action.is_final ? 'success' : 'primary'"
          >
            <el-card shadow="never" class="timeline-card">
              <p class="action-text">{{ action.action_text }}</p>
              <p class="action-meta">
                <span>处置人：{{ action.action_name }}</span>
                <el-tag v-if="action.is_final" type="success" size="small" style="margin-left: 8px">结案</el-tag>
              </p>
            </el-card>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无处置记录" />
      </el-card>
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
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sqIncidentApi } from '@/api/sq'
import LevelBadge from '@/components/LevelBadge.vue'
import { formatDateTime as formatTime } from '@/utils/datetime'

const route = useRoute()
const router = useRouter()

const incidentId = route.params.id

// 状态字典
const statusMap = { open: '待处理', processing: '处理中', closed: '已结案', cancelled: '已取消' }

const statusTagType = (status) => {
  const map = { open: 'danger', processing: 'warning', closed: 'success', cancelled: 'info' }
  return map[status] || 'info'
}



// 详情数据
const detail = ref({})
const loading = ref(false)

const fetchDetail = async () => {
  loading.value = true
  try {
    const data = await sqIncidentApi.get(incidentId)
    detail.value = data || {}
  } catch (e) {
    console.error('获取事件详情失败', e)
  } finally {
    loading.value = false
  }
}

// 处置弹窗
const handleDialogVisible = ref(false)
const handleForm = reactive({ action_text: '' })

// 结案弹窗
const closeDialogVisible = ref(false)
const closeForm = reactive({ final_action: '' })

const submitting = ref(false)

const openHandleDialog = () => {
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
    await sqIncidentApi.handle(incidentId, { action_text: handleForm.action_text })
    ElMessage.success('处置成功')
    handleDialogVisible.value = false
    fetchDetail()
  } catch (e) {
    ElMessage.error('处置失败')
  } finally {
    submitting.value = false
  }
}

const openCloseDialog = () => {
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
    await sqIncidentApi.close(incidentId, { final_action: closeForm.final_action })
    ElMessage.success('结案成功')
    closeDialogVisible.value = false
    fetchDetail()
  } catch (e) {
    ElMessage.error('结案失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
.incident-detail {
  padding: var(--sh-space-lg);
}
.info-card {
  margin-bottom: var(--sh-space-md);
}
/* .card-header 已在 App.vue 全局定义 */
.action-buttons {
  display: flex;
  gap: var(--sh-space-sm);
}
.timeline-card {
  margin-bottom: 0;
}
.action-text {
  margin: 0 0 var(--sh-space-sm) 0;
}
.action-meta {
  margin: 0;
  color: var(--sh-text-secondary);
  font-size: var(--sh-text-sm);
}
</style>

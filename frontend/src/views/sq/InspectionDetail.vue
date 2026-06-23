<template>
  <div class="inspection-detail">
    <el-page-header @back="router.back()" title="返回" content="巡查详情" />

    <div v-loading="loading" style="margin-top: 20px">
      <!-- 基本信息 -->
      <el-card shadow="never" class="info-card">
        <template #header>
          <div class="card-header">
            <span>巡查基本信息</span>
            <div class="action-buttons">
              <el-button type="primary" @click="router.push('/sq/inspection')">返回列表</el-button>
            </div>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="业务编号">{{ detail.biz_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="巡查类型">{{ detail.inspection_type_text || detail.inspection_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="楼栋">{{ detail.building_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="楼层">{{ detail.floor_no != null ? detail.floor_no + ' 层' : '-' }}</el-descriptions-item>
          <el-descriptions-item label="寝室">{{ detail.room_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="巡查人">{{ detail.inspector_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="巡查时间">{{ formatTime(detail.inspected_at) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(detail.status)">{{ detail.status_text || statusMap[detail.status] || detail.status || '-' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="分数">
            <el-tag v-if="detail.score != null" :type="scoreTagType(detail.score)">{{ detail.score }}</el-tag>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="巡查摘要" :span="2">
            <span class="summary-text">{{ detail.summary || '-' }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 扣分项 -->
      <el-card shadow="never" class="info-card">
        <template #header>扣分项明细</template>
        <el-table
          v-if="detail.deductions && detail.deductions.length"
          :data="detail.deductions"
          stripe
          border
          style="width: 100%"
        >
          <el-table-column type="index" label="#" width="60" align="center" />
          <el-table-column prop="item" label="扣分项" min-width="220" />
          <el-table-column prop="deduction" label="扣分" width="120" align="center">
            <template #default="{ row }">
              <el-tag type="danger">-{{ row.deduction }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="本次巡查无扣分项" />
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sqInspectionApi } from '@/api/sq'
import { formatDateTime as formatTime } from '@/utils/datetime'

const route = useRoute()
const router = useRouter()

const inspectionId = route.params.id

// 状态字典
const statusMap = { draft: '草稿', submitted: '已提交' }

const statusTagType = (status) => {
  const map = { draft: 'info', submitted: 'success' }
  return map[status] || 'info'
}

const scoreTagType = (score) => {
  if (score >= 90) return 'success'
  if (score >= 60) return 'warning'
  return 'danger'
}

// 详情数据
const detail = ref({})
const loading = ref(false)

const fetchDetail = async () => {
  loading.value = true
  try {
    const data = await sqInspectionApi.get(inspectionId)
    detail.value = data || {}
  } catch (e) {
    console.error('获取巡查详情失败', e)
    ElMessage.error('获取巡查详情失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
.inspection-detail {
  padding: var(--sh-space-lg);
}
.info-card {
  margin-bottom: var(--sh-space-md);
}
.action-buttons {
  display: flex;
  gap: var(--sh-space-sm);
}
.summary-text {
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--sh-text-secondary);
}
</style>

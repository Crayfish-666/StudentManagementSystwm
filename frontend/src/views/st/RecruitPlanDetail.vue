<template>
  <div class="page-container" v-loading="loading">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>招新计划详情</span>
          <el-button @click="goBack">返回</el-button>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="编号">{{ plan.biz_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType[plan.status]" size="small">{{ plan.status_text }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="所属社团">{{ plan.association_name }}</el-descriptions-item>
        <el-descriptions-item label="招新季节">{{ plan.season_text }}</el-descriptions-item>
        <el-descriptions-item label="学年">{{ plan.academic_year }}</el-descriptions-item>
        <el-descriptions-item label="目标人数">{{ plan.target_count }}</el-descriptions-item>
        <el-descriptions-item label="考核方式" :span="2">{{ plan.assessment_method || '-' }}</el-descriptions-item>
        <el-descriptions-item label="面试时间">{{ plan.interview_at ? formatDateTime(plan.interview_at) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="结果录入期限">{{ plan.result_deadline || '-' }}</el-descriptions-item>
        <el-descriptions-item label="招新状态">
          <el-tag v-if="plan.is_finished === 1" type="info">已结束（{{ plan.finished_at ? formatDateTime(plan.finished_at) : '' }}）</el-tag>
          <el-tag v-else type="success">招新中</el-tag>
          <span v-if="plan.finished_reason" class="finish-reason">原因：{{ plan.finished_reason }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="投递数 / 录用数">
          <el-tag>{{ plan.apply_count || 0 }}</el-tag> /
          <el-tag type="success">{{ plan.accepted_count || 0 }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(plan.created_at) }}</el-descriptions-item>
      </el-descriptions>

      <div style="margin-top: 20px">
        <el-button v-if="plan.status === 'S0'" type="success" @click="handleSubmit">提交</el-button>
        <el-button v-if="plan.status === 'S1'" type="warning" @click="handleWithdraw">撤回</el-button>
        <el-button v-if="plan.status === 'S1'" type="success" @click="handleApprove">审批通过</el-button>
        <el-button v-if="plan.status === 'S1'" type="danger" @click="handleReject">驳回</el-button>
        <el-button v-if="plan.status === 'S3'" type="primary" @click="handlePublish">发布</el-button>
        <el-button v-if="plan.status === 'S3' && plan.is_finished !== 1" type="warning" @click="handleFinish">结束招新</el-button>
        <el-button v-if="plan.status === 'S3'" type="primary" @click="goApplies">申请列表</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { stRecruitPlanApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'

const route = useRoute()
const router = useRouter()

const planId = computed(() => Number(route.params.id))
const plan = ref({})
const loading = ref(false)
const statusType = { S0: 'info', S1: 'warning', S3: 'success', S4: 'danger' }

async function fetchDetail() {
  loading.value = true
  try {
    const r = await stRecruitPlanApi.get(planId.value)
    plan.value = r
  } catch (e) {
    ElMessage.error('加载详情失败')
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push('/st/recruit-plan')
}
function goApplies() {
  router.push({ path: '/st/recruit-apply', query: { plan_id: planId.value } })
}

async function handleSubmit() {
  await ElMessageBox.confirm('确认提交该招新计划？', '提示', { type: 'warning' })
  await stRecruitPlanApi.submit(planId.value)
  ElMessage.success('已提交')
  fetchDetail()
}
async function handleWithdraw() {
  await ElMessageBox.confirm('确认撤回？', '提示', { type: 'warning' })
  await stRecruitPlanApi.withdraw(planId.value)
  ElMessage.success('已撤回')
  fetchDetail()
}
async function handleApprove() {
  await ElMessageBox.confirm('确认审批通过？', '提示', { type: 'success' })
  await stRecruitPlanApi.approve(planId.value)
  ElMessage.success('已通过')
  fetchDetail()
}
async function handleReject() {
  const { value: opinion } = await ElMessageBox.prompt('请输入驳回意见（至少 10 字）', '驳回', {
    confirmButtonText: '确认驳回',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputValidator: (v) => (v && v.trim().length >= 10) || '驳回意见至少 10 字'
  })
  await stRecruitPlanApi.reject(planId.value, { opinion })
  ElMessage.success('已驳回')
  fetchDetail()
}
async function handlePublish() {
  await ElMessageBox.confirm('确认发布？发布后将开启学生投递通道。', '提示', { type: 'success' })
  await stRecruitPlanApi.publish(planId.value)
  ElMessage.success('已发布')
  fetchDetail()
}
async function handleFinish() {
  const { value: reason } = await ElMessageBox.prompt('请输入结束原因（可空，不超过 200 字）', '提前结束招新', {
    confirmButtonText: '下一步',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputPlaceholder: '例如：招新人数已满足，提前结束',
    inputValidator: (v) => !v || v.trim().length <= 200 || '结束原因不超过 200 字'
  })
  await ElMessageBox.confirm('结束操作不可逆，确认提前结束该招新？结束后学生不可再投递。', '二次确认', { type: 'warning' })
  await stRecruitPlanApi.finish(planId.value, { reason: reason || '' })
  ElMessage.success('已结束招新')
  fetchDetail()
}

onMounted(fetchDetail)
</script>

<style scoped>
.page-container { padding: 16px; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.finish-reason { margin-left: 12px; color: #909399; font-size: 13px; }
</style>

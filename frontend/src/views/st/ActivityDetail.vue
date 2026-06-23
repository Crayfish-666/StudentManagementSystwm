<template>
  <div class="page-container">
    <!-- 活动基本信息 -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>活动详情</span>
          <div>
            <!-- S0: 编辑、提交 -->
            <el-button v-if="act && act.status === 'S0'" type="primary" size="small" @click="goEdit">编辑</el-button>
            <el-button v-if="act && act.status === 'S0'" type="success" size="small" @click="handleSubmit">提交</el-button>
            <!-- S1: 撤回 -->
            <el-button v-if="act && act.status === 'S1'" type="warning" size="small" @click="handleWithdraw">撤回</el-button>
            <!-- S2: 审批 -->
            <el-button v-if="act && act.status === 'S2'" type="success" size="small" @click="openApprove">审批</el-button>
            <!-- S3: 签到、提交总结 -->
            <el-button v-if="act && act.status === 'S3'" type="primary" size="small" @click="goCheckin">签到</el-button>
            <el-button v-if="act && act.status === 'S3' && !summaryData" type="success" size="small" @click="goSummary">提交总结</el-button>
            <el-button @click="goBack">返回</el-button>
          </div>
        </div>
      </template>

      <el-descriptions :column="2" border v-if="act">
        <el-descriptions-item label="编号" :span="2">{{ act.biz_no }}</el-descriptions-item>
        <el-descriptions-item label="活动名称">{{ act.title }}</el-descriptions-item>
        <el-descriptions-item label="所属社团">{{ act.association_name }}</el-descriptions-item>
        <el-descriptions-item label="活动等级">
          <el-tag :type="levelType[act.level]" size="small">{{ act.level }}级</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType[act.status]" size="small">{{ act.status_text }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="预计参与人数">{{ act.expected_participants }} 人</el-descriptions-item>
        <el-descriptions-item label="预算">{{ (act.budget_cents / 100).toFixed(2) }} 元</el-descriptions-item>
        <el-descriptions-item label="活动地点" :span="2">{{ act.location }}</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatDateTime(act.started_at) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatDateTime(act.ended_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(act.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDateTime(act.updated_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 审批时间线 -->
    <el-card shadow="never" class="section-card">
      <template #header>
        <span>审批时间线</span>
      </template>
      <el-timeline v-if="approvalRecords.length > 0">
        <el-timeline-item
          v-for="rec in approvalRecords"
          :key="rec.event_id"
          :type="rec.payload?.result === 'approve' ? 'success' : 'danger'"
          :timestamp="formatDateTime(rec.occurred_at)"
          placement="top"
        >
          <el-card shadow="hover" class="rec-card">
            <div class="rec-head">
              <el-tag :type="rec.payload?.result === 'approve' ? 'success' : 'danger'" size="small">
                {{ stepTextMap[rec.payload?.step_no] || `步骤${rec.payload?.step_no}` }} · {{ rec.payload?.result === 'approve' ? '通过' : '驳回' }}
              </el-tag>
              <span class="rec-meta">{{ rec.payload?.approver }}</span>
            </div>
            <div v-if="rec.payload?.opinion" class="rec-opinion">{{ rec.payload?.opinion }}</div>
          </el-card>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else description="暂无审批记录" :image-size="80" />
    </el-card>

    <!-- 签到记录 & 活动总结 -->
    <el-tabs v-if="act" class="section-card" type="border-card">
      <el-tab-pane label="签到记录">
        <CheckinTable
          :items="checkinList"
          :loading="checkinLoading"
          :total="checkinTotal"
          :page="checkinPage"
          :page-size="checkinPageSize"
          @change="onCheckinPageChange"
        />
      </el-tab-pane>
      <el-tab-pane label="活动总结">
        <div v-if="summaryData">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="实际参与人数">{{ summaryData.participants }} 人</el-descriptions-item>
            <el-descriptions-item label="目标达成度">
              <el-rate v-model="summaryDisplayScore" disabled show-score :max="5" />
            </el-descriptions-item>
            <el-descriptions-item label="改进建议">
              <div class="suggestions-text">{{ summaryData.improvements }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="提交时间">{{ formatDateTime(summaryData.submitted_at) }}</el-descriptions-item>
          </el-descriptions>
          <div class="action-bar">
            <el-button type="primary" size="small" @click="goSummary">编辑总结</el-button>
          </div>
        </div>
        <div v-else>
          <el-empty description="暂未提交总结" :image-size="80">
            <el-button v-if="act.status === 'S3'" type="primary" @click="goSummary">提交总结</el-button>
          </el-empty>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 审批对话框 -->
    <el-dialog
      v-model="approveDialogVisible"
      title="审批活动"
      width="520px"
      :close-on-click-modal="false"
      @close="onApproveDialogClose"
    >
      <el-form ref="approveFormRef" :model="approveForm" :rules="approveRules" label-width="100px">
        <el-form-item label="审批结果" prop="result">
          <el-radio-group v-model="approveForm.result">
            <el-radio value="approve">通过</el-radio>
            <el-radio value="reject">驳回</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="审批意见" prop="opinion">
          <el-input
            v-model="approveForm.opinion"
            type="textarea"
            :rows="5"
            placeholder="请输入审批意见"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveDialogVisible = false">取消</el-button>
        <el-button
          :type="approveForm.result === 'approve' ? 'success' : 'danger'"
          :loading="approveSubmitting"
          @click="handleApprove"
        >
          提交审批
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { stActivityApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'
import CheckinTable from './components/CheckinTable.vue'

const router = useRouter()
const route = useRoute()

const statusType = { S0: 'info', S1: 'warning', S2: '', S3: 'success', S4: 'danger', cancelled: 'info' }
const levelType = { A: 'danger', B: 'warning', C: '', D: 'success' }
const stepTextMap = {
  1: '指导教师审批',
  2: '院系审批',
  3: '校社联审批',
  4: '校团委审批',
  5: '校领导审批'
}

const act = ref(null)

// 审批时间线
const timelineRecords = ref([])

// 仅审批事件（过滤掉创建/提交/撤回等非审批事件）
const approvalRecords = computed(() =>
  timelineRecords.value.filter((r) =>
    ['StActivityApproved', 'StActivityRejected'].includes(r.event_type)
  )
)

// 签到记录
const checkinList = ref([])
const checkinLoading = ref(false)
const checkinTotal = ref(0)
const checkinPage = ref(1)
const checkinPageSize = ref(20)

// 活动总结
const summaryData = ref(null)
const summaryDisplayScore = computed(() => {
  if (!summaryData.value) return 0
  return summaryData.value.goal_score || 0
})

// 审批对话框
const approveDialogVisible = ref(false)
const approveSubmitting = ref(false)
const approveFormRef = ref(null)
const approveForm = reactive({
  result: 'approve',
  opinion: ''
})

const opinionMinLen = (rule, value, callback) => {
  if (!value || value.trim().length === 0) {
    callback(new Error('请填写审批意见'))
  } else if (approveForm.result === 'reject' && value.trim().length < 30) {
    callback(new Error('驳回时审批意见至少 30 字'))
  } else {
    callback()
  }
}

const approveRules = {
  result: [{ required: true, message: '请选择审批结果' }],
  opinion: [{ required: true, validator: opinionMinLen, trigger: 'blur' }]
}

async function fetchDetail() {
  try {
    act.value = await stActivityApi.get(route.params.id)
  } catch (e) {
    console.error('获取活动详情失败', e)
  }
}

async function fetchTimeline() {
  try {
    const data = await stActivityApi.timeline(route.params.id)
    timelineRecords.value = data || []
  } catch (e) {
    console.error('获取审批时间线失败', e)
  }
}

async function loadCheckins() {
  checkinLoading.value = true
  try {
    const data = await stActivityApi.listCheckins(route.params.id, {
      page: checkinPage.value,
      page_size: checkinPageSize.value
    })
    checkinList.value = data.items || []
    checkinTotal.value = data.total || 0
  } catch (e) {
    console.error('获取签到记录失败', e)
  } finally {
    checkinLoading.value = false
  }
}

function onCheckinPageChange({ page, pageSize }) {
  checkinPage.value = page
  checkinPageSize.value = pageSize
  loadCheckins()
}

async function fetchSummary() {
  try {
    const data = await stActivityApi.getSummary(route.params.id)
    if (data && data.id) {
      summaryData.value = data
    }
  } catch (e) {
    // 404 表示尚未提交总结，忽略
  }
}

function goEdit() {
  router.push(`/st/activity/${route.params.id}/edit`)
}

function goCheckin() {
  router.push(`/st/activity/${route.params.id}/checkin`)
}

function goSummary() {
  router.push(`/st/activity/${route.params.id}/summary`)
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/st/activity')
  }
}

async function handleSubmit() {
  try {
    await ElMessageBox.confirm('确认提交此活动？提交后将进入审批流程。', '提交确认')
    await stActivityApi.submit(route.params.id)
    ElMessage.success('提交成功')
    fetchDetail()
    fetchTimeline()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

async function handleWithdraw() {
  try {
    await ElMessageBox.confirm('确认撤回此活动？撤回后将回到草稿状态。', '撤回确认', { type: 'warning' })
    await stActivityApi.withdraw(route.params.id)
    ElMessage.success('已撤回')
    fetchDetail()
    fetchTimeline()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

function openApprove() {
  approveForm.result = 'approve'
  approveForm.opinion = ''
  approveDialogVisible.value = true
}

function onApproveDialogClose() {
  approveFormRef.value?.clearValidate()
}

async function handleApprove() {
  if (!approveFormRef.value) return
  const valid = await approveFormRef.value.validate().catch(() => false)
  if (!valid) return

  if (approveForm.result === 'reject') {
    try {
      await ElMessageBox.confirm('确认驳回此活动？驳回后活动将终止。', '驳回确认', { type: 'warning' })
    } catch {
      return
    }
  }

  approveSubmitting.value = true
  try {
    await stActivityApi.approve(route.params.id, {
      step_no: act.value.current_step_no,
      result: approveForm.result,
      opinion: approveForm.opinion
    })
    ElMessage.success('审批已提交')
    approveDialogVisible.value = false
    fetchDetail()
    fetchTimeline()
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    approveSubmitting.value = false
  }
}

onMounted(() => {
  fetchDetail()
  fetchTimeline()
  loadCheckins()
  fetchSummary()
})
</script>

<style scoped>
/* .card-header, .action-bar 已在 App.vue 全局定义 */
.section-card {
  margin-top: var(--sh-space-md);
}
.rec-card {
  margin-bottom: var(--sh-space-xs);
}
.rec-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--sh-space-sm);
  margin-bottom: var(--sh-space-xs);
}
.rec-meta {
  color: var(--sh-text-regular);
  font-size: var(--sh-text-sm);
}
.rec-opinion {
  white-space: pre-wrap;
  word-break: break-all;
  line-height: var(--sh-leading-normal);
  color: var(--sh-text-primary);
}
.suggestions-text {
  white-space: pre-wrap;
  word-break: break-all;
  line-height: var(--sh-leading-normal);
}
</style>

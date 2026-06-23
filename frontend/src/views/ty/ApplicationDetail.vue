<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>入团申请详情</span>
          <div class="header-actions">
            <el-button v-if="canEdit" type="primary" @click="goEdit">编辑</el-button>
            <el-button v-if="canSubmit" type="success" @click="handleSubmit">提交申请</el-button>
            <el-button v-if="canWithdraw" type="warning" @click="handleWithdraw">撤回申请</el-button>
            <el-button
              v-if="approvableStep"
              :type="approvableStep === 'school' ? 'danger' : 'success'"
              @click="openApprove"
            >
              {{ approvableLabel }}
            </el-button>
            <el-button @click="goBack">返回列表</el-button>
          </div>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="业务编号">{{ app.biz_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType[app.status]" size="small">
            {{ statusMap[app.status] || app.status }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="申请人">{{ app.student_name }}</el-descriptions-item>
        <el-descriptions-item label="学号">{{ app.student_no }}</el-descriptions-item>
        <el-descriptions-item label="团支部">{{ app.branch_name }}</el-descriptions-item>
        <el-descriptions-item label="院系">{{ app.college_name }}</el-descriptions-item>
        <el-descriptions-item label="申请日期">{{ app.apply_date }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(app.created_at) }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">思想政治表现自述</el-divider>
      <div class="statement-content">{{ app.self_statement }}</div>

      <template v-if="app.family_members_json">
        <el-divider content-position="left">家庭成员信息</el-divider>
        <el-table :data="parsedFamilyMembers" border size="small" style="max-width: 500px">
          <el-table-column prop="name" label="姓名" width="100" />
          <el-table-column prop="relation" label="关系" width="80" />
          <el-table-column prop="political_status" label="政治面貌">
            <template #default="{ row }">{{ row.political_status }}</template>
          </el-table-column>
        </el-table>
      </template>

      <template v-if="app.rewards_punishments">
        <el-divider content-position="left">奖惩情况</el-divider>
        <div class="statement-content">{{ app.rewards_punishments }}</div>
      </template>

      <el-divider content-position="left">发展轨迹</el-divider>
      <div v-if="loadingTrack" style="padding: 20px 0">
        <el-skeleton :rows="4" animated />
      </div>
      <DevelopmentTrack v-else :track="trackData" />

      <div class="track-actions">
        <el-button type="primary" @click="goTrackPage">查看完整发展轨迹</el-button>
      </div>
    </el-card>

    <ApprovalDialog
      v-model="dialogVisible"
      :application="app"
      :step="approvableStep"
      @success="onApproveSuccess"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { tyApplicationApi } from '@/api/ty'
import { useAuthStore } from '@/stores/auth'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import DevelopmentTrack from '@/components/DevelopmentTrack.vue'
import { formatDateTime as formatTime } from '@/utils/datetime'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const appId = Number(route.params.id)

const statusMap = { S0: '草稿', S1: '待审', S2: '审批中', S3: '通过', S4: '驳回' }
const statusType = { S0: 'info', S1: 'warning', S2: 'warning', S3: 'success', S4: 'danger' }

const app = ref({})
const dialogVisible = ref(false)
const trackData = ref({ student_name: '', political_status: '', political_status_text: '', entries: [] })
const loadingTrack = ref(true)

const isCreator = computed(() => {
  return authStore.user && app.value && app.value.created_by === authStore.user.id
})
const canEdit = computed(() => app.value.status === 'S0' && isCreator.value)
const canSubmit = computed(() => app.value.status === 'S0' && isCreator.value)
const canWithdraw = computed(() => app.value.status === 'S1' && isCreator.value)

// 解析家庭成员 JSON
const parsedFamilyMembers = computed(() => {
  try {
    const raw = app.value.family_members_json
    if (!raw) return []
    return typeof raw === 'string' ? JSON.parse(raw) : raw
  } catch {
    return []
  }
})

// 计算当前用户可执行的审批步骤
const approvableStep = computed(() => {
  const roles = authStore.roles || []
  const status = app.value.status
  if (status === 'S1') {
    if (
      roles.includes('R-COL-COUN') ||
      roles.includes('R-COL-LEAGUE') ||
      roles.includes('R-SY-ADMIN')
    ) {
      return 'counselor'
    }
  }
  if (status === 'S2') {
    // 已有 college 通过 → 校级审；否则院系审
    const hasCollegeApproved = approvalRecords.value.some(
      (r) => r.step === 'college' && r.result === 'approve'
    )
    if (hasCollegeApproved) {
      if (roles.includes('R-SY-LEAGUE') || roles.includes('R-SY-ADMIN')) return 'school'
    } else {
      if (roles.includes('R-COL-LEAGUE') || roles.includes('R-SY-ADMIN')) return 'college'
    }
  }
  return ''
})

const approvableLabel = computed(() => {
  switch (approvableStep.value) {
    case 'counselor':
      return '辅导员/团支部初审'
    case 'college':
      return '院系团委复核'
    case 'school':
      return '校团委终审'
    default:
      return ''
  }
})

async function fetchDetail() {
  try {
    const data = await tyApplicationApi.get(appId)
    app.value = data || {}
    // 获取详情后加载发展轨迹
    fetchTrack()
  } catch (e) {
    console.error('获取申请详情失败', e)
  }
}

function goEdit() {
  router.push(`/ty/application/${appId}/edit`)
}
function goBack() {
  router.push('/ty/application')
}

async function handleSubmit() {
  try {
    await ElMessageBox.confirm('确认提交此申请？提交后将进入审批流程。', '提交确认')
    await tyApplicationApi.submit(appId)
    ElMessage.success('提交成功')
    fetchDetail()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

async function handleWithdraw() {
  try {
    const { value } = await ElMessageBox.prompt('请输入撤回原因', '撤回申请', {
      confirmButtonText: '确认撤回',
      cancelButtonText: '取消',
      inputPlaceholder: '请说明撤回原因'
    })
    await tyApplicationApi.withdraw(appId, value || '')
    ElMessage.success('已撤回')
    fetchDetail()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

function openApprove() {
  dialogVisible.value = true
}

function onApproveSuccess() {
  fetchDetail()
  fetchTrack()
}

async function fetchTrack() {
  loadingTrack.value = true
  try {
    const studentId = app.value.student_id
    if (!studentId) {
      loadingTrack.value = false
      return
    }
    const data = await tyApplicationApi.developmentTrack(studentId)
    trackData.value = data || { student_name: '', political_status: '', political_status_text: '', entries: [] }
  } catch (e) {
    // 发展轨迹为增强功能，加载失败不影响主流程（后端可能未重启）
    console.debug('[DevelopmentTrack] 轨迹数据暂不可用', e.message)
  } finally {
    loadingTrack.value = false
  }
}

function goTrackPage() {
  const studentId = app.value.student_id
  if (studentId) {
    router.push(`/ty/students/${studentId}/development-track`)
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
/* .card-header / .statement-content 已在全局定义 */
.header-actions {
  display: flex;
  gap: var(--sh-space-sm);
  flex-wrap: wrap;
}
.track-actions {
  margin-top: var(--sh-space-md);
  text-align: center;
}
</style>

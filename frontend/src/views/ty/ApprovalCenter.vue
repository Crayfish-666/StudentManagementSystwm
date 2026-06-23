<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="header">
          <span class="title">入团申请审批中心</span>
          <el-tag size="small" type="info">{{ roleLabel }}</el-tag>
        </div>
      </template>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="待我审批" name="pending">
          <el-table v-loading="loading" :data="items" stripe>
            <el-table-column prop="biz_no" label="申请编号" width="160" />
            <el-table-column prop="student_name" label="申请人" width="100" />
            <el-table-column prop="student_no" label="学号" width="120" />
            <el-table-column prop="college_name" label="院系" width="160" />
            <el-table-column prop="branch_name" label="团支部" min-width="160" />
            <el-table-column prop="apply_date" label="申请日期" width="120" />
            <el-table-column label="当前阶段" width="200">
              <template #default="{ row }">
                <el-tag size="small" :type="statusType[row.status]">{{ row.status_text }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="240" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link @click="goDetail(row.id)">查看</el-button>
                <el-button
                  v-if="canApprove(row)"
                  type="success"
                  size="small"
                  @click="openApprove(row)"
                >
                  审批
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-if="total > 0"
            class="pager"
            :current-page="page"
            :page-size="pageSize"
            :total="total"
            layout="total, prev, pager, next"
            @current-change="handlePageChange"
          />
        </el-tab-pane>

        <el-tab-pane label="历史" name="history">
          <el-table v-loading="loading" :data="historyItems" stripe>
            <el-table-column prop="biz_no" label="申请编号" width="160" />
            <el-table-column prop="student_name" label="申请人" width="100" />
            <el-table-column prop="college_name" label="院系" width="160" />
            <el-table-column prop="branch_name" label="团支部" min-width="160" />
            <el-table-column label="状态" width="120">
              <template #default="{ row }">
                <el-tag size="small" :type="statusType[row.status]">{{ statusMap[row.status] }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="updated_at" label="更新时间" width="180">
              <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link @click="goDetail(row.id)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-if="historyTotal > 0"
            class="pager"
            :current-page="historyPage"
            :page-size="pageSize"
            :total="historyTotal"
            layout="total, prev, pager, next"
            @current-change="handleHistoryPageChange"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <ApprovalDialog
      v-model="dialogVisible"
      :application="currentApp"
      :step="currentStep"
      @success="onApproveSuccess"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { tyApplicationApi } from '@/api/ty'
import { useAuthStore } from '@/stores/auth'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import { formatDateTime as formatTime } from '@/utils/datetime'

const router = useRouter()
const authStore = useAuthStore()

const activeTab = ref('pending')
const loading = ref(false)

const items = ref([])
const total = ref(0)
const page = ref(1)

const historyItems = ref([])
const historyTotal = ref(0)
const historyPage = ref(1)

const pageSize = 20

const dialogVisible = ref(false)
const currentApp = ref(null)
const currentStep = ref('')

const statusMap = { S0: '草稿', S1: '待审', S2: '审批中', S3: '通过', S4: '驳回' }
const statusType = { S0: 'info', S1: 'warning', S2: 'warning', S3: 'success', S4: 'danger' }

const roleLabel = computed(() => {
  const roles = authStore.roles
  if (roles.includes('R-SY-ADMIN')) return '系统管理员（全权审批）'
  if (roles.includes('R-SY-LEAGUE')) return '校团委（终审）'
  if (roles.includes('R-COL-LEAGUE')) return '院系团委（复核 / 初审）'
  if (roles.includes('R-COL-COUN')) return '辅导员（初审）'
  return '审批人员'
})

function canApprove(row) {
  // 后端已按角色 + 步骤过滤 pending 列表，这里仅做兜底
  return ['S1', 'S2'].includes(row.status)
}

function nextStepFor(row) {
  const roles = authStore.roles
  if (row.status === 'S1') return 'counselor'
  if (row.status === 'S2') {
    // 没有 college 通过：当前应为 college；否则 school
    // 但前端无法直接判断，采用回退优先 college；admin/sy_league 视角则 school
    if (roles.includes('R-SY-LEAGUE') || roles.includes('R-SY-ADMIN')) return 'school'
    return 'college'
  }
  return ''
}

async function loadPending(p = 1) {
  loading.value = true
  try {
    const data = await tyApplicationApi.listPending({ page: p, page_size: pageSize })
    items.value = data?.items || []
    total.value = data?.total || 0
    page.value = data?.page || p
  } finally {
    loading.value = false
  }
}

async function loadHistory(p = 1) {
  loading.value = true
  try {
    // 历史：S3 + S4 合并查询（两次请求）
    const [done, rejected] = await Promise.all([
      tyApplicationApi.list({ status: 'S3', page: p, page_size: pageSize }),
      tyApplicationApi.list({ status: 'S4', page: p, page_size: pageSize })
    ])
    const merged = [...(done?.items || []), ...(rejected?.items || [])]
    merged.sort((a, b) => (a.updated_at < b.updated_at ? 1 : -1))
    historyItems.value = merged
    historyTotal.value = (done?.total || 0) + (rejected?.total || 0)
    historyPage.value = p
  } finally {
    loading.value = false
  }
}

function handleTabChange(name) {
  if (name === 'pending') loadPending(1)
  else loadHistory(1)
}

function handlePageChange(p) {
  loadPending(p)
}
function handleHistoryPageChange(p) {
  loadHistory(p)
}

function goDetail(id) {
  router.push(`/ty/application/${id}`)
}

function openApprove(row) {
  currentApp.value = row
  currentStep.value = nextStepFor(row)
  dialogVisible.value = true
}

function onApproveSuccess() {
  loadPending(page.value)
}

onMounted(() => {
  loadPending(1)
})
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--sh-text-xl);
  font-weight: 600;
  color: var(--sh-text-primary);
}
.title {
  font-size: var(--sh-text-xl);
  font-weight: 600;
  color: var(--sh-text-primary);
}
.pager {
  margin-top: var(--sh-space-md);
  text-align: right;
  justify-content: flex-end;
  display: flex;
}
</style>

<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>推优大会列表</span>
          <el-button type="primary" @click="goCreate">新建大会</el-button>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select ref="decisionSelectRef" v-model="filterDecision" placeholder="决议结果筛选" clearable style="width: 160px" @change="fetchList">
          <el-option label="通过" value="pass" />
          <el-option label="不通过" value="reject" />
        </el-select>
        <el-select ref="branchSelectRef" v-model="filterBranch" placeholder="团支部筛选" clearable style="width: 180px" @change="fetchList">
          <el-option v-for="b in branches" :key="b.id" :label="b.name" :value="b.id" />
        </el-select>
      </div>

      <el-table :data="list" stripe v-loading="loading" :key="'tbl-' + list.length + '-' + total" class="rd-table" table-layout="auto">
        <el-table-column prop="biz_no" label="业务编号" min-width="190" />
        <el-table-column prop="student_name" label="申请人" min-width="120" />
        <el-table-column prop="meeting_at" label="会议时间" min-width="200">
          <template #default="{ row }">{{ formatDateTime(row.meeting_at) }}</template>
        </el-table-column>
        <el-table-column prop="location" label="地点" min-width="160" />
        <el-table-column label="实到/应到" min-width="140">
          <template #default="{ row }">
            {{ row.actual_count }} / {{ row.expected_count }}
          </template>
        </el-table-column>
        <el-table-column prop="decision" label="决议" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.decision === 'pass' ? 'success' : 'danger'" size="small">
              {{ row.decision === 'pass' ? '通过' : '不通过' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="120">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">查看详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 详情弹窗：克制的现代工业风 -->
    <el-dialog v-model="detailVisible" width="640px" destroy-on-close>
      <template #header>
        <div class="rd-head">
          <span class="rd-head-title">推优大会详情</span>
          <span class="rd-head-bizno">{{ currentDetail?.biz_no }}</span>
        </div>
      </template>

      <div v-if="currentDetail" class="rd">
        <!-- 申请人主信息 -->
        <div class="rd-applicant">
          <div class="rd-eyebrow">申请人 / Applicant</div>
          <div class="rd-name">{{ currentDetail.student_name }}</div>
        </div>

        <div class="rd-rule" />

        <!-- 元信息：2 列 grid -->
        <div class="rd-grid">
          <div class="rd-field">
            <div class="rd-eyebrow">会议时间</div>
            <div class="rd-value">{{ formatDateTime(currentDetail.meeting_at) }}</div>
          </div>
          <div class="rd-field">
            <div class="rd-eyebrow">会议地点</div>
            <div class="rd-value">{{ currentDetail.location || '—' }}</div>
          </div>
          <div class="rd-field">
            <div class="rd-eyebrow">应到人数</div>
            <div class="rd-value">{{ currentDetail.expected_count }}</div>
          </div>
          <div class="rd-field">
            <div class="rd-eyebrow">实到人数</div>
            <div class="rd-value">{{ currentDetail.actual_count }}</div>
          </div>
        </div>

        <div class="rd-rule" />

        <!-- 投票结果：3 列 metric -->
        <div class="rd-eyebrow">投票结果</div>
        <div class="rd-votes">
          <div class="rd-vote">
            <div class="rd-vote-num">{{ currentDetail.vote?.approve_count ?? 0 }}</div>
            <div class="rd-vote-lbl">赞成</div>
          </div>
          <div class="rd-vote">
            <div class="rd-vote-num">{{ currentDetail.vote?.against_count ?? 0 }}</div>
            <div class="rd-vote-lbl">反对</div>
          </div>
          <div class="rd-vote">
            <div class="rd-vote-num">{{ currentDetail.vote?.abstain_count ?? 0 }}</div>
            <div class="rd-vote-lbl">弃权</div>
          </div>
        </div>

        <div class="rd-rule" />

        <!-- 决议 -->
        <div class="rd-eyebrow">决议</div>
        <div class="rd-decision">
          <el-tag
            :type="currentDetail.decision === 'pass' ? 'success' : 'danger'"
            size="large"
            effect="dark"
            class="rd-decision-tag"
          >
            {{ currentDetail.decision === 'pass' ? '通过' : '不通过' }}
          </el-tag>
          <p class="rd-reason">{{ currentDetail.decision_reason }}</p>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { tyRecommendationMeetingApi, tyBranchApi } from '@/api/ty'
import { formatDateTime } from '@/utils/datetime'

const router = useRouter()

// 列表数据
const list = ref([])
const loading = ref(false)
const filterDecision = ref('')
const filterBranch = ref('')
const decisionSelectRef = ref(null)
const branchSelectRef = ref(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 详情弹窗
const detailVisible = ref(false)
const currentDetail = ref(null)

// 团支部下拉
const branches = ref([])

// 获取列表
async function fetchList() {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value
    }
    if (filterDecision.value) params.decision = filterDecision.value
    if (filterBranch.value) params.branch_id = filterBranch.value
    const data = await tyRecommendationMeetingApi.list(params)
    list.value = Array.isArray(data?.items) ? data.items : []
    total.value = typeof data?.total === 'number' ? data.total : 0
  } catch (e) {
    console.error('获取推优大会列表失败', e)
  } finally {
    loading.value = false
  }
}

// 获取支部列表
async function fetchBranches() {
  try {
    const data = await tyBranchApi.list()
    branches.value = data || []
  } catch (e) {
    console.error('获取团支部列表失败', e)
  }
}

function goCreate() {
  router.push('/ty/recommendation-meeting/new')
}

// 打开详情：列表项不含 vote 字段，必须调详情接口拿完整数据。
async function showDetail(row) {
  detailVisible.value = true
  currentDetail.value = row // 先用列表数据打底，避免空白
  try {
    const detail = await tyRecommendationMeetingApi.get(row.id)
    currentDetail.value = { ...row, ...detail }
  } catch (e) {
    // 错误已由 http 拦截器提示；这里只保留列表数据
    console.error('获取推优大会详情失败', e)
  }
}

onMounted(() => {
  fetchList()
  fetchBranches()
})

onBeforeUnmount(() => {
  decisionSelectRef.value?.blur?.()
  branchSelectRef.value?.blur?.()
})
</script>

<style scoped>
/* .card-header / .filter-bar / .pagination-wrap 已在全局定义 */

/* 让 el-table 整体铺满容器（不要 table-layout: fixed，保留弹性分配） */
.rd-table {
  width: 100% !important;
}

/* ============================================================
   详情弹窗：克制的现代工业风
   设计原则：只用 Element Plus 主题色 + 间距 + 字号层级
   ============================================================ */

.rd {
  color: var(--sh-text-primary);
  font-family: var(--sh-font-body);
}

/* 收紧 el-dialog 内边距，确保内容不高过 max-height */
:deep(.el-dialog__header) {
  padding: 18px 20px 12px !important;
  margin-right: 0 !important;
}
:deep(.el-dialog__body) {
  padding: 0 20px 8px !important;
  /* 让内容撑开，但不超过视口 90% */
  max-height: calc(90vh - 110px);
}

/* 顶部 header：左标题 + 右业务编号 */
.rd-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
}
.rd-head-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--sh-text-primary);
}
.rd-head-bizno {
  font-family: ui-monospace, 'JetBrains Mono', 'Fira Code', SFMono-Regular, monospace;
  font-size: 12.5px;
  color: var(--sh-text-secondary);
  letter-spacing: 0.04em;
}

/* 眉题：字段标题 */
.rd-eyebrow {
  font-size: 11px;
  color: var(--sh-text-placeholder);
  letter-spacing: 0.08em;
  margin-bottom: 6px;
}

/* 申请人主信息 */
.rd-applicant {
  margin: 4px 0 4px;
}
.rd-name {
  font-size: 22px;
  font-weight: 600;
  line-height: 1.3;
  color: var(--sh-text-primary);
}

/* 分隔线：收紧 */
.rd-divider {
  margin: 18px 0;
}

/* 2 列 grid */
.rd-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 28px;
}
.rd-value {
  font-size: 14px;
  color: var(--sh-text-regular);
  font-variant-numeric: tabular-nums;
}

/* 投票：3 列 metric */
.rd-votes {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 4px;
}
.rd-vote {
  padding: 16px 12px;
  background: var(--sh-bg-elevated);
  border-radius: 6px;
  text-align: center;
  border: 1px solid var(--sh-border-light);
}
.rd-vote-num {
  font-size: 26px;
  font-weight: 600;
  line-height: 1.2;
  color: var(--sh-text-primary);
  font-variant-numeric: tabular-nums;
}
.rd-vote-lbl {
  font-size: 12px;
  color: var(--sh-text-secondary);
  margin-top: 4px;
  letter-spacing: 0.06em;
}

/* 决议 */
.rd-decision {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  margin-top: 4px;
}
.rd-decision-tag {
  flex-shrink: 0;
  min-width: 60px;
  text-align: center;
  font-weight: 600;
  letter-spacing: 0.1em;
}
.rd-reason {
  flex: 1;
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--sh-text-regular);
  border-left: 2px solid var(--sh-border);
  padding-left: 14px;
}

/* 窄屏堆叠 */
@media (max-width: 640px) {
  .rd-votes { grid-template-columns: 1fr; }
  .rd-grid { grid-template-columns: 1fr; }
  .rd-decision { flex-direction: column; }
}
</style>

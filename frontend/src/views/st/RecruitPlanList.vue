<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>招新计划管理</span>
          <el-button type="primary" @click="goCreate">新建招新计划</el-button>
        </div>
      </template>

      <div class="filter-bar">
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 140px" @change="fetchList">
          <el-option label="草稿" value="S0" />
          <el-option label="待审" value="S1" />
          <el-option label="已通过" value="S3" />
          <el-option label="已驳回" value="S4" />
        </el-select>
        <el-select v-model="filterAssoc" placeholder="社团筛选" clearable style="width: 180px; margin-left: 12px" @change="fetchList">
          <el-option v-for="a in assocs" :key="a.id" :label="a.name" :value="a.id" />
        </el-select>
        <el-input v-model="filterYear" placeholder="学年 (如 2025-2026)" clearable style="width: 180px; margin-left: 12px" @keyup.enter="fetchList" @clear="fetchList" />
        <el-button style="margin-left: 12px" @click="fetchList">查询</el-button>
      </div>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="biz_no" label="编号" width="160" />
        <el-table-column prop="association_name" label="所属社团" min-width="160" />
        <el-table-column label="季节" width="100">
          <template #default="{ row }">{{ row.season_text }}</template>
        </el-table-column>
        <el-table-column prop="academic_year" label="学年" width="120" />
        <el-table-column prop="target_count" label="目标人数" width="100" align="center" />
        <el-table-column label="面试时间" width="170">
          <template #default="{ row }">{{ row.interview_at ? formatDateTime(row.interview_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType[row.status]" size="small">{{ row.status_text }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="招新状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="phaseType[row.recruit_phase]" size="small">{{ row.recruit_phase_text }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="投递/录用" width="120" align="center">
          <template #default="{ row }">
            <span v-if="row.status === 'S3'">
              <el-tag size="small">{{ row.apply_count }}</el-tag>
              /
              <el-tag size="small" type="success">{{ row.accepted_count }}</el-tag>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="360" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'S0'" link type="primary" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button v-if="row.status === 'S0'" link type="success" size="small" @click="handleSubmit(row.id)">提交</el-button>
            <el-button v-if="row.status === 'S1'" link type="warning" size="small" @click="handleWithdraw(row.id)">撤回</el-button>
            <el-button v-if="row.status === 'S1'" link type="success" size="small" @click="handleApprove(row.id)">通过</el-button>
            <el-button v-if="row.status === 'S1'" link type="danger" size="small" @click="handleReject(row.id)">驳回</el-button>
            <el-button v-if="row.status === 'S3'" link type="primary" size="small" @click="handlePublish(row.id)">发布</el-button>
            <el-button v-if="row.status === 'S3' && row.is_finished !== 1" link type="warning" size="small" @click="handleFinish(row.id)">结束招新</el-button>
            <el-button v-if="row.status === 'S3'" link type="primary" size="small" @click="goApplies(row.id)">申请列表</el-button>
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { stRecruitPlanApi, stAssociationApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'

const router = useRouter()

const statusType = { S0: 'info', S1: 'warning', S3: 'success', S4: 'danger' }
// 招新阶段枚举对应标签颜色：未发布=info，招新中=success，已结束=info（灰色）
const phaseType = { not_open: 'info', ongoing: 'success', finished: 'info' }

const list = ref([])
const loading = ref(false)
const filterStatus = ref('')
const filterAssoc = ref(null)
const filterYear = ref('')
const assocs = ref([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (filterStatus.value) params.status = filterStatus.value
    if (filterAssoc.value) params.association_id = filterAssoc.value
    if (filterYear.value) params.academic_year = filterYear.value
    const r = await stRecruitPlanApi.list(params)
    list.value = r.items || []
    total.value = r.total || 0
  } catch (e) {
    ElMessage.error('获取招新计划失败')
  } finally {
    loading.value = false
  }
}

async function loadAssocs() {
  try {
    const r = await stAssociationApi.list({ page: 1, page_size: 200 })
    assocs.value = r.items || []
  } catch (e) {
    // 社团加载失败不影响列表
  }
}

function goCreate() {
  router.push('/st/recruit-plan/new')
}
function goEdit(id) {
  router.push(`/st/recruit-plan/${id}/edit`)
}
function goDetail(id) {
  router.push(`/st/recruit-plan/${id}`)
}
function goApplies(id) {
  router.push({ path: '/st/recruit-apply', query: { plan_id: id } })
}

async function handleSubmit(id) {
  await ElMessageBox.confirm('确认提交该招新计划？提交后将进入待审状态。', '提示', { type: 'warning' })
  await stRecruitPlanApi.submit(id)
  ElMessage.success('已提交')
  fetchList()
}
async function handleWithdraw(id) {
  await ElMessageBox.confirm('确认撤回该招新计划？撤回后将回到草稿状态。', '提示', { type: 'warning' })
  await stRecruitPlanApi.withdraw(id)
  ElMessage.success('已撤回')
  fetchList()
}
async function handleApprove(id) {
  await ElMessageBox.confirm('确认审批通过该招新计划？', '提示', { type: 'success' })
  await stRecruitPlanApi.approve(id)
  ElMessage.success('已审批通过')
  fetchList()
}
async function handleReject(id) {
  const { value: opinion } = await ElMessageBox.prompt('请输入驳回意见（至少 10 字）', '驳回', {
    confirmButtonText: '确认驳回',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputValidator: (v) => (v && v.trim().length >= 10) || '驳回意见至少 10 字'
  })
  await stRecruitPlanApi.reject(id, { opinion })
  ElMessage.success('已驳回')
  fetchList()
}
async function handlePublish(id) {
  await ElMessageBox.confirm('确认发布该招新计划？发布后将开启学生投递通道。', '提示', { type: 'success' })
  await stRecruitPlanApi.publish(id)
  ElMessage.success('已发布')
  fetchList()
}
async function handleFinish(id) {
  const { value: reason } = await ElMessageBox.prompt('请输入结束原因（可空，不超过 200 字）', '提前结束招新', {
    confirmButtonText: '确认结束',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputPlaceholder: '例如：招新人数已满足，提前结束',
    inputValidator: (v) => !v || v.trim().length <= 200 || '结束原因不超过 200 字'
  })
  await ElMessageBox.confirm('结束操作不可逆，确认提前结束该招新？结束后学生不可再投递。', '二次确认', { type: 'warning' })
  await stRecruitPlanApi.finish(id, { reason: reason || '' })
  ElMessage.success('已结束招新')
  fetchList()
}

onMounted(() => {
  loadAssocs()
  fetchList()
})
</script>

<style scoped>
.page-container { padding: 16px; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.filter-bar { margin-bottom: 16px; display: flex; align-items: center; flex-wrap: wrap; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>

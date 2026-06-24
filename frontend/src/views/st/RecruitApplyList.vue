<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>招新申请管理</span>
          <el-button type="primary" @click="openCreate">学生投递</el-button>
        </div>
      </template>

      <div class="filter-bar">
        <el-select v-model="filterPlan" placeholder="按招新计划筛选" clearable style="width: 220px" @change="fetchList">
          <el-option v-for="p in plans" :key="p.id" :label="`${p.biz_no} · ${p.association_name || ''}`" :value="p.id" />
        </el-select>
        <el-select v-model="filterResult" placeholder="按结果筛选" clearable style="width: 140px; margin-left: 12px" @change="fetchList">
          <el-option label="待面试" value="pending" />
          <el-option label="已录用" value="accepted" />
          <el-option label="未通过" value="rejected" />
        </el-select>
      </div>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="plan_biz_no" label="招新计划" min-width="180">
          <template #default="{ row }">
            <div class="plan-cell">
              <span class="plan-biz">{{ row.plan_biz_no || `PLAN-${row.plan_id}` }}</span>
              <span v-if="row.association_name" class="plan-assoc">{{ row.association_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="!isStudent" prop="student_no" label="学号" width="140" />
        <el-table-column v-if="!isStudent" prop="student_name" label="姓名" width="120" />
        <el-table-column label="结果" width="120">
          <template #default="{ row }">
            <el-tag :type="resultType[row.result]" size="small">{{ row.result_text }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="投递时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column v-if="!isStudent" label="录入时间" width="170">
          <template #default="{ row }">{{ row.result_at ? formatDateTime(row.result_at) : '-' }}</template>
        </el-table-column>
        <el-table-column v-if="!isStudent" label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.result === 'pending'" link type="success" size="small" @click="handleResult(row.id, 'accepted')">录用</el-button>
            <el-button v-if="row.result === 'pending'" link type="danger" size="small" @click="handleResult(row.id, 'rejected')">不通过</el-button>
            <el-tag v-else type="info" size="small">已录入</el-tag>
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

    <el-dialog v-model="createDialog" title="学生投递" width="480px">
      <el-form ref="formRef" :model="createForm" :rules="rules" label-width="100px">
        <el-form-item label="招新计划" prop="plan_id">
          <el-select v-model="createForm.plan_id" placeholder="请选择" filterable style="width: 100%">
            <el-option v-for="p in availablePlans" :key="p.id" :label="`${p.biz_no} · ${p.association_name || ''}`" :value="p.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">投递</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { stRecruitApplyApi, stRecruitPlanApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'

const route = useRoute()
const authStore = useAuthStore()
// 学生视角：登录用户关联了学生主体（user.student_id 非 0/空）即为学生
const isStudent = computed(() => {
  const u = authStore.user
  return !!(u && u.student_id)
})
const resultType = { pending: 'warning', accepted: 'success', rejected: 'danger' }

const list = ref([])
const loading = ref(false)
const filterPlan = ref(null)
const filterResult = ref('')
const plans = ref([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const createDialog = ref(false)
const formRef = ref(null)
const createForm = reactive({ plan_id: null })
const rules = { plan_id: [{ required: true, message: '请选择招新计划', trigger: 'change' }] }

const availablePlans = computed(() => plans.value.filter((p) => p.status === 'S3' && p.is_finished !== 1))

async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (filterPlan.value) params.plan_id = filterPlan.value
    if (filterResult.value) params.result = filterResult.value
    // 不再传 scope=all：后端默认按当前用户的 student_id 过滤
    // 教师账号无 student_id → 后端不过滤 → 返回全量（管理员视图）
    // 学生账号有 student_id → 后端按自己过滤（学生视图）
    const r = await stRecruitApplyApi.list(params)
    list.value = r.items || []
    total.value = r.total || 0
  } catch (e) {
    ElMessage.error('获取申请列表失败')
  } finally {
    loading.value = false
  }
}

async function loadPlans() {
  try {
    const r = await stRecruitPlanApi.list({ page: 1, page_size: 200 })
    plans.value = r.items || []
  } catch (e) {
    // ignore
  }
}

function openCreate() {
  createForm.plan_id = null
  createDialog.value = true
}

async function handleCreate() {
  await formRef.value.validate()
  await stRecruitApplyApi.create({ plan_id: createForm.plan_id })
  ElMessage.success('投递成功')
  createDialog.value = false
  fetchList()
}

async function handleResult(id, result) {
  const text = result === 'accepted' ? '录用' : '不通过'
  await ElMessageBox.confirm(`确认将该申请标记为「${text}」？`, '提示', { type: 'warning' })
  await stRecruitApplyApi.submitResult(id, { result })
  ElMessage.success('已录入')
  fetchList()
}

onMounted(async () => {
  if (route.query.plan_id) {
    filterPlan.value = Number(route.query.plan_id)
  }
  await loadPlans()
  await fetchList()
})
</script>

<style scoped>
.page-container { padding: 16px; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.filter-bar { margin-bottom: 16px; display: flex; align-items: center; flex-wrap: wrap; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.plan-cell { display: flex; flex-direction: column; line-height: 1.4; }
.plan-biz { font-weight: 500; color: #303133; }
.plan-assoc { font-size: 12px; color: #909399; }
</style>

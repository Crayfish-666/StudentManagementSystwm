<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>岗位管理</span>
          <el-button type="primary" @click="openCreateDialog">新增岗位</el-button>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-input
          v-model="filterKeyword"
          placeholder="搜索部门名称"
          clearable
          style="width: 200px"
          @keyup.enter="fetchList"
          @clear="fetchList"
        />
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 140px; margin-left: 12px" @change="fetchList">
          <el-option label="草稿" value="S0" />
          <el-option label="待审" value="S1" />
          <el-option label="院系通过" value="S2" />
          <el-option label="终审通过" value="S3" />
          <el-option label="已驳回" value="S4" />
          <el-option label="已关闭" value="closed" />
        </el-select>
        <el-button type="primary" style="margin-left: 12px" @click="fetchList">查询</el-button>
      </div>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="biz_no" label="业务编号" width="160" />
        <el-table-column prop="dept_name" label="部门" min-width="120" />
        <el-table-column prop="title" label="岗位名称" min-width="140" />
        <el-table-column prop="headcount" label="人数" width="80" align="center" />
        <el-table-column prop="weekly_hours_limit" label="周工时上限" width="110" align="center" />
        <el-table-column label="时薪(元)" width="100" align="center">
          <template #default="{ row }">
            {{ (row.hourly_rate_cents / 100).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusType[row.status]" size="small">
              {{ statusMap[row.status] || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'S0'" link type="success" size="small" @click="handleSubmit(row.id)">提交</el-button>
            <el-button v-if="row.status === 'S1' || row.status === 'S2'" link type="primary" size="small" @click="openApproveDialog(row)">审批</el-button>
            <el-button v-if="row.status === 'S1' || row.status === 'S2'" link type="warning" size="small" @click="openRejectDialog(row)">驳回</el-button>
            <el-popconfirm v-if="row.status === 'S0'" title="确认删除此岗位？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button link type="danger" size="small">删除</el-button>
              </template>
            </el-popconfirm>
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

    <!-- 新增岗位对话框 -->
    <el-dialog v-model="createVisible" title="新增岗位" width="600px" :close-on-click-modal="false" @close="resetCreateForm">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="110px">
        <el-form-item label="部门类型" prop="dept_type">
          <el-select v-model="createForm.dept_type" placeholder="请选择部门类型" style="width: 100%">
            <el-option label="图书馆" value="library" />
            <el-option label="食堂" value="canteen" />
            <el-option label="行政" value="admin" />
            <el-option label="实验室" value="lab" />
          </el-select>
        </el-form-item>
        <el-form-item label="部门名称" prop="dept_name">
          <el-input v-model="createForm.dept_name" placeholder="请输入部门名称" maxlength="50" />
        </el-form-item>
        <el-form-item label="岗位名称" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入岗位名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="createForm.description" type="textarea" :rows="3" placeholder="请输入岗位描述" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="人数" prop="headcount">
          <el-input-number v-model="createForm.headcount" :min="1" :max="999" style="width: 100%" />
        </el-form-item>
        <el-form-item label="周工时上限" prop="weekly_hours_limit">
          <el-input-number v-model="createForm.weekly_hours_limit" :min="1" :max="40" style="width: 100%" />
        </el-form-item>
        <el-form-item label="时薪(元)" prop="hourly_rate_input">
          <el-input-number v-model="createForm.hourly_rate_input" :min="0.01" :precision="2" :step="0.5" style="width: 100%" />
        </el-form-item>
        <el-form-item label="开始时间" prop="start_at">
          <el-date-picker v-model="createForm.start_at" type="date" placeholder="选择开始时间" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="结束时间" prop="end_at">
          <el-date-picker v-model="createForm.end_at" type="date" placeholder="选择结束时间" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="风险提示" prop="risk_notes">
          <el-input v-model="createForm.risk_notes" type="textarea" :rows="2" placeholder="如有风险请填写，无则留空" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreate">确认新增</el-button>
      </template>
    </el-dialog>

    <!-- 审批对话框 -->
    <el-dialog v-model="approveVisible" title="岗位审批" width="480px" :close-on-click-modal="false" @close="resetApproveForm">
      <el-form ref="approveFormRef" :model="approveForm" :rules="approveRules" label-width="90px">
        <el-form-item label="业务编号">
          <span>{{ approveRow?.biz_no }}</span>
        </el-form-item>
        <el-form-item label="岗位名称">
          <span>{{ approveRow?.title }}</span>
        </el-form-item>
        <el-form-item label="审批层级" prop="level">
          <el-select v-model="approveForm.level" placeholder="请选择审批层级" style="width: 100%">
            <el-option label="院系" value="college" />
            <el-option label="校级" value="school" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" :loading="approveLoading" @click="handleApprove">确认审批</el-button>
      </template>
    </el-dialog>

    <!-- 驳回对话框 -->
    <el-dialog v-model="rejectVisible" title="岗位驳回" width="480px" :close-on-click-modal="false" @close="resetRejectForm">
      <el-form ref="rejectFormRef" :model="rejectForm" :rules="rejectRules" label-width="90px">
        <el-form-item label="业务编号">
          <span>{{ rejectRow?.biz_no }}</span>
        </el-form-item>
        <el-form-item label="岗位名称">
          <span>{{ rejectRow?.title }}</span>
        </el-form-item>
        <el-form-item label="驳回意见" prop="opinion">
          <el-input v-model="rejectForm.opinion" type="textarea" :rows="4" placeholder="请输入驳回意见" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" @click="handleReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { qgPositionApi, qgApplyApi } from '@/api/qg'

// 状态映射
const statusMap = { S0: '草稿', S1: '待审', S2: '院系通过', S3: '终审通过', S4: '已驳回', closed: '已关闭' }
const statusType = { S0: 'info', S1: 'warning', S2: '', S3: 'success', S4: 'danger', closed: 'info' }

// 列表数据
const list = ref([])
const loading = ref(false)
const filterKeyword = ref('')
const filterStatus = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 获取列表
async function fetchList() {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value
    }
    if (filterKeyword.value && filterKeyword.value.trim()) {
      params.keyword = filterKeyword.value.trim()
    }
    if (filterStatus.value) params.status = filterStatus.value
    const data = await qgPositionApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取岗位列表失败', e)
  } finally {
    loading.value = false
  }
}

// ========== 新增岗位 ==========
const createVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref(null)
const createForm = reactive({
  dept_type: '',
  dept_name: '',
  title: '',
  description: '',
  headcount: 1,
  weekly_hours_limit: 8,
  hourly_rate_input: 15.00,
  start_at: '',
  end_at: '',
  risk_notes: ''
})

const createRules = {
  dept_type: [{ required: true, message: '请选择部门类型', trigger: 'change' }],
  dept_name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
  title: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }],
  headcount: [{ required: true, message: '请输入人数', trigger: 'blur' }],
  weekly_hours_limit: [{ required: true, message: '请输入周工时上限', trigger: 'blur' }],
  hourly_rate_input: [{ required: true, message: '请输入时薪', trigger: 'blur' }],
  start_at: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  end_at: [{ required: true, message: '请选择结束时间', trigger: 'change' }]
}

function openCreateDialog() {
  createVisible.value = true
}

function resetCreateForm() {
  createFormRef.value?.clearValidate()
  Object.assign(createForm, {
    dept_type: '',
    dept_name: '',
    title: '',
    description: '',
    headcount: 1,
    weekly_hours_limit: 8,
    hourly_rate_input: 15.00,
    start_at: '',
    end_at: '',
    risk_notes: ''
  })
}

async function handleCreate() {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      createLoading.value = true
      const payload = {
        dept_type: createForm.dept_type,
        dept_name: createForm.dept_name,
        title: createForm.title,
        description: createForm.description,
        headcount: createForm.headcount,
        weekly_hours_limit: createForm.weekly_hours_limit,
        hourly_rate_cents: Math.round(createForm.hourly_rate_input * 100),
        start_at: createForm.start_at,
        end_at: createForm.end_at,
        risk_notes: createForm.risk_notes || null
      }
      await qgPositionApi.create(payload)
      ElMessage.success('岗位创建成功')
      createVisible.value = false
      fetchList()
    } catch (e) {
      // 错误已由 http 拦截器处理
    } finally {
      createLoading.value = false
    }
  })
}

// ========== 提交岗位 ==========
async function handleSubmit(id) {
  try {
    await ElMessageBox.confirm('确认提交此岗位？提交后将进入审批流程。', '提交确认')
    await qgPositionApi.submit(id)
    ElMessage.success('提交成功')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

// ========== 审批岗位 ==========
const approveVisible = ref(false)
const approveLoading = ref(false)
const approveFormRef = ref(null)
const approveRow = ref(null)
const approveForm = reactive({
  level: ''
})

const approveRules = {
  level: [{ required: true, message: '请选择审批层级', trigger: 'change' }]
}

function openApproveDialog(row) {
  approveRow.value = row
  approveVisible.value = true
}

function resetApproveForm() {
  approveFormRef.value?.clearValidate()
  approveForm.level = ''
  approveRow.value = null
}

async function handleApprove() {
  if (!approveFormRef.value) return
  await approveFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      approveLoading.value = true
      await qgPositionApi.approve(approveRow.value.id, { level: approveForm.level })
      ElMessage.success('审批通过')
      approveVisible.value = false
      fetchList()
    } catch (e) {
      // 错误已由 http 拦截器处理
    } finally {
      approveLoading.value = false
    }
  })
}

// ========== 驳回岗位 ==========
const rejectVisible = ref(false)
const rejectLoading = ref(false)
const rejectFormRef = ref(null)
const rejectRow = ref(null)
const rejectForm = reactive({
  opinion: ''
})

const rejectRules = {
  opinion: [
    { required: true, message: '请输入驳回意见', trigger: 'blur' },
    { min: 2, message: '驳回意见至少 2 字', trigger: 'blur' }
  ]
}

function openRejectDialog(row) {
  rejectRow.value = row
  rejectVisible.value = true
}

function resetRejectForm() {
  rejectFormRef.value?.clearValidate()
  rejectForm.opinion = ''
  rejectRow.value = null
}

async function handleReject() {
  if (!rejectFormRef.value) return
  await rejectFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      rejectLoading.value = true
      await qgPositionApi.reject(rejectRow.value.id, { opinion: rejectForm.opinion })
      ElMessage.success('已驳回')
      rejectVisible.value = false
      fetchList()
    } catch (e) {
      // 错误已由 http 拦截器处理
    } finally {
      rejectLoading.value = false
    }
  })
}

// ========== 删除岗位 ==========
async function handleDelete(id) {
  try {
    await qgPositionApi.delete(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    // 错误已由 http 拦截器处理
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
/* .card-header, .filter-bar, .pagination-wrap 已在 App.vue 全局定义 */
</style>

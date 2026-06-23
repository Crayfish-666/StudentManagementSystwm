<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>困难认定管理</span>
          <el-button type="primary" @click="openCreateDialog">新增认定</el-button>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterLevel" placeholder="困难等级" clearable style="width: 140px">
          <el-option label="特别困难" value="special" />
          <el-option label="困难" value="hard" />
          <el-option label="一般困难" value="normal" />
          <el-option label="不困难" value="none" />
        </el-select>
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 140px; margin-left: 12px">
          <el-option label="草稿" value="S0" />
          <el-option label="待审" value="S1" />
          <el-option label="院系通过" value="S2" />
          <el-option label="终审通过" value="S3" />
          <el-option label="已驳回" value="S4" />
        </el-select>
        <el-button type="primary" style="margin-left: 12px" @click="fetchList">查询</el-button>
      </div>

      <!-- 表格 -->
      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="biz_no" label="业务编号" width="160" />
        <el-table-column prop="student_name" label="学生姓名" width="100" />
        <el-table-column prop="student_no" label="学号" width="130" />
        <el-table-column prop="academic_year" label="学年" width="110" />
        <el-table-column prop="level" label="困难等级" width="110">
          <template #default="{ row }">
            <el-tag :type="levelType[row.level]" size="small">
              {{ levelMap[row.level] || row.level }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType[row.status]" size="small">
              {{ statusMap[row.status] || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'S0'" link type="success" size="small" @click="handleSubmit(row.id)">提交</el-button>
            <el-button v-if="row.status === 'S1' || row.status === 'S2'" link type="primary" size="small" @click="openApproveDialog(row)">审批</el-button>
            <el-button v-if="row.status === 'S1' || row.status === 'S2'" link type="warning" size="small" @click="openRejectDialog(row)">驳回</el-button>
            <el-popconfirm v-if="row.status === 'S0'" title="确认删除此认定记录？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button link type="danger" size="small">删除</el-button>
              </template>
            </el-popconfirm>
            <el-button link type="primary" size="small" @click="openDetailDialog(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
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

    <!-- 新增认定对话框 -->
    <el-dialog v-model="createVisible" title="新增困难认定" width="520px" :close-on-click-modal="false" @close="onCreateClose">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="90px">
        <el-form-item label="学生ID" prop="student_id">
          <el-input v-model.number="createForm.student_id" placeholder="请输入学生ID" />
        </el-form-item>
        <el-form-item label="学年" prop="academic_year">
          <el-input v-model="createForm.academic_year" placeholder="例如：2025-2026" />
        </el-form-item>
        <el-form-item label="困难等级" prop="level">
          <el-select v-model="createForm.level" placeholder="请选择困难等级" style="width: 100%">
            <el-option label="特别困难" value="special" />
            <el-option label="困难" value="hard" />
            <el-option label="一般困难" value="normal" />
            <el-option label="不困难" value="none" />
          </el-select>
        </el-form-item>
        <el-form-item label="证书文件" prop="cert_files">
          <el-input v-model="createForm.cert_files" type="textarea" :rows="3" placeholder="请输入证书文件路径，多个用逗号分隔" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreate">确认新增</el-button>
      </template>
    </el-dialog>

    <!-- 审批对话框 -->
    <el-dialog v-model="approveVisible" title="审批" width="520px" :close-on-click-modal="false" @close="onApproveClose">
      <el-form ref="approveFormRef" :model="approveForm" :rules="approveRules" label-width="90px">
        <el-form-item label="业务编号">
          <span>{{ currentRow?.biz_no }}</span>
        </el-form-item>
        <el-form-item label="学生姓名">
          <span>{{ currentRow?.student_name }}（{{ currentRow?.student_no }}）</span>
        </el-form-item>
        <el-form-item label="审批层级" prop="level">
          <el-select v-model="approveForm.level" placeholder="请选择审批层级" style="width: 100%">
            <el-option label="院系审批" value="college" />
            <el-option label="学校审批" value="school" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="success" :loading="approveLoading" @click="handleApprove">确认通过</el-button>
      </template>
    </el-dialog>

    <!-- 驳回对话框 -->
    <el-dialog v-model="rejectVisible" title="驳回" width="520px" :close-on-click-modal="false" @close="onRejectClose">
      <el-form ref="rejectFormRef" :model="rejectForm" :rules="rejectRules" label-width="90px">
        <el-form-item label="业务编号">
          <span>{{ currentRow?.biz_no }}</span>
        </el-form-item>
        <el-form-item label="学生姓名">
          <span>{{ currentRow?.student_name }}（{{ currentRow?.student_no }}）</span>
        </el-form-item>
        <el-form-item label="审批层级" prop="level">
          <el-select v-model="rejectForm.level" placeholder="请选择审批层级" style="width: 100%">
            <el-option label="院系审批" value="college" />
            <el-option label="学校审批" value="school" />
          </el-select>
        </el-form-item>
        <el-form-item label="驳回原因" prop="reject_reason">
          <el-input v-model="rejectForm.reject_reason" type="textarea" :rows="4" placeholder="请输入驳回原因" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" @click="handleReject">确认驳回</el-button>
      </template>
    </el-dialog>
    <!-- 查看详情对话框 -->
    <el-dialog v-model="detailVisible" title="困难认定详情" width="560px" :close-on-click-modal="false">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="业务编号">{{ detailData?.biz_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType[detailData?.status]" size="small">
            {{ statusMap[detailData?.status] || detailData?.status }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="学生姓名">{{ detailData?.student_name }}</el-descriptions-item>
        <el-descriptions-item label="学号">{{ detailData?.student_no }}</el-descriptions-item>
        <el-descriptions-item label="学年" :span="2">{{ detailData?.academic_year }}</el-descriptions-item>
        <el-descriptions-item label="困难等级" :span="2">
          <el-tag :type="levelType[detailData?.level]" size="small">
            {{ levelMap[detailData?.level] || detailData?.level }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="证书文件" :span="2">{{ detailData?.cert_files || '无' }}</el-descriptions-item>
        <el-descriptions-item v-if="detailData?.reject_reason" label="驳回原因" :span="2">
          <span style="color: #f56c6c">{{ detailData.reject_reason }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(detailData?.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDateTime(detailData?.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { qgDifficultyApi } from '@/api/qg'
import { formatDateTime } from '@/utils/datetime'

// 状态映射
const statusMap = { S0: '草稿', S1: '待审', S2: '院系通过', S3: '终审通过', S4: '已驳回' }
const statusType = { S0: 'info', S1: 'warning', S2: '', S3: 'success', S4: 'danger' }

// 困难等级映射
const levelMap = { special: '特别困难', hard: '困难', normal: '一般困难', none: '不困难' }
const levelType = { special: 'danger', hard: 'warning', normal: '', none: 'info' }

// 列表数据
const list = ref([])
const loading = ref(false)
const filterLevel = ref('')
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
    if (filterLevel.value) params.level = filterLevel.value
    if (filterStatus.value) params.status = filterStatus.value
    const data = await qgDifficultyApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取困难认定列表失败', e)
  } finally {
    loading.value = false
  }
}

// 提交认定（S0 → S1）
async function handleSubmit(id) {
  try {
    await ElMessageBox.confirm('确认提交此认定？提交后将进入审批流程。', '提交确认')
    await qgDifficultyApi.submit(id)
    ElMessage.success('提交成功')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

// 删除认定（仅 S0）
async function handleDelete(id) {
  try {
    await qgDifficultyApi.delete(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    // 错误已由 http 拦截器处理
  }
}

// ========== 新增认定 ==========
const createVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref(null)
const createForm = reactive({
  student_id: '',
  academic_year: '',
  level: '',
  cert_files: ''
})
const createRules = {
  student_id: [{ required: true, message: '请输入学生ID', trigger: 'blur' }],
  academic_year: [{ required: true, message: '请输入学年', trigger: 'blur' }],
  level: [{ required: true, message: '请选择困难等级', trigger: 'change' }]
}

function openCreateDialog() {
  createForm.student_id = ''
  createForm.academic_year = ''
  createForm.level = ''
  createForm.cert_files = ''
  createVisible.value = true
}

function onCreateClose() {
  createFormRef.value?.clearValidate()
}

async function handleCreate() {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      createLoading.value = true
      await qgDifficultyApi.create({
        student_id: createForm.student_id,
        academic_year: createForm.academic_year,
        level: createForm.level,
        cert_files: createForm.cert_files
      })
      ElMessage.success('新增成功')
      createVisible.value = false
      fetchList()
    } finally {
      createLoading.value = false
    }
  })
}

// ========== 审批对话框 ==========
const approveVisible = ref(false)
const approveLoading = ref(false)
const approveFormRef = ref(null)
const currentRow = ref(null)
const approveForm = reactive({
  level: ''
})
const approveRules = {
  level: [{ required: true, message: '请选择审批层级', trigger: 'change' }]
}

function openApproveDialog(row) {
  currentRow.value = row
  approveForm.level = row.status === 'S1' ? 'college' : 'school'
  approveVisible.value = true
}

function onApproveClose() {
  approveFormRef.value?.clearValidate()
}

async function handleApprove() {
  if (!approveFormRef.value) return
  await approveFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      approveLoading.value = true
      await qgDifficultyApi.approve(currentRow.value.id, { level: approveForm.level })
      ElMessage.success('审批通过')
      approveVisible.value = false
      fetchList()
    } finally {
      approveLoading.value = false
    }
  })
}

// ========== 驳回对话框 ==========
const rejectVisible = ref(false)
const rejectLoading = ref(false)
const rejectFormRef = ref(null)
const rejectForm = reactive({
  level: '',
  reject_reason: ''
})
const rejectRules = {
  level: [{ required: true, message: '请选择审批层级', trigger: 'change' }],
  reject_reason: [{ required: true, message: '请输入驳回原因', trigger: 'blur' }]
}

function openRejectDialog(row) {
  currentRow.value = row
  rejectForm.level = row.status === 'S1' ? 'college' : 'school'
  rejectForm.reject_reason = ''
  rejectVisible.value = true
}

function onRejectClose() {
  rejectFormRef.value?.clearValidate()
}

async function handleReject() {
  if (!rejectFormRef.value) return
  await rejectFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      rejectLoading.value = true
      await qgDifficultyApi.reject(currentRow.value.id, {
        level: rejectForm.level,
        reject_reason: rejectForm.reject_reason
      })
      ElMessage.success('已驳回')
      rejectVisible.value = false
      fetchList()
    } finally {
      rejectLoading.value = false
    }
  })
}

// ========== 查看详情 ==========
const detailVisible = ref(false)
const detailData = ref(null)

function openDetailDialog(row) {
  detailData.value = row
  detailVisible.value = true
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
/* .card-header, .filter-bar, .pagination-wrap 已在 App.vue 全局定义 */
</style>

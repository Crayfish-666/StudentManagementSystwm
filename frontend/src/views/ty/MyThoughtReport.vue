<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>我的思想汇报</span>
          <el-button type="primary" size="small" :disabled="!canSubmit" @click="openDialog">
            <el-icon><Plus /></el-icon>新增汇报
          </el-button>
        </div>
      </template>

      <el-alert
        v-if="!hasApplication"
        type="info"
        :closable="false"
        title="暂无可关联的入团申请"
        description="思想汇报需关联一条入团申请单；如尚未提交入团申请，请先完成入团申请后再来提交汇报。"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-table :data="reportList" stripe v-loading="loading" empty-text="暂无思想汇报">
        <el-table-column prop="biz_no" label="编号" width="180" />
        <el-table-column label="学号" width="120">
          <template #default="{ row }">
            <span>{{ row.student_no || myStudentInfo?.student_no || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="汇报人" width="100">
          <template #default="{ row }">
            <span>{{ row.student_name || myStudentInfo?.name || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="quarter" label="季度" width="100" />
        <el-table-column prop="ai_similarity" label="AI相似度" width="100">
          <template #default="{ row }">
            {{ row.ai_similarity != null ? (row.ai_similarity * 100).toFixed(1) + '%' : '未检测' }}
          </template>
        </el-table-column>
        <el-table-column prop="is_qualified" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_qualified ? 'success' : 'danger'" size="small">
              {{ row.is_qualified ? '合格' : '不合格' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" min-width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > pageSize"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top: 16px; justify-content: flex-end; display: flex"
        @current-change="fetchList"
        @size-change="fetchList"
      />
    </el-card>

    <!-- 新增思想汇报弹窗 -->
    <el-dialog v-model="dialogVisible" title="提交思想汇报" width="720px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="关联申请" prop="application_id">
          <el-select
            v-model="form.application_id"
            placeholder="请选择入团申请"
            style="width: 100%"
            filterable
          >
            <el-option
              v-for="app in applications"
              :key="app.id"
              :label="`${app.biz_no || ''}（${app.status}）`"
              :value="app.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="汇报人">
          <el-input :model-value="studentLabel" disabled placeholder="—" />
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入汇报标题" maxlength="80" show-word-limit />
        </el-form-item>
        <el-form-item label="季度" prop="quarter">
          <el-select v-model="form.quarter" placeholder="请选择季度" style="width: 100%">
            <el-option label="第一季度" value="Q1" />
            <el-option label="第二季度" value="Q2" />
            <el-option label="第三季度" value="Q3" />
            <el-option label="第四季度" value="Q4" />
          </el-select>
        </el-form-item>
        <el-form-item label="汇报内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="14"
            placeholder="请详细撰写思想汇报内容（不少于 1000 字）"
            maxlength="20000"
            show-word-limit
          />
          <div class="word-count" :class="{ warning: contentLength < 1000 }">
            已输入 {{ contentLength }} 字（最少 1000 字）
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">提交汇报</el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="思想汇报详情" width="720px" destroy-on-close>
      <el-descriptions v-if="current" :column="2" border>
        <el-descriptions-item label="编号" :span="2">{{ current.biz_no }}</el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ current.title }}</el-descriptions-item>
        <el-descriptions-item label="季度">{{ current.quarter }}</el-descriptions-item>
        <el-descriptions-item label="AI相似度">
          {{ current.ai_similarity != null ? (current.ai_similarity * 100).toFixed(1) + '%' : '未检测' }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="current.is_qualified ? 'success' : 'danger'" size="small">
            {{ current.is_qualified ? '合格' : '不合格' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ formatDateTime(current.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <el-divider content-position="left">汇报正文</el-divider>
      <div class="content-block">{{ current?.content }}</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { tyThoughtReportApi, tyApplicationApi } from '@/api/ty'
import { studentApi } from '@/api/idx'
import { formatDateTime } from '@/utils/datetime'

// === 当前学生身份 ===
const myStudentId = ref(null)
const myStudentInfo = ref(null)

const studentLabel = computed(() => {
  const s = myStudentInfo.value
  if (!s) return ''
  return `${s.student_no || ''} ${s.name || ''}`.trim()
})

// === 列表 ===
const reportList = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

async function fetchList() {
  loading.value = true
  try {
    const res = await tyThoughtReportApi.list({ page: page.value, page_size: pageSize.value })
    reportList.value = res.items || []
    total.value = res.total || 0
  } catch (e) {
    console.error('获取思想汇报列表失败', e)
    ElMessage.error(`获取思想汇报列表失败：${e?.message || e}`)
  } finally {
    loading.value = false
  }
}

// === 我的入团申请 ===
const applications = ref([])
const hasApplication = computed(() => applications.value.length > 0)
const canSubmit = computed(() => hasApplication.value && !!myStudentId.value)

async function fetchApplications() {
  if (!myStudentId.value) {
    applications.value = []
    return
  }
  try {
    const res = await tyApplicationApi.list({ student_id: myStudentId.value, page_size: 200 })
    applications.value = res.items || []
  } catch (e) {
    console.error('获取入团申请失败', e)
    applications.value = []
  }
}

// === 加载学生身份 ===
async function loadMyStudent() {
  try {
    const profile = await studentApi.getMyProfile()
    myStudentId.value = profile?.id || null
    myStudentInfo.value = profile || null
  } catch (e) {
    // 当前用户未关联学生身份：静默处理
    myStudentId.value = null
    myStudentInfo.value = null
  }
}

// === 新增弹窗 ===
const dialogVisible = ref(false)
const saving = ref(false)
const formRef = ref()
const form = ref({ application_id: null, title: '', content: '', quarter: '' })
const formRules = {
  application_id: [{ required: true, message: '请选择入团申请', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  quarter: [{ required: true, message: '请选择季度', trigger: 'change' }],
  content: [{ required: true, message: '请填写汇报内容', trigger: 'blur' }]
}
const contentLength = computed(() => (form.value.content || '').length)

function openDialog() {
  if (!canSubmit.value) {
    ElMessage.warning('暂无可关联的入团申请，无法提交')
    return
  }
  form.value = { application_id: null, title: '', content: '', quarter: '' }
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  if (contentLength.value < 1000) {
    ElMessage.warning('汇报内容不少于 1000 字')
    return
  }
  saving.value = true
  try {
    // 汇报人由后端从当前登录用户自动注入，前端不传 student_id
    await tyThoughtReportApi.create(form.value)
    ElMessage.success('思想汇报已提交')
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    // 错误提示由 http 拦截器统一处理
  } finally {
    saving.value = false
  }
}

// === 详情弹窗 ===
const detailVisible = ref(false)
const current = ref(null)
function showDetail(row) {
  current.value = row
  detailVisible.value = true
}

onMounted(async () => {
  await loadMyStudent()
  await Promise.all([fetchApplications(), fetchList()])
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.word-count {
  font-size: 12px;
  color: var(--el-color-success);
  margin-top: 4px;
  text-align: right;
}
.word-count.warning {
  color: var(--el-color-danger);
}
.content-block {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
  padding: 8px 4px;
  font-size: 14px;
  color: var(--el-text-color-regular);
  max-height: 50vh;
  overflow-y: auto;
}
</style>

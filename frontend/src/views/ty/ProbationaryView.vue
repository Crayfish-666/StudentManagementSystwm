<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>转正流程</span>
          <el-select
            v-model="activeAppFilter"
            placeholder="全部（不筛选）"
            clearable
            style="width: 260px"
            @change="onAppFilterChange"
          >
            <template #prefix><span class="filter-label">关联申请：</span></template>
            <el-option
              v-for="app in applications"
              :key="app.id"
              :label="app.student_name"
              :value="app.id"
            />
          </el-select>
        </div>
      </template>

      <el-tabs v-model="activeTab" type="border-card">
        <!-- 预备期考察 Tab -->
        <el-tab-pane label="预备期考察" name="probationary-record">
          <div class="tab-header">
            <h4>季度考察记录</h4>
            <el-button type="primary" size="small" @click="openRecordDialog">新增考察记录</el-button>
          </div>

          <el-table :data="recordList" stripe v-loading="recordLoading">
            <el-table-column prop="application_id" label="申请ID" width="80" />
            <el-table-column prop="student_name" label="学生" min-width="120">
              <template #default="{ row }">
                <span v-if="row.student_name">{{ row.student_name }}</span>
                <span v-else class="text-muted">—</span>
                <span v-if="row.student_no" class="student-no">（{{ row.student_no }}）</span>
              </template>
            </el-table-column>
            <el-table-column prop="record_year" label="年份" width="80" />
            <el-table-column prop="record_quarter" label="季度" width="80">
              <template #default="{ row }">
                第{{ row.record_quarter }}季度
              </template>
            </el-table-column>
            <el-table-column prop="summary" label="考察总结" min-width="250" show-overflow-tooltip />
            <el-table-column prop="created_at" label="记录时间" min-width="180">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openRecordView(row)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 转正大会 Tab -->
        <el-tab-pane label="转正大会" name="probationary-meeting">
          <div class="tab-header">
            <h4>转正大会记录</h4>
            <el-button type="primary" size="small" @click="openMeetingDialog">召开转正大会</el-button>
          </div>

          <el-table :data="meetingList" stripe v-loading="meetingLoading">
            <el-table-column prop="biz_no" label="业务编号" width="170" />
            <el-table-column prop="student_name" label="申请人" min-width="140">
              <template #default="{ row }">
                <span v-if="row.student_name">{{ row.student_name }}</span>
                <span v-else class="text-muted">—</span>
                <span v-if="row.student_no" class="student-no">（{{ row.student_no }}）</span>
              </template>
            </el-table-column>
            <el-table-column prop="meeting_at" label="会议时间" min-width="180">
              <template #default="{ row }">{{ formatDateTime(row.meeting_at) }}</template>
            </el-table-column>
            <el-table-column label="到会情况" width="120">
              <template #default="{ row }">
                {{ row.actual_count }} / {{ row.expected_count }}
              </template>
            </el-table-column>
            <el-table-column label="票数" width="180">
              <template #default="{ row }">
                赞成{{ row.approve_count }} / 反对{{ (row.actual_count - row.approve_count) }}
              </template>
            </el-table-column>
            <el-table-column prop="decision" label="决议" width="90">
              <template #default="{ row }">
                <el-tag :type="row.decision === 'pass' ? 'success' : 'danger'" size="small">
                  {{ row.decision === 'pass' ? '通过' : '不通过' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="formal_join_at" label="正式入团时间" width="170">
              <template #default="{ row }">
                {{ row.formal_join_at ? formatDate(row.formal_join_at) : '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openMeetingView(row)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 新增考察记录弹窗 -->
    <el-dialog v-model="recordDialogVisible" title="新增预备期考察记录" width="640px" destroy-on-close>
      <el-form ref="recordFormRef" :model="recordForm" :rules="recordFormRules" label-width="100px">
        <!-- 绑定学生区域（选择申请后即显示对应学生） -->
        <el-form-item label="关联申请" prop="application_id">
          <el-select
            v-model="recordForm.application_id"
            placeholder="请选择入团申请（选择后自动绑定学生）"
            style="width: 100%"
            filterable
          >
            <el-option
              v-for="app in applications"
              :key="app.id"
              :label="`${app.student_name}（${app.student_no || '无学号'} · ${app.branch_name || '未分配支部'}）`"
              :value="app.id"
            />
          </el-select>
        </el-form-item>
        <el-card v-if="recordBoundStudent" shadow="never" class="student-card">
          <template #header>
            <div class="student-card-header">
              <el-icon><User /></el-icon>
              <span>已绑定学生</span>
            </div>
          </template>
          <el-descriptions :column="3" size="default" border>
            <el-descriptions-item label="姓名">{{ recordBoundStudent.student_name }}</el-descriptions-item>
            <el-descriptions-item label="学号">{{ recordBoundStudent.student_no || '—' }}</el-descriptions-item>
            <el-descriptions-item label="学生ID">{{ recordBoundStudent.student_id || '—' }}</el-descriptions-item>
            <el-descriptions-item label="所在支部" :span="2">{{ recordBoundStudent.branch_name || '—' }}</el-descriptions-item>
            <el-descriptions-item label="申请编号">{{ recordBoundStudent.biz_no || '—' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
        <el-alert
          v-else
          title="请先选择「关联申请」以绑定学生"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 16px"
        />

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="年份" prop="record_year">
              <el-input-number v-model="recordForm.record_year" :min="2020" :max="2030" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="季度" prop="record_quarter">
              <el-select v-model="recordForm.record_quarter" placeholder="选择季度" style="width: 100%">
                <el-option label="第1季度" :value="1" />
                <el-option label="第2季度" :value="2" />
                <el-option label="第3季度" :value="3" />
                <el-option label="第4季度" :value="4" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="考察总结" prop="summary">
          <el-input
            v-model="recordForm.summary"
            type="textarea"
            :rows="6"
            placeholder="请详细撰写该季度的考察总结（不少于100字）"
            show-word-limit
          />
          <div class="word-count" :class="{ warning: recordSummaryLength < 100 }">
            已输入 {{ recordSummaryLength }} 字（最少 100 字）
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateRecord" :loading="recordSaving" :disabled="!recordBoundStudent">保存</el-button>
      </template>
    </el-dialog>

    <!-- 召开转正大会弹窗 -->
    <el-dialog v-model="meetingDialogVisible" title="召开转正大会" width="680px" destroy-on-close>
      <el-alert
        title="转正大会通过后，该同志将转为正式团员。"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      />

      <el-form ref="meetingFormRef" :model="meetingForm" :rules="meetingFormRules" label-width="110px">
        <el-form-item label="关联申请" prop="application_id">
          <el-select
            v-model="meetingForm.application_id"
            placeholder="请选择预备团员（选择后自动绑定学生）"
            style="width: 100%"
            filterable
          >
            <el-option
              v-for="app in applications"
              :key="app.id"
              :label="`${app.student_name}（${app.student_no || '无学号'} · ${app.branch_name || '未分配支部'}）`"
              :value="app.id"
            />
          </el-select>
        </el-form-item>
        <el-card v-if="meetingBoundStudent" shadow="never" class="student-card">
          <template #header>
            <div class="student-card-header">
              <el-icon><User /></el-icon>
              <span>已绑定学生</span>
            </div>
          </template>
          <el-descriptions :column="3" size="default" border>
            <el-descriptions-item label="姓名">{{ meetingBoundStudent.student_name }}</el-descriptions-item>
            <el-descriptions-item label="学号">{{ meetingBoundStudent.student_no || '—' }}</el-descriptions-item>
            <el-descriptions-item label="学生ID">{{ meetingBoundStudent.student_id || '—' }}</el-descriptions-item>
            <el-descriptions-item label="所在支部" :span="2">{{ meetingBoundStudent.branch_name || '—' }}</el-descriptions-item>
            <el-descriptions-item label="申请编号">{{ meetingBoundStudent.biz_no || '—' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
        <el-alert
          v-else
          title="请先选择「关联申请」以绑定学生"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 16px"
        />

        <el-form-item label="转正申请书" prop="self_application_path">
          <el-input v-model="meetingForm.self_application_path" placeholder="转正申请书材料路径/编号" />
        </el-form-item>

        <el-form-item label="会议时间" prop="meeting_at">
          <el-date-picker
            v-model="meetingForm.meeting_at"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="选择会议时间"
            style="width: 100%"
          />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="应到人数" prop="expected_count">
              <el-input-number v-model="meetingForm.expected_count" :min="1" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="实到人数" prop="actual_count">
              <el-input-number v-model="meetingForm.actual_count" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="赞成票数" prop="approve_count">
              <el-input-number v-model="meetingForm.approve_count" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="决议结果" prop="decision">
              <el-radio-group v-model="meetingForm.decision">
                <el-radio value="pass">通过（转正）</el-radio>
                <el-radio value="reject">不通过</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="meetingDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateMeeting" :loading="meetingSaving" :disabled="!meetingBoundStudent">提交大会记录</el-button>
      </template>
    </el-dialog>

    <!-- 预备期考察记录 详情弹窗 -->
    <el-dialog v-model="recordViewVisible" title="预备期考察记录详情" width="640px" destroy-on-close>
      <div v-loading="recordViewLoading">
        <!-- 绑定学生信息卡片 -->
        <el-card v-if="recordView.student_name" shadow="never" class="student-card">
          <template #header>
            <div class="student-card-header">
              <el-icon><User /></el-icon>
              <span>绑定学生</span>
            </div>
          </template>
          <el-descriptions :column="2" size="default" border>
            <el-descriptions-item label="姓名">{{ recordView.student_name }}</el-descriptions-item>
            <el-descriptions-item label="学号">{{ recordView.student_no || '—' }}</el-descriptions-item>
            <el-descriptions-item label="学生ID">{{ recordView.student_id || '—' }}</el-descriptions-item>
            <el-descriptions-item label="申请ID">{{ recordView.application_id }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
        <el-alert
          v-else
          title="该记录未关联到有效学生"
          type="warning"
          :closable="false"
          show-icon
        />

        <el-divider content-position="left">考察记录</el-divider>

        <el-descriptions :column="2" size="default" border>
          <el-descriptions-item label="记录ID">{{ recordView.id }}</el-descriptions-item>
          <el-descriptions-item label="年份 / 季度">
            {{ recordView.record_year }} 年 · 第{{ recordView.record_quarter }}季度
          </el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(recordView.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="考察总结" :span="2">
            <div class="summary-content">{{ recordView.summary || '—' }}</div>
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="recordViewVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 转正大会记录 详情弹窗 -->
    <el-dialog v-model="meetingViewVisible" title="转正大会记录详情" width="680px" destroy-on-close>
      <div v-loading="meetingViewLoading">
        <!-- 绑定学生信息卡片 -->
        <el-card v-if="meetingView.student_name" shadow="never" class="student-card">
          <template #header>
            <div class="student-card-header">
              <el-icon><User /></el-icon>
              <span>绑定学生</span>
            </div>
          </template>
          <el-descriptions :column="2" size="default" border>
            <el-descriptions-item label="姓名">{{ meetingView.student_name }}</el-descriptions-item>
            <el-descriptions-item label="学号">{{ meetingView.student_no || '—' }}</el-descriptions-item>
            <el-descriptions-item label="学生ID">{{ meetingView.student_id || '—' }}</el-descriptions-item>
            <el-descriptions-item label="申请ID">{{ meetingView.application_id }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
        <el-alert
          v-else
          title="该记录未关联到有效学生"
          type="warning"
          :closable="false"
          show-icon
        />

        <el-divider content-position="left">大会信息</el-divider>

        <el-descriptions :column="2" size="default" border>
          <el-descriptions-item label="大会ID">{{ meetingView.id }}</el-descriptions-item>
          <el-descriptions-item label="业务编号">{{ meetingView.biz_no || '—' }}</el-descriptions-item>
          <el-descriptions-item label="会议时间">{{ formatDateTime(meetingView.meeting_at) }}</el-descriptions-item>
          <el-descriptions-item label="决议">
            <el-tag :type="meetingView.decision === 'pass' ? 'success' : 'danger'" size="small">
              {{ meetingView.decision_text || (meetingView.decision === 'pass' ? '通过' : '不通过') }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="到会情况">
            实到 {{ meetingView.actual_count }} / 应到 {{ meetingView.expected_count }}
          </el-descriptions-item>
          <el-descriptions-item label="票数">
            赞成 {{ meetingView.approve_count }} / 反对 {{ (meetingView.actual_count - meetingView.approve_count) }}
          </el-descriptions-item>
          <el-descriptions-item label="正式入团时间" :span="2">
            {{ meetingView.formal_join_at ? formatDate(meetingView.formal_join_at) : '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="转正申请书" :span="2">
            {{ meetingView.self_application_path || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(meetingView.created_at) }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="meetingViewVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { User } from '@element-plus/icons-vue'
import { tyProbationaryRecordApi, tyProbationaryMeetingApi, tyApplicationApi } from '@/api/ty'
import { formatDateTime, formatDate } from '@/utils/datetime'

const route = useRoute()
const activeTab = ref('probationary-record')

// 列表过滤器：关联入团申请（共用）
const activeAppFilter = ref(null)
const applications = ref([])

async function fetchApplications() {
  try {
    const data = await tyApplicationApi.list({ page_size: 200 })
    applications.value = data.items || []
  } catch (e) {
    console.error('获取入团申请列表失败', e)
  }
}

// 过滤器变化：按当前 Tab 重新查询
function onAppFilterChange() {
  if (activeTab.value === 'probationary-record') {
    fetchRecords()
  } else if (activeTab.value === 'probationary-meeting') {
    fetchMeetings()
  }
}

// ========== 预备期考察 ==========
const recordList = ref([])
const recordLoading = ref(false)
async function fetchRecords() {
  recordLoading.value = true
  try {
    const params = { page: 1, page_size: 100 }
    if (activeAppFilter.value) params.application_id = activeAppFilter.value
    const data = await tyProbationaryRecordApi.list(params)
    recordList.value = data.items || []
  } catch (e) {
    console.error('获取考察记录失败', e)
  } finally {
    recordLoading.value = false
  }
}

const recordDialogVisible = ref(false)
const recordSaving = ref(false)
const recordFormRef = ref()
const recordForm = ref({
  application_id: null,
  record_year: new Date().getFullYear(),
  record_quarter: Math.ceil((new Date().getMonth() + 1) / 3),
  summary: ''
})
const recordFormRules = {
  application_id: [{ required: true, message: '请选择入团申请', trigger: 'change' }],
  record_year: [{ required: true, message: '请填写年份', trigger: 'blur' }],
  record_quarter: [{ required: true, message: '请选择季度', trigger: 'change' }],
  summary: [
    { required: true, message: '请填写考察总结', trigger: 'blur' },
    { min: 100, message: '考察总结不少于100字', trigger: 'blur' }
  ]
}
const recordSummaryLength = computed(() => (recordForm.value.summary || '').length)
// 绑定的学生：依据所选入团申请从 applications 列表中查找
const recordBoundStudent = computed(() => {
  const id = recordForm.value.application_id
  if (!id) return null
  return applications.value.find((a) => a.id === id) || null
})

function openRecordDialog() {
  recordForm.value = {
    application_id: null,
    record_year: new Date().getFullYear(),
    record_quarter: Math.ceil((new Date().getMonth() + 1) / 3),
    summary: ''
  }
  recordDialogVisible.value = true
}
async function handleCreateRecord() {
  try { await recordFormRef.value.validate() } catch { return }
  if ((recordForm.value.summary || '').length < 100) {
    ElMessage.warning('考察总结不少于100字')
    return
  }
  recordSaving.value = true
  try {
    await tyProbationaryRecordApi.create(recordForm.value)
    ElMessage.success('考察记录已保存')
    recordDialogVisible.value = false
    fetchRecords()
  } catch (e) {} finally { recordSaving.value = false }
}

// ========== 转正大会 ==========
const meetingList = ref([])
const meetingLoading = ref(false)
async function fetchMeetings() {
  meetingLoading.value = true
  try {
    const params = { page: 1, page_size: 100 }
    if (activeAppFilter.value) params.application_id = activeAppFilter.value
    const data = await tyProbationaryMeetingApi.list(params)
    meetingList.value = data.items || []
  } catch (e) {
    console.error('获取转正大会列表失败', e)
  } finally {
    meetingLoading.value = false
  }
}

const meetingDialogVisible = ref(false)
const meetingSaving = ref(false)
const meetingFormRef = ref()
const meetingForm = ref({
  application_id: null,
  self_application_path: '',
  meeting_at: '',
  expected_count: null,
  actual_count: null,
  approve_count: 0,
  decision: ''
})
const meetingFormRules = {
  application_id: [{ required: true, message: '请选择预备团员', trigger: 'change' }],
  self_application_path: [{ required: true, message: '请填写转正申请书路径', trigger: 'blur' }],
  meeting_at: [{ required: true, message: '请选择会议时间', trigger: 'change' }],
  expected_count: [{ required: true, message: '请输入应到人数', trigger: 'blur' }],
  actual_count: [{ required: true, message: '请输入实到人数', trigger: 'blur' }],
  decision: [{ required: true, message: '请选择决议结果', trigger: 'change' }]
}
// 绑定的学生：依据所选入团申请从 applications 列表中查找
const meetingBoundStudent = computed(() => {
  const id = meetingForm.value.application_id
  if (!id) return null
  return applications.value.find((a) => a.id === id) || null
})

function openMeetingDialog() {
  meetingForm.value = {
    application_id: null,
    self_application_path: '',
    meeting_at: '',
    expected_count: null,
    actual_count: null,
    approve_count: 0,
    decision: ''
  }
  meetingDialogVisible.value = true
}
async function handleCreateMeeting() {
  try { await meetingFormRef.value.validate() } catch { return }
  if (meetingForm.value.decision === 'pass') {
    try {
      await ElMessageBox.confirm(
        '决议为「通过」后，该同志将转为正式团员。是否确认？',
        '转正确认',
        { type: 'warning' }
      )
    } catch { return }
  }
  meetingSaving.value = true
  try {
    await tyProbationaryMeetingApi.create(meetingForm.value)
    ElMessage.success('转正大会记录已提交')
    meetingDialogVisible.value = false
    fetchMeetings()
  } catch (e) {} finally { meetingSaving.value = false }
}

// ========== 详情查看（绑定学生） ==========
// 预备期考察记录
const recordViewVisible = ref(false)
const recordViewLoading = ref(false)
const recordView = ref({})

async function openRecordView(row) {
  // 优先用列表行内已有字段快速回显，避免重复请求
  recordView.value = { ...row }
  recordViewVisible.value = true
  // 调详情接口拿最新 + 完整数据（再次补齐学生信息，防止列表未带）
  recordViewLoading.value = true
  try {
    const detail = await tyProbationaryRecordApi.get(row.id)
    recordView.value = { ...row, ...detail }
  } catch (e) {
    console.error('获取预备期考察记录详情失败', e)
  } finally {
    recordViewLoading.value = false
  }
}

// 转正大会
const meetingViewVisible = ref(false)
const meetingViewLoading = ref(false)
const meetingView = ref({})

async function openMeetingView(row) {
  meetingView.value = { ...row }
  meetingViewVisible.value = true
  meetingViewLoading.value = true
  try {
    const detail = await tyProbationaryMeetingApi.get(row.id)
    meetingView.value = { ...row, ...detail }
  } catch (e) {
    console.error('获取转正大会详情失败', e)
  } finally {
    meetingViewLoading.value = false
  }
}

// ========== 共享数据 ==========

onMounted(async () => {
  // 路由带 application_id 时预选过滤器（来自发展轨迹的「召开转正大会」按钮）
  if (route.query.application_id) {
    const id = Number(route.query.application_id)
    if (!Number.isNaN(id) && id > 0) activeAppFilter.value = id
  }
  await fetchApplications()
  // 预选后按当前 Tab 重新拉取列表（覆盖默认 fetch）
  if (activeTab.value === 'probationary-record') {
    await fetchRecords()
  } else {
    await fetchMeetings()
  }
})

watch(activeTab, (val) => {
  if (val === 'probationary-meeting') fetchMeetings()
})
</script>

<style scoped>
.card-header-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--sh-space-md);
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: var(--sh-space-xs);
}
.filter-label {
  color: var(--sh-text-regular);
  font-size: var(--sh-text-sm);
  white-space: nowrap;
}
/* .tab-header / .word-count 已在全局定义 */
.text-muted {
  color: var(--sh-text-placeholder, #c0c4cc);
}
.student-no {
  color: var(--sh-text-secondary, #909399);
  font-size: 12px;
  margin-left: 2px;
}
.student-card {
  margin-bottom: 8px;
}
.student-card :deep(.el-card__header) {
  padding: 10px 16px;
}
.student-card-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: var(--sh-text-primary, #303133);
}
.summary-content {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}
</style>

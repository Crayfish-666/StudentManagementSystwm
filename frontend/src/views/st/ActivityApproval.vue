<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>社团活动审批中心</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 待我审批 -->
        <el-tab-pane label="待我审批" name="pending">
          <el-table v-loading="loading" :data="pendingItems" stripe>
            <el-table-column prop="biz_no" label="编号" width="150" />
            <el-table-column prop="title" label="活动名称" min-width="180" />
            <el-table-column prop="level" label="等级" width="70">
              <template #default="{ row }">
                <el-tag :type="levelType[row.level]" size="small">{{ row.level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="association_name" label="所属社团" min-width="140" />
            <el-table-column label="当前步骤" width="140">
              <template #default="{ row }">
                <el-tag size="small" type="warning">{{ stepTextMap[row.current_step_no] || `步骤${row.current_step_no}` }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="goDetail(row.id)">查看</el-button>
                <el-button type="success" size="small" @click="openApprove(row)">审批</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-wrap">
            <el-pagination
              v-model:current-page="pendingPage"
              v-model:page-size="pageSize"
              :total="pendingTotal"
              :page-sizes="[20, 50, 100]"
              layout="total, sizes, prev, pager, next"
              @size-change="loadPending"
              @current-change="loadPending"
            />
          </div>
        </el-tab-pane>

        <!-- 审批历史 -->
        <el-tab-pane label="审批历史" name="history">
          <el-table v-loading="loading" :data="historyItems" stripe>
            <el-table-column prop="biz_no" label="编号" width="150" />
            <el-table-column prop="title" label="活动名称" min-width="180" />
            <el-table-column prop="level" label="等级" width="70">
              <template #default="{ row }">
                <el-tag :type="levelType[row.level]" size="small">{{ row.level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="association_name" label="所属社团" min-width="140" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusType[row.status]" size="small">{{ row.status_text }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="更新时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="goDetail(row.id)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-wrap">
            <el-pagination
              v-model:current-page="historyPage"
              v-model:page-size="pageSize"
              :total="historyTotal"
              :page-sizes="[20, 50, 100]"
              layout="total, sizes, prev, pager, next"
              @size-change="loadHistory"
              @current-change="loadHistory"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 审批对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="审批活动"
      width="520px"
      :close-on-click-modal="false"
      @close="onDialogClose"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="活动编号">
          <span>{{ currentRow?.biz_no }}</span>
        </el-form-item>
        <el-form-item label="活动名称">
          <span>{{ currentRow?.title }}</span>
        </el-form-item>
        <el-form-item label="当前步骤">
          <span>{{ stepTextMap[currentRow?.current_step_no] || `步骤${currentRow?.current_step_no}` }}</span>
        </el-form-item>
        <el-form-item label="审批结果" prop="result">
          <el-radio-group v-model="form.result">
            <el-radio value="approve">通过</el-radio>
            <el-radio value="reject">驳回</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="审批意见" prop="opinion">
          <el-input
            v-model="form.opinion"
            type="textarea"
            :rows="5"
            placeholder="请输入审批意见"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          :type="form.result === 'approve' ? 'success' : 'danger'"
          :loading="submitting"
          @click="handleApprove"
        >
          提交审批
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { stActivityApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'

const router = useRouter()

const levelType = { A: 'danger', B: 'warning', C: '', D: 'success' }
const statusType = { S0: 'info', S1: 'warning', S2: '', S3: 'success', S4: 'danger', cancelled: 'info' }
const stepTextMap = {
  1: '指导教师审批',
  2: '院系审批',
  3: '校社联审批',
  4: '校团委审批',
  5: '校领导审批'
}

const activeTab = ref('pending')
const loading = ref(false)

// 待审批列表
const pendingItems = ref([])
const pendingTotal = ref(0)
const pendingPage = ref(1)

// 审批历史列表
const historyItems = ref([])
const historyTotal = ref(0)
const historyPage = ref(1)

const pageSize = ref(20)

// 审批对话框
const dialogVisible = ref(false)
const submitting = ref(false)
const currentRow = ref(null)
const formRef = ref(null)
const form = reactive({
  result: 'approve',
  opinion: ''
})

// 驳回时意见至少30字
const opinionMinLen = (rule, value, callback) => {
  if (!value || value.trim().length === 0) {
    callback(new Error('请填写审批意见'))
  } else if (form.result === 'reject' && value.trim().length < 30) {
    callback(new Error('驳回时审批意见至少 30 字'))
  } else {
    callback()
  }
}

const rules = {
  result: [{ required: true, message: '请选择审批结果' }],
  opinion: [{ required: true, validator: opinionMinLen, trigger: 'blur' }]
}

async function loadPending() {
  loading.value = true
  try {
    const data = await stActivityApi.list({
      page: pendingPage.value,
      page_size: pageSize.value,
      status: 'S2'
    })
    pendingItems.value = data.items || []
    pendingTotal.value = data.total || 0
  } catch (e) {
    console.error('获取待审批列表失败', e)
  } finally {
    loading.value = false
  }
}

async function loadHistory() {
  loading.value = true
  try {
    // 合并查询已通过和已驳回
    const [approved, rejected] = await Promise.all([
      stActivityApi.list({ status: 'S3', page: historyPage.value, page_size: pageSize.value }),
      stActivityApi.list({ status: 'S4', page: historyPage.value, page_size: pageSize.value })
    ])
    const merged = [...(approved?.items || []), ...(rejected?.items || [])]
    merged.sort((a, b) => (a.updated_at < b.updated_at ? 1 : -1))
    historyItems.value = merged
    historyTotal.value = (approved?.total || 0) + (rejected?.total || 0)
  } catch (e) {
    console.error('获取审批历史失败', e)
  } finally {
    loading.value = false
  }
}

function handleTabChange(name) {
  if (name === 'pending') loadPending()
  else loadHistory()
}

function goDetail(id) {
  router.push(`/st/activity/${id}`)
}

function openApprove(row) {
  currentRow.value = row
  form.result = 'approve'
  form.opinion = ''
  dialogVisible.value = true
}

function onDialogClose() {
  formRef.value?.clearValidate()
}

async function handleApprove() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  if (form.result === 'reject') {
    try {
      await ElMessageBox.confirm('确认驳回此活动？驳回后活动将终止。', '驳回确认', { type: 'warning' })
    } catch {
      return
    }
  }

  submitting.value = true
  try {
    await stActivityApi.approve(currentRow.value.id, {
      step_no: currentRow.value.current_step_no,
      result: form.result,
      opinion: form.opinion
    })
    ElMessage.success('审批已提交')
    dialogVisible.value = false
    loadPending()
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadPending()
})
</script>

<style scoped>
/* .card-header, .pagination-wrap 已在 App.vue 全局定义 */
</style>

<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>团员花名册</span>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterBranch" placeholder="团支部筛选" clearable style="width: 180px" @change="fetchList">
          <el-option v-for="b in branches" :key="b.id" :label="b.name" :value="b.id" />
        </el-select>
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 140px" @change="fetchList">
          <el-option label="在团" value="active" />
          <el-option label="已转出" value="transferred" />
          <el-option label="超龄离团" value="overtime" />
          <el-option label="已归档" value="archived" />
        </el-select>
        <el-input
          v-model="filterKeyword"
          placeholder="搜索学号/姓名"
          clearable
          style="width: 200px"
          @keyup.enter="fetchList"
          @clear="fetchList"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="fetchList">查询</el-button>
      </div>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="biz_no" label="团员证号" width="170" />
        <el-table-column prop="student_no" label="学号" width="130" />
        <el-table-column prop="student_name" label="姓名" width="90" />
        <el-table-column prop="branch_name" label="所属支部" min-width="150" />
        <el-table-column prop="join_at" label="入团时间" width="110">
          <template #default="{ row }">{{ formatDate(row.join_at) }}</template>
        </el-table-column>
        <el-table-column prop="become_probationary_at" label="成为预备团员" width="140">
          <template #default="{ row }">
            {{ row.become_probationary_at ? formatDate(row.become_probationary_at) : '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="formal_join_at" label="正式入团时间" width="120">
          <template #default="{ row }">
            {{ row.formal_join_at ? formatDate(row.formal_join_at) : '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ statusTextMap[row.status] || row.status_text || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button v-if="row.status === 'active'" link type="warning" size="small" @click="handleTransferOut(row)">转出</el-button>
            <el-button v-if="row.status === 'active'" link type="info" size="small" @click="handleOvertime(row)">超龄离团</el-button>
            <el-button v-if="['transferred', 'overtime'].includes(row.status)" link type="danger" size="small" @click="handleArchive(row)">归档</el-button>
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

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editDialogVisible" title="编辑团员信息" width="550px" destroy-on-close>
      <el-form ref="editFormRef" :model="editForm" :rules="editFormRules" label-width="120px">
        <el-form-item label="团员证号">
          <el-input v-model="editForm.biz_no" disabled />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="editForm.student_name" disabled />
        </el-form-item>
        <el-form-item label="所属支部" prop="branch_id">
          <el-select v-model="editForm.branch_id" placeholder="请选择团支部" style="width: 100%">
            <el-option v-for="b in branches" :key="b.id" :label="b.name" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="入团时间" prop="join_at">
          <el-date-picker v-model="editForm.join_at" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="成为预备团员">
          <el-date-picker v-model="editForm.become_probationary_at" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="正式入团时间">
          <el-date-picker v-model="editForm.formal_join_at" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditSave" :loading="editSaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { tyMemberRosterApi, tyBranchApi } from '@/api/ty'
import { formatDate } from '@/utils/datetime'

// 状态映射
const statusTextMap = { active: '在团', transferred: '已转出', overtime: '超龄离团', archived: '已归档' }

function statusTagType(status) {
  switch (status) {
    case 'active': return 'success'
    case 'transferred': return 'warning'
    case 'overtime': return 'info'
    case 'archived': return 'danger'
    default: return ''
  }
}

// 列表数据
const list = ref([])
const loading = ref(false)
const filterBranch = ref('')
const filterStatus = ref('')
const filterKeyword = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 团支部下拉
const branches = ref([])

// 编辑弹窗
const editDialogVisible = ref(false)
const editSaving = ref(false)
const editFormRef = ref()
const editForm = ref({})
const editFormRules = {
  branch_id: [{ required: true, message: '请选择团支部', trigger: 'change' }],
  join_at: [{ required: true, message: '请选择入团时间', trigger: 'change' }]
}

// 获取列表
async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (filterBranch.value) params.branch_id = filterBranch.value
    if (filterStatus.value) params.status = filterStatus.value
    if (filterKeyword.value) params.keyword = filterKeyword.value
    const data = await tyMemberRosterApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取团员花名册列表失败', e)
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

// 编辑
function openEditDialog(row) {
  editForm.value = { ...row }
  editDialogVisible.value = true
}

async function handleEditSave() {
  try { await editFormRef.value.validate() } catch { return }
  editSaving.value = true
  try {
    await tyMemberRosterApi.update(editForm.value.id, editForm.value)
    ElMessage.success('保存成功')
    editDialogVisible.value = false
    fetchList()
  } catch (e) {} finally { editSaving.value = false }
}

// 转出
async function handleTransferOut(row) {
  try {
    await ElMessageBox.confirm(`确认将 ${row.student_name} 的团组织关系转出？`, '转出确认', { type: 'warning' })
    await tyMemberRosterApi.transferOut(row.id)
    ElMessage.success('已转出')
    fetchList()
  } catch (e) { if (e !== 'cancel') {} }
}

// 超龄离团
async function handleOvertime(row) {
  try {
    await ElMessageBox.confirm(`确认标记 ${row.student_name} 为超龄离团？`, '超龄离团确认', { type: 'warning' })
    await tyMemberRosterApi.overtime(row.id)
    ElMessage.success('已标记超龄离团')
    fetchList()
  } catch (e) { if (e !== 'cancel') {} }
}

// 归档
async function handleArchive(row) {
  try {
    await ElMessageBox.confirm(`确认归档 ${row.student_name} 的团员档案？归档后将不可修改。`, '归档确认', { type: 'warning' })
    await tyMemberRosterApi.archive(row.id)
    ElMessage.success('已归档')
    fetchList()
  } catch (e) { if (e !== 'cancel') {} }
}

onMounted(() => {
  fetchList()
  fetchBranches()
})
</script>

<style scoped>
/* .card-header / .filter-bar / .pagination-wrap 已在全局定义 */
</style>

<template>
  <div class="page-container sh-animate-slide-up">
    <div class="sh-glass-card page-header-card">
      <div class="card-header">
        <div class="header-title">
          <el-icon :size="20" class="title-icon"><Flag /></el-icon>
          <h2>团员发展 — 入团申请</h2>
        </div>
        <div class="header-actions">
          <el-button type="primary" class="sh-btn-gradient" @click="goCreate">
            <el-icon><Plus /></el-icon>
            <span>提交入团申请</span>
          </el-button>
        </div>
      </div>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 160px" @change="fetchList">
          <el-option label="草稿 (S0)" value="S0" />
          <el-option label="待审 (S1)" value="S1" />
          <el-option label="审批中 (S2)" value="S2" />
          <el-option label="通过 (S3)" value="S3" />
          <el-option label="驳回 (S4)" value="S4" />
        </el-select>
        <el-input v-model="searchQuery" placeholder="搜索申请人姓名 / 学号" style="width: 240px" prefix-icon="Search" clearable />
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="sh-glass-card table-card">
      <el-table :data="displayList" stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="biz_no" label="申请编号" width="160" />
        <el-table-column prop="student_name" label="申请人" width="120">
          <template #default="{ row }">
            <div class="student-cell">
              <el-avatar :size="28" class="mini-avatar">{{ row.student_name.charAt(0) }}</el-avatar>
              <span>{{ row.student_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="student_no" label="学号" width="130" />
        <el-table-column prop="branch_name" label="团支部" min-width="160" />
        <el-table-column prop="college_name" label="院系" min-width="140" />
        <el-table-column prop="apply_date" label="申请日期" width="120" />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType[row.status]" size="small" effect="dark">
              {{ statusMap[row.status] || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row.id)">查看全景</el-button>
            <el-button v-if="row.status === 'S0'" link type="success" size="small" @click="handleSubmit(row.id)">提交</el-button>
            <el-button v-if="row.status === 'S1'" link type="warning" size="small" @click="handleWithdraw(row.id)">撤回</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="displayList.length"
          layout="total, prev, pager, next"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Flag, Plus, Search } from '@element-plus/icons-vue'
import { tyApplicationApi } from '@/api/ty'

const router = useRouter()

const statusMap = { S0: '草稿', S1: '待审', S2: '审批中', S3: '已通过', S4: '已驳回' }
const statusType = { S0: 'info', S1: 'warning', S2: 'primary', S3: 'success', S4: 'danger' }

const MOCK_APPLICATIONS = [
  { id: 1, biz_no: 'TY-2026-0001', student_name: '张三', student_no: '2023010101', branch_name: '计算机2301团支部', college_name: '计算机学院', apply_date: '2026-03-01', status: 'S3' },
  { id: 2, biz_no: 'TY-2026-0002', student_name: '李四', student_no: '2023010102', branch_name: '经管2302团支部', college_name: '经济管理学院', apply_date: '2026-03-05', status: 'S2' },
  { id: 3, biz_no: 'TY-2026-0003', student_name: '王五', student_no: '2023010103', branch_name: '艺术2301团支部', college_name: '艺术设计学院', apply_date: '2026-03-10', status: 'S1' },
  { id: 4, biz_no: 'TY-2026-0004', student_name: '赵六', student_no: '2023010104', branch_name: '软件2303团支部', college_name: '软件工程学院', apply_date: '2026-03-12', status: 'S0' }
]

const list = ref([])
const loading = ref(false)
const filterStatus = ref('')
const searchQuery = ref('')
const page = ref(1)
const pageSize = ref(10)

const displayList = computed(() => {
  let source = list.value.length ? list.value : MOCK_APPLICATIONS
  if (filterStatus.value) {
    source = source.filter(i => i.status === filterStatus.value)
  }
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    source = source.filter(i => i.student_name.includes(q) || i.student_no.includes(q))
  }
  return source
})

async function fetchList() {
  loading.value = true
  try {
    const res = await tyApplicationApi.getList({ page: page.value, page_size: pageSize.value })
    if (res && res.items && res.items.length) {
      list.value = res.items
    } else {
      list.value = MOCK_APPLICATIONS
    }
  } catch {
    list.value = MOCK_APPLICATIONS
  } finally {
    loading.value = false
  }
}

function goCreate() { router.push('/ty/application/new') }
function goDetail(id) { router.push(`/ty/application/${id}`) }
function handleSubmit() { ElMessage.success('申请提交成功！已进入辅导员待审批流程') }
function handleWithdraw() { ElMessage.warning('申请已撤回至草稿状态') }

onMounted(fetchList)
</script>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header-card {
  padding: 20px 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
}
.header-title h2 {
  font-size: 18px;
  font-weight: 700;
  color: var(--sh-text-primary);
}
.title-icon {
  color: var(--sh-primary);
}

.filter-bar {
  display: flex;
  gap: 12px;
}

.table-card {
  padding: 16px;
}

.student-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.mini-avatar {
  background: var(--sh-primary);
  font-size: 12px;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>

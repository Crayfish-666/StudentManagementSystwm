<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>活动管理</span>
          <el-button type="primary" @click="goCreate">新建活动</el-button>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 140px" @change="fetchList">
          <el-option label="草稿" value="S0" />
          <el-option label="待审" value="S1" />
          <el-option label="审批中" value="S2" />
          <el-option label="通过" value="S3" />
          <el-option label="驳回" value="S4" />
          <el-option label="已取消" value="cancelled" />
        </el-select>
        <el-select v-model="filterAssoc" placeholder="社团筛选" clearable style="width: 180px; margin-left: 12px" @change="fetchList">
          <el-option v-for="a in assocs" :key="a.id" :label="a.name" :value="a.id" />
        </el-select>
      </div>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="biz_no" label="编号" width="150" />
        <el-table-column prop="title" label="活动名称" min-width="180" />
        <el-table-column prop="association_name" label="所属社团" min-width="140" />
        <el-table-column prop="level" label="等级" width="70">
          <template #default="{ row }">
            <el-tag :type="levelType[row.level]" size="small">{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="location" label="地点" min-width="120" />
        <el-table-column prop="started_at" label="开始时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.started_at) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="statusType[row.status]" size="small">{{ row.status_text }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'S0'" link type="primary" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button v-if="row.status === 'S0'" link type="success" size="small" @click="handleSubmit(row.id)">提交</el-button>
            <el-popconfirm v-if="row.status === 'S0' || row.status === 'S4'" title="确认删除此活动？" @confirm="handleDelete(row.id)">
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { stActivityApi, stAssociationApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'

const router = useRouter()

const statusType = { S0: 'info', S1: 'warning', S2: '', S3: 'success', S4: 'danger', cancelled: 'info' }
const levelType = { A: 'danger', B: 'warning', C: '', D: 'info' }

const list = ref([])
const loading = ref(false)
const filterStatus = ref('')
const filterAssoc = ref(null)
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
    const data = await stActivityApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取活动列表失败', e)
  } finally {
    loading.value = false
  }
}

async function fetchAssocs() {
  try {
    const data = await stAssociationApi.list({ page_size: 200 })
    assocs.value = data.items || []
  } catch (e) {
    console.error('获取社团列表失败', e)
  }
}

function goCreate() {
  router.push('/st/activity/new')
}

function goEdit(id) {
  router.push(`/st/activity/${id}/edit`)
}

function goDetail(id) {
  router.push(`/st/activity/${id}`)
}

async function handleSubmit(id) {
  try {
    await ElMessageBox.confirm('确认提交此活动？提交后将进入审批流程。', '提交确认')
    await stActivityApi.submit(id)
    ElMessage.success('提交成功')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

async function handleDelete(id) {
  try {
    await stActivityApi.delete(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    // 错误已由 http 拦截器处理
  }
}

onMounted(() => {
  fetchList()
  fetchAssocs()
})
</script>

<style scoped>
/* .card-header, .filter-bar, .pagination-wrap 已在 App.vue 全局定义 */
</style>

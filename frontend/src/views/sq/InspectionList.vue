<template>
  <div class="page-container">
    <h2>巡查记录</h2>

    <!-- 筛选栏 -->
    <el-form :inline="true" :model="query" class="filter-bar">
      <el-form-item label="巡查类型">
        <el-select v-model="query.inspection_type" placeholder="全部" clearable>
          <el-option
            v-for="(label, key) in typeDict"
            :key="key"
            :label="label"
            :value="key"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="楼栋">
        <el-select v-model="query.building_id" placeholder="全部" clearable>
          <el-option
            v-for="b in buildings"
            :key="b.id"
            :label="b.name"
            :value="b.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="fetchList">搜索</el-button>
      </el-form-item>
    </el-form>

    <!-- 操作栏 -->
    <div class="action-bar">
      <el-button type="primary" @click="router.push('/sq/inspection/new')">新增巡查</el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="biz_no" label="业务编号" min-width="160" />
      <el-table-column prop="inspection_type_text" label="巡查类型" min-width="100" />
      <el-table-column prop="building_name" label="楼栋" min-width="100" />
      <el-table-column prop="floor_no" label="楼层" min-width="80" />
      <el-table-column prop="room_no" label="寝室" min-width="80" />
      <el-table-column prop="inspector_name" label="巡查人" min-width="90" />
      <el-table-column prop="inspected_at" label="巡查时间" min-width="170">
        <template #default="{ row }">{{ formatDateTime(row.inspected_at) }}</template>
      </el-table-column>
      <el-table-column label="分数" min-width="80">
        <template #default="{ row }">
          <el-tag :type="scoreTagType(row.score)">{{ row.score }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status_text" label="状态" min-width="80" />
      <el-table-column label="操作" min-width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/sq/inspection/${row.id}`)">查看详情</el-button>
          <el-popconfirm title="确认删除该记录？" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sqInspectionApi, sqBuildingApi } from '@/api/sq'
import { formatDateTime } from '@/utils/datetime'

const router = useRouter()

// 巡查类型字典
const typeDict = {
  hygiene: '卫生巡查',
  late_return: '晚归检查',
  appliance: '违规电器',
  safety: '安全隐患',
  fire_lane: '消防通道'
}

// 查询参数
const query = reactive({
  inspection_type: '',
  building_id: '',
  page: 1,
  page_size: 10
})

// 列表数据
const list = ref([])
const total = ref(0)
const loading = ref(false)

// 楼栋下拉
const buildings = ref([])

// 分数标签颜色
function scoreTagType(score) {
  if (score >= 90) return 'success'
  if (score >= 60) return 'warning'
  return 'danger'
}

// 获取楼栋列表
async function fetchBuildings() {
  try {
    const data = await sqBuildingApi.list()
    buildings.value = data?.items || []
  } catch {
    // 静默处理
  }
}

// 获取巡查列表
async function fetchList() {
  loading.value = true
  try {
    const params = { ...query }
    // 清除空值参数
    Object.keys(params).forEach(k => {
      if (params[k] === '' || params[k] === null || params[k] === undefined) {
        delete params[k]
      }
    })
    const data = await sqInspectionApi.list(params)
    list.value = data?.items || []
    total.value = data?.total || 0
  } catch {
    ElMessage.error('获取巡查列表失败')
  } finally {
    loading.value = false
  }
}

// 删除
async function handleDelete(id) {
  try {
    await sqInspectionApi.delete(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  fetchBuildings()
  fetchList()
})
</script>

<style scoped>
/* .page-container, .filter-bar, .action-bar, .pagination-wrap 已在 App.vue 全局定义 */
.filter-bar :deep(.el-select) {
  width: 180px;
}
</style>

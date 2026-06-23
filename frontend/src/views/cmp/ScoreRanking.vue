<template>
  <div class="page-container score-ranking">
    <h2 class="page-title">综合分排行</h2>

    <el-card shadow="never">
      <!-- 筛选栏 -->
      <el-form :inline="true" :model="query" class="filter-bar">
        <el-form-item label="学年">
          <el-input
            v-model="query.term"
            placeholder="如 2025-2026"
            clearable
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="院系">
          <el-select
            v-model="query.college_id"
            placeholder="全部院系"
            clearable
            style="width: 200px"
          >
            <el-option
              v-for="c in colleges"
              :key="c.id"
              :label="c.name"
              :value="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
        <el-form-item style="float: right">
          <el-button type="warning" :loading="batchLoading" @click="onBatchRecompute">
            <el-icon><Refresh /></el-icon>
            <span>批量重算</span>
          </el-button>
          <el-button type="success" @click="onExportExcel">
            <el-icon><Download /></el-icon>
            <span>导出 Excel</span>
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table
        :data="list"
        v-loading="loading"
        border
        stripe
        :default-sort="{ prop: 'total_score', order: 'descending' }"
      >
        <el-table-column type="index" label="排名" width="70" fixed="left" />
        <el-table-column prop="student_no" label="学号" width="120" />
        <el-table-column prop="student_name" label="姓名" width="100" fixed="left" />
        <el-table-column prop="college_name" label="院系" min-width="140" />
        <el-table-column prop="college_class_name" label="班级" min-width="140" />
        <el-table-column prop="academic_year" label="学年" width="120" />
        <el-table-column prop="total_score" label="综合分" width="120" sortable>
          <template #default="{ row }">
            <span class="score-num" :class="scoreClass(row.total_score)">
              {{ formatScore(row.total_score) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="rank_in_class" label="班级排名" width="100">
          <template #default="{ row }">
            <span v-if="row.rank_in_class">第 {{ row.rank_in_class }} 名</span>
            <span v-else class="empty">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="rank_in_college" label="院系排名" width="100">
          <template #default="{ row }">
            <span v-if="row.rank_in_college">第 {{ row.rank_in_college }} 名</span>
            <span v-else class="empty">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="computed_at" label="计算时间" min-width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.computed_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="onViewDetail(row)">
              查看明细
            </el-button>
            <el-button link type="warning" size="small" @click="onRecompute(row)">
              重算
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 明细对话框 -->
    <el-dialog
      v-model="detailVisible"
      :title="`综合分明细：${detail.student_name || ''}`"
      width="720px"
      destroy-on-close
    >
      <el-descriptions :column="2" border size="small" v-if="detailVisible">
        <el-descriptions-item label="学号">{{ detail.student_no }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.student_name }}</el-descriptions-item>
        <el-descriptions-item label="院系">{{ detail.college_name }}</el-descriptions-item>
        <el-descriptions-item label="班级">{{ detail.college_class_name }}</el-descriptions-item>
        <el-descriptions-item label="学年">{{ detail.academic_year }}</el-descriptions-item>
        <el-descriptions-item label="规则版本">{{ detail.rule_version }}</el-descriptions-item>
        <el-descriptions-item label="总分">
          <span class="score-num" :class="scoreClass(detail.total_score)">
            {{ formatScore(detail.total_score) }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="计算时间">
          {{ formatDateTime(detail.computed_at) }}
        </el-descriptions-item>
      </el-descriptions>

      <el-table
        :data="detail.details || []"
        border
        size="small"
        style="margin-top: 12px"
      >
        <el-table-column prop="dimension_zh" label="维度" width="120" />
        <el-table-column prop="sub_item" label="子项" min-width="140" />
        <el-table-column prop="raw_value" label="原始数据" min-width="120" />
        <el-table-column prop="source_module" label="来源" width="80" />
        <el-table-column prop="score" label="得分" width="90">
          <template #default="{ row }">{{ formatScore(row.score) }}</template>
        </el-table-column>
        <el-table-column prop="max" label="满分" width="80">
          <template #default="{ row }">{{ formatScore(row.max) }}</template>
        </el-table-column>
        <el-table-column prop="weight" label="权重" width="80">
          <template #default="{ row }">{{ (row.weight * 100).toFixed(0) }}%</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Download } from '@element-plus/icons-vue'
import { cmpScoreApi } from '@/api/cmp'
import { collegeApi } from '@/api/sys-org'
import { formatDateTime } from '@/utils/datetime'

const query = reactive({
  term: '',
  college_id: null,
  page: 1,
  page_size: 20
})

const list = ref([])
const total = ref(0)
const loading = ref(false)
const batchLoading = ref(false)
const colleges = ref([])

const detailVisible = ref(false)
const detail = reactive({
  student_no: '',
  student_name: '',
  college_name: '',
  college_class_name: '',
  academic_year: '',
  rule_version: '',
  total_score: 0,
  computed_at: '',
  details: []
})

function formatScore(n) {
  if (n == null || isNaN(n)) return '0.00'
  return Number(n).toFixed(2)
}

function scoreClass(s) {
  if (s >= 85) return 'high'
  if (s >= 70) return 'mid'
  return 'low'
}

async function fetchList() {
  loading.value = true
  try {
    const params = {
      page: query.page,
      page_size: query.page_size
    }
    if (query.term) params.term = query.term
    if (query.college_id) params.college_id = query.college_id
    const data = await cmpScoreApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取综合分列表失败', e)
    ElMessage.error('加载综合分列表失败')
  } finally {
    loading.value = false
  }
}

async function fetchColleges() {
  try {
    const data = await collegeApi.list()
    colleges.value = data.items || data || []
  } catch (e) {
    console.error('获取院系列表失败', e)
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

function onReset() {
  query.term = ''
  query.college_id = null
  query.page = 1
  fetchList()
}

async function onViewDetail(row) {
  try {
    const data = await cmpScoreApi.get(row.student_id, row.academic_year)
    Object.assign(detail, {
      student_no: data.student_no || row.student_no,
      student_name: data.student_name || row.student_name,
      college_name: data.college_name || row.college_name,
      college_class_name: data.college_class_name || row.college_class_name,
      academic_year: data.academic_year || row.academic_year,
      rule_version: data.rule_version || '-',
      total_score: data.total_score || 0,
      computed_at: data.computed_at || row.computed_at,
      details: data.details || []
    })
    detailVisible.value = true
  } catch (e) {
    console.error('获取综合分明细失败', e)
    ElMessage.error('获取综合分明细失败')
  }
}

async function onRecompute(row) {
  try {
    await ElMessageBox.confirm(
      `确认重算学生「${row.student_name}」的综合分？`,
      '重算确认',
      { type: 'warning' }
    )
  } catch {
    return
  }
  try {
    await cmpScoreApi.recompute(row.student_id, row.academic_year)
    ElMessage.success('重算完成')
    fetchList()
  } catch (e) {
    ElMessage.error('重算失败')
  }
}

async function onBatchRecompute() {
  try {
    await ElMessageBox.confirm(
      '确认批量重算所选范围学生的综合分？此操作可能耗时较长。',
      '批量重算',
      { type: 'warning' }
    )
  } catch {
    return
  }
  batchLoading.value = true
  try {
    const payload = {}
    if (query.term) payload.term = query.term
    if (query.college_id) payload.college_id = query.college_id
    const res = await cmpScoreApi.batchCompute(payload)
    ElMessage.success(
      `批量重算完成：${res.recomputed_count || 0} 名学生（学年 ${res.academic_year || '-'}）`
    )
    fetchList()
  } catch (e) {
    ElMessage.error('批量重算失败')
  } finally {
    batchLoading.value = false
  }
}

function onExportExcel() {
  // 简单 CSV 导出（Excel 可打开），避免引入额外依赖
  if (!list.value.length) {
    ElMessage.warning('暂无数据可导出')
    return
  }
  const headers = ['排名', '学号', '姓名', '院系', '班级', '学年', '综合分', '班级排名', '院系排名', '计算时间']
  const rows = list.value.map((r, i) => [
    i + 1,
    r.student_no,
    r.student_name,
    r.college_name,
    r.college_class_name,
    r.academic_year,
    formatScore(r.total_score),
    r.rank_in_class || '-',
    r.rank_in_college || '-',
    formatDateTime(r.computed_at)
  ])
  const csv = [headers, ...rows]
    .map((line) => line.map((c) => `"${String(c == null ? '' : c).replace(/"/g, '""')}"`).join(','))
    .join('\n')
  // 加 BOM 避免 Excel 打开中文乱码
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `综合分排行_${query.term || '全部'}_${formatDateTime(new Date(), 'yyyyMMddHHmmss')}.csv`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

onMounted(() => {
  fetchColleges()
  fetchList()
})
</script>

<style scoped>
.score-ranking {
  padding: var(--sh-space-lg);
}
/* .page-title 已在 App.vue 全局定义 */
/* .filter-bar, .pagination-wrap 已在 App.vue 全局定义 */
.score-num {
  font-weight: 700;
}
.score-num.high { color: var(--sh-success); }
.score-num.mid { color: var(--sh-warning); }
.score-num.low { color: var(--sh-danger); }
.empty {
  color: var(--sh-text-disabled);
}
</style>

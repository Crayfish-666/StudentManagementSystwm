<template>
  <div class="page-container dashboard">
    <h2 class="page-title">综合看板</h2>

    <el-skeleton v-if="loading" :rows="10" animated />

    <template v-else>
      <!-- 顶部 KPI 卡片 -->
      <el-row :gutter="16" class="kpi-row">
        <el-col :span="6" v-for="kpi in kpiCards" :key="kpi.key">
          <el-card shadow="hover" class="kpi-card" :body-style="{ padding: '16px' }">
            <div class="kpi-label">{{ kpi.label }}</div>
            <div class="kpi-value" :style="{ color: kpi.color }">
              {{ kpi.value }}
              <span v-if="kpi.unit" class="kpi-unit">{{ kpi.unit }}</span>
            </div>
            <div class="kpi-foot">{{ kpi.foot }}</div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 图表三联 -->
      <el-row :gutter="16" class="chart-row">
        <el-col :span="12">
          <el-card shadow="hover">
            <template #header>
              <div class="card-header">
                <span>推优通过率趋势（近 {{ trendRangeLabel }}）</span>
                <el-radio-group v-model="trendRange" size="small" @change="fetchTrend">
                  <el-radio-button label="3m">3月</el-radio-button>
                  <el-radio-button label="6m">6月</el-radio-button>
                  <el-radio-button label="12m">12月</el-radio-button>
                </el-radio-group>
              </div>
            </template>
            <div ref="trendEl" class="chart-box" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="hover">
            <template #header>
              <div class="card-header">
                <span>各院系活跃社团数</span>
              </div>
            </template>
            <div ref="barEl" class="chart-box" />
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" class="chart-row">
        <el-col :span="12">
          <el-card shadow="hover">
            <template #header>
              <div class="card-header">
                <span>事件等级分布</span>
              </div>
            </template>
            <div ref="pieEl" class="chart-box" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="hover">
            <template #header>
              <div class="card-header">
                <span>综合分分布</span>
                <el-radio-group v-model="distDim" size="small" @change="fetchDistribution">
                  <el-radio-button label="college">按院系</el-radio-button>
                  <el-radio-button label="gender">按性别</el-radio-button>
                  <el-radio-button label="grade">按年级</el-radio-button>
                  <el-radio-button label="score_range">按分数段</el-radio-button>
                </el-radio-group>
              </div>
            </template>
            <div ref="distEl" class="chart-box" />
          </el-card>
        </el-col>
      </el-row>

      <!-- Top10 学生综合分 -->
      <el-card shadow="hover" class="top-card">
        <template #header>
          <div class="card-header">
            <span>综合分 Top10（{{ topList.length }} / {{ topTotal }}）</span>
            <el-button link type="primary" @click="goRanking">查看完整排行 →</el-button>
          </div>
        </template>
        <el-table :data="topList" border stripe size="small">
          <el-table-column type="index" label="排名" width="70" fixed="left" />
          <el-table-column prop="student_no" label="学号" width="120" />
          <el-table-column prop="student_name" label="姓名" width="100" />
          <el-table-column prop="college_name" label="院系" min-width="140" />
          <el-table-column prop="college_class_name" label="班级" min-width="140" />
          <el-table-column prop="academic_year" label="学年" width="120" />
          <el-table-column prop="total_score" label="综合分" width="120">
            <template #default="{ row }">
              <span class="score-num" :class="scoreClass(row.total_score)">
                {{ formatScore(row.total_score) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="computed_at" label="计算时间" min-width="160">
            <template #default="{ row }">{{ formatDateTime(row.computed_at) }}</template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { cmpDashboardApi, cmpScoreApi } from '@/api/cmp'
import { createChart, bindResize, disposeChart } from '@/utils/echarts'
import { formatDateTime } from '@/utils/datetime'

const router = useRouter()

const loading = ref(true)
const kpi = reactive({
  active_assoc: 0,
  ty_pass_rate: 0,
  l4_incidents_30d: 0,
  qg_payroll_amount_cents: 0,
  excellent_count: 0,
  student_count: 0,
  total_scored: 0
})

const kpiCards = computed(() => [
  {
    key: 'active_assoc',
    label: '活跃社团数',
    value: kpi.active_assoc,
    unit: '个',
    color: '#67c23a',
    foot: `覆盖学生 ${kpi.student_count} 人`
  },
  {
    key: 'ty_pass_rate',
    label: '推优通过率',
    value: (kpi.ty_pass_rate * 100).toFixed(1),
    unit: '%',
    color: '#e6a23c',
    foot: `已评优 ${kpi.excellent_count} 人（综合分 ≥ 85）`
  },
  {
    key: 'l4_incidents_30d',
    label: '近 30 天 L4 事件',
    value: kpi.l4_incidents_30d,
    unit: '起',
    color: '#f56c6c',
    foot: '需教师角色闭环处置'
  },
  {
    key: 'qg_payroll_amount',
    label: '本月薪酬总额',
    value: (kpi.qg_payroll_amount_cents / 100).toLocaleString('zh-CN', { maximumFractionDigits: 2 }),
    unit: '元',
    color: '#409eff',
    foot: `已计算综合分 ${kpi.total_scored} 人`
  }
])

const trendRange = ref('12m')
const trendRangeLabel = computed(() => {
  const map = { '3m': '3 个月', '6m': '6 个月', '12m': '12 个月' }
  return map[trendRange.value] || '12 个月'
})

const distDim = ref('college')

const topList = ref([])
const topTotal = ref(0)

const trendEl = ref(null)
const barEl = ref(null)
const pieEl = ref(null)
const distEl = ref(null)

let trendChart = null
let barChart = null
let pieChart = null
let distChart = null
let offTrend = () => {}
let offBar = () => {}
let offPie = () => {}
let offDist = () => {}

function formatScore(n) {
  if (n == null || isNaN(n)) return '0.00'
  return Number(n).toFixed(2)
}

function scoreClass(s) {
  if (s >= 85) return 'high'
  if (s >= 70) return 'mid'
  return 'low'
}

async function fetchKpi() {
  const data = await cmpDashboardApi.kpi()
  Object.assign(kpi, data)
}

async function fetchTrend() {
  const data = await cmpDashboardApi.trends('ty_pass_rate', trendRange.value)
  await nextTick()
  renderTrend(data.points || [])
}

async function fetchActiveAssoc() {
  const data = await cmpDashboardApi.activeAssocByCollege()
  await nextTick()
  renderBar(data.buckets || [])
}

async function fetchIncidentLevel() {
  const data = await cmpDashboardApi.incidentLevel()
  await nextTick()
  renderPie(data.buckets || [])
}

async function fetchDistribution() {
  const data = await cmpDashboardApi.distribution(distDim.value)
  await nextTick()
  renderDist(data.buckets || [])
}

async function fetchTop() {
  const data = await cmpScoreApi.list({ page: 1, page_size: 10 })
  topList.value = data.items || []
  topTotal.value = data.total || 0
}

function renderTrend(points) {
  if (!trendEl.value) return
  disposeChart(trendChart)
  const option = {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 30, bottom: 30 },
    xAxis: {
      type: 'category',
      data: points.map((p) => p.label),
      axisLabel: { color: '#606266' }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: (v) => `${(v * 100).toFixed(0)}%`,
        color: '#606266'
      },
      max: 1
    },
    series: [
      {
        name: '推优通过率',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 7,
        data: points.map((p) => Number(p.value.toFixed(4))),
        lineStyle: { color: '#409eff', width: 2 },
        itemStyle: { color: '#409eff' },
        areaStyle: { color: 'rgba(64,158,255,0.18)' }
      }
    ]
  }
  trendChart = createChart(trendEl.value, option)
  offTrend = bindResize(trendChart)
}

function renderBar(buckets) {
  if (!barEl.value) return
  disposeChart(barChart)
  const option = {
    tooltip: { trigger: 'axis' },
    grid: { left: 60, right: 20, top: 20, bottom: 60 },
    xAxis: {
      type: 'category',
      data: buckets.map((b) => b.label),
      axisLabel: {
        color: '#606266',
        interval: 0,
        rotate: buckets.length > 6 ? 30 : 0
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#606266' }
    },
    series: [
      {
        name: '活跃社团数',
        type: 'bar',
        data: buckets.map((b) => b.value),
        barWidth: '50%',
        itemStyle: {
          color: '#67c23a',
          borderRadius: [4, 4, 0, 0]
        },
        label: {
          show: true,
          position: 'top',
          color: '#606266'
        }
      }
    ]
  }
  barChart = createChart(barEl.value, option)
  offBar = bindResize(barChart)
}

function renderPie(buckets) {
  if (!pieEl.value) return
  disposeChart(pieChart)
  const option = {
    tooltip: { trigger: 'item', formatter: '{b}：{c}（{d}%）' },
    legend: {
      orient: 'vertical',
      left: 'left',
      top: 'middle',
      textStyle: { color: '#606266' }
    },
    series: [
      {
        name: '事件等级',
        type: 'pie',
        radius: ['38%', '70%'],
        center: ['65%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        label: { show: true, formatter: '{b}\n{c}' },
        data: buckets.map((b) => ({ name: b.label, value: b.value })),
        color: ['#67c23a', '#e6a23c', '#f56c6c', '#909399']
      }
    ]
  }
  pieChart = createChart(pieEl.value, option)
  offPie = bindResize(pieChart)
}

function renderDist(buckets) {
  if (!distEl.value) return
  disposeChart(distChart)
  const option = {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 30, bottom: 30 },
    xAxis: {
      type: 'category',
      data: buckets.map((b) => b.label),
      axisLabel: { color: '#606266' }
    },
    yAxis: { type: 'value', axisLabel: { color: '#606266' } },
    series: [
      {
        name: '人数',
        type: 'bar',
        data: buckets.map((b) => b.value),
        barWidth: '50%',
        itemStyle: { color: '#409eff', borderRadius: [4, 4, 0, 0] }
      }
    ]
  }
  distChart = createChart(distEl.value, option)
  offDist = bindResize(distChart)
}

function goRanking() {
  router.push('/cmp/ranking')
}

onMounted(async () => {
  loading.value = true
  try {
    // 使用 allSettled 容错：单个 API 失败不影响其他数据展示
    const results = await Promise.allSettled([
      cmpDashboardApi.kpi(),
      cmpDashboardApi.trends('ty_pass_rate', trendRange.value),
      cmpDashboardApi.activeAssocByCollege(),
      cmpDashboardApi.incidentLevel(),
      cmpDashboardApi.distribution(distDim.value),
      cmpScoreApi.list({ page: 1, page_size: 10 })
    ])

    const [kpiR, trendR, assocR, incidentR, distR, topR] = results

    // 填充数据（每个都判断是否成功）
    if (kpiR.status === 'fulfilled') Object.assign(kpi, kpiR.value)
    if (topR.status === 'fulfilled') {
      topList.value = topR.value.items || []
      topTotal.value = topR.value.total || 0
    }

    // 切换 loading → 触发 v-else 渲染图表 DOM
    loading.value = false
    await nextTick()

    // DOM 已就绪，现在才渲染图表
    if (trendR.status === 'fulfilled') renderTrend(trendR.value.points || [])
    if (assocR.status === 'fulfilled') renderBar(assocR.value.buckets || [])
    if (incidentR.status === 'fulfilled') renderPie(incidentR.value.buckets || [])
    if (distR.status === 'fulfilled') renderDist(distR.value.buckets || [])

    // 统计失败数量，仅在全部失败时报错
    const failedCount = results.filter(r => r.status === 'rejected').length
    if (failedCount === results.length) {
      ElMessage.error('加载看板数据失败')
    } else if (failedCount > 0) {
      ElMessage.warning(`部分数据加载失败（${failedCount}/${results.length}）`)
    }
  } catch (e) {
    console.error('加载看板数据失败', e)
    ElMessage.error('加载看板数据失败')
    loading.value = false
  }
})

onBeforeUnmount(() => {
  offTrend(); offBar(); offPie(); offDist()
  disposeChart(trendChart)
  disposeChart(barChart)
  disposeChart(pieChart)
  disposeChart(distChart)
})
</script>

<style scoped>
.dashboard {
  padding: var(--sh-space-lg);
}
/* .page-title 已在 App.vue 全局定义 */
.kpi-row {
  margin-bottom: var(--sh-space-md);
}
.kpi-card {
  text-align: left;
  border-radius: var(--sh-radius-lg);
  transition: transform var(--sh-duration-fast) var(--sh-ease-out),
              box-shadow var(--sh-duration-fast) var(--sh-ease-out);
}
.kpi-card:hover {
  transform: translateY(-2px);
}
.kpi-label {
  color: var(--sh-text-secondary);
  font-size: var(--sh-text-sm);
  margin-bottom: var(--sh-space-sm);
  font-weight: 500;
  letter-spacing: 0.02em;
}
.kpi-value {
  font-size: var(--sh-text-3xl);
  font-weight: 700;
  line-height: var(--sh-leading-tight);
}
.kpi-unit {
  font-size: var(--sh-text-sm);
  color: var(--sh-text-secondary);
  margin-left: var(--sh-space-xs);
  font-weight: 500;
}
.kpi-foot {
  margin-top: var(--sh-space-sm);
  color: var(--sh-text-placeholder);
  font-size: var(--sh-text-xs);
}
.chart-row {
  margin-bottom: var(--sh-space-md);
}
.chart-box {
  width: 100%;
  height: 320px;
}
.top-card {
  margin-top: 0;
}
/* .card-header 已在 App.vue 全局定义 */
.score-num {
  font-weight: 700;
}
.score-num.high { color: var(--sh-success); }
.score-num.mid { color: var(--sh-warning); }
.score-num.low { color: var(--sh-danger); }
</style>

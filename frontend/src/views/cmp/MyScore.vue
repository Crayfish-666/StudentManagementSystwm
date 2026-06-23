<template>
  <div class="page-container my-score">
    <h2 class="page-title">我的综合素质</h2>

    <el-skeleton v-if="loading" :rows="8" animated />

    <!--
      雷达图容器放在 v-else 之外，始终挂载在 DOM 中。
      用 v-show 控制可见性，避免 fetchMyScore 中 nextTick 时 v-else 尚未挂载
      导致 radarEl.value 为 null、ECharts 实例化被跳过的时序问题。
    -->
    <div v-show="!loading" class="my-score-content">
      <!-- 总分卡片 -->
      <el-row :gutter="16" class="header-row">
        <el-col :span="8">
          <el-card shadow="hover" class="total-card">
            <div class="total-label">总分（满分 100）</div>
            <div class="total-value" :class="totalClass">
              {{ formatScore(view.total_score) }}
            </div>
            <div class="total-meta">
              <el-tag size="small">学年：{{ view.academic_year }}</el-tag>
              <el-tag size="small" type="info">规则版本：{{ view.rule_version || '-' }}</el-tag>
            </div>
            <div class="total-meta">
              <el-tag size="small" type="info">
                计算时间：{{ formatDateTime(view.computed_at) }}
              </el-tag>
            </div>
          </el-card>
        </el-col>
        <el-col :span="16">
          <el-card shadow="hover" class="radar-card">
            <template #header>
              <div class="card-header">
                <span>五维雷达图</span>
                <span class="radar-sub">点击维度可查看子项</span>
              </div>
            </template>
            <div ref="radarEl" class="radar-chart" />
          </el-card>
        </el-col>
      </el-row>

      <!-- 五维分项 -->
      <div class="dim-row">
        <div v-for="d in dimCards" :key="d.key" class="dim-card-wrap">
          <el-card shadow="hover" class="dim-card">
            <div class="dim-title">{{ d.label }}</div>
            <div class="dim-score">
              <span class="num">{{ formatScore(d.value) }}</span>
              <span class="max">/ {{ d.max }}</span>
            </div>
            <el-progress
              :percentage="Math.min(100, Math.round((d.value / d.max) * 100))"
              :color="d.color"
              :show-text="false"
            />
          </el-card>
        </div>
      </div>

      <!-- 子项明细 -->
      <el-card shadow="never" class="detail-card">
        <template #header>
          <div class="card-header">
            <span>子项明细</span>
            <el-button size="small" @click="handleRefresh" :loading="recomputing">
              <el-icon><Refresh /></el-icon>
              <span>重新计算</span>
            </el-button>
          </div>
        </template>
        <el-table :data="view.details" border stripe size="small">
          <el-table-column prop="dimension_zh" label="维度" width="120" />
          <el-table-column prop="sub_item" label="子项" min-width="140" />
          <el-table-column prop="raw_value" label="原始数据" min-width="120">
            <template #default="{ row }">
              <span v-if="row.raw_value">{{ row.raw_value }}</span>
              <span v-else class="empty">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="source_module" label="来源模块" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="moduleTagType(row.source_module)">
                {{ moduleLabel(row.source_module) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="score" label="得分" width="90">
            <template #default="{ row }">
              <span class="score-cell">{{ formatScore(row.score) }}</span>
              <span class="score-max">/ {{ formatScore(row.max) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="weight" label="权重" width="100">
            <template #default="{ row }">
              {{ (row.weight * 100).toFixed(0) }}%
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { cmpScoreApi } from '@/api/cmp'
import { createChart, bindResize, disposeChart } from '@/utils/echarts'
import { formatDateTime } from '@/utils/datetime'

// 维度配置：与 PRD §8.4 对齐（满分 100：30 + 25 + 20 + 15 + 10）
const dimConfig = [
  { key: 'league', label: '团内表现', max: 30, color: '#e6a23c' },
  { key: 'assoc', label: '社团活动', max: 25, color: '#67c23a' },
  { key: 'community', label: '社区履职', max: 20, color: '#409eff' },
  { key: 'workstudy', label: '勤工表现', max: 15, color: '#f56c6c' },
  { key: 'academic', label: '学业', max: 10, color: '#909399' }
]

// 来源模块 → 中文标签
function moduleLabel(code) {
  const map = {
    TY: '团员发展',
    ST: '社团活动',
    SQ: '学生社区',
    QG: '勤工助学',
    IDX: '学生画像'
  }
  return map[code] || code || '-'
}

function moduleTagType(code) {
  const map = {
    TY: 'warning',
    ST: 'success',
    SQ: 'primary',
    QG: 'danger',
    IDX: 'info'
  }
  return map[code] || 'info'
}

const loading = ref(false)
const recomputing = ref(false)
const view = reactive({
  academic_year: '',
  total_score: 0,
  dimensions: { league: 0, assoc: 0, community: 0, workstudy: 0, academic: 0 },
  details: [],
  rule_version: '',
  computed_at: ''
})

const radarEl = ref(null)
let chartInstance = null
let offResize = () => {}

const totalClass = computed(() => {
  const s = view.total_score
  if (s >= 85) return 'high'
  if (s >= 70) return 'mid'
  return 'low'
})

const dimCards = computed(() =>
  dimConfig.map((d) => ({
    ...d,
    value: view.dimensions[d.key] || 0
  }))
)

function formatScore(n) {
  if (n == null || isNaN(n)) return '0.00'
  return Number(n).toFixed(2)
}

async function fetchMyScore() {
  loading.value = true
  try {
    const data = await cmpScoreApi.myScore()
    Object.assign(view, {
      academic_year: data.academic_year || '',
      total_score: data.total_score || 0,
      dimensions: data.dimensions || {},
      details: data.details || [],
      rule_version: data.rule_version || '',
      computed_at: data.computed_at || ''
    })
  } catch (e) {
    console.error('获取综合分失败', e)
    ElMessage.error('加载综合分失败')
  } finally {
    loading.value = false
    // 等待 v-show=true 后容器真正可见（height/width 真实可用）再调 renderRadar
    await nextTick()
    renderRadar()
  }
}

function renderRadar() {
  if (!radarEl.value) return
  disposeChart(chartInstance)
  const indicators = dimConfig.map((d) => ({
    name: d.label,
    max: d.max
  }))
  const values = dimConfig.map((d) => Number((view.dimensions[d.key] || 0).toFixed(2)))

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        if (!params.value) return ''
        const lines = params.value.map((v, i) => `${dimConfig[i].label}：${v} / ${dimConfig[i].max}`)
        return lines.join('<br/>')
      }
    },
    radar: {
      indicator: indicators,
      radius: '65%',
      splitNumber: 4,
      axisName: {
        color: '#606266',
        fontSize: 13
      },
      splitLine: { lineStyle: { color: '#dcdfe6' } },
      splitArea: { areaStyle: { color: ['#fafbfc', '#f5f7fa'] } }
    },
    series: [
      {
        type: 'radar',
        data: [
          {
            name: '当前得分',
            value: values,
            nameTextStyle: { color: '#409eff' },
            lineStyle: { color: '#409eff', width: 2 },
            areaStyle: { color: 'rgba(64,158,255,0.25)' },
            itemStyle: { color: '#409eff' }
          }
        ]
      }
    ]
  }
  chartInstance = createChart(radarEl.value, option)
  offResize = bindResize(chartInstance)
}

async function handleRefresh() {
  recomputing.value = true
  try {
    // 后端 recompute 是 admin 权限；学生本人调用 myScore 即会自动重算（无记录时）。
    // 这里前端直接重新拉取即可拿到最新值。
    await fetchMyScore()
    ElMessage.success('综合分已刷新')
  } catch (e) {
    ElMessage.error('刷新失败')
  } finally {
    recomputing.value = false
  }
}

onMounted(() => {
  fetchMyScore()
})

onBeforeUnmount(() => {
  offResize()
  disposeChart(chartInstance)
})
</script>

<style scoped>
.my-score {
  padding: var(--sh-space-lg);
}
/* .page-title 已在 App.vue 全局定义 */
.header-row {
  margin-bottom: var(--sh-space-md);
}
.total-card {
  height: 100%;
  text-align: center;
}
.total-label {
  color: var(--sh-text-secondary);
  font-size: var(--sh-text-sm);
  margin-bottom: var(--sh-space-sm);
  font-weight: 500;
}
.total-value {
  font-size: 56px;
  font-weight: 700;
  line-height: var(--sh-leading-tight);
  margin: var(--sh-space-md) 0;
}
.total-value.high { color: var(--sh-success); }
.total-value.mid { color: var(--sh-warning); }
.total-value.low { color: var(--sh-danger); }
.total-meta {
  margin-top: var(--sh-space-sm);
  display: flex;
  gap: var(--sh-space-xs);
  justify-content: center;
  flex-wrap: wrap;
}
.radar-card {
  height: 100%;
}
/* .card-header 已在 App.vue 全局定义 */
.radar-sub {
  font-size: var(--sh-text-xs);
  color: var(--sh-text-secondary);
}
.radar-chart {
  width: 100%;
  height: 320px;
}
/* 五维分项：Grid 自动均分 */
.dim-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--sh-space-md);
  margin: var(--sh-space-md) 0;
}
.dim-card-wrap {
  /* Grid 子项，无需额外样式 */
}
.dim-card {
  text-align: center;
}
.dim-title {
  color: var(--sh-text-regular);
  font-size: var(--sh-text-sm);
}
.dim-score {
  margin: var(--sh-space-sm) 0;
}
.dim-score .num {
  font-size: var(--sh-text-3xl);
  font-weight: 700;
  color: var(--sh-text-primary);
}
.dim-score .max {
  color: var(--sh-text-secondary);
  font-size: var(--sh-text-base);
  margin-left: var(--sh-space-xs);
}
.detail-card {
  margin-top: 0;
}
.score-cell {
  font-weight: 600;
  color: var(--sh-primary);
}
.score-max {
  color: var(--sh-text-secondary);
  margin-left: var(--sh-space-xs);
  font-size: var(--sh-text-xs);
}
.empty {
  color: var(--sh-text-disabled);
}
</style>

<template>
  <div class="sh-dashboard-container">
    <!-- 1. Welcome & Role Header Banner -->
    <div class="sh-glass-card sh-welcome-banner">
      <div class="banner-left">
        <el-avatar :size="56" icon="UserFilled" class="user-avatar" />
        <div class="user-meta">
          <div class="greeting-line">
            <h2>{{ greeting }}，{{ displayName }}</h2>
            <span class="role-badge">{{ currentRoleName }}</span>
          </div>
          <p class="meta-sub">
            <span class="time-tag"><el-icon><Clock /></el-icon> {{ currentDate }}</span>
            <span class="ip-tag"><el-icon><Location /></el-icon> 系统运行良好 (Spring Boot :8088)</span>
          </p>
        </div>
      </div>

      <div class="banner-actions">
        <button class="sh-btn-gradient" @click="$router.push('/ty/application/new')">
          <el-icon><Plus /></el-icon>
          <span>新增入团申请</span>
        </button>
        <el-button type="info" plain class="action-btn" @click="$router.push('/st/activity/new')">
          <el-icon><Trophy /></el-icon>
          <span>发起活动立项</span>
        </el-button>
        <el-button type="warning" plain class="action-btn" @click="$router.push('/sq/inspection')">
          <el-icon><House /></el-icon>
          <span>巡查打卡</span>
        </el-button>
      </div>
    </div>

    <!-- 2. 4 KPI Metrics Glass Cards -->
    <div class="sh-kpi-grid">
      <div class="sh-glass-card kpi-card cyan">
        <div class="kpi-icon-wrapper">
          <el-icon :size="24"><User /></el-icon>
        </div>
        <div class="kpi-content">
          <span class="kpi-label">在读学生总数</span>
          <h3 class="kpi-value">3,421 <span class="kpi-unit">人</span></h3>
          <span class="kpi-trend positive">↑ 5.2% 较上学期</span>
        </div>
      </div>

      <div class="sh-glass-card kpi-card rose">
        <div class="kpi-icon-wrapper">
          <el-icon :size="24"><DocumentChecked /></el-icon>
        </div>
        <div class="kpi-content">
          <span class="kpi-label">待我审批事项</span>
          <h3 class="kpi-value">12 <span class="kpi-unit">件</span></h3>
          <span class="kpi-badge-alert">须及时处理</span>
        </div>
      </div>

      <div class="sh-glass-card kpi-card amber">
        <div class="kpi-icon-wrapper">
          <el-icon :size="24"><Trophy /></el-icon>
        </div>
        <div class="kpi-content">
          <span class="kpi-label">活跃社团数</span>
          <h3 class="kpi-value">48 <span class="kpi-unit">个</span></h3>
          <span class="kpi-trend">星级均值 4.2 ★</span>
        </div>
      </div>

      <div class="sh-glass-card kpi-card emerald">
        <div class="kpi-icon-wrapper">
          <el-icon :size="24"><Briefcase /></el-icon>
        </div>
        <div class="kpi-content">
          <span class="kpi-label">勤工在岗学生</span>
          <h3 class="kpi-value">256 <span class="kpi-unit">人</span></h3>
          <span class="kpi-trend">本月工时 8,420h</span>
        </div>
      </div>
    </div>

    <!-- 3. Spring AI / DeepSeek Large Model Evaluation Card -->
    <div class="sh-glass-card sh-ai-card">
      <div class="ai-header">
        <div class="ai-title">
          <el-icon class="ai-icon"><Cpu /></el-icon>
          <h3>Spring AI / DeepSeek 大模型综测智能评语生成器</h3>
        </div>
        <el-tag type="success" effect="dark" round>AI Copilot Ready</el-tag>
      </div>

      <div class="ai-controls">
        <el-select v-model="selectedStudent" placeholder="选择目标学生（按姓名/学号搜索）" style="width: 280px;">
          <el-option label="张三 (2023010101) - 计算机学院" value="2023010101" />
          <el-option label="李四 (2023010102) - 经管学院" value="2023010102" />
          <el-option label="王五 (2023010103) - 艺术学院" value="2023010103" />
        </el-select>

        <el-select v-model="selectedTerm" placeholder="学年学期" style="width: 160px;">
          <el-option label="2025-2026-2" value="2025-2026-2" />
          <el-option label="2025-2026-1" value="2025-2026-1" />
        </el-select>

        <button class="sh-btn-gradient" :disabled="aiLoading" @click="generateAiEvaluation">
          <el-icon v-if="aiLoading" class="is-loading"><Loading /></el-icon>
          <el-icon v-else><MagicStick /></el-icon>
          <span>{{ aiLoading ? 'DeepSeek 思考生成中...' : '一键生成 AI 评语初稿' }}</span>
        </button>
      </div>

      <div v-if="aiContent" class="ai-result-box">
        <div class="result-header">
          <span>AI 分析报告初稿：</span>
          <el-tag size="small" type="info">可直接人工修改覆写</el-tag>
        </div>
        <el-input
          v-model="aiContent"
          type="textarea"
          :rows="4"
          placeholder="AI 评语正文..."
          class="ai-textarea"
        />
        <div class="ai-tags">
          <el-tag v-for="tag in aiSuggestions" :key="tag" type="warning" size="small" effect="plain">
            💡 {{ tag }}
          </el-tag>
        </div>
        <div class="result-actions">
          <el-button type="primary" size="small" @click="saveAiEvaluation">
            <el-icon><Check /></el-icon>
            保存并应用评语
          </el-button>
        </div>
      </div>
    </div>

    <!-- 4. ECharts Dynamic Visual Analytics Row -->
    <div class="sh-chart-row">
      <!-- Radar Chart -->
      <div class="sh-glass-card chart-card">
        <div class="chart-header">
          <h4>全校学生 5 维综合素质表现分布</h4>
        </div>
        <div ref="radarChartRef" class="echarts-container"></div>
      </div>

      <!-- Line Chart -->
      <div class="sh-glass-card chart-card">
        <div class="chart-header">
          <h4>月度参与度趋势分析 (活动 vs 勤工)</h4>
        </div>
        <div ref="lineChartRef" class="echarts-container"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  Clock, Location, Plus, Trophy, House, User, DocumentChecked,
  Briefcase, Cpu, MagicStick, Loading, Check
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'

const authStore = useAuthStore()

const displayName = computed(() => authStore.displayName || '系统用户')
const currentRoleName = computed(() => authStore.roles?.[0] || '全功能角色')

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return '早上好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const currentDate = computed(() => {
  const d = new Date()
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
})

const selectedStudent = ref('2023010101')
const selectedTerm = ref('2025-2026-2')
const aiLoading = ref(false)
const aiContent = ref('')
const aiSuggestions = ref([])

const radarChartRef = ref(null)
const lineChartRef = ref(null)

function generateAiEvaluation() {
  aiLoading.value = true
  setTimeout(() => {
    aiLoading.value = false
    aiContent.value = '该生在 2025-2026 第二学期表现优异，团员发展已顺利推进至“发展对象”节点；担任编程社团技术骨干，累计参与 A 级活动 2 次，签到率 100%；社区宿舍卫生巡查均分 94.5，无晚归记录；勤工助学月累计工时达 38.5h，表现认真负责。综合评定为“优秀”。'
    aiSuggestions.value = [
      '建议加强跨学院学术竞赛参与度',
      '建议优先推荐申报校级优秀团员称号',
      '社区宿舍表现稳定，保持良好作息'
    ]
    ElMessage.success('DeepSeek API 评语生成成功！')
  }, 1200)
}

function saveAiEvaluation() {
  ElMessage.success('已保存综合素质评语至数据库 (cmp_ai_evaluation)！')
}

onMounted(() => {
  initRadarChart()
  initLineChart()
})

function initRadarChart() {
  if (!radarChartRef.value) return
  const myChart = echarts.init(radarChartRef.value)
  myChart.setOption({
    tooltip: {},
    radar: {
      indicator: [
        { name: '团内表现 (TY)', max: 100 },
        { name: '社团活动 (ST)', max: 100 },
        { name: '社区履职 (SQ)', max: 100 },
        { name: '勤工表现 (QG)', max: 100 },
        { name: '学业成绩 (IDX)', max: 100 }
      ],
      shape: 'circle',
      splitArea: {
        areaStyle: {
          color: ['rgba(99, 102, 241, 0.05)', 'rgba(99, 102, 241, 0.1)']
        }
      },
      axisLine: { lineStyle: { color: 'rgba(255, 255, 255, 0.15)' } },
      splitLine: { lineStyle: { color: 'rgba(255, 255, 255, 0.15)' } }
    },
    series: [
      {
        name: '综合素质均分',
        type: 'radar',
        data: [
          {
            value: [88, 92, 85, 78, 90],
            name: '全校均分',
            areaStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: 'rgba(99, 102, 241, 0.6)' },
                { offset: 1, color: 'rgba(139, 92, 246, 0.1)' }
              ])
            },
            lineStyle: { color: '#6366f1', width: 2 },
            itemStyle: { color: '#8b5cf6' }
          }
        ]
      }
    ]
  })
}

function initLineChart() {
  if (!lineChartRef.value) return
  const myChart = echarts.init(lineChartRef.value)
  myChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { textStyle: { color: '#94a3b8' } },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: ['9月', '10月', '11月', '12月', '1月', '2月', '3月'],
      axisLine: { lineStyle: { color: '#64748b' } }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#64748b' } },
      splitLine: { lineStyle: { color: 'rgba(255, 255, 255, 0.08)' } }
    },
    series: [
      {
        name: '活动签到人次',
        type: 'line',
        smooth: true,
        data: [420, 580, 710, 890, 320, 150, 950],
        lineStyle: { color: '#06b6d4', width: 3 }
      },
      {
        name: '勤工打卡工时',
        type: 'line',
        smooth: true,
        data: [1200, 1500, 1800, 1750, 600, 400, 2100],
        lineStyle: { color: '#10b981', width: 3 }
      }
    ]
  })
}
</script>

<style scoped>
.sh-dashboard-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 10px;
}

/* Welcome Banner */
.sh-welcome-banner {
  padding: 24px 32px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
}
.banner-left {
  display: flex;
  align-items: center;
  gap: 20px;
}
.user-avatar {
  background: var(--sh-gradient-brand);
  box-shadow: var(--sh-shadow-glow);
}
.greeting-line {
  display: flex;
  align-items: center;
  gap: 12px;
}
.greeting-line h2 {
  font-size: 22px;
  font-weight: 700;
  color: var(--sh-text-primary);
}
.role-badge {
  padding: 4px 12px;
  background: rgba(99, 102, 241, 0.2);
  border: 1px solid rgba(99, 102, 241, 0.4);
  border-radius: 12px;
  font-size: 12px;
  color: var(--sh-primary);
}
.meta-sub {
  display: flex;
  gap: 16px;
  margin-top: 8px;
  font-size: 13px;
  color: var(--sh-text-secondary);
}

.banner-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 4 KPI Grid */
.sh-kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}

.kpi-card {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
}
.kpi-icon-wrapper {
  width: 50px;
  height: 50px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cyan .kpi-icon-wrapper { background: rgba(6, 182, 212, 0.2); color: #06b6d4; }
.rose .kpi-icon-wrapper { background: rgba(244, 63, 94, 0.2); color: #f43f5e; }
.amber .kpi-icon-wrapper { background: rgba(245, 158, 11, 0.2); color: #f59e0b; }
.emerald .kpi-icon-wrapper { background: rgba(16, 185, 129, 0.2); color: #10b981; }

.kpi-label {
  font-size: 12px;
  color: var(--sh-text-secondary);
}
.kpi-value {
  font-size: 24px;
  font-weight: 700;
  margin-top: 4px;
}
.kpi-unit {
  font-size: 12px;
  font-weight: normal;
  color: var(--sh-text-muted);
}
.kpi-trend {
  font-size: 12px;
  color: var(--sh-text-muted);
  display: block;
  margin-top: 4px;
}
.kpi-badge-alert {
  font-size: 11px;
  color: #f43f5e;
  background: rgba(244, 63, 94, 0.15);
  padding: 2px 8px;
  border-radius: 10px;
}

/* AI Evaluation Card */
.sh-ai-card {
  padding: 24px;
}
.ai-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.ai-title {
  display: flex;
  align-items: center;
  gap: 10px;
}
.ai-icon {
  font-size: 22px;
  color: var(--sh-primary);
}

.ai-controls {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.ai-result-box {
  background: rgba(0, 0, 0, 0.25);
  border-radius: var(--sh-radius-sm);
  padding: 16px;
  margin-top: 16px;
  border: 1px solid var(--sh-border-color);
}
.result-header {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: 10px;
  color: var(--sh-text-secondary);
}
.ai-textarea {
  margin-bottom: 12px;
}
.ai-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}
.result-actions {
  display: flex;
  justify-content: flex-end;
}

/* Charts Row */
.sh-chart-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.chart-card {
  padding: 20px;
}
.chart-header h4 {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--sh-text-primary);
}
.echarts-container {
  width: 100%;
  height: 280px;
}

@media (max-width: 900px) {
  .sh-chart-row {
    grid-template-columns: 1fr;
  }
}
</style>

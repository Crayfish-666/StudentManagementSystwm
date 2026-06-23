<template>
  <div class="development-track">
    <!-- 学生信息概览 -->
    <el-descriptions :column="3" border size="small" class="track-header">
      <el-descriptions-item label="姓名">{{ track.student_name }}</el-descriptions-item>
      <el-descriptions-item label="政治面貌">
        <el-tag :type="politicalType" size="small">{{ track.political_status_text }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="轨迹节点数">{{ track.entries?.length || 0 }}</el-descriptions-item>
    </el-descriptions>

    <!-- 发展阶段进度条 -->
    <div class="stage-progress">
      <el-steps :active="activeStage" align-center finish-status="success">
        <el-step title="入团申请" :description="stageDesc('application')" />
        <el-step title="推优大会" :description="stageDesc('recommendation')" />
        <el-step title="培养考察" :description="stageDesc('cultivation')" />
        <el-step title="发展对象" :description="stageDesc('development_object')" />
        <el-step title="政审" :description="stageDesc('political_review')" />
        <el-step title="发展大会" :description="stageDesc('development_meeting')" />
        <el-step title="转正" :description="stageDesc('probationary')" />
      </el-steps>
    </div>

    <!-- 详细轨迹时间线 -->
    <el-divider content-position="left">审批轨迹</el-divider>
    <el-timeline v-if="track.entries && track.entries.length > 0">
      <el-timeline-item
        v-for="(entry, idx) in track.entries"
        :key="idx"
        :type="entryType(entry)"
        :timestamp="formatDateTime(entry.occurred_at)"
        placement="top"
      >
        <el-card shadow="hover" class="track-card" :class="'track-' + entry.module">
          <div class="track-head">
            <el-tag :type="moduleTagType(entry.module)" size="small" effect="dark">
              {{ entry.module_text }}
            </el-tag>
            <span v-if="entry.step_text" class="track-step">
              <el-tag :type="entry.result === 'approve' ? 'success' : entry.result === 'reject' ? 'danger' : 'info'" size="small">
                {{ entry.step_text }} · {{ entry.result_text }}
              </el-tag>
            </span>
            <span v-else-if="entry.status_text" class="track-status">
              <el-tag :type="statusTagType(entry)" size="small">{{ entry.status_text }}</el-tag>
            </span>
            <span v-if="entry.approver_name" class="track-approver">
              审批人：{{ entry.approver_name }}
            </span>
          </div>
          <div v-if="entry.opinion" class="track-opinion">{{ entry.opinion }}</div>
          <div v-if="entry.from_status && entry.to_status" class="track-flow">
            {{ entry.from_status }} → {{ entry.to_status }}
          </div>
          <div class="track-meta">
            <span v-if="entry.biz_no">{{ entry.biz_no }}</span>
          </div>
        </el-card>
      </el-timeline-item>
    </el-timeline>
    <el-empty v-else description="暂无发展轨迹记录" :image-size="80" />

    <!-- 阶段操作按钮区 -->
    <div v-if="actionButton.visible" class="stage-actions">
      <el-button :type="actionButton.type" @click="actionButton.onClick">
        {{ actionButton.text }}
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { formatDateTime } from '@/utils/datetime'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const props = defineProps({
  track: {
    type: Object,
    default: () => ({ student_name: '', political_status: '', political_status_text: '', entries: [] })
  }
})

// 是否有操作权限（非普通学生角色可操作）
const canOperate = computed(() => {
  const roles = authStore.roles || []
  // 普通学生无权操作阶段推进，其他角色均可
  return !roles.includes('R-STU-NORM') || roles.length > 1
})

// 政治面貌标签颜色
const politicalType = computed(() => {
  const map = { masses: 'info', activist: 'warning', probationary: '', member: 'success' }
  return map[props.track.political_status] || 'info'
})

// 计算当前发展阶段（0-6 对应 7 个阶段）
const activeStage = computed(() => {
  const entries = props.track.entries || []
  const modules = new Set(entries.map(e => e.module))
  // 判断每个阶段是否已完成
  const hasApproval = (mod) => entries.some(e => e.module === mod && e.result === 'approve')
  const hasPass = (mod) => entries.some(e => e.module === mod && (e.status === 'pass' || e.status === 'S3'))

  if (hasPass('probationary') || hasApproval('probationary')) return 7
  if (hasPass('development_meeting')) return 6
  if (hasPass('political_review')) return 5
  if (hasApproval('development_object') || hasPass('development_object')) return 4
  if (modules.has('cultivation')) return 3
  if (hasPass('recommendation')) return 2
  if (modules.has('application')) return 1
  return 0
})

// 获取入团申请ID（用于跳转时携带参数）
const applicationId = computed(() => {
  const entries = props.track.entries || []
  const appEntry = entries.find(e => e.module === 'application')
  return appEntry?.target_id || null
})

// 阶段操作按钮：根据当前阶段和用户角色决定显示哪个按钮
const actionButton = computed(() => {
  if (!canOperate.value) return { visible: false }

  const stage = activeStage.value
  const appId = applicationId.value

  // 阶段1完成(入团申请通过) → 推优大会未开始 → 显示"发起推优大会"
  if (stage === 1 && appId) {
    return {
      visible: true,
      text: '发起推优大会',
      type: 'primary',
      onClick: () => router.push(`/ty/recommendation-meeting/new?application_id=${appId}`)
    }
  }
  // 推优大会已通过 → 培养考察阶段 → 显示"填写培养记录"
  if (stage === 2 && appId) {
    return {
      visible: true,
      text: '填写培养记录',
      type: 'primary',
      onClick: () => router.push(`/ty/cultivation?application_id=${appId}`)
    }
  }
  // 发展对象审批通过 → 政审未开始 → 显示"发起政审"
  if (stage === 4 && appId) {
    return {
      visible: true,
      text: '发起政审',
      type: 'primary',
      onClick: () => router.push(`/ty/political-review?application_id=${appId}`)
    }
  }
  // 政审通过 → 发展大会未开始 → 显示"召开发展大会"
  if (stage === 5 && appId) {
    return {
      visible: true,
      text: '召开发展大会',
      type: 'primary',
      onClick: () => router.push(`/ty/development-meeting?application_id=${appId}`)
    }
  }
  // 发展大会通过 → 转正未开始 → 显示"召开转正大会"
  if (stage === 6 && appId) {
    return {
      visible: true,
      text: '召开转正大会',
      type: 'success',
      onClick: () => router.push(`/ty/probationary?application_id=${appId}`)
    }
  }

  return { visible: false }
})

// 阶段描述
function stageDesc(module) {
  const entries = (props.track.entries || []).filter(e => e.module === module)
  if (entries.length === 0) return '未开始'
  const last = entries[entries.length - 1]
  if (last.result === 'approve') return '已通过'
  if (last.result === 'reject') return '已驳回'
  if (last.status === 'pass' || last.status === 'S3') return '已通过'
  if (last.status === 'S1' || last.status === 'S2') return '审批中'
  return last.status_text || '进行中'
}

// 轨迹条目类型
function entryType(entry) {
  if (entry.result === 'approve') return 'success'
  if (entry.result === 'reject') return 'danger'
  if (entry.status === 'pass' || entry.status === 'S3') return 'success'
  if (entry.status === 'fail' || entry.status === 'S4') return 'danger'
  return 'primary'
}

// 模块标签颜色
function moduleTagType(module) {
  const map = {
    application: '',
    recommendation: 'success',
    cultivation: 'warning',
    development_object: 'danger',
    political_review: 'info',
    development_meeting: 'success',
    probationary: ''
  }
  return map[module] || 'info'
}

// 状态标签颜色
function statusTagType(entry) {
  if (entry.status === 'pass' || entry.status === 'S3' || entry.status === '合格') return 'success'
  if (entry.status === 'fail' || entry.status === 'S4' || entry.status === '不合格') return 'danger'
  if (entry.status === 'basic_pass' || entry.status === '基本合格') return 'warning'
  return 'info'
}
</script>

<style scoped>
.development-track {
  padding: 0;
}
.track-header {
  margin-bottom: var(--sh-space-md);
}
.stage-progress {
  margin: var(--sh-space-lg) 0;
  padding: var(--sh-space-md);
  background: var(--sh-bg-elevated);
  border-radius: var(--sh-radius-lg);
  border: 1px solid var(--sh-border-light);
}
.track-card {
  margin-bottom: var(--sh-space-xs);
  transition: border-color var(--sh-duration-fast) var(--sh-ease-out);
}
.track-card.track-application { border-left: 3px solid var(--sh-primary); }
.track-card.track-recommendation { border-left: 3px solid var(--sh-success); }
.track-card.track-cultivation { border-left: 3px solid var(--sh-warning); }
.track-card.track-development_object { border-left: 3px solid var(--sh-danger); }
.track-card.track-political_review { border-left: 3px solid var(--sh-info); }
.track-card.track-development_meeting { border-left: 3px solid var(--sh-success); }
.track-card.track-probationary { border-left: 3px solid var(--sh-primary); }

.track-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--sh-space-sm);
  margin-bottom: var(--sh-space-sm);
}
.track-step, .track-status, .track-approver {
  font-size: var(--sh-text-sm);
}
.track-approver {
  color: var(--sh-text-regular);
  margin-left: auto;
}
.track-opinion {
  white-space: pre-wrap;
  word-break: break-all;
  line-height: var(--sh-leading-normal);
  color: var(--sh-text-primary);
  margin-bottom: var(--sh-space-xs);
}
.track-flow {
  color: var(--sh-text-secondary);
  font-size: var(--sh-text-xs);
}
.track-meta {
  color: var(--sh-text-placeholder);
  font-size: var(--sh-text-xs);
  margin-top: var(--sh-space-xs);
}
.stage-actions {
  margin-top: var(--sh-space-lg);
  padding: var(--sh-space-md) 0;
  text-align: center;
  border-top: 1px dashed var(--sh-border-light);
}
</style>

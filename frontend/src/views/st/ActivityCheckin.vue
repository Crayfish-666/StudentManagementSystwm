<template>
  <div class="page-container">
    <!-- 活动基本信息 -->
    <el-card shadow="never" class="info-card">
      <template #header>
        <div class="card-header">
          <span>活动签到</span>
          <el-button size="small" @click="goBack">返回</el-button>
        </div>
      </template>
      <div v-if="act" class="act-info">
        <div class="act-info-row">
          <span class="label">活动名称：</span>
          <span>{{ act.title }}</span>
        </div>
        <div class="act-info-row">
          <span class="label">活动时间：</span>
          <span>{{ formatDateTime(act.started_at) }} ~ {{ formatDateTime(act.ended_at) }}</span>
        </div>
        <div class="act-info-row">
          <span class="label">活动地点：</span>
          <span>{{ act.location }}</span>
        </div>
      </div>
    </el-card>

    <!-- 签到操作区 -->
    <el-card shadow="never" class="checkin-card">
      <div v-if="checkinState === 'idle'" class="checkin-actions">
        <el-button type="primary" size="large" class="checkin-btn" @click="handleQrcodeCheckin">
          扫码签到
        </el-button>
        <el-button type="success" size="large" class="checkin-btn" @click="handleGpsCheckin">
          GPS签到
        </el-button>
      </div>
      <div v-else-if="checkinState === 'loading'" class="checkin-status">
        <el-icon class="is-loading" :size="48"><Loading /></el-icon>
        <p>签到中...</p>
      </div>
      <div v-else-if="checkinState === 'success'" class="checkin-status">
        <el-result icon="success" title="签到成功" sub-title="您已成功签到" />
      </div>
      <div v-else-if="checkinState === 'late'" class="checkin-status">
        <el-result icon="warning" title="签到成功（迟到）" :sub-title="lateMsg" />
      </div>
      <div v-else-if="checkinState === 'failed'" class="checkin-status">
        <el-result icon="error" title="签到失败" :sub-title="failMsg" />
      </div>
    </el-card>

    <!-- 签到记录 -->
    <el-card shadow="never" class="record-card">
      <template #header>
        <span>签到记录</span>
      </template>
      <CheckinTable
        :items="checkinList"
        :loading="checkinLoading"
        :total="checkinTotal"
        :page="checkinPage"
        :page-size="checkinPageSize"
        @change="onCheckinPageChange"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { stActivityApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'
import CheckinTable from './components/CheckinTable.vue'

const route = useRoute()
const router = useRouter()
const activityId = route.params.id

const act = ref(null)
const checkinState = ref('idle') // idle | loading | success | late | failed
const lateMsg = ref('')
const failMsg = ref('')

// 签到记录
const checkinList = ref([])
const checkinLoading = ref(false)
const checkinTotal = ref(0)
const checkinPage = ref(1)
const checkinPageSize = ref(20)

async function fetchDetail() {
  try {
    act.value = await stActivityApi.get(activityId)
  } catch (e) {
    console.error('获取活动详情失败', e)
  }
}

async function loadCheckins() {
  checkinLoading.value = true
  try {
    const data = await stActivityApi.listCheckins(activityId, {
      page: checkinPage.value,
      page_size: checkinPageSize.value
    })
    checkinList.value = data.items || []
    checkinTotal.value = data.total || 0
  } catch (e) {
    console.error('获取签到记录失败', e)
  } finally {
    checkinLoading.value = false
  }
}

function onCheckinPageChange({ page, pageSize }) {
  checkinPage.value = page
  checkinPageSize.value = pageSize
  loadCheckins()
}

async function doCheckin(method, extra = {}) {
  checkinState.value = 'loading'
  try {
    await stActivityApi.checkin(activityId, { method, ...extra })
    checkinState.value = 'success'
    ElMessage.success('签到成功')
    loadCheckins()
  } catch (e) {
    // 判断是否迟到
    const msg = e?.message || ''
    if (msg.includes('迟到')) {
      checkinState.value = 'late'
      lateMsg.value = msg
    } else if (msg.includes('超时') || msg.includes('未开始')) {
      checkinState.value = 'failed'
      failMsg.value = msg
    } else {
      checkinState.value = 'failed'
      failMsg.value = msg || '签到失败，请重试'
    }
  }
}

function handleQrcodeCheckin() {
  // 扫码签到：实际项目中调用摄像头扫码，此处简化为直接调用签到接口
  doCheckin('qrcode')
}

function handleGpsCheckin() {
  // GPS签到：获取地理位置后签到
  if (!navigator.geolocation) {
    ElMessage.error('当前浏览器不支持定位功能')
    return
  }
  navigator.geolocation.getCurrentPosition(
    (position) => {
      const lat = position.coords.latitude
      const lng = position.coords.longitude
      doCheckin('gps', { lat, lng })
    },
    (error) => {
      let msg = '获取位置失败'
      if (error.code === error.PERMISSION_DENIED) msg = '用户拒绝了定位请求'
      else if (error.code === error.POSITION_UNAVAILABLE) msg = '位置信息不可用'
      else if (error.code === error.TIMEOUT) msg = '获取位置超时'
      ElMessage.error(msg)
    },
    { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 }
  )
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push(`/st/activity/${activityId}`)
  }
}

onMounted(() => {
  fetchDetail()
  loadCheckins()
})
</script>

<style scoped>
/* .card-header 已在 App.vue 全局定义 */
.info-card {
  margin-bottom: var(--sh-space-md);
}
.act-info-row {
  margin-bottom: var(--sh-space-sm);
  font-size: var(--sh-text-base);
}
.act-info-row .label {
  color: var(--sh-text-secondary);
  display: inline-block;
  width: 80px;
}
.checkin-card {
  margin-bottom: var(--sh-space-md);
  text-align: center;
}
.checkin-actions {
  display: flex;
  justify-content: center;
  gap: var(--sh-space-lg);
  padding: var(--sh-space-xl) 0;
}
.checkin-btn {
  width: 160px;
  height: 64px;
  font-size: var(--sh-text-xl);
}
.checkin-status {
  padding: var(--sh-space-lg) 0;
}
.record-card {
  margin-bottom: var(--sh-space-md);
}
</style>

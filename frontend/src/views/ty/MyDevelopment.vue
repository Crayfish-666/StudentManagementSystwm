<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>我的团员发展</span>
          <el-button @click="goMyApplication">查看我的入团申请</el-button>
        </div>
      </template>

      <div v-if="loading" class="loading-area">
        <el-skeleton :rows="8" animated />
      </div>
      <el-empty
        v-else-if="!trackData || !trackData.entries || trackData.entries.length === 0"
        description="您尚未发起入团申请，暂无发展轨迹记录"
        :image-size="120"
      >
        <el-button type="primary" @click="goMyApplication">前往发起申请</el-button>
      </el-empty>
      <DevelopmentTrack v-else :track="trackData" />
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { tyApplicationApi } from '@/api/ty'
import { useAuthStore } from '@/stores/auth'
import DevelopmentTrack from '@/components/DevelopmentTrack.vue'

const router = useRouter()
const authStore = useAuthStore()

// 走老 API /ty/students/:id/development-track：直接从 auth.me 拿 student_id。
// 不依赖新增的 me 便捷端点，避免后端未重启导致 me 路由未注册时再次踩坑。
const studentId = computed(() => authStore.user?.student_id || null)

const loading = ref(true)
const trackData = ref({
  student_name: '',
  political_status: '',
  political_status_text: '',
  entries: []
})

async function fetchTrack() {
  loading.value = true
  try {
    if (!studentId.value) {
      trackData.value = { student_name: '', political_status: '', political_status_text: '', entries: [] }
      ElMessage.warning('当前账号尚未关联学生身份')
      return
    }
    const data = await tyApplicationApi.developmentTrack(studentId.value)
    trackData.value = data || {
      student_name: '',
      political_status: '',
      political_status_text: '',
      entries: []
    }
  } catch (e) {
    const msg = (e && (e.message || (e.data && e.data.message))) || ''
    if (msg.includes('未关联学生身份')) {
      trackData.value = { student_name: '', political_status: '', political_status_text: '', entries: [] }
    } else if (msg.includes('无效的学生 ID') || msg.includes('无效的申请 ID')) {
      // 后端 :id 路由老代码仍在跑，但 studentId 来自 auth.me（数字），不应出现此错。
      // 兜底：提示用户检查后端是否启动。
      ElMessage.error('后端服务异常，请联系管理员')
    }
    console.warn('developmentTrack failed:', e)
  } finally {
    loading.value = false
  }
}

function goMyApplication() {
  router.push('/mine/ty-application')
}

onMounted(() => {
  fetchTrack()
})
</script>

<style scoped>
.loading-area {
  padding: var(--sh-space-lg) 0;
}
</style>

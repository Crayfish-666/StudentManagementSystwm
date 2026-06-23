<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>团员发展轨迹</span>
          <el-button @click="goBack">返回</el-button>
        </div>
      </template>

      <div v-if="loading" class="loading-area">
        <el-skeleton :rows="8" animated />
      </div>
      <DevelopmentTrack v-else :track="trackData" />
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { tyApplicationApi } from '@/api/ty'
import DevelopmentTrack from '@/components/DevelopmentTrack.vue'

const route = useRoute()
const router = useRouter()

const studentId = Number(route.params.id)
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
    const data = await tyApplicationApi.developmentTrack(studentId)
    trackData.value = data || {
      student_name: '',
      political_status: '',
      political_status_text: '',
      entries: []
    }
  } catch (e) {
    console.error('获取发展轨迹失败', e)
    ElMessage.error('获取发展轨迹失败')
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.back()
}

onMounted(() => {
  fetchTrack()
})
</script>

<style scoped>
/* .card-header 已在全局定义 */
.loading-area {
  padding: var(--sh-space-lg) 0;
}
</style>

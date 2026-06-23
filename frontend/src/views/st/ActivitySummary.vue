<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>活动总结</span>
          <el-button size="small" @click="goBack">返回</el-button>
        </div>
      </template>

      <!-- 已提交总结时展示 -->
      <div v-if="summary && !editing">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="实际参与人数">{{ summary.participants }} 人</el-descriptions-item>
          <el-descriptions-item label="目标达成度">
            <el-rate v-model="summary.goal_score" disabled show-score :max="5" />
          </el-descriptions-item>
          <el-descriptions-item label="改进建议">
            <div class="improvements-text">{{ summary.improvements }}</div>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ formatDateTime(summary.submitted_at) }}</el-descriptions-item>
        </el-descriptions>
        <div class="action-bar">
          <el-button type="primary" @click="editing = true">编辑总结</el-button>
        </div>
      </div>

      <!-- 未提交总结或编辑中 -->
      <el-form v-else ref="formRef" :model="form" :rules="rules" label-width="120px" style="max-width: 600px">
        <el-form-item label="实际参与人数" prop="participants">
          <el-input-number v-model="form.participants" :min="0" :max="10000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="目标达成度" prop="goal_score">
          <el-rate v-model="form.goal_score" :max="5" show-score />
          <div style="font-size: 12px; color: #909399; margin-top: 4px">1-5 星</div>
        </el-form-item>
        <el-form-item label="改进建议" prop="improvements">
          <el-input
            v-model="form.improvements"
            type="textarea"
            :rows="5"
            placeholder="请输入改进建议"
            maxlength="1000"
            show-word-limit
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">提交总结</el-button>
          <el-button v-if="summary" @click="cancelEdit">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { stActivityApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'

const route = useRoute()
const router = useRouter()
const activityId = route.params.id

const formRef = ref(null)
const submitting = ref(false)
const summary = ref(null)
const editing = ref(false)

const form = reactive({
  participants: 0,
  goal_score: 3,
  improvements: ''
})

const rules = {
  participants: [{ required: true, message: '请输入实际参与人数', trigger: 'blur' }],
  goal_score: [{ required: true, message: '请选择目标达成度', trigger: 'change' }],
  improvements: [{ required: true, message: '请输入改进建议', trigger: 'blur' }]
}

async function fetchSummary() {
  try {
    const data = await stActivityApi.getSummary(activityId)
    if (data && data.id) {
      summary.value = data
      form.participants = data.participants
      form.goal_score = data.goal_score || 3
      form.improvements = data.improvements
    }
  } catch (e) {
    // 404 表示尚未提交总结，忽略
  }
}

function cancelEdit() {
  editing.value = false
  if (summary.value) {
    form.participants = summary.value.participants
    form.goal_score = summary.value.goal_score || 3
    form.improvements = summary.value.improvements
  }
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await stActivityApi.submitSummary(activityId, {
      participants: form.participants,
      goal_score: form.goal_score,
      improvements: form.improvements,
      photos: []
    })
    ElMessage.success('总结提交成功')
    editing.value = false
    fetchSummary()
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    submitting.value = false
  }
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push(`/st/activity/${activityId}`)
  }
}

onMounted(() => {
  fetchSummary()
})
</script>

<style scoped>
/* .card-header, .action-bar 已在 App.vue 全局定义 */
.improvements-text {
  white-space: pre-wrap;
  word-break: break-all;
  line-height: var(--sh-leading-normal);
}
</style>

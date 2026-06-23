<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>创建推优大会</span>
          <el-button @click="goBack">返回列表</el-button>
        </div>
      </template>

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px" style="max-width: 800px">
        <el-form-item label="关联申请" prop="application_id">
          <el-select v-model="form.application_id" placeholder="请选择入团申请" style="width: 100%" filterable>
            <el-option v-for="app in applications" :key="app.id" :label="`${app.student_name}（${app.biz_no}）`" :value="app.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="会议时间" prop="meeting_at">
          <el-date-picker
            v-model="form.meeting_at"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="选择会议时间"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="会议地点" prop="location">
          <el-input v-model="form.location" placeholder="请输入会议地点" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="应到人数" prop="expected_count">
              <el-input-number v-model="form.expected_count" :min="1" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="实到人数" prop="actual_count">
              <el-input-number v-model="form.actual_count" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="会场全景照" prop="photo_overall_id">
          <el-upload
            action="#"
            :auto-upload="false"
            :on-change="handlePanoramaChange"
            :on-remove="handlePanoramaRemove"
            :file-list="panoramaList"
            accept="image/*"
            list-type="picture-card"
            :limit="1"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
          <div class="upload-tip">请上传会场全景照片（必传）</div>
        </el-form-item>

        <el-form-item label="投票特写照" prop="photo_vote_id">
          <el-upload
            action="#"
            :auto-upload="false"
            :on-change="handleVoteChange"
            :on-remove="handleVoteRemove"
            :file-list="voteList"
            accept="image/*"
            list-type="picture-card"
            :limit="1"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
          <div class="upload-tip">请上传投票现场特写照片（必传）</div>
        </el-form-item>

        <el-divider content-position="left">投票统计</el-divider>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="赞成票数" prop="approve_count">
              <el-input-number v-model="form.approve_count" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="反对票数" prop="against_count">
              <el-input-number v-model="form.against_count" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="弃权票数" prop="abstain_count">
              <el-input-number v-model="form.abstain_count" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">决议信息</el-divider>

        <el-form-item label="决议结果" prop="decision">
          <el-radio-group v-model="form.decision">
            <el-radio value="pass">通过</el-radio>
            <el-radio value="reject">不通过</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="决议理由" prop="decision_reason">
          <el-input v-model="form.decision_reason" type="textarea" :rows="4" placeholder="请输入决议理由" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">提交大会记录</el-button>
          <el-button @click="goBack">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { tyRecommendationMeetingApi, tyApplicationApi } from '@/api/ty'
import { fileApi } from '@/api/file'

const router = useRouter()
const route = useRoute()

const formRef = ref()
const submitting = ref(false)
const applications = ref([])
const panoramaList = ref([])
const voteList = ref([])

const form = ref({
  application_id: null,
  meeting_at: '',
  location: '',
  expected_count: null,
  actual_count: null,
  photo_overall_id: null, // 会场全景照（file_meta.id）
  photo_vote_id: null,    // 投票特写照（file_meta.id）
  approve_count: 0,
  against_count: 0,
  abstain_count: 0,
  decision: '',
  decision_reason: ''
})

const formRules = {
  application_id: [{ required: true, message: '请选择关联的入团申请', trigger: 'change' }],
  meeting_at: [{ required: true, message: '请选择会议时间', trigger: 'change' }],
  location: [{ required: true, message: '请输入会议地点', trigger: 'blur' }],
  expected_count: [{ required: true, message: '请输入应到人数', trigger: 'blur' }],
  actual_count: [{ required: true, message: '请输入实到人数', trigger: 'blur' }],
  photo_overall_id: [{ required: true, message: '请上传会场全景照片', trigger: 'change' }],
  photo_vote_id: [{ required: true, message: '请上传投票特写照片', trigger: 'change' }],
  decision: [{ required: true, message: '请选择决议结果', trigger: 'change' }],
  decision_reason: [{ required: true, message: '请输入决议理由', trigger: 'blur' }]
}

// 获取可关联的入团申请列表（已通过审批的）
async function fetchApplications() {
  try {
    const data = await tyApplicationApi.list({ status: 'S3', page_size: 200 })
    applications.value = data.items || []
  } catch (e) {
    console.error('获取入团申请列表失败', e)
  }
}

// 会场全景照：选中后立即上传到 file_meta，回填 file_id。
async function handlePanoramaChange(file) {
  // 只在 status='ready'（新文件首次进入）时上传，避免重复触发
  if (file?.status && file.status !== 'ready') {
    return
  }
  panoramaList.value = [file]
  // 标记为上传中，触发 el-upload 自身的 loading 态
  file.status = 'uploading'
  const fd = new FormData()
  fd.append('file', file.raw)
  fd.append('module', 'TY')
  fd.append('biz_type', 'rec_meeting_panorama')
  try {
    const res = await fileApi.upload(fd)
    // 调试：后端是否已包含 file_id 字段
    console.log('[upload panorama] response =', res)
    const fileId = Number(res?.file_id ?? res?.id ?? 0)
    if (!Number.isFinite(fileId) || fileId <= 0) {
      throw new Error('后端未返回 file_id，请确认后端已用最新源码重启')
    }
    form.value.photo_overall_id = fileId
    file.status = 'success'
    // 程序赋值不会触发 el-form 'change' 校验，主动清理 / 重检
    formRef.value?.clearValidate(['photo_overall_id'])
    formRef.value?.validateField?.(['photo_overall_id'])
  } catch (e) {
    ElMessage.error('全景照上传失败：' + (e?.message || '未知错误'))
    form.value.photo_overall_id = null
    file.status = 'fail'
    panoramaList.value = []
    formRef.value?.clearValidate(['photo_overall_id'])
  }
}

// 投票特写照：选中后立即上传到 file_meta，回填 file_id。
async function handleVoteChange(file) {
  if (file?.status && file.status !== 'ready') {
    return
  }
  voteList.value = [file]
  file.status = 'uploading'
  const fd = new FormData()
  fd.append('file', file.raw)
  fd.append('module', 'TY')
  fd.append('biz_type', 'rec_meeting_vote')
  try {
    const res = await fileApi.upload(fd)
    console.log('[upload vote] response =', res)
    const fileId = Number(res?.file_id ?? res?.id ?? 0)
    if (!Number.isFinite(fileId) || fileId <= 0) {
      throw new Error('后端未返回 file_id，请确认后端已用最新源码重启')
    }
    form.value.photo_vote_id = fileId
    file.status = 'success'
    formRef.value?.clearValidate(['photo_vote_id'])
    formRef.value?.validateField?.(['photo_vote_id'])
  } catch (e) {
    ElMessage.error('投票特写照上传失败：' + (e?.message || '未知错误'))
    form.value.photo_vote_id = null
    file.status = 'fail'
    voteList.value = []
    formRef.value?.clearValidate(['photo_vote_id'])
  }
}

function handlePanoramaRemove() {
  form.value.photo_overall_id = null
  // 移除后强制重检，让"请上传…"红字重新出现
  formRef.value?.validateField?.(['photo_overall_id'])
}
function handleVoteRemove() {
  form.value.photo_vote_id = null
  formRef.value?.validateField?.(['photo_vote_id'])
}

// 提交前校验
async function validateBeforeSubmit() {
  // 照片必传校验
  if (!panoramaList.value.length) {
    ElMessage.warning('请上传会场全景照片')
    return false
  }
  if (!voteList.value.length) {
    ElMessage.warning('请上传投票特写照片')
    return false
  }
  // 到会率 ≥ 2/3 校验
  const rate = form.value.actual_count / form.value.expected_count
  if (rate < 2 / 3) {
    ElMessage.warning(`到会率 ${(rate * 100).toFixed(1)}%，不足 2/3，无法召开推优大会`)
    return false
  }
  // 赞成票过半校验
  const totalVotes = form.value.approve_count + form.value.against_count + form.value.abstain_count
  if (totalVotes > 0 && form.value.approve_count <= totalVotes / 2) {
    ElMessage.warning('赞成票未过半数，无法通过决议')
    return false
  }
  return true
}

// 提交
async function handleSubmit() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  if (!await validateBeforeSubmit()) return

  submitting.value = true
  try {
    await tyRecommendationMeetingApi.create(form.value)
    ElMessage.success('推优大会记录创建成功')
    router.push('/ty/recommendation-meeting')
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    submitting.value = false
  }
}

function goBack() {
  router.push('/ty/recommendation-meeting')
}

onMounted(() => {
  fetchApplications()
  // 如果路由带 application_id 参数，自动填充
  if (route.query.application_id) {
    form.value.application_id = Number(route.query.application_id)
  }
})
</script>

<style scoped>
/* .card-header / .upload-tip 已在全局定义 */
</style>

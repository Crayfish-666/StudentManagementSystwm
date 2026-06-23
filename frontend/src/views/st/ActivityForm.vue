<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑活动' : '新建活动' }}</span>
        </div>
      </template>

      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" style="max-width: 700px">
        <el-form-item label="所属社团" prop="association_id">
          <el-select v-model="form.association_id" placeholder="请选择社团" style="width: 100%">
            <el-option v-for="a in assocs" :key="a.id" :label="a.name" :value="a.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="活动名称" prop="title">
          <el-input v-model="form.title" placeholder="请输入活动名称" />
        </el-form-item>

        <el-form-item label="活动等级" prop="level">
          <el-select v-model="form.level" placeholder="请选择等级" style="width: 100%">
            <el-option label="A级（跨校/省/全国）" value="A" />
            <el-option label="B级（跨院系/500人+）" value="B" />
            <el-option label="C级（院系内/100人+）" value="C" />
            <el-option label="D级（100人以下）" value="D" />
          </el-select>
        </el-form-item>

        <el-form-item label="预计参与人数" prop="expected_participants">
          <el-input-number v-model="form.expected_participants" :min="1" :max="10000" style="width: 100%" />
        </el-form-item>

        <el-form-item label="预算（分）" prop="budget_cents">
          <el-input-number v-model="form.budget_cents" :min="0" :step="1000" style="width: 100%" />
          <div style="font-size: 12px; color: #909399; margin-top: 4px">单位：分（1元=100分）</div>
        </el-form-item>

        <el-form-item label="活动地点" prop="location">
          <el-input v-model="form.location" placeholder="请输入活动地点" />
        </el-form-item>

        <el-form-item label="开始时间" prop="started_at">
          <el-date-picker v-model="form.started_at" type="datetime" placeholder="选择开始时间" format="YYYY-MM-DD HH:mm" value-format="YYYY-MM-DDTHH:mm:ss+08:00" style="width: 100%" />
        </el-form-item>

        <el-form-item label="结束时间" prop="ended_at">
          <el-date-picker v-model="form.ended_at" type="datetime" placeholder="选择结束时间" format="YYYY-MM-DD HH:mm" value-format="YYYY-MM-DDTHH:mm:ss+08:00" style="width: 100%" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="submitting">保存</el-button>
          <el-button @click="router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { stActivityApi, stAssociationApi } from '@/api/st'

const router = useRouter()
const route = useRoute()
const isEdit = !!route.params.id

const formRef = ref()
const submitting = ref(false)
const assocs = ref([])

const form = reactive({
  association_id: null,
  title: '',
  level: 'D',
  expected_participants: 10,
  budget_cents: 0,
  location: '',
  started_at: '',
  ended_at: ''
})

const rules = {
  association_id: [{ required: true, message: '请选择所属社团', trigger: 'change' }],
  title: [{ required: true, message: '请输入活动名称', trigger: 'blur' }],
  level: [{ required: true, message: '请选择活动等级', trigger: 'change' }],
  expected_participants: [{ required: true, message: '请输入预计参与人数', trigger: 'blur' }],
  location: [{ required: true, message: '请输入活动地点', trigger: 'blur' }],
  started_at: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  ended_at: [{ required: true, message: '请选择结束时间', trigger: 'change' }]
}

async function fetchAssocs() {
  try {
    const data = await stAssociationApi.list({ page_size: 200 })
    assocs.value = data.items || []
  } catch (e) {
    console.error('获取社团列表失败', e)
  }
}

async function fetchDetail() {
  if (!isEdit) return
  try {
    const data = await stActivityApi.get(route.params.id)
    form.association_id = data.association_id
    form.title = data.title
    form.level = data.level
    form.expected_participants = data.expected_participants
    form.budget_cents = data.budget_cents
    form.location = data.location
    form.started_at = data.started_at
    form.ended_at = data.ended_at
  } catch (e) {
    ElMessage.error('获取活动信息失败')
  }
}

async function handleSave() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEdit) {
      await stActivityApi.update(route.params.id, {
        title: form.title,
        level: form.level,
        expected_participants: form.expected_participants,
        budget_cents: form.budget_cents,
        location: form.location,
        started_at: form.started_at,
        ended_at: form.ended_at
      })
      ElMessage.success('更新成功')
    } else {
      await stActivityApi.create({
        association_id: form.association_id,
        title: form.title,
        level: form.level,
        expected_participants: form.expected_participants,
        budget_cents: form.budget_cents,
        location: form.location,
        started_at: form.started_at,
        ended_at: form.ended_at
      })
      ElMessage.success('创建成功')
    }
    router.push('/st/activity')
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchAssocs()
  if (isEdit) fetchDetail()
})
</script>

<style scoped>
/* .card-header 已在 App.vue 全局定义 */
</style>

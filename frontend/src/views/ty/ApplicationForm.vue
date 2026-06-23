<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑入团申请' : '新增入团申请' }}</span>
          <el-button @click="goBack">返回列表</el-button>
        </div>
      </template>

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px" style="max-width: 800px">
        <el-form-item label="申请团支部" prop="branch_id">
          <el-select v-model="form.branch_id" placeholder="请选择团支部" style="width: 100%">
            <el-option v-for="b in branches" :key="b.id" :label="b.name" :value="b.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="申请日期" prop="apply_date">
          <el-date-picker
            v-model="form.apply_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择日期"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="思想政治表现自述" prop="self_statement">
          <el-input
            v-model="form.self_statement"
            type="textarea"
            :rows="10"
            placeholder="请详细描述你的思想政治表现（不少于500字）"
            show-word-limit
          />
          <div class="word-count" :class="{ warning: statementLength < 500 }">
            已输入 {{ statementLength }} 字（最少 500 字）
          </div>
        </el-form-item>

        <el-form-item label="家庭成员信息">
          <el-input
            v-model="form.family_members_json"
            type="textarea"
            :rows="4"
            placeholder="请填写主要家庭成员信息（选填）"
          />
        </el-form-item>

        <el-form-item label="奖惩情况">
          <el-input
            v-model="form.rewards_punishments"
            type="textarea"
            :rows="3"
            placeholder="请填写在校期间的奖惩情况（选填）"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">
            {{ isEdit ? '保存修改' : '保存草稿' }}
          </el-button>
          <el-button v-if="isEdit && appStatus === 'S0'" type="success" @click="handleSaveAndSubmit" :loading="submitting">
            保存并提交
          </el-button>
          <el-button v-if="!isEdit" type="success" @click="handleSaveAndSubmit" :loading="submitting">
            保存并提交
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { tyApplicationApi, tyBranchApi } from '@/api/ty'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.id)
const appId = computed(() => route.params.id ? Number(route.params.id) : null)
const appStatus = ref('S0')

// 表单
const formRef = ref()
const saving = ref(false)
const submitting = ref(false)
const form = ref({
  branch_id: null,
  apply_date: '',
  self_statement: '',
  family_members_json: '',
  rewards_punishments: ''
})

const formRules = {
  branch_id: [{ required: true, message: '请选择团支部', trigger: 'change' }],
  apply_date: [{ required: true, message: '请选择申请日期', trigger: 'change' }],
  self_statement: [
    { required: true, message: '请填写思想政治表现自述', trigger: 'blur' },
    { min: 500, message: '自述内容不少于500字', trigger: 'blur' }
  ]
}

// 自述字数
const statementLength = computed(() => (form.value.self_statement || '').length)

// 团支部下拉
const branches = ref([])

// 获取团支部列表
async function fetchBranches() {
  try {
    const data = await tyBranchApi.list()
    branches.value = data || []
  } catch (e) {
    console.error('获取团支部列表失败', e)
  }
}

// 加载已有申请数据（编辑模式）
async function fetchApplication() {
  if (!appId.value) return
  try {
    const data = await tyApplicationApi.get(appId.value)
    appStatus.value = data.status
    form.value = {
      branch_id: data.branch_id,
      apply_date: data.apply_date,
      self_statement: data.self_statement,
      family_members_json: data.family_members_json || '',
      rewards_punishments: data.rewards_punishments || ''
    }
  } catch (e) {
    console.error('获取申请详情失败', e)
  }
}

// 保存草稿
async function handleSave() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await tyApplicationApi.update(appId.value, form.value)
      ElMessage.success('保存成功')
    } else {
      await tyApplicationApi.create(form.value)
      ElMessage.success('草稿已保存')
    }
    router.push('/ty/application')
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    saving.value = false
  }
}

// 保存并提交
async function handleSaveAndSubmit() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    let id = appId.value
    if (isEdit.value) {
      await tyApplicationApi.update(id, form.value)
    } else {
      const data = await tyApplicationApi.create(form.value)
      id = data.id
    }
    // 提交
    await tyApplicationApi.submit(id)
    ElMessage.success('提交成功，申请已进入审批流程')
    router.push('/ty/application')
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    submitting.value = false
  }
}

// 返回列表
function goBack() {
  router.push('/ty/application')
}

onMounted(() => {
  fetchBranches()
  if (isEdit.value) {
    fetchApplication()
  }
})
</script>

<style scoped>
/* .card-header / .word-count 已在全局定义 */
</style>

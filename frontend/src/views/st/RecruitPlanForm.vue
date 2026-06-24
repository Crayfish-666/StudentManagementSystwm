<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑招新计划' : '新建招新计划' }}</span>
          <el-button @click="goBack">返回</el-button>
        </div>
      </template>

      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" style="max-width: 720px">
        <el-form-item label="所属社团" prop="association_id">
          <el-select v-model="form.association_id" placeholder="请选择社团" filterable :disabled="isEdit" style="width: 100%">
            <el-option v-for="a in assocs" :key="a.id" :label="a.name" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="招新季节" prop="season">
          <el-radio-group v-model="form.season">
            <el-radio value="autumn">秋季招新</el-radio>
            <el-radio value="spring">春季补招</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="学年" prop="academic_year">
          <el-input v-model="form.academic_year" placeholder="如 2025-2026" />
        </el-form-item>
        <el-form-item label="目标人数" prop="target_count">
          <el-input-number v-model="form.target_count" :min="1" :max="999" />
        </el-form-item>
        <el-form-item label="考核方式">
          <el-input v-model="form.assessment_method" placeholder="如：简历筛选 + 面试 + 作品展示" />
        </el-form-item>
        <el-form-item label="面试时间">
          <el-date-picker
            v-model="form.interview_at"
            type="datetime"
            placeholder="选择面试时间"
            value-format="YYYY-MM-DDTHH:mm:ss+08:00"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSave">保存</el-button>
          <el-button @click="goBack">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { stRecruitPlanApi, stAssociationApi } from '@/api/st'

const route = useRoute()
const router = useRouter()

const formRef = ref(null)
const assocs = ref([])
const planId = computed(() => (route.params.id ? Number(route.params.id) : null))
const isEdit = computed(() => Boolean(planId.value))

const form = reactive({
  association_id: null,
  season: 'autumn',
  academic_year: '',
  target_count: 20,
  assessment_method: '',
  interview_at: null
})

const rules = {
  association_id: [{ required: true, message: '请选择社团', trigger: 'change' }],
  season: [{ required: true, message: '请选择招新季节', trigger: 'change' }],
  academic_year: [{ required: true, message: '请填写学年', trigger: 'blur' }],
  target_count: [{ required: true, type: 'number', min: 1, message: '目标人数须 ≥ 1', trigger: 'blur' }]
}

async function loadAssocs() {
  const r = await stAssociationApi.list({ page: 1, page_size: 200 })
    assocs.value = r.items || []
}

async function loadDetail() {
  if (!isEdit.value) return
  const r = await stRecruitPlanApi.get(planId.value)
  const v = r
  form.association_id = v.association_id
  form.season = v.season
  form.academic_year = v.academic_year
  form.target_count = v.target_count
  form.assessment_method = v.assessment_method || ''
  form.interview_at = v.interview_at || null
}

function goBack() {
  router.push('/st/recruit-plan')
}

async function handleSave() {
  await formRef.value.validate()
  try {
    if (isEdit.value) {
      await stRecruitPlanApi.update(planId.value, form)
      ElMessage.success('已保存')
    } else {
      await stRecruitPlanApi.create(form)
      ElMessage.success('已创建')
    }
    goBack()
  } catch (e) {
    ElMessage.error(e?.message || '保存失败')
  }
}

onMounted(async () => {
  await loadAssocs()
  await loadDetail()
})
</script>

<style scoped>
.page-container { padding: 16px; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
</style>

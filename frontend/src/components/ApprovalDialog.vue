<template>
  <el-dialog
    v-model="visible"
    :title="`审批 - ${stepLabel}`"
    width="520px"
    :close-on-click-modal="false"
    @close="onClose"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
      <el-form-item label="申请编号">
        <span>{{ application?.biz_no }}</span>
      </el-form-item>
      <el-form-item label="申请人">
        <span>{{ application?.student_name }}（{{ application?.student_no }}）</span>
      </el-form-item>
      <el-form-item label="审批结果" prop="result">
        <el-radio-group v-model="form.result">
          <el-radio value="approve">通过</el-radio>
          <el-radio value="reject">驳回</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="审批意见" prop="opinion">
        <el-input
          v-model="form.opinion"
          type="textarea"
          :rows="5"
          placeholder="请输入审批意见（≥5 字）"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button
        :type="form.result === 'approve' ? 'success' : 'danger'"
        :loading="loading"
        @click="handleSubmit"
      >
        提交审批
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { tyApplicationApi } from '@/api/ty'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  application: { type: Object, default: () => ({}) },
  step: { type: String, default: '' } // counselor / college / school
})

const emit = defineEmits(['update:modelValue', 'success'])

const stepTextMap = {
  counselor: '辅导员/团支部初审',
  college: '院系团委复核',
  school: '校团委终审'
}

const visible = ref(false)
const loading = ref(false)
const formRef = ref(null)
const form = reactive({
  result: 'approve',
  opinion: ''
})

const rules = {
  result: [{ required: true, message: '请选择审批结果' }],
  opinion: [
    { required: true, message: '请填写审批意见', trigger: 'blur' },
    { min: 5, message: '审批意见至少 5 字', trigger: 'blur' }
  ]
}

const stepLabel = computed(() => stepTextMap[props.step] || props.step)

watch(
  () => props.modelValue,
  (v) => {
    visible.value = v
    if (v) {
      form.result = 'approve'
      form.opinion = ''
    }
  }
)
watch(visible, (v) => emit('update:modelValue', v))

function onClose() {
  formRef.value?.clearValidate()
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    if (form.result === 'reject') {
      try {
        await ElMessageBox.confirm('确认驳回此申请？驳回后将进入 S4 终止状态。', '驳回确认', {
          type: 'warning'
        })
      } catch {
        return
      }
    }
    try {
      loading.value = true
      await tyApplicationApi.approve(props.application.id, {
        step: props.step,
        result: form.result,
        opinion: form.opinion
      })
      ElMessage.success('审批已提交')
      visible.value = false
      emit('success')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
:deep(.el-form-item__label) {
  color: var(--sh-text-secondary);
  font-weight: 500;
}
:deep(.el-radio__label) {
  color: var(--sh-text-regular);
}
</style>

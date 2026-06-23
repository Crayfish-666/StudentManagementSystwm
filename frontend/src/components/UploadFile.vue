<template>
  <el-upload
    :action="uploadUrl"
    :headers="uploadHeaders"
    :data="uploadData"
    :before-upload="handleBeforeUpload"
    :on-success="handleSuccess"
    :on-error="handleError"
    :on-remove="handleRemove"
    :file-list="fileList"
    :limit="limit"
    :accept="accept"
    :disabled="disabled"
    v-bind="$attrs"
  >
    <el-button type="primary" :disabled="disabled">选择文件</el-button>
    <template #tip>
      <div class="el-upload__tip">
        支持 jpg/png/pdf/doc/docx 格式，单文件不超过 50MB
      </div>
    </template>
  </el-upload>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  module: { type: String, required: true },
  bizType: { type: String, required: true },
  limit: { type: Number, default: 5 },
  accept: { type: String, default: '.jpg,.jpeg,.png,.pdf,.doc,.docx' },
  disabled: { type: Boolean, default: false },
  modelValue: { type: Array, default: () => [] }
})

const emit = defineEmits(['update:modelValue', 'success', 'remove'])

const fileList = ref([])

const uploadUrl = computed(() => '/api/v1/files/upload')

const uploadHeaders = computed(() => {
  const token = localStorage.getItem('access_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
})

const uploadData = computed(() => ({
  module: props.module,
  biz_type: props.bizType
}))

const maxSize = 50 * 1024 * 1024 // 50MB

const allowedTypes = [
  'image/jpeg', 'image/png', 'image/gif', 'image/webp',
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'text/plain'
]

function handleBeforeUpload(file) {
  if (file.size > maxSize) {
    ElMessage.error('文件大小不能超过 50MB')
    return false
  }
  if (!allowedTypes.includes(file.type) && !file.type.startsWith('image/')) {
    ElMessage.error('不支持的文件类型')
    return false
  }
  return true
}

function handleSuccess(response, uploadFile, uploadFiles) {
  // http 拦截器已解包，response 就是 data
  const fileKey = response?.key || response?.data?.key
  if (fileKey) {
    emit('success', { key: fileKey, name: uploadFile.name, response })
    const current = [...props.modelValue]
    current.push({ key: fileKey, name: uploadFile.name })
    emit('update:modelValue', current)
  }
}

function handleError() {
  ElMessage.error('文件上传失败')
}

function handleRemove(uploadFile) {
  const current = props.modelValue.filter(f => f.name !== uploadFile.name)
  emit('update:modelValue', current)
  emit('remove', uploadFile)
}
</script>

<style scoped>
:deep(.el-upload__tip) {
  color: var(--sh-text-placeholder);
  font-size: var(--sh-text-xs);
  margin-top: var(--sh-space-xs);
}
</style>

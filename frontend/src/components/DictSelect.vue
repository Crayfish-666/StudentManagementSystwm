<template>
  <el-select
    v-model="modelValue"
    :placeholder="placeholder || '请选择'"
    :loading="loading"
    clearable
    @update:modelValue="$emit('update:modelValue', $event)"
  >
    <el-option
      v-for="item in items"
      :key="item.code"
      :label="item.name_zh"
      :value="item.code"
    />
  </el-select>
</template>

<script setup>
import { useDict } from '@/stores/dict'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  category: { type: String, required: true },
  placeholder: { type: String, default: '' }
})

defineEmits(['update:modelValue'])

const { items, loading } = useDict(props.category)
</script>

<style scoped>
:deep(.el-select) {
  width: 100%;
}
</style>

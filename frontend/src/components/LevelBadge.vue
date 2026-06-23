<template>
  <el-tag :type="tagType" :effect="effect" :size="size">
    {{ levelText }}
  </el-tag>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  level: { type: String, required: true },
  size: { type: String, default: 'default' },
  effect: { type: String, default: 'light' }
})

const levelMap = {
  L1: { text: 'L1-常规报修', type: 'info' },
  L2: { text: 'L2-违规/矛盾', type: 'warning' },
  L3: { text: 'L3-聚众/打架/隐患', type: 'danger' },
  L4: { text: 'L4-火警/群体/伤害', type: 'danger' }
}

const tagType = computed(() => levelMap[props.level]?.type || 'info')
const levelText = computed(() => levelMap[props.level]?.text || props.level)
</script>

<style scoped>
/* 级别颜色映射 — 使用设计令牌覆盖 Element Plus Tag 默认色 */
:deep(.el-tag--info) {
  --el-tag-bg-color: var(--sh-info-light);
  --el-tag-border-color: var(--sh-info);
  --el-tag-text-color: var(--sh-info);
}
:deep(.el-tag--warning) {
  --el-tag-bg-color: var(--sh-warning-light);
  --el-tag-border-color: var(--sh-warning);
  --el-tag-text-color: var(--sh-warning);
}
:deep(.el-tag--danger) {
  --el-tag-bg-color: var(--sh-danger-light);
  --el-tag-border-color: var(--sh-danger);
  --el-tag-text-color: var(--sh-danger);
}
</style>

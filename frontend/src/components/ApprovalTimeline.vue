<template>
  <el-timeline v-if="records.length > 0">
    <el-timeline-item
      v-for="rec in records"
      :key="rec.id"
      :type="rec.result === 'approve' ? 'success' : 'danger'"
      :timestamp="formatTime(rec.occurred_at)"
      placement="top"
    >
      <el-card shadow="hover" class="rec-card">
        <div class="rec-head">
          <el-tag :type="rec.result === 'approve' ? 'success' : 'danger'" size="small">
            {{ rec.step_text }} · {{ rec.result_text }}
          </el-tag>
          <span class="rec-meta">
            {{ rec.approver_name }}（{{ rec.approver_role }}）
          </span>
          <span class="rec-meta rec-status">
            {{ rec.from_status }} → {{ rec.to_status }}
          </span>
        </div>
        <div class="rec-opinion">{{ rec.opinion }}</div>
      </el-card>
    </el-timeline-item>
  </el-timeline>
  <el-empty v-else description="暂无审批记录" :image-size="80" />
</template>

<script setup>
import { formatDateTime as formatTime } from '@/utils/datetime'

defineProps({
  records: {
    type: Array,
    default: () => []
  }
})
</script>

<style scoped>
.rec-card {
  margin-bottom: var(--sh-space-xs);
  border-left: 3px solid var(--sh-primary);
  transition: border-color var(--sh-duration-fast) var(--sh-ease-out);
}
.rec-card:hover {
  border-left-color: var(--sh-accent);
}
.rec-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--sh-space-md);
  margin-bottom: var(--sh-space-sm);
}
.rec-meta {
  color: var(--sh-text-regular);
  font-size: var(--sh-text-sm);
}
.rec-status {
  color: var(--sh-text-secondary);
}
.rec-opinion {
  white-space: pre-wrap;
  word-break: break-all;
  line-height: var(--sh-leading-normal);
  color: var(--sh-text-primary);
}
</style>

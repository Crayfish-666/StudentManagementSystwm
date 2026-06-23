<template>
  <el-table :data="items" stripe v-loading="loading">
    <el-table-column prop="student_no" label="学号" width="130" />
    <el-table-column prop="student_name" label="姓名" width="100" />
    <el-table-column label="签到时间" width="170">
      <template #default="{ row }">{{ formatDateTime(row.checkin_at) }}</template>
    </el-table-column>
    <el-table-column label="签到方式" width="100">
      <template #default="{ row }">{{ methodText[row.method] || row.method }}</template>
    </el-table-column>
    <el-table-column label="是否迟到" width="90">
      <template #default="{ row }">
        <el-tag v-if="row.is_late" type="danger" size="small">迟到</el-tag>
        <el-tag v-else type="success" size="small">准时</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="late_minutes" label="迟到(分)" width="100">
      <template #default="{ row }">{{ row.is_late ? row.late_minutes : '-' }}</template>
    </el-table-column>
    <el-table-column label="是否出勤" width="90">
      <template #default="{ row }">
        <el-tag v-if="row.counted_as_absent" type="info" size="small">缺勤</el-tag>
        <el-tag v-else type="success" size="small">出勤</el-tag>
      </template>
    </el-table-column>
  </el-table>
  <div class="pagination-wrap" v-if="total > 0">
    <el-pagination
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      :page-sizes="[20, 50, 100]"
      layout="total, sizes, prev, pager, next"
      @current-change="onPageChange"
      @size-change="onSizeChange"
    />
  </div>
</template>

<script setup>
import { formatDateTime } from '@/utils/datetime'

const props = defineProps({
  items: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  total: { type: Number, default: 0 },
  page: { type: Number, default: 1 },
  pageSize: { type: Number, default: 20 }
})

const emit = defineEmits(['change'])

const methodText = {
  qrcode: '扫码',
  gps: 'GPS定位',
  manual: '手动补登'
}

function onPageChange(newPage) {
  emit('change', { page: newPage, pageSize: props.pageSize })
}

function onSizeChange(newSize) {
  emit('change', { page: 1, pageSize: newSize })
}
</script>

<style scoped>
/* .pagination-wrap 已在 App.vue 全局定义 */
</style>

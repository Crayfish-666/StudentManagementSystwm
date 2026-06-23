<template>
  <div class="student-import">
    <el-card shadow="hover">
      <template #header>
        <span style="font-size: 16px; font-weight: 600">学生批量导入</span>
      </template>

      <el-alert
        title="CSV 格式说明"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      >
        <template #default>
          <p>支持 <strong>UTF-8</strong>（推荐，含/不含 BOM）或 <strong>GBK / GB18030</strong>（Excel "另存为 CSV" 默认编码）；首行为表头，字段顺序如下：</p>
          <code style="font-size: 13px; line-height: 1.8">
            学号,姓名,性别,身份证号,手机号,院系ID,专业ID,班级ID,年级,邮箱
          </code>
          <p style="margin-top: 6px; color: #909399">
            性别填写：M=男 / F=女；院系ID、专业ID、班级ID 需先在组织管理中创建。
          </p>
          <p style="margin-top: 4px; color: #909399">
            如导入后中文显示为 "?"，请将文件用记事本"另存为"时编码切换为 UTF-8 后重新上传。
          </p>
        </template>
      </el-alert>

      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :limit="1"
        accept=".csv"
        :on-change="onFileChange"
        :on-exceed="onExceed"
        drag
      >
        <el-icon style="font-size: 48px; color: #c0c4cc"><UploadFilled /></el-icon>
        <div style="margin-top: 8px">将 CSV 文件拖到此处，或 <em>点击上传</em></div>
        <template #tip>
          <div style="color: #909399; font-size: 12px">仅支持 .csv 文件，单次限传 1 个</div>
        </template>
      </el-upload>

      <div style="margin-top: 20px; text-align: right">
        <el-button type="primary" :loading="importing" :disabled="!file" @click="handleImport">
          开始导入
        </el-button>
      </div>

      <!-- 导入结果 -->
      <el-result
        v-if="result"
        :icon="result.success ? 'success' : 'warning'"
        :title="result.success ? '导入完成' : '导入完成（部分失败）'"
        style="margin-top: 20px"
      >
        <template #extra>
          <div class="result-detail">
            <p>成功导入：<strong>{{ result.success_count }}</strong> 条</p>
            <p v-if="result.fail_count > 0">
              失败：<strong style="color: #f56c6c">{{ result.fail_count }}</strong> 条
            </p>
            <div v-if="result.errors && result.errors.length" class="error-list">
              <p style="color: #f56c6c; margin-top: 8px">错误详情：</p>
              <ul>
                <li v-for="(err, idx) in result.errors" :key="idx">{{ err }}</li>
              </ul>
            </div>
          </div>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { studentApi } from '@/api/idx'

const uploadRef = ref()
const file = ref(null)
const importing = ref(false)
const result = ref(null)

function onFileChange(uploadFile) {
  file.value = uploadFile.raw
  result.value = null
}

function onExceed() {
  ElMessage.warning('仅支持上传 1 个文件，请先移除已选文件')
}

async function handleImport() {
  if (!file.value) {
    ElMessage.warning('请先选择 CSV 文件')
    return
  }
  importing.value = true
  result.value = null
  try {
    const data = await studentApi.importCSV(file.value)
    result.value = data
    if (data.fail_count === 0) {
      ElMessage.success(`成功导入 ${data.success_count} 条学生数据`)
    } else {
      ElMessage.warning(`导入完成：成功 ${data.success_count} 条，失败 ${data.fail_count} 条`)
    }
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.student-import {
  max-width: 700px;
  padding: var(--sh-space-lg);
}
.result-detail {
  text-align: left;
  font-size: var(--sh-text-base);
  line-height: var(--sh-leading-relaxed);
}
.error-list {
  max-height: 200px;
  overflow-y: auto;
  text-align: left;
}
.error-list ul {
  margin: var(--sh-space-xs) 0 0 0;
  padding-left: 20px;
}
.error-list li {
  font-size: var(--sh-text-sm);
  line-height: var(--sh-leading-relaxed);
}
</style>

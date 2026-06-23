<template>
  <div class="my-profile">
    <el-card shadow="hover" v-loading="loading">
      <template #header>
        <span style="font-size: 16px; font-weight: 600">我的档案</span>
      </template>

      <template v-if="profile">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="学号">{{ profile.student_no }}</el-descriptions-item>
          <el-descriptions-item label="姓名">{{ profile.name }}</el-descriptions-item>
          <el-descriptions-item label="性别">{{ genderMap[profile.gender] || profile.gender }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ profile.id_card_masked || '-' }}</el-descriptions-item>
          <el-descriptions-item label="院系">{{ profile.college_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="专业">{{ profile.major_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="班级">{{ profile.class_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="年级">{{ profile.grade || '-' }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ profile.phone_masked || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ profile.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="政治面貌">{{ politicalMap[profile.political_status] || profile.political_status || '-' }}</el-descriptions-item>
          <el-descriptions-item label="入学日期">{{ profile.enrollment_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType[profile.status]" size="small">
              {{ statusMap[profile.status] || profile.status || '-' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="出生日期">{{ profile.birth_date || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>

      <el-empty v-else-if="!loading" description="暂无档案信息" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { studentApi } from '@/api/idx'

const profile = ref(null)
const loading = ref(false)

const genderMap = { M: '男', F: '女', U: '未知' }
const statusMap = { enrolled: '在读', suspended: '休学', graduated: '毕业', withdrawn: '退学' }
const statusType = { enrolled: 'success', suspended: 'warning', graduated: 'info', withdrawn: 'danger' }
const politicalMap = {
  masses: '群众',
  activist: '入团积极分子',
  probationary: '预备团员',
  member: '共青团员',
  party_probationary: '预备党员',
  party_member: '中共党员'
}

async function fetchProfile() {
  loading.value = true
  try {
    const data = await studentApi.getMyProfile()
    profile.value = data
  } catch (e) {
    // 404 表示该用户无关联学生档案，静默处理
    profile.value = null
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchProfile()
})
</script>

<style scoped>
.my-profile {
  max-width: 800px;
  padding: var(--sh-space-lg);
}
</style>

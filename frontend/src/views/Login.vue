<template>
  <div class="sh-login-wrapper">
    <!-- Animated Ambient Glowing Orbs -->
    <div class="sh-orb orb-1"></div>
    <div class="sh-orb orb-2"></div>
    <div class="sh-orb orb-3"></div>

    <div class="sh-login-card sh-glass-card sh-animate-slide-up">
      <!-- Left Panel: Branding & Slogan -->
      <div class="sh-login-brand">
        <div class="brand-badge">
          <span class="badge-dot"></span>
          <span>StudentHub v2.1 SpringBoot</span>
        </div>
        <h1 class="brand-title">学生“一站式”自主管理过程管理系统</h1>
        <p class="brand-desc">
          基于全周期过程档案与时间戳记录，覆盖团员发展、社团活动、社区自治与勤工助学四大模块。
        </p>

        <div class="brand-features">
          <div class="feature-item">
            <el-icon class="feature-icon"><Check /></el-icon>
            <span>多角色 RBAC / ABAC 权限动态切控</span>
          </div>
          <div class="feature-item">
            <el-icon class="feature-icon"><Check /></el-icon>
            <span>MinIO 对象存储分片上传与时效预览</span>
          </div>
          <div class="feature-item">
            <el-icon class="feature-icon"><Check /></el-icon>
            <span>Spring AI / DeepSeek 大模型综测自动考评</span>
          </div>
        </div>

        <div class="quick-preset-box">
          <span class="preset-title">快速测试账户：</span>
          <div class="preset-btns">
            <el-tag class="preset-tag" effect="dark" type="primary" @click="fillAccount('admin', 'admin@123')">
              系统管理员 (admin)
            </el-tag>
            <el-tag class="preset-tag" effect="dark" type="success" @click="fillAccount('2023010101', 'student@123')">
              普通学生 (2023010101)
            </el-tag>
          </div>
        </div>
      </div>

      <!-- Right Panel: Interactive Form -->
      <div class="sh-login-form-area">
        <div class="form-header">
          <h2>账号身份验证</h2>
          <p>请使用学号 / 工号 / 管理员账号登录</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          size="large"
          class="sh-login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="学号 / 工号 / 用户名"
              :prefix-icon="User"
              clearable
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              :prefix-icon="Lock"
              show-password
            />
          </el-form-item>

          <div class="form-options">
            <el-checkbox v-model="rememberMe" label="记住当前设备" />
            <el-link type="primary" :underline="false">忘记密码？</el-link>
          </div>

          <button
            type="button"
            class="sh-btn-gradient login-submit-btn"
            :disabled="loading"
            @click="handleLogin"
          >
            <el-icon v-if="loading" class="is-loading"><Loading /></el-icon>
            <el-icon v-else><Right /></el-icon>
            <span>{{ loading ? '正在进行全身份验证...' : '立 即 登 录' }}</span>
          </button>
        </el-form>

        <div class="form-footer">
          <span>StudentHub Management Center &copy; 2026</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, Check, Right, Loading } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useMenuStore } from '@/stores/menu'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const menuStore = useMenuStore()

const formRef = ref(null)
const loading = ref(false)
const rememberMe = ref(true)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入学号/工号/用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

function fillAccount(u, p) {
  form.username = u
  form.password = p
}

async function handleLogin() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await authStore.login(form.username, form.password)
      ElMessage.success(`登录成功！欢迎回来，${authStore.displayName}`)
      await menuStore.fetchMenus(router)
      router.push('/dashboard')
    } catch (err) {
      ElMessage.error(err?.message || '登录失败，请检查账号密码')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.sh-login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--sh-bg-main);
  position: relative;
  overflow: hidden;
  padding: 20px;
}

/* Glowing Ambient Orbs */
.sh-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px);
  opacity: 0.45;
  pointer-events: none;
}
.orb-1 {
  width: 450px;
  height: 450px;
  background: #6366f1;
  top: -100px;
  left: -100px;
}
.orb-2 {
  width: 550px;
  height: 550px;
  background: #8b5cf6;
  bottom: -150px;
  right: -150px;
}
.orb-3 {
  width: 350px;
  height: 350px;
  background: #06b6d4;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.sh-login-card {
  width: 100%;
  max-width: 980px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  overflow: hidden;
  z-index: 10;
}

/* Left Panel */
.sh-login-brand {
  padding: 48px;
  background: linear-gradient(135deg, rgba(30, 27, 75, 0.6) 0%, rgba(15, 23, 42, 0.8) 100%);
  border-right: 1px solid var(--sh-border-color);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.brand-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 20px;
  font-size: 12px;
  color: var(--sh-primary);
  width: fit-content;
}
.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--sh-primary);
  box-shadow: 0 0 8px var(--sh-primary);
}

.brand-title {
  font-family: 'Outfit', sans-serif;
  font-size: 26px;
  font-weight: 700;
  line-height: 1.35;
  margin-top: 24px;
  background: linear-gradient(135deg, #ffffff 0%, #cbd5e1 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.brand-desc {
  font-size: 14px;
  color: var(--sh-text-secondary);
  margin-top: 16px;
  line-height: 1.6;
}

.brand-features {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 32px;
}
.feature-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--sh-text-primary);
}
.feature-icon {
  color: var(--sh-accent-emerald);
}

.quick-preset-box {
  margin-top: 36px;
  padding-top: 20px;
  border-top: 1px dashed var(--sh-border-color);
}
.preset-title {
  font-size: 12px;
  color: var(--sh-text-muted);
  display: block;
  margin-bottom: 10px;
}
.preset-btns {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.preset-tag {
  cursor: pointer;
  transition: transform 0.2s;
}
.preset-tag:hover {
  transform: scale(1.05);
}

/* Right Panel */
.sh-login-form-area {
  padding: 48px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.form-header h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--sh-text-primary);
}
.form-header p {
  font-size: 13px;
  color: var(--sh-text-secondary);
  margin-top: 6px;
}

.sh-login-form {
  margin-top: 32px;
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.login-submit-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  justify-content: center;
}

.form-footer {
  text-align: center;
  font-size: 12px;
  color: var(--sh-text-muted);
  margin-top: 24px;
}

@media (max-width: 768px) {
  .sh-login-card {
    grid-template-columns: 1fr;
  }
  .sh-login-brand {
    display: none;
  }
}
</style>

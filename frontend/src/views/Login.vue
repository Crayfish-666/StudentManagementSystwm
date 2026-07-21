<template>
  <div class="stitch-login-canvas">
    <div class="login-card sh-stitch-card">
      <!-- Left Panel: Nexus Campus Branding -->
      <div class="brand-panel">
        <div class="campus-badge">
          <span class="badge-dot"></span>
          <span>Nexus Campus - StudentHub v2.1</span>
        </div>

        <div class="brand-hero">
          <h1 class="hero-title">学生一站式自主管理过程管理系统</h1>
          <p class="hero-sub">
            围绕“学生主体 + 过程档案”，覆盖团员发展、社团活动、学生社区、勤工助学与综合素质考评。
          </p>
        </div>

        <div class="feature-bullets">
          <div class="bullet-item">
            <div class="bullet-icon"><el-icon><Select /></el-icon></div>
            <div>
              <div class="bullet-title">三角色动态切控</div>
              <div class="bullet-desc">支持系统管理员、院系辅导员、普通学生权限隔离</div>
            </div>
          </div>

          <div class="bullet-item">
            <div class="bullet-icon"><el-icon><Select /></el-icon></div>
            <div>
              <div class="bullet-title">Spring AI / DeepSeek 综测</div>
              <div class="bullet-desc">整合五维量化评分与大模型智能综合评价</div>
            </div>
          </div>
        </div>

        <div class="preset-account-area">
          <span class="preset-label">快捷身份预设（一键填充）：</span>
          <div class="preset-chips">
            <div class="sh-chip sh-chip-navy cursor-pointer" @click="fillAccount('admin', 'admin@123')">
              系统管理员 (admin)
            </div>
            <div class="sh-chip sh-chip-amber cursor-pointer" @click="fillAccount('counselor', 'counselor@123')">
              院系辅导员 (counselor)
            </div>
            <div class="sh-chip sh-chip-emerald cursor-pointer" @click="fillAccount('2023010101', 'student@123')">
              普通学生 (2023010101)
            </div>
          </div>
        </div>
      </div>

      <!-- Right Panel: Stitch Form -->
      <div class="form-panel">
        <div class="form-title-group">
          <h2>账号身份验证</h2>
          <p>请选择学号 / 辅导员工号 / 管理员账号登录</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          size="large"
          class="stitch-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="学号 / 辅导员工号 / 用户名"
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

          <div class="form-row-options">
            <el-checkbox v-model="rememberMe" label="记住登录设备" />
            <el-link type="primary" :underline="false">忘记密码？</el-link>
          </div>

          <button
            type="button"
            class="sh-btn-stitch submit-btn"
            :disabled="loading"
            @click="handleLogin"
          >
            <el-icon v-if="loading" class="is-loading"><Loading /></el-icon>
            <el-icon v-else><Right /></el-icon>
            <span>{{ loading ? '正在验证身份...' : '安全登录' }}</span>
          </button>
        </el-form>

        <div class="stitch-form-footer">
          <span>Nexus Campus Management System &copy; 2026</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, Select, Right, Loading } from '@element-plus/icons-vue'
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
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
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
.stitch-login-canvas {
  min-height: 100vh;
  background: var(--sh-bg-main);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 960px;
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  overflow: hidden;
}

/* Left Brand Panel */
.brand-panel {
  padding: 40px;
  background: linear-gradient(135deg, #00164e 0%, #00236f 100%);
  color: #ffffff;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.campus-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  width: fit-content;
}
.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #6ffbbe;
}

.hero-title {
  font-size: 24px;
  font-weight: 700;
  line-height: 1.4;
  margin-top: 24px;
}
.hero-sub {
  font-size: 13px;
  color: #b6c4ff;
  margin-top: 12px;
  line-height: 1.6;
}

.feature-bullets {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 24px;
}
.bullet-item {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}
.bullet-icon {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: rgba(111, 251, 190, 0.2);
  color: #6ffbbe;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  flex-shrink: 0;
}
.bullet-title {
  font-size: 14px;
  font-weight: 600;
}
.bullet-desc {
  font-size: 12px;
  color: #dce1ff;
  margin-top: 2px;
}

.preset-account-area {
  margin-top: 28px;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.15);
}
.preset-label {
  font-size: 12px;
  color: #b6c4ff;
  display: block;
  margin-bottom: 10px;
}
.preset-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.cursor-pointer {
  cursor: pointer;
  transition: transform 0.15s;
}
.cursor-pointer:hover {
  transform: scale(1.04);
}

/* Right Form Panel */
.form-panel {
  padding: 40px;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.form-title-group h2 {
  font-size: 22px;
  font-weight: 700;
  color: var(--sh-primary);
}
.form-title-group p {
  font-size: 13px;
  color: var(--sh-text-secondary);
  margin-top: 4px;
}

.stitch-form {
  margin-top: 28px;
}

.form-row-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.submit-btn {
  width: 100%;
  height: 46px;
  justify-content: center;
  font-size: 15px;
}

.stitch-form-footer {
  text-align: center;
  font-size: 12px;
  color: var(--sh-text-muted);
  margin-top: 20px;
}

@media (max-width: 768px) {
  .login-card {
    grid-template-columns: 1fr;
  }
  .brand-panel {
    display: none;
  }
}
</style>

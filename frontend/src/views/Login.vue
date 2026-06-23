<template>
  <div class="login-page">
    <!-- 左侧：品牌展示区（轮播背景） -->
    <div class="login-brand">
      <div class="brand-slider">
        <img class="slide-img slide-1" src="/images/img01.jpg" alt="校园风光" />
        <img class="slide-img slide-2" src="/images/img02.jpg" alt="校园风光" />
      </div>
      <div class="brand-overlay">
        <div class="brand-header">
          <img
            class="brand-logo"
            src="/images/logo.png"
            alt="福州软件职业技术学院"
          />
          <div class="brand-text">
            <span class="brand-divider">|</span>
            <span class="brand-subtitle">学生一站式自主管理平台</span>
          </div>
        </div>
        <!-- 底部标语 -->
        <div class="brand-slogan">
          <div class="slogan-line" />
          <span class="slogan-text">一个入口 · 一套身份 · 一条主线</span>
        </div>
      </div>
    </div>

    <!-- 右侧：登录表单区 -->
    <div class="login-form-panel">
      <div class="form-container">
        <div class="form-brand">
          <img src="/images/logo.png" alt="Logo" class="form-logo" />
        </div>
        <h2 class="form-title">欢迎回来</h2>
        <p class="form-subtitle">请登录您的账号以继续</p>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="0"
          size="large"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="工号 / 学号 / 用户名"
              :prefix-icon="User"
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="密码"
              :prefix-icon="Lock"
              show-password
            />
          </el-form-item>

          <el-button
            type="primary"
            class="login-btn"
            :loading="loading"
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form>

        <div class="form-footer">
          <span>福州软件职业技术学院</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref(null)
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入工号 / 学号 / 用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    router.push('/dashboard')
  } catch {
    // http.js 已处理错误提示
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  min-height: 100vh;
  background: var(--sh-bg-base);
}

/* ===== 左侧品牌区 + 轮播 ===== */
.login-brand {
  position: relative;
  flex: 1;
  min-width: 420px;
  overflow: hidden;
}

/* 轮播容器 */
.brand-slider {
  position: absolute;
  inset: 0;
}

.slide-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  animation-duration: 10s;
  animation-iteration-count: infinite;
  animation-timing-function: ease-in-out;
}

.slide-1 {
  opacity: 1;
  animation-name: fade1;
}

.slide-2 {
  opacity: 0;
  animation-name: fade2;
}

@keyframes fade1 {
  0%, 44%   { opacity: 1; }
  50%, 94% { opacity: 0; }
  100%      { opacity: 1; }
}

@keyframes fade2 {
  0%, 44%   { opacity: 0; }
  50%, 94% { opacity: 1; }
  100%      { opacity: 0; }
}

.brand-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    180deg,
    rgba(27, 58, 75, 0.25) 0%,
    rgba(45, 106, 122, 0.08) 40%,
    rgba(27, 58, 75, 0.35) 100%
  );
  z-index: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.brand-header {
  position: absolute;
  top: 32px;
  left: 36px;
  display: flex;
  align-items: center;
  gap: 12px;
  z-index: 2;
}

.brand-logo {
  height: 44px;
  width: auto;
  object-fit: contain;
  filter: brightness(0) invert(1);
}

.brand-text {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
  top: -2px;
}

.brand-divider {
  font-size: 28px;
  font-weight: 300;
  color: rgba(255, 255, 255, 0.6);
}

.brand-subtitle {
  font-size: 22px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 2px;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

/* 底部标语 */
.brand-slogan {
  position: absolute;
  bottom: 40px;
  left: 36px;
  right: 36px;
  display: flex;
  align-items: center;
  gap: 16px;
  z-index: 2;
}
.slogan-line {
  flex: 1;
  height: 1px;
  background: rgba(255, 255, 255, 0.3);
}
.slogan-text {
  color: rgba(255, 255, 255, 0.85);
  font-size: var(--sh-text-sm);
  letter-spacing: 4px;
  font-weight: 500;
  white-space: nowrap;
}

/* ===== 右侧表单区 ===== */
.login-form-panel {
  flex: 0 0 480px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 60px;
  background: var(--sh-bg-white);
}

.form-container {
  width: 100%;
  max-width: 340px;
}

.form-brand {
  margin-bottom: var(--sh-space-xl);
}
.form-logo {
  height: 40px;
  width: auto;
  object-fit: contain;
}

.form-title {
  margin: 0 0 8px 0;
  font-size: var(--sh-text-3xl);
  font-weight: 700;
  color: var(--sh-text-primary);
  letter-spacing: -0.01em;
}

.form-subtitle {
  margin: 0 0 var(--sh-space-xl) 0;
  font-size: var(--sh-text-md);
  color: var(--sh-text-secondary);
}

/* 表单项样式覆盖 */
.form-container :deep(.el-form-item) {
  margin-bottom: 20px;
}

.form-container :deep(.el-input__wrapper) {
  border-radius: var(--sh-radius-lg);
  padding: 4px 16px;
  box-shadow: 0 0 0 1px var(--sh-border) inset;
  transition: box-shadow var(--sh-duration-fast) var(--sh-ease-out);
}

.form-container :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--sh-border-dark) inset;
}

.form-container :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1.5px var(--sh-primary), 0 0 0 4px rgba(var(--sh-primary-rgb), 0.1);
}

.form-container :deep(.el-input__inner) {
  height: 44px;
  font-size: var(--sh-text-md);
  color: var(--sh-text-primary);
}

.form-container :deep(.el-input__prefix .el-icon) {
  font-size: 17px;
  color: var(--sh-text-placeholder);
}

.form-container :deep(.el-input__inner::placeholder) {
  color: var(--sh-text-placeholder);
}

/* 登录按钮 */
.login-btn {
  width: 100%;
  height: 48px;
  border-radius: var(--sh-radius-lg);
  font-size: var(--sh-text-lg);
  font-weight: 600;
  letter-spacing: 6px;
  background: var(--sh-primary);
  border: none;
  transition: all var(--sh-duration-fast) var(--sh-ease-out);
}

.login-btn:hover {
  background: var(--sh-primary-light);
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(var(--sh-primary-rgb), 0.35);
}

.login-btn:active {
  transform: translateY(0);
  background: var(--sh-primary-dark);
}

/* 底部 */
.form-footer {
  margin-top: var(--sh-space-2xl);
  text-align: center;
  font-size: var(--sh-text-xs);
  color: var(--sh-text-placeholder);
  letter-spacing: 1px;
}

/* ===== 响应式适配 ===== */
@media (max-width: 900px) {
  .login-brand {
    display: none;
  }

  .login-form-panel {
    flex: 1;
    min-width: 0;
    padding: 32px 28px;
  }
}
</style>

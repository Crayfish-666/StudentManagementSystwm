<template>
  <div class="error-page">
    <div class="error-content">
      <div class="error-code">403</div>
      <div class="error-divider" />
      <h2 class="error-title">访问受限</h2>
      <p class="error-desc">抱歉，您没有权限访问此页面</p>
      <el-button type="primary" size="large" @click="$router.push('/dashboard')">
        返回首页
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.error-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80vh;
  padding: var(--sh-space-xl);
}
.error-content {
  text-align: center;
  animation: sh-fade-in var(--sh-duration-slow) var(--sh-ease-out);
}
.error-code {
  font-size: 120px;
  font-weight: 800;
  color: var(--sh-danger);
  line-height: 1;
  letter-spacing: -4px;
  opacity: 0.15;
}
.error-divider {
  width: 48px;
  height: 3px;
  background: var(--sh-accent);
  border-radius: 2px;
  margin: var(--sh-space-md) auto;
}
.error-title {
  margin: 0 0 var(--sh-space-sm);
  font-size: var(--sh-text-2xl);
  font-weight: 700;
  color: var(--sh-text-primary);
}
.error-desc {
  margin: 0 0 var(--sh-space-xl);
  font-size: var(--sh-text-md);
  color: var(--sh-text-secondary);
}
</style>

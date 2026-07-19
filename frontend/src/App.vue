<template>
  <router-view />
</template>

<script setup>
// 根组件：仅承载路由视图。
</script>

<style>
/* ============================================================
   StudentHub 全局设计系统 — "学术雅致" (Academic Elegance)
   色彩基调：深靛青 + 暖琥珀 + 色调化中性色
   ============================================================ */

:root {
  /* ── 主色 ── */
  --sh-primary: #2D6A7A;
  --sh-primary-light: #3A8A9D;
  --sh-primary-lighter: #E8F4F7;
  --sh-primary-dark: #1B3A4B;
  --sh-primary-rgb: 45, 106, 122;

  /* ── 强调色 ── */
  --sh-accent: #D4915C;
  --sh-accent-light: #E8B88A;
  --sh-accent-lighter: #FDF3EB;
  --sh-accent-dark: #B87440;

  /* ── 语义色 ── */
  --sh-success: #3D9970;
  --sh-success-light: #E6F5ED;
  --sh-warning: #D4915C;
  --sh-warning-light: #FDF3EB;
  --sh-danger: #C0392B;
  --sh-danger-light: #FDECEB;
  --sh-info: #5B8FA8;
  --sh-info-light: #EBF2F6;

  /* ── 中性色（靛青色调化） ── */
  --sh-text-primary: #1A2B36;
  --sh-text-regular: #3B5060;
  --sh-text-secondary: #6B8294;
  --sh-text-placeholder: #9FB3C0;
  --sh-text-disabled: #B8C9D3;

  --sh-bg-base: #F3F6F8;
  --sh-bg-white: #FFFFFF;
  --sh-bg-elevated: #FAFBFC;
  --sh-bg-overlay: rgba(26, 43, 54, 0.45);

  --sh-border-light: #E2E8ED;
  --sh-border: #CBD5DD;
  --sh-border-dark: #A8B8C3;

  /* ── 间距 ── */
  --sh-space-xs: 4px;
  --sh-space-sm: 8px;
  --sh-space-md: 16px;
  --sh-space-lg: 24px;
  --sh-space-xl: 32px;
  --sh-space-2xl: 48px;
  --sh-space-3xl: 64px;

  /* ── 圆角 ── */
  --sh-radius-sm: 4px;
  --sh-radius-md: 8px;
  --sh-radius-lg: 12px;
  --sh-radius-xl: 16px;

  /* ── 阴影 ── */
  --sh-shadow-sm: 0 1px 3px rgba(26, 43, 54, 0.06);
  --sh-shadow-md: 0 4px 12px rgba(26, 43, 54, 0.08);
  --sh-shadow-lg: 0 8px 24px rgba(26, 43, 54, 0.1);
  --sh-shadow-xl: 0 16px 48px rgba(26, 43, 54, 0.12);

  /* ── 排版 ── */
  --sh-font-display: 'Plus Jakarta Sans', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif;
  --sh-font-body: 'PingFang SC', 'Microsoft YaHei', 'Helvetica Neue', system-ui, sans-serif;
  --sh-font-mono: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;

  --sh-text-xs: 12px;
  --sh-text-sm: 13px;
  --sh-text-base: 14px;
  --sh-text-md: 15px;
  --sh-text-lg: 16px;
  --sh-text-xl: 18px;
  --sh-text-2xl: 22px;
  --sh-text-3xl: 28px;

  --sh-leading-tight: 1.3;
  --sh-leading-normal: 1.6;
  --sh-leading-relaxed: 1.8;

  /* ── 动画 ── */
  --sh-ease-out: cubic-bezier(0.25, 1, 0.5, 1);
  --sh-ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
  --sh-duration-fast: 150ms;
  --sh-duration-normal: 250ms;
  --sh-duration-slow: 400ms;

  /* ── 布局 ── */
  --sh-header-height: 64px;
  --sh-sidebar-width: 220px;
  --sh-sidebar-collapsed: 64px;
}

html,
body,
#app {
  height: 100%;
  margin: 0;
  font-family: var(--sh-font-body);
  font-size: var(--sh-text-base);
  color: var(--sh-text-primary);
  background: var(--sh-bg-base);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* ── Element Plus 全局覆盖 ── */

/* 主色覆盖 */
:root {
  --el-color-primary: var(--sh-primary);
  --el-color-primary-light-3: var(--sh-primary-light);
  --el-color-primary-light-5: rgba(45, 106, 122, 0.5);
  --el-color-primary-light-7: var(--sh-primary-lighter);
  --el-color-primary-light-8: rgba(45, 106, 122, 0.2);
  --el-color-primary-light-9: rgba(45, 106, 122, 0.1);
  --el-color-primary-dark-2: var(--sh-primary-dark);

  --el-color-success: var(--sh-success);
  --el-color-success-light-3: #6AB893;
  --el-color-success-light-5: rgba(61, 153, 112, 0.5);
  --el-color-success-light-7: var(--sh-success-light);
  --el-color-success-light-9: rgba(61, 153, 112, 0.1);

  --el-color-warning: var(--sh-warning);
  --el-color-warning-light-3: var(--sh-accent-light);
  --el-color-warning-light-5: rgba(212, 145, 92, 0.5);
  --el-color-warning-light-7: var(--sh-warning-light);
  --el-color-warning-light-9: rgba(212, 145, 92, 0.1);

  --el-color-danger: var(--sh-danger);
  --el-color-danger-light-3: #D46458;
  --el-color-danger-light-5: rgba(192, 57, 43, 0.5);
  --el-color-danger-light-7: var(--sh-danger-light);
  --el-color-danger-light-9: rgba(192, 57, 43, 0.1);

  --el-color-info: var(--sh-info);
  --el-color-info-light-3: #7DAEC2;
  --el-color-info-light-5: rgba(91, 143, 168, 0.5);
  --el-color-info-light-7: var(--sh-info-light);
  --el-color-info-light-9: rgba(91, 143, 168, 0.1);

  --el-text-color-primary: var(--sh-text-primary);
  --el-text-color-regular: var(--sh-text-regular);
  --el-text-color-secondary: var(--sh-text-secondary);
  --el-text-color-placeholder: var(--sh-text-placeholder);
  --el-text-color-disabled: var(--sh-text-disabled);

  --el-bg-color: var(--sh-bg-white);
  --el-bg-color-page: var(--sh-bg-base);
  --el-bg-color-overlay: var(--sh-bg-white);

  --el-border-color: var(--sh-border);
  --el-border-color-light: var(--sh-border-light);
  --el-border-color-lighter: #EDF1F4;
  --el-border-color-dark: var(--sh-border-dark);

  --el-border-radius-base: var(--sh-radius-md);
  --el-border-radius-small: var(--sh-radius-sm);
  --el-border-radius-round: 20px;

  --el-font-size-base: var(--sh-text-base);
  --el-font-size-small: var(--sh-text-sm);
  --el-font-size-extra-small: var(--sh-text-xs);

  --el-fill-color-light: var(--sh-bg-elevated);
  --el-fill-color-lighter: #F7F9FA;
  --el-fill-color-blank: var(--sh-bg-white);

  --el-box-shadow: var(--sh-shadow-md);
  --el-box-shadow-light: var(--sh-shadow-sm);
  --el-box-shadow-lighter: var(--sh-shadow-sm);
  --el-box-shadow-dark: var(--sh-shadow-lg);

  --el-transition-duration: var(--sh-duration-normal);
  --el-transition-duration-fast: var(--sh-duration-fast);
}

/* ── 全局页面容器 ── */
.page-container {
  padding: var(--sh-space-lg);
  min-height: 100%;
}

/* ── 卡片头部通用 ── */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--sh-text-xl);
  font-weight: 600;
  color: var(--sh-text-primary);
  letter-spacing: 0.02em;
}

/* ── 筛选栏 ── */
.filter-bar {
  margin-bottom: var(--sh-space-md);
  display: flex;
  gap: var(--sh-space-sm);
  align-items: center;
  flex-wrap: wrap;
}

/* ── 分页 ── */
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--sh-space-md);
}

/* ── Element Plus 卡片美化 ── */
.el-card {
  border-radius: var(--sh-radius-lg) !important;
  border: 1px solid var(--sh-border-light) !important;
  box-shadow: var(--sh-shadow-sm) !important;
  transition: box-shadow var(--sh-duration-normal) var(--sh-ease-out) !important;
}
.el-card:hover {
  box-shadow: var(--sh-shadow-md) !important;
}
.el-card__header {
  border-bottom: 1px solid var(--sh-border-light) !important;
  padding: var(--sh-space-md) var(--sh-space-lg) !important;
}
.el-card__body {
  padding: var(--sh-space-lg) !important;
}

/* ── 表格美化 ── */
.el-table {
  --el-table-border-color: var(--sh-border-light);
  --el-table-header-bg-color: var(--sh-bg-elevated);
  --el-table-row-hover-bg-color: var(--sh-primary-lighter);
  font-size: var(--sh-text-sm);
}
.el-table th.el-table__cell {
  font-weight: 600;
  color: var(--sh-text-secondary);
  font-size: var(--sh-text-xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

/* ── 按钮微调 ── */
.el-button--primary {
  --el-button-bg-color: var(--sh-primary);
  --el-button-border-color: var(--sh-primary);
  --el-button-hover-bg-color: var(--sh-primary-light);
  --el-button-hover-border-color: var(--sh-primary-light);
  --el-button-active-bg-color: var(--sh-primary-dark);
  --el-button-active-border-color: var(--sh-primary-dark);
  border-radius: var(--sh-radius-md);
  font-weight: 500;
  transition: all var(--sh-duration-fast) var(--sh-ease-out);
}
.el-button--primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(var(--sh-primary-rgb), 0.3);
}
.el-button--primary:active {
  transform: translateY(0);
}

.el-button--success {
  --el-button-bg-color: var(--sh-success);
  --el-button-border-color: var(--sh-success);
  border-radius: var(--sh-radius-md);
}
.el-button--warning {
  --el-button-bg-color: var(--sh-warning);
  --el-button-border-color: var(--sh-warning);
  border-radius: var(--sh-radius-md);
}
.el-button--danger {
  --el-button-bg-color: var(--sh-danger);
  --el-button-border-color: var(--sh-danger);
  border-radius: var(--sh-radius-md);
}

.el-button--default {
  border-radius: var(--sh-radius-md);
  border-color: var(--sh-border);
  color: var(--sh-text-regular);
}
.el-button--default:hover {
  color: var(--sh-primary);
  border-color: var(--sh-primary-light);
  background: var(--sh-primary-lighter);
}

/* ── Tag 美化 ── */
.el-tag {
  border-radius: var(--sh-radius-sm);
  font-weight: 500;
  letter-spacing: 0.02em;
}

/* ─ 输入框美化 ── */
.el-input__wrapper {
  border-radius: var(--sh-radius-md) !important;
  transition: box-shadow var(--sh-duration-fast) var(--sh-ease-out) !important;
}
.el-input__wrapper:hover {
  box-shadow: 0 0 0 1px var(--sh-border-dark) inset !important;
}
.el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px var(--sh-primary) inset, 0 0 0 3px rgba(var(--sh-primary-rgb), 0.12) !important;
}
/* 排除 el-input-number 内部的 input wrapper，防止数字区域被挤压 */
.el-input-number .el-input__wrapper {
  border-radius: 0 !important;
}
.el-input-number .el-input__wrapper:hover {
  box-shadow: none !important;
}
.el-input-number .el-input__wrapper.is-focus {
  box-shadow: none !important;
}

/* ── 对话框美化 ── */
.el-dialog {
  border-radius: var(--sh-radius-xl) !important;
  overflow: hidden;
}
.el-dialog__header {
  padding: var(--sh-space-lg) var(--sh-space-xl) var(--sh-space-md) !important;
  border-bottom: 1px solid var(--sh-border-light);
  margin-right: 0 !important;
}
.el-dialog__title {
  font-size: var(--sh-text-xl) !important;
  font-weight: 600 !important;
  color: var(--sh-text-primary) !important;
}
.el-dialog__body {
  padding: var(--sh-space-lg) var(--sh-space-xl) !important;
}
.el-dialog__footer {
  padding: var(--sh-space-md) var(--sh-space-xl) var(--sh-space-lg) !important;
  border-top: 1px solid var(--sh-border-light);
}

/* ── Tabs 美化 ── */
.el-tabs__item {
  font-weight: 500 !important;
  color: var(--sh-text-secondary) !important;
  transition: color var(--sh-duration-fast) var(--sh-ease-out) !important;
}
.el-tabs__item.is-active {
  color: var(--sh-primary) !important;
  font-weight: 600 !important;
}
.el-tabs__active-bar {
  background-color: var(--sh-primary) !important;
  height: 3px !important;
  border-radius: 2px !important;
}

/* ── 分页美化 ── */
.el-pagination {
  --el-pagination-button-bg-color: var(--sh-bg-white);
  --el-pagination-hover-color: var(--sh-primary);
}
.el-pager li.is-active {
  color: var(--sh-primary) !important;
  font-weight: 600;
}

/* ── Descriptions 美化 ── */
.el-descriptions__label {
  font-weight: 500 !important;
  color: var(--sh-text-secondary) !important;
  background: var(--sh-bg-elevated) !important;
}
.el-descriptions__content {
  color: var(--sh-text-primary) !important;
}

/* ── Dropdown 美化 ── */
.el-dropdown-menu {
  background: var(--sh-bg-white) !important;
  border-radius: var(--sh-radius-lg) !important;
  box-shadow: none !important;
  border: none !important;
  padding: 6px !important;
  min-width: 180px !important;
}
.el-popper.is-light {
  border: none !important;
  box-shadow: none !important;
  background: transparent !important;
}
.el-dropdown-menu__item {
  display: flex !important;
  align-items: center !important;
  gap: 6px !important;
  margin: 2px 0 !important;
  padding: 8px 16px !important;
  border-radius: var(--sh-radius-md) !important;
  color: var(--sh-text-regular) !important;
  font-size: var(--sh-text-sm);
  line-height: 1.5 !important;
  white-space: nowrap !important;
  transition: background-color var(--sh-duration-fast), color var(--sh-duration-fast) !important;
}
.el-dropdown-menu__item:not(.is-disabled):hover {
  background-color: var(--sh-primary-lighter) !important;
  color: var(--sh-primary) !important;
}
.el-dropdown-menu__item.is-disabled {
  cursor: default !important;
  opacity: 0.7 !important;
}

/* ── Select 美化 ── */
.el-select .el-select__wrapper {
  border-radius: var(--sh-radius-md) !important;
}
/* Select 下拉面板 */
.el-select-dropdown {
  border: none !important;
  border-radius: var(--sh-radius-lg) !important;
  box-shadow: var(--sh-shadow-md) !important;
  background-color: var(--sh-bg-white) !important;
}
.el-select-dropdown__item {
  font-size: var(--sh-text-sm);
  color: var(--sh-text-regular) !important;
  transition: all var(--sh-duration-fast) var(--sh-ease-out);
}
.el-select-dropdown__item:hover,
.el-select-dropdown__item.is-hovering {
  background-color: var(--sh-primary-lighter) !important;
  color: var(--sh-primary) !important;
}

/* ── Divider 美化 ── */
.el-divider__text {
  color: var(--sh-text-secondary) !important;
  font-weight: 500;
  font-size: var(--sh-text-sm);
}

/* ── Skeleton 美化 ── */
.el-skeleton__item {
  border-radius: var(--sh-radius-md) !important;
}

/* ── 全局滚动条 ── */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: var(--sh-border);
  border-radius: 3px;
}
::-webkit-scrollbar-thumb:hover {
  background: var(--sh-border-dark);
}

/* ── 页面标题 ── */
.page-title {
  margin: 0 0 var(--sh-space-lg);
  font-size: var(--sh-text-2xl);
  font-weight: 700;
  color: var(--sh-text-primary);
  letter-spacing: -0.01em;
}

/* ── Tab 内头部 ── */
.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--sh-space-md);
}
.tab-header h4 {
  margin: 0;
  font-size: var(--sh-text-md);
  font-weight: 600;
  color: var(--sh-text-primary);
}

/* ── 字数统计 ── */
.word-count {
  margin-top: var(--sh-space-xs);
  font-size: var(--sh-text-xs);
  color: var(--sh-text-placeholder);
}
.word-count.warning {
  color: var(--sh-warning);
}

/* ── 内容展示区 ── */
.statement-content,
.thought-content {
  padding: var(--sh-space-md) var(--sh-space-lg);
  background: var(--sh-bg-elevated);
  border-radius: var(--sh-radius-md);
  border-left: 3px solid var(--sh-primary);
  line-height: var(--sh-leading-relaxed);
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--sh-text-regular);
}

/* ── 统计栏 ── */
.stats-bar {
  display: flex;
  gap: var(--sh-space-2xl);
  margin-bottom: var(--sh-space-lg);
  padding: var(--sh-space-lg);
  background: var(--sh-bg-elevated);
  border-radius: var(--sh-radius-md);
  border: 1px solid var(--sh-border-light);
}

/* ── 操作栏 ── */
.action-bar {
  margin-bottom: var(--sh-space-md);
}

/* ── 表单提示 ── */
.form-tip,
.upload-tip {
  margin-top: var(--sh-space-xs);
  font-size: var(--sh-text-xs);
  color: var(--sh-text-placeholder);
}

/* ── 动画 ── */
@keyframes sh-fade-in {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes sh-slide-up {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

.page-container {
  animation: sh-fade-in var(--sh-duration-slow) var(--sh-ease-out);
}

/* ── 减弱动画偏好 ── */
@media (prefers-reduced-motion: reduce) {
  .page-container {
    animation: none;
  }
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
</style>

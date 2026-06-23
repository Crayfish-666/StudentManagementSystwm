/**
 * StudentHub 前端 ESLint 配置
 * 依据: docs/02_ADR.md §3.4.2 Vue/TS 代码风格
 * 规则集: vue3-essential (基础规则, 误报少, 便于快速收敛历史问题)
 *         后续可升级到 vue3-recommended
 */
module.exports = {
  root: true,
  env: {
    browser: true,
    es2022: true,
    node: true
  },
  parser: 'vue-eslint-parser',
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
    extraFileExtensions: ['.vue']
  },
  extends: [
    'eslint:recommended',
    'plugin:vue/vue3-essential'
  ],
  rules: {
    // ADR §3.4.2 通用约束
    'no-console': 'off', // 允许 console, 业务侧按需
    'no-debugger': 'warn',
    'no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    'vue/multi-word-component-names': 'off', // 单页面组件命名放宽
    'vue/no-v-html': 'warn' // ADR §3.11 安全规范, 警告级别
  },
  ignorePatterns: [
    'dist/**',
    'node_modules/**',
    'coverage/**',
    '*.config.js',
    '*.config.cjs'
  ]
}

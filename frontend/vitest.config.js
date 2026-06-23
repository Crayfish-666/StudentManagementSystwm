// =============================================================================
// StudentHub 前端 vitest 配置
// 依据: docs/02_ADR.md §3.9 测试规范
// 目标覆盖率: composables / 关键组件 >= 60%
// =============================================================================
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  test: {
    globals: true,
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{js,mjs,vue}'],
    // 当前仓库尚无业务测试, 允许空跑通过, 后续随测试补齐逐步收紧
    passWithNoTests: true,
    // 覆盖率阈值 (与 ADR §3.9 对齐)
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov', 'html'],
      include: ['src/**/*.{js,vue}'],
      exclude: [
        'src/**/*.{test,spec}.{js,vue}',
        'src/main.js',
        'src/router/**',
        'src/api/**',
        'src/views/**'
      ],
      thresholds: {
        // V1 阶段保留宽松阈值, 后续逐步收紧至 60%
        lines: 0,
        statements: 0,
        functions: 0,
        branches: 0
      }
    }
  }
})

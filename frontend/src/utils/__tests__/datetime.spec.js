// =============================================================================
// StudentHub 前端最小测试样例
// 覆盖: src/utils/datetime.js 时区格式化纯函数
// 目的: 验证 vitest 链路通畅 + 保护核心 utils
// =============================================================================
import { describe, it, expect } from 'vitest'
import { formatDate, formatDateTime, formatDateTimeMinute } from '@/utils/datetime'

describe('utils/datetime', () => {
  it('formatDateTime 应正确格式化带时区的 ISO 字符串', () => {
    // 2026-06-15 18:27:02 +08:00
    expect(formatDateTime('2026-06-15T18:27:02+08:00')).toBe('2026-06-15 18:27:02')
  })

  it('formatDateTime 应正确解析无时区的 ISO 字符串 (按北京时间)', () => {
    // 字符串无时区, 按北京时间解析
    expect(formatDateTime('2026-06-15 18:27:02')).toBe('2026-06-15 18:27:02')
  })

  it('formatDate 应只保留日期部分', () => {
    expect(formatDate('2026-06-15T18:27:02+08:00')).toBe('2026-06-15')
  })

  it('formatDateTimeMinute 应省略秒', () => {
    expect(formatDateTimeMinute('2026-06-15T18:27:02+08:00')).toBe('2026-06-15 18:27')
  })

  it('formatDateTime 对非法输入应返回 fallback', () => {
    expect(formatDateTime('not-a-date', '--')).toBe('--')
    expect(formatDateTime(null, '')).toBe('')
  })
})

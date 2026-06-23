// 统一的北京时间（Asia/Shanghai, UTC+8）格式化工具
// 后端返回的时间字符串通常带 +08:00 或为 UTC，本模块将其按北京时间显示，避免浏览器本地时区差异。

const BEIJING_OFFSET_MS = 8 * 60 * 60 * 1000

/**
 * 解析任意时间值为 Date 对象
 * 兼容：
 * - ISO 字符串（带或不带时区）
 * - 时间戳（秒 / 毫秒）
 * - Date 对象
 * 对于不带时区的 ISO 字符串（如 "2026-06-15 18:27:02"），按北京时间解析。
 */
function toDate(input) {
  if (input == null || input === '') return null
  if (input instanceof Date) {
    return isNaN(input.getTime()) ? null : input
  }
  if (typeof input === 'number') {
    // 兼容秒级时间戳
    const ms = input < 1e12 ? input * 1000 : input
    const d = new Date(ms)
    return isNaN(d.getTime()) ? null : d
  }
  if (typeof input === 'string') {
    const s = input.trim()
    // 检测是否带时区信息（Z 或 +HH:MM / -HH:MM）
    const hasTZ = /(?:Z|[+-]\d{2}:?\d{2})$/.test(s)
    if (hasTZ) {
      const d = new Date(s)
      return isNaN(d.getTime()) ? null : d
    }
    // 无时区：按北京时间解析
    // 接受 "YYYY-MM-DD HH:mm:ss" 或 "YYYY-MM-DDTHH:mm:ss" 等格式
    const m = s.match(/^(\d{4})-(\d{2})-(\d{2})(?:[T ](\d{2}):(\d{2})(?::(\d{2}))?(?:\.\d+)?)?$/)
    if (m) {
      const [, y, mo, da, h = '00', mi = '00', se = '00'] = m
      // 直接构造对应的 UTC 毫秒（北京时间 = UTC + 8h）
      const utcMs = Date.UTC(+y, +mo - 1, +da, +h, +mi, +se) - BEIJING_OFFSET_MS
      return new Date(utcMs)
    }
    const d = new Date(s)
    return isNaN(d.getTime()) ? null : d
  }
  return null
}

/**
 * 将 Date 转换为「北京时间」对应的各字段
 */
function getBeijingParts(date) {
  // 通过偏移把 UTC 时刻转成「北京时间挂在 UTC 字段上」的 Date
  const shifted = new Date(date.getTime() + BEIJING_OFFSET_MS)
  return {
    year: shifted.getUTCFullYear(),
    month: shifted.getUTCMonth() + 1,
    day: shifted.getUTCDate(),
    hour: shifted.getUTCHours(),
    minute: shifted.getUTCMinutes(),
    second: shifted.getUTCSeconds()
  }
}

function pad2(n) {
  return n < 10 ? '0' + n : '' + n
}

/**
 * 格式化为北京时间日期时间字符串
 * @param {*} input 任意时间输入
 * @param {string} fallback 输入无效时的兜底值
 * @returns 形如 "2026-06-15 18:27:02"
 */
export function formatDateTime(input, fallback = '') {
  const d = toDate(input)
  if (!d) return fallback
  const p = getBeijingParts(d)
  return `${p.year}-${pad2(p.month)}-${pad2(p.day)} ${pad2(p.hour)}:${pad2(p.minute)}:${pad2(p.second)}`
}

/**
 * 格式化为北京时间日期字符串
 * @returns 形如 "2026-06-15"
 */
export function formatDate(input, fallback = '') {
  const d = toDate(input)
  if (!d) return fallback
  const p = getBeijingParts(d)
  return `${p.year}-${pad2(p.month)}-${pad2(p.day)}`
}

/**
 * 格式化为北京时间「年-月-日 时:分」（无秒）
 */
export function formatDateTimeMinute(input, fallback = '') {
  const d = toDate(input)
  if (!d) return fallback
  const p = getBeijingParts(d)
  return `${p.year}-${pad2(p.month)}-${pad2(p.day)} ${pad2(p.hour)}:${pad2(p.minute)}`
}

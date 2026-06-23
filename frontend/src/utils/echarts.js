// ECharts 共享工具：按需引入、实例创建、Resize 监听、销毁。
import * as echarts from 'echarts/core'
import {
  // 图表类型
  BarChart,
  LineChart,
  PieChart,
  RadarChart
} from 'echarts/charts'
import {
  // 组件
  GridComponent,
  TooltipComponent,
  TitleComponent,
  LegendComponent,
  DataZoomComponent
} from 'echarts/components'
import {
  CanvasRenderer
} from 'echarts/renderers'

// 注册用到的模块
echarts.use([
  BarChart,
  LineChart,
  PieChart,
  RadarChart,
  GridComponent,
  TooltipComponent,
  TitleComponent,
  LegendComponent,
  DataZoomComponent,
  CanvasRenderer
])

/**
 * 在指定 DOM 元素上初始化 ECharts 实例。
 * @param {HTMLElement} el
 * @param {object} option
 * @returns {echarts.ECharts}
 */
export function createChart(el, option) {
  if (!el) return null
  const instance = echarts.init(el, null, { renderer: 'canvas' })
  instance.setOption(option)
  return instance
}

/**
 * 监听窗口尺寸变化并自动 resize。
 * @param {echarts.ECharts} instance
 * @returns {Function} 取消监听
 */
export function bindResize(instance) {
  if (!instance) return () => {}
  const handler = () => instance.resize()
  window.addEventListener('resize', handler)
  return () => window.removeEventListener('resize', handler)
}

/**
 * 安全销毁 ECharts 实例。
 * @param {echarts.ECharts} instance
 */
export function disposeChart(instance) {
  if (instance && !instance.isDisposed()) {
    instance.dispose()
  }
}

export default echarts

import http from './http'

// 定时任务 API
export const jobApi = {
  // 手动触发任务
  runJob(name) {
    return http.post(`/sys/jobs/${name}/run`)
  },
  // 查询任务执行记录
  listRuns(params) {
    return http.get('/sys/jobs/runs', { params })
  }
}

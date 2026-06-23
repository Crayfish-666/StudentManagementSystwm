package scheduler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"student-system/internal/models"
	"student-system/pkg/response"
)

// JobHandler 定时任务 HTTP 接口处理器。
type JobHandler struct {
	sched *Scheduler
	db    *gorm.DB
}

// NewJobHandler 创建任务处理器。
func NewJobHandler(sched *Scheduler, db *gorm.DB) *JobHandler {
	return &JobHandler{sched: sched, db: db}
}

// RunJob 手动触发指定任务。POST /api/v1/sys/jobs/:name/run
func (h *JobHandler) RunJob(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Fail(c, 40001, "缺少任务名称")
		return
	}

	jobRun, err := h.sched.RunJobManual(name)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}

	response.OK(c, gin.H{
		"job_run_id":  jobRun.ID,
		"status":      jobRun.Status,
		"elapsed_ms":  jobRun.DurationMs,
	})
}

// ListRuns 查询任务执行记录（分页）。GET /api/v1/sys/jobs/runs
func (h *JobHandler) ListRuns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	jobName := c.Query("job_name")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&models.JobRun{}).Where("1 = 1")
	if jobName != "" {
		query = query.Where("job_name = ?", jobName)
	}

	var total int64
	query.Count(&total)

	var runs []models.JobRun
	if err := query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&runs).Error; err != nil {
		response.Fail(c, 1500, "查询任务执行记录失败")
		return
	}

	response.OK(c, gin.H{
		"items": runs,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 注册任务路由。
func (h *JobHandler) RegisterRoutes(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	jobs := rg.Group("/sys/jobs")
	{
		jobs.POST("/:name/run", adminOnly, h.RunJob)
		jobs.GET("/runs", adminOnly, h.ListRuns)
	}
}

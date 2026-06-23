package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/ty/service"
	"student-system/pkg/response"
)

// PoliticalReviewHandler 政审记录接口处理器。
type PoliticalReviewHandler struct {
	svc *service.PoliticalReviewService
}

// NewPoliticalReviewHandler 创建政审记录处理器。
func NewPoliticalReviewHandler(svc *service.PoliticalReviewService) *PoliticalReviewHandler {
	return &PoliticalReviewHandler{svc: svc}
}

// Create 创建政审记录。POST /api/v1/ty/political-reviews
func (h *PoliticalReviewHandler) Create(c *gin.Context) {
	userID, name, _, _, _ := actorFromCtx(c)

	var req service.CreatePoliticalReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	result, err := h.svc.Create(userID, &req, name, "", "", "")
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "发展对象不存在":
			code = 1404
		default:
			if contains(msg, "尚未完成审批流程") {
				code = 40901
			} else if contains(msg, "无效的审查对象关系") || contains(msg, "无效的审查方式") || contains(msg, "无效的结论") {
				code = 40001
			}
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// List 政审记录列表。GET /api/v1/ty/political-reviews
func (h *PoliticalReviewHandler) List(c *gin.Context) {
	var developmentID int64
	if v := c.Query("development_id"); v != "" {
		developmentID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if developmentID > 0 {
		// 按发展对象 ID 查询
		result, err := h.svc.ListByDevelopmentID(developmentID)
		if err != nil {
			response.Fail(c, 1500, "查询政审记录失败")
			return
		}
		response.OK(c, gin.H{"items": result, "total": int64(len(result)), "page": page, "page_size": pageSize})
		return
	}

	// 无 development_id 时返回空列表（前端一般会传此参数）
	response.OK(c, gin.H{"items": []interface{}{}, "total": int64(0), "page": page, "page_size": pageSize})
}

// Get 政审记录详情。GET /api/v1/ty/political-reviews/:id
func (h *PoliticalReviewHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的政审记录 ID")
		return
	}

	result, err := h.svc.GetByID(id)
	if err != nil {
		response.Fail(c, 1404, "政审记录不存在")
		return
	}
	response.OK(c, result)
}

// Summary 政审汇总结果。GET /api/v1/ty/political-reviews/summary
func (h *PoliticalReviewHandler) Summary(c *gin.Context) {
	developmentID, err := strconv.ParseInt(c.Query("development_id"), 10, 64)
	if err != nil || developmentID == 0 {
		response.Fail(c, 40002, "缺少 development_id 参数")
		return
	}

	result, err := h.svc.ProcessConclusion(developmentID)
	if err != nil {
		response.Fail(c, 1500, "获取政审汇总失败: "+err.Error())
		return
	}
	response.OK(c, result)
}

// RegisterRoutes 注册政审记录相关路由。
func (h *PoliticalReviewHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	ty := rg.Group("/ty")
	{
		ty.GET("/political-reviews", h.List)
		ty.GET("/political-reviews/:id", h.Get)
		ty.GET("/political-reviews/summary", h.Summary)
		ty.POST("/political-reviews", h.Create)
	}
}

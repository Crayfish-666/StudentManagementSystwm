package api

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/ty/service"
	"student-system/pkg/response"
)

// RecommendationHandler 推优大会接口处理器。
type RecommendationHandler struct {
	svc *service.RecommendationService
}

// NewRecommendationHandler 创建推优大会处理器。
func NewRecommendationHandler(svc *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{svc: svc}
}

// Create 创建推优大会。POST /api/v1/ty/recommendation-meetings
func (h *RecommendationHandler) Create(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	var req service.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	result, err := h.svc.Create(userID, &req)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "入团申请不存在":
			code = 1404
		case "仅 S3 状态的申请可召开推优大会":
			code = 40901
		default:
			if contains(msg, "到会人数不足") {
				code = 2501
			} else if contains(msg, "必须上传会议") {
				code = 2502
			} else if contains(msg, "赞成票数未超过") {
				code = 2503
			} else if contains(msg, "3 个月内已进行过推优") {
				code = 2504
			}
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, result)
}

// List 推优大会列表。GET /api/v1/ty/recommendation-meetings
func (h *RecommendationHandler) List(c *gin.Context) {
	var branchID int64
	if v := c.Query("branch_id"); v != "" {
		branchID, _ = strconv.ParseInt(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(branchID, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询推优大会列表失败")
		return
	}
	response.OK(c, result)
}

// Get 推优大会详情。GET /api/v1/ty/recommendation-meetings/:id
func (h *RecommendationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的推优大会 ID")
		return
	}

	result, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, 1404, "推优大会不存在")
		return
	}
	response.OK(c, result)
}

// GetByApplication 按申请ID查询推优大会。GET /api/v1/ty/recommendation-meetings/application/:applicationId
func (h *RecommendationHandler) GetByApplication(c *gin.Context) {
	applicationID, err := strconv.ParseInt(c.Param("applicationId"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的申请 ID")
		return
	}

	result, err := h.svc.GetByApplication(applicationID)
	if err != nil {
		response.Fail(c, 1404, err.Error())
		return
	}
	response.OK(c, result)
}

// RegisterRoutes 注册推优大会相关路由。
func (h *RecommendationHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	ty := rg.Group("/ty")
	{
		ty.GET("/recommendation-meetings", h.List)
		ty.GET("/recommendation-meetings/:id", h.Get)
		ty.GET("/recommendation-meetings/application/:applicationId", h.GetByApplication)
		ty.POST("/recommendation-meetings", h.Create)
	}
}

// contains 判断字符串是否包含子串（用于错误码匹配）。
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

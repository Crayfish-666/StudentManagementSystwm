package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"student-system/internal/models"
	"student-system/pkg/cachex"
	"student-system/pkg/response"
)

// DictHandler 字典接口处理器。
type DictHandler struct {
	db    *gorm.DB
	cache *cachex.Cache
}

// NewDictHandler 创建字典处理器。
func NewDictHandler(db *gorm.DB, cache *cachex.Cache) *DictHandler {
	return &DictHandler{db: db, cache: cache}
}

// ListItems 按分类查询字典项。GET /api/v1/sys/dicts/:category/items
func (h *DictHandler) ListItems(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		response.Fail(c, 40001, "缺少字典分类")
		return
	}

	// 尝试从缓存读取
	cacheKey := "sys:dict:" + category
	if cached, ok := h.cache.Get(cacheKey); ok {
		response.OK(c, cached)
		return
	}

	var items []models.SysDict
	if err := h.db.Where("category = ? AND is_deleted = 0 AND is_active = 1", category).
		Order("sort ASC, id ASC").
		Find(&items).Error; err != nil {
		response.Fail(c, 1500, "查询字典失败")
		return
	}

	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"code":    item.Code,
			"name_zh": item.NameZh,
			"name_en": item.NameEn,
			"sort":    item.Sort,
		})
	}

	// 写入缓存
	h.cache.Set(cacheKey, result)
	response.OK(c, gin.H{"items": result})
}

// ListCategories 列出所有字典分类。GET /api/v1/sys/dicts
func (h *DictHandler) ListCategories(c *gin.Context) {
	var categories []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	h.db.Model(&models.SysDict{}).
		Select("category, count(*) as count").
		Where("is_deleted = 0").
		Group("category").
		Find(&categories)

	// 同时返回每个分类下的条目
	type catDetail struct {
		Category string        `json:"category"`
		Count    int64         `json:"count"`
		Items    []models.SysDict `json:"items"`
	}

	var result []catDetail
	for _, cat := range categories {
		var items []models.SysDict
		h.db.Where("category = ? AND is_deleted = 0", cat.Category).
			Order("sort ASC, id ASC").
			Find(&items)
		result = append(result, catDetail{
			Category: cat.Category,
			Count:    cat.Count,
			Items:    items,
		})
	}

	response.OK(c, gin.H{"categories": result})
}

// CreateItem 新增字典项。POST /api/v1/sys/dicts/items
func (h *DictHandler) CreateItem(c *gin.Context) {
	var req struct {
		Category  string `json:"category" binding:"required"`
		Code      string `json:"code" binding:"required"`
		NameZh    string `json:"name_zh" binding:"required"`
		NameEn    string `json:"name_en"`
		Sort      int    `json:"sort"`
		ExtraJSON string `json:"extra_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整")
		return
	}

	item := models.SysDict{
		Category:  req.Category,
		Code:      req.Code,
		NameZh:    req.NameZh,
		NameEn:    req.NameEn,
		Sort:      req.Sort,
		ExtraJSON: req.ExtraJSON,
		IsActive:  1,
	}

	if err := h.db.Create(&item).Error; err != nil {
		response.Fail(c, 1409, "字典项已存在或创建失败")
		return
	}

	// 失效缓存
	h.cache.InvalidatePrefix("sys:dict:")
	response.OK(c, item)
}

// UpdateItem 修改字典项。PUT /api/v1/sys/dicts/items/:id
func (h *DictHandler) UpdateItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的 ID")
		return
	}

	var item models.SysDict
	if err := h.db.Where("id = ? AND is_deleted = 0", id).First(&item).Error; err != nil {
		response.Fail(c, 1404, "字典项不存在")
		return
	}

	var req struct {
		Code      *string `json:"code"`
		NameZh    *string `json:"name_zh"`
		NameEn    *string `json:"name_en"`
		Sort      *int    `json:"sort"`
		IsActive  *int    `json:"is_active"`
		ExtraJSON *string `json:"extra_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.NameZh != nil {
		updates["name_zh"] = *req.NameZh
	}
	if req.NameEn != nil {
		updates["name_en"] = *req.NameEn
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ExtraJSON != nil {
		updates["extra_json"] = *req.ExtraJSON
	}

	if len(updates) > 0 {
		h.db.Model(&item).Updates(updates)
	}

	h.cache.InvalidatePrefix("sys:dict:")
	h.db.Where("id = ? AND is_deleted = 0", id).First(&item)
	response.OK(c, item)
}

// DeleteItem 删除字典项。DELETE /api/v1/sys/dicts/items/:id
func (h *DictHandler) DeleteItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的 ID")
		return
	}

	if err := h.db.Model(&models.SysDict{}).Where("id = ?", id).
		Update("is_deleted", 1).Error; err != nil {
		response.Fail(c, 1500, "删除失败")
		return
	}

	h.cache.InvalidatePrefix("sys:dict:")
	response.OK(c, gin.H{"id": id})
}

// RegisterRoutes 注册字典路由。
func (h *DictHandler) RegisterRoutes(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	dicts := rg.Group("/sys/dicts")
	{
		// 公开：按分类查询字典项
		dicts.GET("/:category/items", h.ListItems)
		// 管理类接口：仅 R-SY-ADMIN
		dicts.GET("", adminOnly, h.ListCategories)
		dicts.POST("/items", adminOnly, h.CreateItem)
		dicts.PUT("/items/:id", adminOnly, h.UpdateItem)
		dicts.DELETE("/items/:id", adminOnly, h.DeleteItem)
	}
}

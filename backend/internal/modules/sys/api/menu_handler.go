package api

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"student-system/internal/models"
	"student-system/pkg/cachex"
	"student-system/pkg/response"
)

// MenuHandler 菜单接口处理器。
type MenuHandler struct {
	db    *gorm.DB
	cache *cachex.Cache
}

// NewMenuHandler 创建菜单处理器。
func NewMenuHandler(db *gorm.DB, cache *cachex.Cache) *MenuHandler {
	return &MenuHandler{db: db, cache: cache}
}

// menuItemVO 菜单项视图对象。
type menuItemVO struct {
	ID       int64         `json:"id"`
	Code     string        `json:"code"`
	Title    string        `json:"title"`
	Icon     string        `json:"icon"`
	Path     string        `json:"path"`
	Component string       `json:"component"`
	Sort     int           `json:"sort"`
	Children []*menuItemVO `json:"children,omitempty"`
}

// Mine 获取当前用户可见菜单树。GET /api/v1/sys/menus/mine
func (h *MenuHandler) Mine(c *gin.Context) {
	userRoles, _ := c.Get("user_roles")
	roles, _ := userRoles.([]string)

	// 尝试从缓存读取
	cacheKey := "sys:menu:mine:" + rolesKey(roles)
	if cached, ok := h.cache.Get(cacheKey); ok {
		response.OK(c, gin.H{"menus": cached})
		return
	}

	// 查询所有可见菜单
	var menus []models.SysMenu
	if err := h.db.Where("is_deleted = 0 AND visible = 1").
		Order("sort ASC, id ASC").
		Find(&menus).Error; err != nil {
		response.Fail(c, 1500, "查询菜单失败")
		return
	}

	// 按角色过滤
	filtered := filterMenusByRoles(menus, roles)

	// 构建树
	tree := buildMenuTree(filtered, nil)

	// 写入缓存
	h.cache.Set(cacheKey, tree)
	response.OK(c, gin.H{"menus": tree})
}

// filterMenusByRoles 根据角色过滤菜单。
// Roles 字段为 JSON 数组，空数组表示所有角色可见。
func filterMenusByRoles(menus []models.SysMenu, userRoles []string) []models.SysMenu {
	roleSet := make(map[string]struct{}, len(userRoles))
	for _, r := range userRoles {
		roleSet[r] = struct{}{}
	}

	var result []models.SysMenu
	for _, m := range menus {
		if isMenuVisible(m.Roles, roleSet) {
			result = append(result, m)
		}
	}
	return result
}

// isMenuVisible 判断菜单对当前角色是否可见。
func isMenuVisible(rolesJSON string, userRoles map[string]struct{}) bool {
	if rolesJSON == "" || rolesJSON == "[]" || rolesJSON == "null" {
		return true // 无角色限制，所有用户可见
	}
	var allowedRoles []string
	if err := json.Unmarshal([]byte(rolesJSON), &allowedRoles); err != nil {
		return true // 解析失败默认可见
	}
	for _, r := range allowedRoles {
		if _, ok := userRoles[r]; ok {
			return true
		}
	}
	return false
}

// buildMenuTree 构建菜单树。
func buildMenuTree(menus []models.SysMenu, parentID *int64) []*menuItemVO {
	var tree []*menuItemVO
	for _, m := range menus {
		if eqParentID(m.ParentID, parentID) {
			node := &menuItemVO{
				ID:        m.ID,
				Code:      m.Code,
				Title:     m.Title,
				Icon:      m.Icon,
				Path:      m.Path,
				Component: m.Component,
				Sort:      m.Sort,
			}
			node.Children = buildMenuTree(menus, &m.ID)
			tree = append(tree, node)
		}
	}
	return tree
}

// eqParentID 判断 parentID 是否相等（处理 nil）。
func eqParentID(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// rolesKey 生成角色列表的缓存 key。
func rolesKey(roles []string) string {
	b, _ := json.Marshal(roles)
	return string(b)
}

// RegisterRoutes 注册菜单路由。
func (h *MenuHandler) RegisterRoutes(rg *gin.RouterGroup) {
	menus := rg.Group("/sys/menus")
	{
		menus.GET("/mine", h.Mine)
	}
}

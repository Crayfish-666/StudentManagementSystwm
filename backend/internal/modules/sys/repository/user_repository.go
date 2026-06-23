// Package repository 数据访问层：用户/角色管理（SYS 子域）。
//
// 与 auth 子域的 user_repository 严格分离：本仓库仅服务于"系统管理 → 用户管理"
// 的 CRUD / 角色分配 / 状态切换 / 重置密码等场景。认证登录请使用 auth 子域。
package repository

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// UserRepository 用户管理仓库。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户管理仓库。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ListFilter 列表查询过滤条件。
type ListFilter struct {
	Keyword  string // 匹配 username / display_name / staff_no
	Status   string // active / locked / disabled / 空=全部
	RoleCode string // 角色 code 过滤（可选）
	Page     int
	PageSize int
}

// ListPage 分页查询用户（带角色聚合）。
func (r *UserRepository) ListPage(f ListFilter) ([]UserWithRoles, int64, error) {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 200 {
		f.PageSize = 20
	}

	q := r.db.Model(&models.SysUser{}).Where("sys_user.is_deleted = 0")
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("sys_user.username LIKE ? OR sys_user.display_name LIKE ? OR sys_user.staff_no LIKE ?", like, like, like)
	}
	if f.Status != "" {
		q = q.Where("sys_user.status = ?", f.Status)
	}

	// 角色过滤：先查出满足角色的 user_id 集合
	if f.RoleCode != "" {
		var uidRows []int64
		if err := r.db.Table("sys_user_role ur").
			Select("ur.user_id").
			Joins("JOIN sys_role r ON r.id = ur.role_id AND r.is_deleted = 0").
			Joins("JOIN sys_user u ON u.id = ur.user_id AND u.is_deleted = 0").
			Where("r.code = ? AND ur.is_deleted = 0", f.RoleCode).
			Pluck("ur.user_id", &uidRows).Error; err != nil {
			return nil, 0, err
		}
		if len(uidRows) == 0 {
			return nil, 0, nil
		}
		q = q.Where("sys_user.id IN ?", uidRows)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.SysUser
	if err := q.Order("sys_user.id ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 聚合角色
	result := make([]UserWithRoles, 0, len(users))
	if len(users) > 0 {
		ids := make([]int64, 0, len(users))
		for _, u := range users {
			ids = append(ids, u.ID)
		}
		roleMap, err := r.findRolesByUserIDs(ids)
		if err != nil {
			return nil, 0, err
		}
		for _, u := range users {
			result = append(result, UserWithRoles{User: u, Roles: roleMap[u.ID]})
		}
	}
	return result, total, nil
}

// GetByID 根据 ID 获取用户（含角色）。
func (r *UserRepository) GetByID(id int64) (*UserWithRoles, error) {
	var u models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&u).Error; err != nil {
		return nil, err
	}
	roleMap, err := r.findRolesByUserIDs([]int64{id})
	if err != nil {
		return nil, err
	}
	return &UserWithRoles{User: u, Roles: roleMap[id]}, nil
}

// FindByUsername 用户名校验（创建/编辑时唯一性检查）。
func (r *UserRepository) FindByUsername(username string) (*models.SysUser, error) {
	var u models.SysUser
	err := r.db.Where("username = ? AND is_deleted = 0", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Create 创建用户。
func (r *UserRepository) Create(u *models.SysUser) error {
	return r.db.Create(u).Error
}

// UpdateBasic 更新用户基本信息（不含 username / password）。
func (r *UserRepository) UpdateBasic(id int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.Model(&models.SysUser{}).Where("id = ? AND is_deleted = 0", id).Updates(updates).Error
}

// UpdatePassword 更新密码。
func (r *UserRepository) UpdatePassword(id int64, passwordHash string) error {
	return r.db.Model(&models.SysUser{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(map[string]interface{}{
			"password_hash": passwordHash,
			"updated_at":    time.Now(),
		}).Error
}

// SetStatus 设置用户状态（active / locked / disabled）。
func (r *UserRepository) SetStatus(id int64, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if status == "active" {
		updates["failed_attempts"] = 0
		updates["lock_until"] = nil
	}
	return r.db.Model(&models.SysUser{}).Where("id = ? AND is_deleted = 0", id).Updates(updates).Error
}

// SoftDelete 软删用户。
func (r *UserRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.SysUser{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// ListRoles 列出所有角色。
func (r *UserRepository) ListRoles() ([]models.SysRole, error) {
	var roles []models.SysRole
	if err := r.db.Where("is_deleted = 0").Order("scope ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoleByID 根据 ID 获取角色。
func (r *UserRepository) GetRoleByID(id int64) (*models.SysRole, error) {
	var role models.SysRole
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// ReplaceUserRoles 事务：撤销原角色 → 授予新角色。
// 角色 ID 列表为空表示清空。
func (r *UserRepository) ReplaceUserRoles(userID int64, roleIDs []int64, grantedBy int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 软删旧关联
		if err := tx.Model(&models.SysUserRole{}).
			Where("user_id = ? AND is_deleted = 0", userID).
			Updates(map[string]interface{}{"is_deleted": 1, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		// 授予新角色
		now := time.Now()
		for _, rid := range roleIDs {
			ur := &models.SysUserRole{
				UserID:    userID,
				RoleID:    rid,
				GrantedAt: now,
				GrantedBy: &grantedBy,
				IsDeleted: 0,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.Create(ur).Error; err != nil {
				// 唯一约束冲突：用户已有该角色（被软删后又插入），跳过即可
				if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
					continue
				}
				return err
			}
		}
		// 更新用户 updated_at
		if err := tx.Model(&models.SysUser{}).
			Where("id = ?", userID).
			Update("updated_at", now).Error; err != nil {
			return err
		}
		return nil
	})
}

// RemoveUserRole 撤销单个角色。
func (r *UserRepository) RemoveUserRole(userID, roleID int64) error {
	res := r.db.Model(&models.SysUserRole{}).
		Where("user_id = ? AND role_id = ? AND is_deleted = 0", userID, roleID).
		Updates(map[string]interface{}{"is_deleted": 1, "updated_at": time.Now()})
	return res.Error
}

// findRolesByUserIDs 批量查询用户的角色。
func (r *UserRepository) findRolesByUserIDs(userIDs []int64) (map[int64][]models.SysRole, error) {
	result := make(map[int64][]models.SysRole, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	// 第一步：取出 (user_id, role_id) 关联
	type relRow struct {
		UserID int64
		RoleID int64
	}
	var rels []relRow
	if err := r.db.Table("sys_user_role").
		Select("user_id, role_id").
		Where("user_id IN ? AND is_deleted = 0", userIDs).
		Scan(&rels).Error; err != nil {
		return nil, err
	}
	if len(rels) == 0 {
		return result, nil
	}
	// 第二步：批量取角色
	roleIDSet := make(map[int64]struct{}, len(rels))
	for _, rel := range rels {
		roleIDSet[rel.RoleID] = struct{}{}
	}
	roleIDs := make([]int64, 0, len(roleIDSet))
	for id := range roleIDSet {
		roleIDs = append(roleIDs, id)
	}
	var roles []models.SysRole
	if err := r.db.Where("id IN ? AND is_deleted = 0", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}
	roleByID := make(map[int64]models.SysRole, len(roles))
	for _, role := range roles {
		roleByID[role.ID] = role
	}
	// 第三步：组装
	for _, rel := range rels {
		role, ok := roleByID[rel.RoleID]
		if !ok {
			continue
		}
		result[rel.UserID] = append(result[rel.UserID], role)
	}
	return result, nil
}

// UserWithRoles 用户及其角色列表。
type UserWithRoles struct {
	User  models.SysUser
	Roles []models.SysRole
}

// IsNotFound 判断是否为未找到错误。
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// UserRepository 用户数据访问层。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername 根据用户名查找用户（未删除）。
func (r *UserRepository) FindByUsername(username string) (*models.SysUser, error) {
	var user models.SysUser
	err := r.db.Where("username = ? AND is_deleted = 0", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByLoginIdentifier 登录标识查询：支持 username / staff_no / student_no。
//
// 主路径：username 精确匹配（新用户 username 即为工号或学号）。
// 兼容路径：staff_no / student_no JOIN 查询，兜底老数据。
func (r *UserRepository) FindByLoginIdentifier(identifier string) (*models.SysUser, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, gorm.ErrRecordNotFound
	}

	// 1. username（主路径：新用户 username = 工号/学号）
	var user models.SysUser
	err := r.db.Where("username = ? AND is_deleted = 0", identifier).First(&user).Error
	if err == nil {
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. staff_no（兜底：老数据 username ≠ staff_no 的情况）
	err = r.db.Where("staff_no = ? AND staff_no <> '' AND is_deleted = 0", identifier).First(&user).Error
	if err == nil {
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 3. student_no JOIN idx_student（兜底：老数据 username ≠ 学号的情况）
	err = r.db.Table("sys_user").
		Select("sys_user.*").
		Joins("LEFT JOIN idx_student ON idx_student.id = sys_user.student_id AND idx_student.is_deleted = 0").
		Where("idx_student.student_no = ? AND sys_user.is_deleted = 0", identifier).
		First(&user).Error
	if err == nil {
		return &user, nil
	}
	return nil, err
}

// FindByID 根据 ID 查找用户。
func (r *UserRepository) FindByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateLoginSuccess 更新登录成功状态。
func (r *UserRepository) UpdateLoginSuccess(id int64) error {
	now := time.Now()
	return r.db.Model(&models.SysUser{}).Where("id = ?", id).Updates(map[string]interface{}{
		"failed_attempts": 0,
		"lock_until":      nil,
		"last_login_at":   now,
	}).Error
}

// UpdateLoginFail 更新登录失败次数，超过阈值则锁定。
func (r *UserRepository) UpdateLoginFail(id int64, maxAttempts int, lockDuration time.Duration) error {
	var user models.SysUser
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return err
	}

	attempts := user.FailedAttempts + 1
	updates := map[string]interface{}{
		"failed_attempts": attempts,
	}

	// 超过最大尝试次数则锁定账户
	if attempts >= maxAttempts {
		lockUntil := time.Now().Add(lockDuration)
		updates["lock_until"] = lockUntil
		updates["status"] = "locked"
	}

	return r.db.Model(&models.SysUser{}).Where("id = ?", id).Updates(updates).Error
}

// FindUserRoles 查询用户角色列表。
func (r *UserRepository) FindUserRoles(userID int64) ([]models.SysRole, error) {
	var roles []models.SysRole
	err := r.db.Table("sys_role").
		Select("sys_role.*").
		Joins("JOIN sys_user_role ON sys_user_role.role_id = sys_role.id").
		Where("sys_user_role.user_id = ? AND sys_user_role.is_deleted = 0 AND sys_role.is_deleted = 0", userID).
		Find(&roles).Error
	return roles, err
}

// CreateUser 创建用户。
func (r *UserRepository) CreateUser(user *models.SysUser) error {
	return r.db.Create(user).Error
}

// CreateRole 创建角色。
func (r *UserRepository) CreateRole(role *models.SysRole) error {
	return r.db.Create(role).Error
}

// CreateUserRole 创建用户-角色关联。
func (r *UserRepository) CreateUserRole(ur *models.SysUserRole) error {
	return r.db.Create(ur).Error
}

// FindRoleByCode 根据角色编码查找角色。
func (r *UserRepository) FindRoleByCode(code string) (*models.SysRole, error) {
	var role models.SysRole
	err := r.db.Where("code = ? AND is_deleted = 0", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// UserCount 返回用户总数。
func (r *UserRepository) UserCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.SysUser{}).Where("is_deleted = 0").Count(&count).Error
	return count, err
}

// LookupTokenVersion 轻量查询用户的 token_version（中间件用）。
// 仅查一列，避免引入完整对象开销。
func (r *UserRepository) LookupTokenVersion(id int64) (int, error) {
	var v int
	err := r.db.Model(&models.SysUser{}).
		Where("id = ? AND is_deleted = 0", id).
		Select("token_version").
		Scan(&v).Error
	return v, err
}

// UpdatePasswordAndBumpTokenVersion 改密并 token_version+1（ADR-005 决策细化）。
// 一次事务，保证旧 RT 全部失效的原子性。
func (r *UserRepository) UpdatePasswordAndBumpTokenVersion(id int64, newPasswordHash string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 读出当前 token_version（行锁），避免并发改密导致 +1 丢失
		var current int
		if err := tx.Model(&models.SysUser{}).
			Where("id = ? AND is_deleted = 0", id).
			Select("token_version").
			Scan(&current).Error; err != nil {
			return fmt.Errorf("read token_version: %w", err)
		}
		return tx.Model(&models.SysUser{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"password_hash": newPasswordHash,
				"token_version": current + 1,
			}).Error
	})
}

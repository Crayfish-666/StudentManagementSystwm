// Package service 业务逻辑层：用户/角色管理（SYS 子域）。
package service

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"student-system/internal/models"
	sysrepo "student-system/internal/modules/sys/repository"
)

// 业务错误码（与 SRD §13.1 + 现有前端提示对齐）。
const (
	ErrUserNotFound  = 1404
	ErrUserExisted   = 1409
	ErrInvalidParam  = 40001
	ErrInvalidStatus = 14002
	ErrNoSuchRole    = 14003
	ErrCannotOpSelf  = 14004
)

// UserService 用户管理服务。
type UserService struct {
	repo *sysrepo.UserRepository
}

// NewUserService 创建服务。
func NewUserService(repo *sysrepo.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUserRequest 创建用户请求。
type CreateUserRequest struct {
	Username    string  `json:"username" binding:"required"`
	Password    string  `json:"password" binding:"required,min=6"`
	DisplayName string  `json:"display_name" binding:"required"`
	StaffNo     string  `json:"staff_no"`
	StudentID   *int64  `json:"student_id"`
	AvatarURL   string  `json:"avatar_url"`
	Status      string  `json:"status"` // 默认 active
	RoleIDs     []int64 `json:"role_ids"`
}

// UpdateUserRequest 更新用户请求。
type UpdateUserRequest struct {
	DisplayName *string `json:"display_name"`
	StaffNo     *string `json:"staff_no"`
	StudentID   *int64  `json:"student_id"`
	AvatarURL   *string `json:"avatar_url"`
}

// ListParams 列表查询参数。
type ListParams struct {
	Keyword  string
	Status   string
	RoleCode string
	Page     int
	PageSize int
}

// UserView 用户视图（含角色）。
type UserView struct {
	ID            int64        `json:"id"`
	Username      string       `json:"username"`
	DisplayName   string       `json:"display_name"`
	StaffNo       string       `json:"staff_no"`
	StudentID     *int64       `json:"student_id,omitempty"`
	AvatarURL     string       `json:"avatar_url"`
	Status        string       `json:"status"`
	LastLoginAt   *string      `json:"last_login_at,omitempty"`
	FailedAttempts int         `json:"failed_attempts"`
	Roles         []RoleView   `json:"roles"`
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
}

// RoleView 角色视图。
type RoleView struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
}

// ListResult 列表结果。
type ListResult struct {
	Items    []UserView `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// List 查询用户列表。
func (s *UserService) List(p ListParams) (*ListResult, error) {
	users, total, err := s.repo.ListPage(sysrepo.ListFilter{
		Keyword:  p.Keyword,
		Status:   p.Status,
		RoleCode: p.RoleCode,
		Page:     p.Page,
		PageSize: p.PageSize,
	})
	if err != nil {
		return nil, err
	}
	page := p.Page
	if page <= 0 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	items := make([]UserView, 0, len(users))
	for _, ur := range users {
		items = append(items, toUserView(ur.User, ur.Roles))
	}
	return &ListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// Get 查询用户详情。
func (s *UserService) Get(id int64) (*UserView, error) {
	ur, err := s.repo.GetByID(id)
	if err != nil {
		if sysrepo.IsNotFound(err) {
			return nil, bizError(ErrUserNotFound, "用户不存在")
		}
		return nil, err
	}
	v := toUserView(ur.User, ur.Roles)
	return &v, nil
}

// Create 创建用户。
func (s *UserService) Create(req *CreateUserRequest, operatorID int64) (*UserView, error) {
	// 1. 字段校验
	if strings.TrimSpace(req.Username) == "" {
		return nil, bizError(ErrInvalidParam, "用户名不能为空")
	}
	if len(req.Password) < 6 {
		return nil, bizError(ErrInvalidParam, "密码长度至少 6 位")
	}
	if strings.TrimSpace(req.DisplayName) == "" {
		return nil, bizError(ErrInvalidParam, "姓名不能为空")
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	if !isValidUserStatus(status) {
		return nil, bizError(ErrInvalidStatus, "用户状态不合法")
	}

	// 2. 用户名唯一性
	if existing, err := s.repo.FindByUsername(req.Username); err == nil && existing != nil {
		return nil, bizError(ErrUserExisted, "用户名已存在")
	} else if err != nil && !sysrepo.IsNotFound(err) {
		return nil, err
	}

	// 3. 角色校验
	roleIDs, err := s.validateRoleIDs(req.RoleIDs)
	if err != nil {
		return nil, err
	}

	// 4. 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 5. 写库（事务：用户 + 角色）
	user := &models.SysUser{
		Username:     strings.TrimSpace(req.Username),
		PasswordHash: string(hash),
		StaffNo:      strings.TrimSpace(req.StaffNo),
		DisplayName:  strings.TrimSpace(req.DisplayName),
		AvatarURL:    req.AvatarURL,
		StudentID:    req.StudentID,
		Status:       status,
		IsDeleted:    0,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 6. 分配角色
	if len(roleIDs) > 0 {
		if err := s.repo.ReplaceUserRoles(user.ID, roleIDs, operatorID); err != nil {
			return nil, fmt.Errorf("分配角色失败: %w", err)
		}
	}

	// 7. 重新查询返回
	return s.Get(user.ID)
}

// Update 更新用户基本信息。
func (s *UserService) Update(id int64, req *UpdateUserRequest, operatorID int64) (*UserView, error) {
	// 操作人不可修改自己状态（可保留，更新姓名/工号是允许的；这里只限制 status 操作）
	updates := map[string]interface{}{}
	if req.DisplayName != nil {
		name := strings.TrimSpace(*req.DisplayName)
		if name == "" {
			return nil, bizError(ErrInvalidParam, "姓名不能为空")
		}
		updates["display_name"] = name
	}
	if req.StaffNo != nil {
		updates["staff_no"] = strings.TrimSpace(*req.StaffNo)
	}
	if req.StudentID != nil {
		updates["student_id"] = *req.StudentID
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}

	if len(updates) == 0 {
		return s.Get(id)
	}

	if _, err := s.repo.GetByID(id); err != nil {
		if sysrepo.IsNotFound(err) {
			return nil, bizError(ErrUserNotFound, "用户不存在")
		}
		return nil, err
	}
	if err := s.repo.UpdateBasic(id, updates); err != nil {
		return nil, err
	}
	_ = operatorID // 当前未使用，预留审计
	return s.Get(id)
}

// ResetPassword 重置用户密码。
func (s *UserService) ResetPassword(id int64, newPassword string, operatorID int64) error {
	if len(newPassword) < 6 {
		return bizError(ErrInvalidParam, "新密码长度至少 6 位")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		if sysrepo.IsNotFound(err) {
			return bizError(ErrUserNotFound, "用户不存在")
		}
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	if err := s.repo.UpdatePassword(id, string(hash)); err != nil {
		return err
	}
	_ = operatorID
	return nil
}

// Lock 锁定用户。
func (s *UserService) Lock(id int64, operatorID int64) error {
	if id == operatorID {
		return bizError(ErrCannotOpSelf, "不能锁定自己的账户")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		if sysrepo.IsNotFound(err) {
			return bizError(ErrUserNotFound, "用户不存在")
		}
		return err
	}
	return s.repo.SetStatus(id, "locked")
}

// Unlock 解锁用户。
func (s *UserService) Unlock(id int64, operatorID int64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if sysrepo.IsNotFound(err) {
			return bizError(ErrUserNotFound, "用户不存在")
		}
		return err
	}
	_ = operatorID
	return s.repo.SetStatus(id, "active")
}

// Disable 禁用用户。
func (s *UserService) Disable(id int64, operatorID int64) error {
	if id == operatorID {
		return bizError(ErrCannotOpSelf, "不能禁用自己的账户")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		if sysrepo.IsNotFound(err) {
			return bizError(ErrUserNotFound, "用户不存在")
		}
		return err
	}
	return s.repo.SetStatus(id, "disabled")
}

// Enable 启用用户。
func (s *UserService) Enable(id int64, operatorID int64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		if sysrepo.IsNotFound(err) {
			return bizError(ErrUserNotFound, "用户不存在")
		}
		return err
	}
	_ = operatorID
	return s.repo.SetStatus(id, "active")
}

// Delete 软删用户。
func (s *UserService) Delete(id int64, operatorID int64) error {
	if id == operatorID {
		return bizError(ErrCannotOpSelf, "不能删除自己的账户")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		if sysrepo.IsNotFound(err) {
			return bizError(ErrUserNotFound, "用户不存在")
		}
		return err
	}
	return s.repo.SoftDelete(id)
}

// ListRoles 列出所有角色。
func (s *UserService) ListRoles() ([]RoleView, error) {
	roles, err := s.repo.ListRoles()
	if err != nil {
		return nil, err
	}
	out := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		out = append(out, toRoleView(r))
	}
	return out, nil
}

// AssignRoles 分配用户角色（覆盖式）。
func (s *UserService) AssignRoles(userID int64, roleIDs []int64, operatorID int64) (*UserView, error) {
	if _, err := s.repo.GetByID(userID); err != nil {
		if sysrepo.IsNotFound(err) {
			return nil, bizError(ErrUserNotFound, "用户不存在")
		}
		return nil, err
	}
	ids, err := s.validateRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceUserRoles(userID, ids, operatorID); err != nil {
		return nil, err
	}
	return s.Get(userID)
}

// RevokeRole 撤销单个角色。
func (s *UserService) RevokeRole(userID, roleID int64) (*UserView, error) {
	if _, err := s.repo.GetByID(userID); err != nil {
		if sysrepo.IsNotFound(err) {
			return nil, bizError(ErrUserNotFound, "用户不存在")
		}
		return nil, err
	}
	if _, err := s.repo.GetRoleByID(roleID); err != nil {
		if sysrepo.IsNotFound(err) {
			return nil, bizError(ErrNoSuchRole, "角色不存在")
		}
		return nil, err
	}
	if err := s.repo.RemoveUserRole(userID, roleID); err != nil {
		return nil, err
	}
	return s.Get(userID)
}

// validateRoleIDs 校验角色 ID 集合是否全部存在。
func (s *UserService) validateRoleIDs(ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	out := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		role, err := s.repo.GetRoleByID(id)
		if err != nil {
			if sysrepo.IsNotFound(err) {
				return nil, bizError(ErrNoSuchRole, fmt.Sprintf("角色 ID=%d 不存在", id))
			}
			return nil, err
		}
		out = append(out, role.ID)
	}
	return out, nil
}

// isValidUserStatus 校验用户状态。
func isValidUserStatus(s string) bool {
	switch s {
	case "active", "locked", "disabled":
		return true
	default:
		return false
	}
}

func toUserView(u models.SysUser, roles []models.SysRole) UserView {
	roleViews := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		roleViews = append(roleViews, toRoleView(r))
	}
	v := UserView{
		ID:             u.ID,
		Username:       u.Username,
		DisplayName:    u.DisplayName,
		StaffNo:        u.StaffNo,
		StudentID:      u.StudentID,
		AvatarURL:      u.AvatarURL,
		Status:         u.Status,
		FailedAttempts: u.FailedAttempts,
		Roles:          roleViews,
		CreatedAt:      u.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		UpdatedAt:      u.UpdatedAt.Format("2006-01-02T15:04:05+08:00"),
	}
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format("2006-01-02T15:04:05+08:00")
		v.LastLoginAt = &s
	}
	return v
}

func toRoleView(r models.SysRole) RoleView {
	return RoleView{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Scope:       r.Scope,
		Description: r.Description,
	}
}

// BizError 业务错误。
type BizError struct {
	Code    int
	Message string
}

func (e *BizError) Error() string { return e.Message }

func bizError(code int, msg string) *BizError {
	return &BizError{Code: code, Message: msg}
}

// AsBizError 提取 BizError。
func AsBizError(err error) (*BizError, bool) {
	var be *BizError
	if errors.As(err, &be) {
		return be, true
	}
	return nil, false
}

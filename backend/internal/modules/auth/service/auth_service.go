package service

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authjwt "student-system/internal/modules/auth/jwt"
	"student-system/internal/modules/auth/repository"
	"student-system/pkg/revokex"
)

// LoginRequest 登录请求 DTO。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 改密请求 DTO。
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UserView 用户视图（登录响应中返回）。
type UserView struct {
	ID                 int64      `json:"id"`
	Username           string     `json:"username"`
	DisplayName        string     `json:"display_name"`
	AvatarURL          string     `json:"avatar_url"`
	Roles              []RoleView `json:"roles"`
	StudentID          *int64     `json:"student_id,omitempty"`
	CollegeID          *int64     `json:"college_id,omitempty"`
	MustChangePassword bool       `json:"must_change_password"`
}

// RoleView 角色视图。
type RoleView struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

// AuthService 认证服务。
type AuthService struct {
	repo       *repository.UserRepository
	jwtManager *authjwt.JWTManager
	revokeStore revokex.Store
}

// NewAuthService 创建认证服务。
func NewAuthService(repo *repository.UserRepository, jwtManager *authjwt.JWTManager, store revokex.Store) *AuthService {
	return &AuthService{
		repo:       repo,
		jwtManager: jwtManager,
		revokeStore: store,
	}
}

const (
	maxLoginAttempts = 5
	lockDuration     = 15 * time.Minute
	bcryptCost       = 12
)

// Login 登录：校验密码 → 签发 Token 对。
//
// 登录标识支持：
//   - 系统管理员等保留用户名（username）
//   - 教师工号（staff_no）
//   - 学生学号（student_no，通过 idx_student 关联）
func (s *AuthService) Login(req *LoginRequest) (*authjwt.TokenPair, *UserView, error) {
	user, err := s.repo.FindByLoginIdentifier(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("账号或密码错误")
		}
		return nil, nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查账户状态
	if user.Status == "disabled" {
		return nil, nil, fmt.Errorf("账号已被禁用")
	}
	if user.Status == "locked" {
		if user.LockUntil != nil && time.Now().Before(*user.LockUntil) {
			remaining := time.Until(*user.LockUntil).Round(time.Minute)
			return nil, nil, fmt.Errorf("账号已被锁定，请 %v 后再试", remaining)
		}
		// 锁定时间已过，允许重新尝试
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// 密码错误，更新失败次数
		_ = s.repo.UpdateLoginFail(user.ID, maxLoginAttempts, lockDuration)
		return nil, nil, fmt.Errorf("账号或密码错误")
	}

	// 登录成功，更新状态
	if err := s.repo.UpdateLoginSuccess(user.ID); err != nil {
		return nil, nil, fmt.Errorf("更新登录状态失败: %w", err)
	}

	// 查询角色
	roles, err := s.repo.FindUserRoles(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("查询角色失败: %w", err)
	}

	roleCodes := make([]string, 0, len(roles))
	roleViews := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		roleCodes = append(roleCodes, r.Code)
		roleViews = append(roleViews, RoleView{Code: r.Code, Name: r.Name, Scope: r.Scope})
	}

	// 签发 Token（绑定当前 token_version）
	accessToken, err := s.jwtManager.GenerateAccess(user.ID, user.DisplayName, roleCodes, user.TokenVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("签发 access_token 失败: %w", err)
	}
	refreshToken, _, err := s.jwtManager.GenerateRefresh(user.ID, user.TokenVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("签发 refresh_token 失败: %w", err)
	}

	pair := &authjwt.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.jwtManager.AccessTTLSeconds(),
	}

	view := &UserView{
		ID:                 user.ID,
		Username:           user.Username,
		DisplayName:        user.DisplayName,
		AvatarURL:          user.AvatarURL,
		Roles:              roleViews,
		StudentID:          user.StudentID,
		MustChangePassword: false,
	}

	return pair, view, nil
}

// Refresh 刷新 Token：校验 jti 黑名单 + token_version + 账号状态 → 轮换签发新对。
func (s *AuthService) Refresh(refreshToken string) (*authjwt.TokenPair, error) {
	claims, exp, err := s.jwtManager.ParseRefresh(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh_token 无效或已过期")
	}

	// 1) jti 黑名单校验
	if s.revokeStore.IsRevoked(claims.ID) {
		return nil, fmt.Errorf("refresh_token 已被吊销")
	}

	// 2) 用户状态校验
	user, err := s.repo.FindByID(claims.UID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if user.Status != "active" {
		return nil, fmt.Errorf("账号状态异常")
	}

	// 3) token_version 校验（防改密 / 禁用后旧 RT 复用）
	if claims.TokenVersion != user.TokenVersion {
		return nil, fmt.Errorf("refresh_token 已被吊销")
	}

	// 4) 查询角色
	roles, err := s.repo.FindUserRoles(user.ID)
	if err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	roleCodes := make([]string, 0, len(roles))
	for _, r := range roles {
		roleCodes = append(roleCodes, r.Code)
	}

	// 5) 签发新 Token 对
	accessToken, err := s.jwtManager.GenerateAccess(user.ID, user.DisplayName, roleCodes, user.TokenVersion)
	if err != nil {
		return nil, fmt.Errorf("签发 access_token 失败: %w", err)
	}
	newRefreshToken, _, err := s.jwtManager.GenerateRefresh(user.ID, user.TokenVersion)
	if err != nil {
		return nil, fmt.Errorf("签发 refresh_token 失败: %w", err)
	}

	// 6) 旧 RT jti 进黑名单（轮换）
	s.revokeStore.Revoke(claims.ID, exp)

	return &authjwt.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.jwtManager.AccessTTLSeconds(),
	}, nil
}

// Logout 把当前 RT 的 jti 加入黑名单直到原 exp。
func (s *AuthService) Logout(refreshToken string) {
	if refreshToken == "" {
		return
	}
	claims, exp, err := s.jwtManager.ParseRefresh(refreshToken)
	if err != nil {
		// 解析失败说明已无效，无需吊销
		return
	}
	s.revokeStore.Revoke(claims.ID, exp)
}

// ChangePassword 改密：bcrypt 校验旧密码 → 更新 hash → token_version+1（使旧 RT 全部失效）。
func (s *AuthService) ChangePassword(uid int64, req *ChangePasswordRequest) error {
	// 1) 强度校验
	if err := validatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	// 2) 读用户
	user, err := s.repo.FindByID(uid)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 3) 校验旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("旧密码错误")
	}

	// 4) 新密码 hash
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}

	// 5) 事务更新（行锁避免并发 +1 丢失）
	if err := s.repo.UpdatePasswordAndBumpTokenVersion(uid, string(newHash)); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}
	return nil
}

// GetCurrentUser 获取当前用户信息。
func (s *AuthService) GetCurrentUser(uid int64) (*UserView, error) {
	user, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	roles, err := s.repo.FindUserRoles(user.ID)
	if err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	roleViews := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		roleViews = append(roleViews, RoleView{Code: r.Code, Name: r.Name, Scope: r.Scope})
	}

	return &UserView{
		ID:                 user.ID,
		Username:           user.Username,
		DisplayName:        user.DisplayName,
		AvatarURL:          user.AvatarURL,
		Roles:              roleViews,
		StudentID:          user.StudentID,
		MustChangePassword: false,
	}, nil
}

// HashPassword 对密码进行 bcrypt 加密。
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

// validatePasswordStrength 密码强度校验：≥8 位、含字母 + 数字。
// 强度规则如需扩展，调用方请先同步 docs/04_SRD_api_specifications.md §5.1.5。
func validatePasswordStrength(pwd string) error {
	if len(pwd) < 8 {
		return fmt.Errorf("新密码长度不能少于 8 位")
	}
	hasLetter, hasDigit := false, false
	for _, r := range pwd {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return fmt.Errorf("新密码必须同时包含字母和数字")
	}
	return nil
}

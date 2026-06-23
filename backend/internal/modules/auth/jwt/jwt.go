// Package jwt 提供 JWT Token 的签发与解析（ADR-005）。
package jwt

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT Claims，与 ADR-005 Token Claims 对齐。
type Claims struct {
	jwtv5.RegisteredClaims
	UID          int64    `json:"uid"`
	Name         string   `json:"name"`
	Roles        []string `json:"roles"`
	TokenVersion int      `json:"tv"` // 与 sys_user.token_version 绑定（ADR-005 决策细化）
}

// TokenPair 登录成功后返回的令牌对（refresh_token 仅在 cookie 模式下不返回 body）。
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // 秒
}

// JWTManager 管理 JWT 签发与解析。
type JWTManager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTManager 创建 JWT 管理器。
func NewJWTManager(secret, issuer string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateAccess 签发 Access Token，tokenVersion 与当前 sys_user.token_version 绑定。
func (m *JWTManager) GenerateAccess(uid int64, name string, roles []string, tokenVersion int) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", uid),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			ID:        fmt.Sprintf("access-%d-%d", uid, now.UnixNano()),
		},
		UID:          uid,
		Name:         name,
		Roles:        roles,
		TokenVersion: tokenVersion,
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateRefresh 签发 Refresh Token，tokenVersion 与当前 sys_user.token_version 绑定。
func (m *JWTManager) GenerateRefresh(uid int64, tokenVersion int) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.refreshTTL)
	claims := Claims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", uid),
			ExpiresAt: jwtv5.NewNumericDate(exp),
			IssuedAt:  jwtv5.NewNumericDate(now),
			ID:        fmt.Sprintf("refresh-%d-%d", uid, now.UnixNano()),
		},
		UID:          uid,
		TokenVersion: tokenVersion,
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// ParseAccess 解析 Access Token 并返回 Claims。
func (m *JWTManager) ParseAccess(tokenStr string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtv5.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// ParseRefresh 解析 Refresh Token。返回 (claims, 过期时间, error)。
func (m *JWTManager) ParseRefresh(tokenStr string) (*Claims, time.Time, error) {
	claims, err := m.ParseAccess(tokenStr)
	if err != nil {
		return nil, time.Time{}, err
	}
	if claims.ExpiresAt == nil {
		return nil, time.Time{}, fmt.Errorf("refresh token missing exp")
	}
	return claims, claims.ExpiresAt.Time, nil
}

// AccessTTLSeconds 返回 Access Token 有效期（秒）。
func (m *JWTManager) AccessTTLSeconds() int64 {
	return int64(m.accessTTL.Seconds())
}

// Package revokex 提供 Refresh Token jti 黑名单（ADR-005 决策细化）。
//
// 存储：进程内 LRU，复用 pkg/cachex，TTL 与 RT 自身 exp 对齐（≤ 7d）。
// 服务重启后黑名单清空，配合 sys_user.token_version 自增可保证安全窗口可控。
package revokex

import (
	"time"

	"student-system/pkg/cachex"
)

// Store 黑名单存储抽象（便于后续替换为 Redis / 持久化方案）。
type Store interface {
	// Revoke 标记 jti 为已吊销，exp 为该 jti 原有的过期时间（早于该时间的判断已无意义）。
	Revoke(jti string, exp time.Time)
	// IsRevoked 查询 jti 是否已被吊销（且尚未到原 exp）。
	IsRevoked(jti string) bool
}

// lruStore 基于 cachex.Cache 的进程内实现。
type lruStore struct {
	cache *cachex.Cache
}

// NewLRU 创建默认 4096 条目、TTL 7 天的黑名单（最大同时在线用户数 × 设备数）。
func NewLRU() Store {
	return &lruStore{cache: cachex.New(4096, 7*24*time.Hour)}
}

// NewLRUWithTTL 自定义 TTL，便于测试。
func NewLRUWithTTL(size int, ttl time.Duration) Store {
	return &lruStore{cache: cachex.New(size, ttl)}
}

// Revoke 写入黑名单，TTL = max(0, exp - now)。
func (s *lruStore) Revoke(jti string, exp time.Time) {
	if jti == "" {
		return
	}
	ttl := time.Until(exp)
	if ttl <= 0 {
		// 已过期的 jti 无需记录
		return
	}
	s.cache.SetWithTTL("rt:"+jti, true, ttl)
}

// IsRevoked 查询 jti 是否在黑名单中。
func (s *lruStore) IsRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	_, ok := s.cache.Get("rt:" + jti)
	return ok
}

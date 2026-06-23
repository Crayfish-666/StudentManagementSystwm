// Package cachex 提供进程内 LRU 缓存（ADR-017）。
// V1 使用 hashicorp/golang-lru；TTL 默认 5 分钟；写操作时主动 Invalidate。
package cachex

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

// entry 缓存条目，带过期时间。
type entry struct {
	value     interface{}
	expiresAt time.Time
}

// Cache 进程内 LRU 缓存。
type Cache struct {
	mu   sync.RWMutex
	lru  *lru.Cache[string, entry]
	ttl  time.Duration
}

// New 创建 LRU 缓存，size 为最大条目数，ttl 为默认过期时间。
func New(size int, ttl time.Duration) *Cache {
	c, _ := lru.New[string, entry](size)
	return &Cache{lru: c, ttl: ttl}
}

// Set 写入缓存。
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL 写入缓存并指定过期时间。
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Add(key, entry{value: value, expiresAt: time.Now().Add(ttl)})
}

// Get 读取缓存，过期返回 nil,false。
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.lru.Get(key)
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		c.lru.Remove(key)
		return nil, false
	}
	return e.value, true
}

// Invalidate 使指定 key 失效。
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Remove(key)
}

// InvalidatePrefix 使指定前缀的所有 key 失效。
func (c *Cache) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := c.lru.Keys()
	for _, k := range keys {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			c.lru.Remove(k)
		}
	}
}

// Purge 清空所有缓存。
func (c *Cache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Purge()
}

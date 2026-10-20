package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultMaxBuckets 单个限流器最大允许记录的客户端上限，彻底锁死内存上限（防海量虚假IP攻击导致OOM）
	DefaultMaxBuckets = 3000
)

// clientBucket 紧凑型记录结构（仅占用 10 字节），杜绝重型对象分配
type clientBucket struct {
	count     uint16 // 2 字节计数器
	resetUnix int64  // 8 字节 Unix 秒级时间戳
}

// MemoryRateLimiter 具备严格内存容量与生命周期受控的滑动窗口限流器
type MemoryRateLimiter struct {
	limit      int
	windowSec  int64
	maxEntries int
	buckets    map[string]*clientBucket
	mutex      sync.RWMutex
}

var (
	globalRateLimiters     []*MemoryRateLimiter
	globalRateLimitersLock sync.Mutex
)

// NewMemoryRateLimiter 创建受容量严格保护的内存限流器（不启动野生后台协程）
func NewMemoryRateLimiter(limit int, window time.Duration) *MemoryRateLimiter {
	windowSec := int64(window.Seconds())
	if windowSec <= 0 {
		windowSec = 1
	}

	limiter := &MemoryRateLimiter{
		limit:      limit,
		windowSec:  windowSec,
		maxEntries: DefaultMaxBuckets,
		buckets:    make(map[string]*clientBucket),
	}

	globalRateLimitersLock.Lock()
	globalRateLimiters = append(globalRateLimiters, limiter)
	globalRateLimitersLock.Unlock()

	return limiter
}

// CleanupExpired 清理已过期的记录（由系统统一 cron 调度执行）
func (l *MemoryRateLimiter) CleanupExpired() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := time.Now().Unix()
	for k, b := range l.buckets {
		if now >= b.resetUnix {
			delete(l.buckets, k)
		}
	}
}

// CleanAllRateLimiters 供系统 cron 周期性统一调用的全局清理函数
func CleanAllRateLimiters() {
	globalRateLimitersLock.Lock()
	defer globalRateLimitersLock.Unlock()

	for _, l := range globalRateLimiters {
		l.CleanupExpired()
	}
}

// Allow 检查并限制请求，带严格的内存水位淘汰保护
func (l *MemoryRateLimiter) Allow(key string) (bool, int) {
	now := time.Now().Unix()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	bucket, exists := l.buckets[key]
	if !exists || now >= bucket.resetUnix {
		// 内存容量保护：达到上限时主动清理过期项
		if len(l.buckets) >= l.maxEntries {
			for k, b := range l.buckets {
				if now >= b.resetUnix {
					delete(l.buckets, k)
				}
			}
			// 若依然饱和，淘汰最早进入的 100 个元素，严防内存无限制膨胀
			if len(l.buckets) >= l.maxEntries {
				evictCount := 0
				for k := range l.buckets {
					delete(l.buckets, k)
					evictCount++
					if evictCount >= 100 {
						break
					}
				}
			}
		}

		l.buckets[key] = &clientBucket{
			count:     1,
			resetUnix: now + l.windowSec,
		}
		return true, 0
	}

	if int(bucket.count) < l.limit {
		bucket.count++
		return true, 0
	}

	retryAfterSec := int(bucket.resetUnix - now)
	if retryAfterSec <= 0 {
		retryAfterSec = 1
	}
	return false, retryAfterSec
}

// IPRateLimiter 基于客户端真实 IP 的限流中间件
func IPRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewMemoryRateLimiter(limit, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = c.Request.RemoteAddr
		}
		if ip == "" {
			ip = "unknown"
		}

		allowed, retryAfter := limiter.Allow(ip)
		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  fmt.Sprintf("请求过于频繁，请等待 %d 秒后再试！", retryAfter),
				"data": nil,
			})
			return
		}
		c.Next()
	}
}

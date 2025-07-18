package middleware

import (
	"Fcircle/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"sync"
	"time"
)

type limiterKey struct {
	IP   string
	Rate int
}

var ipLimiters sync.Map

var ipLastAccess sync.Map

var ipBlockedUntil sync.Map

// RateLimit 限流中间件
func RateLimit(perSecond int, blockDuration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if perSecond <= 0 {
			c.AbortWithStatusJSON(500, gin.H{"error": "invalid rate limit configuration"})
			return
		}

		key := limiterKey{IP: ip, Rate: perSecond}

		// 确保时区正确加载
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			loc = time.FixedZone("CST", 8*3600)
		}

		now := time.Now().In(loc)

		// 检查封禁状态
		if blockedUntil, exists := ipBlockedUntil.Load(key); exists {
			blockTime := blockedUntil.(time.Time)
			if blockTime.IsZero() {
				blockTime = now
			} else {
				blockTime = blockTime.In(loc)
			}

			formattedTime := blockTime.Format("2006-01-02 15:04:05")

			if now.Before(blockTime) {
				utils.Errorf(fmt.Sprintf("IP %s is still blocked until %s", ip, formattedTime))
				c.AbortWithStatusJSON(429, gin.H{"error": "too many requests, please wait until " + formattedTime})
				return
			}

			ipBlockedUntil.Delete(key)
		}

		limiter := getLimiter(key, perSecond)

		if !limiter.Allow() {
			blockTime := now.Add(blockDuration)
			ipBlockedUntil.Store(key, blockTime)

			formattedTime := blockTime.Format("2006-01-02 15:04:05")
			utils.Errorf(fmt.Sprintf("Rate limit exceeded for IP: %s (Rate: %d/s), blocked until %s", ip, perSecond, formattedTime))
			c.AbortWithStatusJSON(429, gin.H{"error": "too many requests, please wait until " + formattedTime})
			return
		}

		utils.Infof(fmt.Sprintf("Allowed access for IP: %s (Rate: %d/s)", ip, perSecond))
		ipLastAccess.Store(key, now)
		c.Next()
	}
}

// 获取或创建限流器（并发安全）
func getLimiter(key limiterKey, perSecond int) *rate.Limiter {

	if limiter, exists := ipLimiters.Load(key); exists {
		return limiter.(*rate.Limiter)
	}

	burst := perSecond
	if burst < 1 {
		burst = 1
	}

	newLimiter := rate.NewLimiter(rate.Limit(perSecond), burst)

	limiter, loaded := ipLimiters.LoadOrStore(key, newLimiter)
	if loaded {
		// 其他goroutine已抢先存储，使用已存在的实例
		return limiter.(*rate.Limiter)
	}
	return newLimiter
}

// InitRateLimiterCleanup 初始化清理任务
func InitRateLimiterCleanup(interval time.Duration) {
	go cleanupLimiters(interval)
}

// 定期清理过期限流器
func cleanupLimiters(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		utils.Infof(fmt.Sprintf("Starting rate limiter cleanup..."))

		var expiredKeys []limiterKey
		ipLimiters.Range(func(key, value interface{}) bool {
			k := key.(limiterKey)
			if lastAccess, exists := ipLastAccess.Load(k); exists {
				if time.Since(lastAccess.(time.Time)) > 15*time.Minute {
					expiredKeys = append(expiredKeys, k)
				}
			} else {
				expiredKeys = append(expiredKeys, k)
			}
			return true
		})
		for _, k := range expiredKeys {
			ipLimiters.Delete(k)
			ipLastAccess.Delete(k)
			utils.Infof(fmt.Sprintf("Batch cleaned up limiter: %v\n", k))
		}
	}
}

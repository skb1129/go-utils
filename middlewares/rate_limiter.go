package middlewares

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skb1129/go-utils/cache"
	"github.com/skb1129/go-utils/config"
	"github.com/skb1129/go-utils/logs"
	"github.com/skb1129/go-utils/request"
	"go.uber.org/zap"
)

func RateLimiter(cache *cache.Cache, redisKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logs.GetLogger()
		origin := c.GetHeader("Origin")
		ctx := context.Background()
		ip := c.ClientIP()
		key := fmt.Sprintf(redisKey, ip)

		configLimit := config.GetInt64("rate-limiter.limit")
		configDuration := time.Second * time.Duration(config.GetInt64("rate-limiter.duration"))
		count, err := cache.Incr(ctx, key, configDuration)
		if err != nil {
			logger.Error("rate limiter redis error", zap.Error(err), zap.String("ip", ip))
			c.Next()
			return
		}
		if count > configLimit {
			logger.Warn("rate limit exceeded",
				zap.String("ip", ip),
				zap.String("origin", origin),
				zap.String("path", c.Request.URL.Path),
				zap.Int64("count", count),
				zap.Int64("limit", configLimit),
			)
			request.SendServiceError(c, request.CreateTooManyRequestsError(nil, "Rate Limit Exceeded"))
			return
		}

		c.Next()
	}
}

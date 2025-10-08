package middleware

import (
	"astron-xmod-shim/pkg/log"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logging 基于zap的HTTP请求日志中间件
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 结束时间和处理时间
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 根据状态码选择日志级别
		logger := log.Info
		if statusCode >= 400 && statusCode < 500 {
			logger = log.Warn
		} else if statusCode >= 500 {
			logger = log.Error
		}

		// 记录请求信息
		logger("HTTP请求处理完成",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("duration", duration),
		)
	}
}

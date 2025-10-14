package server

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-hclog"
)

func HclogMiddleware(logger hclog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if strings.HasSuffix(c.Request.URL.Path, "json") {
			return
		}

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("request",
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"status", status,
			"latency", latency,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

package helper

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// GetIPAddress - get client IP address from request
func GetIPAddress(c *gin.Context) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to ClientIP (Gin's built-in method)
	clientIP := c.ClientIP()
	if clientIP != "" {
		return clientIP
	}

	return "unknown"
}

// GetUserAgent - get user agent from request
func GetUserAgent(c *gin.Context) string {
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		return "unknown"
	}
	return userAgent
}

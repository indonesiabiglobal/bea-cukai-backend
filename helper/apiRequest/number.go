package apiRequest

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ParseInt(c *gin.Context, key string, def int) int {
	if v := strings.TrimSpace(c.Query(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

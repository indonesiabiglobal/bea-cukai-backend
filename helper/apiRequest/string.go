package apiRequest

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func ParseString(c *gin.Context, key, def string) string {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		return def
	}
	return v
}

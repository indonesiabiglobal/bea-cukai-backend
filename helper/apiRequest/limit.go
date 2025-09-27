package apiRequest

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseLimit(ctx *gin.Context, def int) int {
	limStr := ctx.DefaultQuery("limit", strconv.Itoa(def))
	lim, err := strconv.Atoi(limStr)
	if err != nil || lim <= 0 {
		lim = def
	}
	if lim > 1000 {
		lim = 1000
	}
	return lim
}
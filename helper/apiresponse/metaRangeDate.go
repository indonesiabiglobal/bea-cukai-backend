package apiresponse

import (
	"time"

	"github.com/gin-gonic/gin"
)

func MetaRange(from, to time.Time) any {
	return gin.H{
		"from":     from.Format("2006-01-02"),
		"to":       to.Format("2006-01-02"),
		"timezone": "Asia/Jakarta",
	}
}
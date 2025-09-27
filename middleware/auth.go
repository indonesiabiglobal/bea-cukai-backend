package middleware

import (
	"Dashboard-TRDP/helper"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		userData, err := helper.VerifyToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid token",
				"error":   err.Error(),
			})
			return
		}

		fmt.Println("User, ", userData)

		c.Set("userData", userData)
		c.Next()
	}
}

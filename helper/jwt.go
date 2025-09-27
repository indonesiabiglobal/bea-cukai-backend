package helper

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var SECRETKEY string = GetEnv("SECRETKEY")

func GenerateToken(id uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"id":    id,
		"email": email,
		"exp":   time.Now().Add(time.Minute * 15).Unix(),
	}

	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	res, err := jwt.SignedString([]byte(SECRETKEY))

	return res, err
}

func VerifyToken(ctx *gin.Context) (interface{}, error) {
	err := errors.New("please login to get the token")
	auth := ctx.Request.Header.Get("Authorization")
	bearer := strings.HasPrefix(auth, "Bearer")

	if !bearer {
		return nil, err
	}

	tokenStr := strings.Split(auth, "Bearer ")[1]

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRETKEY), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

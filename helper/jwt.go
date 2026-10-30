package helper

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var SECRETKEY string = GetEnv("SECRETKEY")

func GenerateToken(id string, username string) (string, error) {
	if SECRETKEY == "" {
		return "", errors.New("server SECRETKEY is not configured")
	}
	claims := jwt.MapClaims{
		"id": id,
		"username": username,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	res, err := jwt.SignedString([]byte(SECRETKEY))

	return res, err
}

func VerifyToken(ctx *gin.Context) (interface{}, error) {
	// fail closed: secret kosong = siapa pun bisa tanda tangan token sendiri
	if SECRETKEY == "" {
		return nil, errors.New("server SECRETKEY is not configured")
	}

	tokenStr, ok := strings.CutPrefix(ctx.Request.Header.Get("Authorization"), "Bearer ")
	if !ok || tokenStr == "" {
		return nil, errors.New("please login to get the token")
	}

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

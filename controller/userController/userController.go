package userController

import (
	"Dashboard-TRDP/helper"
	"Dashboard-TRDP/model"
	"Dashboard-TRDP/service/userService"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type UserController struct {
	UserService *userService.UserService
}

func NewUserController(userService *userService.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

func (u *UserController) CreateUser(ctx *gin.Context) {
	var userRequest model.UserRequest
	err := ctx.Bind(&userRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail bind data",
			"error":   err.Error(),
		})
		return
	}

	// validate data user
	validator := helper.NewValidator()

	err = validator.Validate(userRequest)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	userResponse, err := u.UserService.CreateUser(userRequest)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "fail create user",
				"error":   err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail create user",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, userResponse)
}

func (u *UserController) LoginUser(ctx *gin.Context) {
	var userRequest model.UserLoginRequest
	err := ctx.Bind(&userRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail bind data",
			"error":   err.Error(),
		})
		return
	}

	// validate data user
	validator := helper.NewValidator()

	err = validator.Validate(userRequest)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	var token string
	token, err = u.UserService.LoginUser(userRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail login",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (u *UserController) UpdateUser(ctx *gin.Context) {
	// bind request data
	var userRequest model.UserUpdateRequest
	err := ctx.Bind(&userRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail bind data",
			"error":   err.Error(),
		})
		return
	}

	// validate data user
	validator := helper.NewValidator()
	err = validator.Validate(userRequest)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// get user data from token
	userData := ctx.MustGet("userData").(jwt.MapClaims)
	userID := uint(userData["id"].(float64))

	// call service to update user
	userResponse, err := u.UserService.UpdateUser(userRequest, userID)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "fail create user",
				"error":   err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail update user",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, userResponse)
}

func (u *UserController) DeleteUser(ctx *gin.Context) {
	userData := ctx.MustGet("userData").(jwt.MapClaims)
	userID := uint(userData["id"].(float64))

	err := u.UserService.DeleteUser(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "user not found",
				"error":   err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail delete user",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success delete user",
	})
}

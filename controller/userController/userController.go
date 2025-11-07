package userController

import (
	"Bea-Cukai/helper"
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/model"
	"Bea-Cukai/service/userService"
	"net/http"
	"time"

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

	// Get IP address and user agent
	ipAddress := helper.GetIPAddress(ctx)
	userAgent := helper.GetUserAgent(ctx)

	userResponse, err := u.UserService.CreateUser(userRequest, ipAddress, userAgent)
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

	// Get IP address and user agent
	ipAddress := helper.GetIPAddress(ctx)
	userAgent := helper.GetUserAgent(ctx)

	var token string
	token, err = u.UserService.LoginUser(userRequest, ipAddress, userAgent)
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
	id := userData["id"].(string)

	// Get IP address and user agent
	ipAddress := helper.GetIPAddress(ctx)
	userAgent := helper.GetUserAgent(ctx)

	// call service to update user
	userResponse, err := u.UserService.UpdateUser(userRequest, id, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "username already exists" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "fail update user",
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
	id := userData["id"].(string)

	// Get IP address and user agent
	ipAddress := helper.GetIPAddress(ctx)
	userAgent := helper.GetUserAgent(ctx)

	err := u.UserService.DeleteUser(id, ipAddress, userAgent)
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

// GetAll retrieves all users with optional filtering and pagination
// Query parameters:
// - id: filter by user ID (partial match)
// - username: filter by username (partial match)
// - level: filter by user level (exact match)
// - page: page number (default: 1)
// - limit: items per page (default: 10)
func (u *UserController) GetAll(ctx *gin.Context) {
	var req model.UserListRequest

	// Parse query parameters
	req.Id = ctx.Query("id")
	req.Username = ctx.Query("username")
	req.Level = ctx.Query("level")

	// Parse pagination parameters
	req.Page = apiRequest.ParseInt(ctx, "page", 1)
	req.Limit = apiRequest.ParseInt(ctx, "limit", 10)

	// Get data from service
	users, total, meta, err := u.UserService.GetAll(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get users",
			"error":   err.Error(),
		})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Users retrieved successfully",
		"data":    users,
		"meta":    meta,
		"total":   total,
	})
}

// GetProfile retrieves the current user's profile information
// Uses the user ID from the JWT token to get profile data
func (u *UserController) GetProfile(ctx *gin.Context) {
	// Get user data from JWT token
	userData := ctx.MustGet("userData").(jwt.MapClaims)
	id := userData["id"].(string)

	// Get profile from service
	profile, err := u.UserService.GetProfile(id)
	if err != nil {
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "User profile not found",
				"error":   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get user profile",
			"error":   err.Error(),
		})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User profile retrieved successfully",
		"data":    profile,
	})
}

// LogoutUser handles user logout
func (u *UserController) LogoutUser(ctx *gin.Context) {
	// Get user data from JWT token
	userData := ctx.MustGet("userData").(jwt.MapClaims)
	id := userData["id"].(string)
	username := userData["username"].(string)

	// Return successful logout response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
		"data": gin.H{
			"id":     id,
			"username":     username,
			"logged_out":  true,
			"logout_time": time.Now().Format(time.RFC3339),
			"note":        "Please remove the token from client storage",
		},
	})
}

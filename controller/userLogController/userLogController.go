package userLogController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/model"
	"Bea-Cukai/service/userLogService"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserLogController struct {
	UserLogService *userLogService.UserLogService
}

func NewUserLogController(userLogService *userLogService.UserLogService) *UserLogController {
	return &UserLogController{
		UserLogService: userLogService,
	}
}

// GetAll retrieves all user logs with optional filtering and pagination
// Query parameters:
// - user_id: filter by user ID
// - username: filter by username (partial match)
// - action: filter by action (login, logout, create, update, delete)
// - status: filter by status (success, failed)
// - start_date: filter by start date (format: 2006-01-02)
// - end_date: filter by end date (format: 2006-01-02)
// - page: page number (default: 1)
// - limit: items per page (default: 20)
func (c *UserLogController) GetAll(ctx *gin.Context) {
	var req model.UserLogListRequest

	// Parse query parameters
	req.UserId = ctx.Query("user_id")
	req.Username = ctx.Query("username")
	req.Action = ctx.Query("action")
	req.Status = ctx.Query("status")
	req.StartDate = ctx.Query("start_date")
	req.EndDate = ctx.Query("end_date")

	// Parse pagination parameters
	req.Page = apiRequest.ParseInt(ctx, "page", 1)
	req.Limit = apiRequest.ParseInt(ctx, "limit", 20)

	// Get data from service
	logs, total, meta, err := c.UserLogService.GetAll(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get user logs",
			"error":   err.Error(),
		})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User logs retrieved successfully",
		"data":    logs,
		"meta":    meta,
		"total":   total,
	})
}

// GetMyLogs retrieves logs for the current authenticated user
// Query parameters:
// - limit: number of logs to retrieve (default: 50)
func (c *UserLogController) GetMyLogs(ctx *gin.Context) {
	// Get user data from JWT token
	userData := ctx.MustGet("userData").(jwt.MapClaims)
	userId := userData["id"].(string)

	// Parse limit parameter
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	// Get logs from service
	logs, err := c.UserLogService.GetByUserId(userId, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get user logs",
			"error":   err.Error(),
		})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User logs retrieved successfully",
		"data":    logs,
		"total":   len(logs),
	})
}

// GetByUserId retrieves logs for a specific user (admin only)
// Path parameter:
// - user_id: the user ID to get logs for
// Query parameters:
// - limit: number of logs to retrieve (default: 50)
func (c *UserLogController) GetByUserId(ctx *gin.Context) {
	userId := ctx.Param("user_id")
	if userId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "User ID is required",
		})
		return
	}

	// Parse limit parameter
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	// Get logs from service
	logs, err := c.UserLogService.GetByUserId(userId, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get user logs",
			"error":   err.Error(),
		})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User logs retrieved successfully",
		"data":    logs,
		"total":   len(logs),
	})
}

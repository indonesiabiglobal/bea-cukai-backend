package userService

import (
	"Bea-Cukai/helper"
	"Bea-Cukai/model"
	"Bea-Cukai/repo/userLogRepository"
	"Bea-Cukai/repo/userRepository"
	"errors"
	"fmt"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo    *userRepository.UserRepository
	userLogRepo *userLogRepository.UserLogRepository
}

func NewUserService(userRepository *userRepository.UserRepository, userLogRepository *userLogRepository.UserLogRepository) *UserService {
	return &UserService{
		userRepo:    userRepository,
		userLogRepo: userLogRepository,
	}
}

// CreateUser implements UserService
func (u *UserService) CreateUser(userRequest model.UserRequest, ipAddress, userAgent string) (model.UserResponse, error) {
	// validate id
	_, err := u.userRepo.GetUserById(userRequest.Id)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound {
		// Log failed attempt
		u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    userRequest.Id,
			Username:  userRequest.Username,
			Action:    "create",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "failed",
			Message:   "ID already exists",
		})
		return model.UserResponse{}, errors.New("id already exists")
	}

	// validate username
	_, err = u.userRepo.GetUserByUsername(userRequest.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound {
		// Log failed attempt
		u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    userRequest.Id,
			Username:  userRequest.Username,
			Action:    "create",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "failed",
			Message:   "Username already exists",
		})
		return model.UserResponse{}, errors.New("username already exists")
	}

	// hash password
	hashedPassword, err := helper.HashPassword(userRequest.Password)
	userRequest.Password = hashedPassword
	if err != nil {
		return model.UserResponse{}, err
	}

	// call repository to save user
	createdUser, err := u.userRepo.CreateUser(userRequest)
	if err != nil {
		// Log failed attempt
		u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    userRequest.Id,
			Username:  userRequest.Username,
			Action:    "create",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "failed",
			Message:   err.Error(),
		})
		return model.UserResponse{}, err
	}

	// Log successful creation
	u.userLogRepo.CreateLog(model.UserLogRequest{
		UserId:    createdUser.Id,
		Username:  createdUser.Username,
		Action:    "create",
		IpAddress: ipAddress,
		UserAgent: userAgent,
		Status:    "success",
		Message:   "User created successfully",
	})

	var userResponse model.UserResponse
	err = copier.Copy(&userResponse, &createdUser)
	if err != nil {
		return model.UserResponse{}, err
	}

	return userResponse, nil
}

// LoginUser implements UserService
func (u *UserService) LoginUser(userLogin model.UserLoginRequest, ipAddress, userAgent string) (string, error) {
	// call repository to get user
	user, err := u.userRepo.LoginUser(userLogin)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Log failed login attempt
			_, logErr := u.userLogRepo.CreateLog(model.UserLogRequest{
				Username:  userLogin.Username,
				Action:    "login",
				IpAddress: ipAddress,
				UserAgent: userAgent,
				Status:    "failed",
				Message:   "User not found",
			})
			if logErr != nil {
				// Print error for debugging (in production, use proper logging)
				// fmt.Printf("Error logging failed login: %v\n", logErr)
			}
			return "", errors.New("Username or Password is incorrect")
		}
		return "", err
	}

	// verify password hash
	match := helper.CheckPasswordHash(userLogin.Password, user.Password)
	if !match {
		// Log failed login attempt
		_, logErr := u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    user.Id,
			Username:  user.Username,
			Action:    "login",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "failed",
			Message:   "Invalid password",
		})
		if logErr != nil {
			// Print error for debugging
			// fmt.Printf("Error logging failed login: %v\n", logErr)
		}
		return "", errors.New("Username or Password is incorrect")
	}

	// Generate token
	token, err := helper.GenerateToken(user.Id, user.Username)
	if err != nil {
		return "", err
	}

	// Update login info (count, last login time, last login IP)
	err = u.userRepo.UpdateLoginInfo(user.Id, ipAddress)
	if err != nil {
		// Log but don't fail the login
		_, logErr := u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    user.Id,
			Username:  user.Username,
			Action:    "login",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "warning",
			Message:   "Failed to update login info: " + err.Error(),
		})
		if logErr != nil {
			// Print error for debugging
			// fmt.Printf("Error creating warning log: %v\n", logErr)
		}
	}

	// Log successful login
	_, logErr := u.userLogRepo.CreateLog(model.UserLogRequest{
		UserId:    user.Id,
		Username:  user.Username,
		Action:    "login",
		IpAddress: ipAddress,
		UserAgent: userAgent,
		Status:    "success",
		Message:   "Login successful",
	})
	fmt.Println(logErr)
	fmt.Println("masuk")
	if logErr != nil {
		// Don't fail login if logging fails, but we should know about it
		// In production, use proper logging framework
		// fmt.Printf("Error logging successful login: %v\n", logErr)
	}

	return token, nil
}

// update user
func (u *UserService) UpdateUser(userRequest model.UserUpdateRequest, id string, ipAddress, userAgent string) (model.UserResponse, error) {
	// validate username
	user, err := u.userRepo.GetUserByUsername(userRequest.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound && user.Id != id {
		// Log failed attempt
		u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    id,
			Username:  userRequest.Username,
			Action:    "update",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "failed",
			Message:   "Username already exists",
		})
		return model.UserResponse{}, errors.New("Username already exists")
	}

	// hash password if provided
	if userRequest.Password != "" {
		hashedPassword, err := helper.HashPassword(userRequest.Password)
		if err != nil {
			return model.UserResponse{}, err
		}
		userRequest.Password = hashedPassword
	}

	// call repository to update user
	updatedUser, err := u.userRepo.UpdateUser(userRequest, id)
	if err != nil {
		// Log failed attempt
		u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    id,
			Username:  userRequest.Username,
			Action:    "update",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "failed",
			Message:   err.Error(),
		})
		return model.UserResponse{}, err
	}

	// Log successful update
	u.userLogRepo.CreateLog(model.UserLogRequest{
		UserId:    updatedUser.Id,
		Username:  updatedUser.Username,
		Action:    "update",
		IpAddress: ipAddress,
		UserAgent: userAgent,
		Status:    "success",
		Message:   "User updated successfully",
	})

	var userResponse model.UserResponse
	err = copier.Copy(&userResponse, &updatedUser)
	if err != nil {
		return model.UserResponse{}, err
	}

	return userResponse, nil
}

// delete user
func (u *UserService) DeleteUser(id string, ipAddress, userAgent string) error {
	// Get user info first for logging
	user, err := u.userRepo.GetUserById(id)
	if err != nil {
		// Log failed attempt
		u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    id,
			Action:    "delete",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "failed",
			Message:   "User not found",
		})
		return err
	}

	// call repository to delete user
	err = u.userRepo.DeleteUser(id)
	if err != nil {
		// Log failed attempt
		u.userLogRepo.CreateLog(model.UserLogRequest{
			UserId:    id,
			Username:  user.Username,
			Action:    "delete",
			IpAddress: ipAddress,
			UserAgent: userAgent,
			Status:    "failed",
			Message:   err.Error(),
		})
		return err
	}

	// Log successful deletion
	u.userLogRepo.CreateLog(model.UserLogRequest{
		UserId:    id,
		Username:  user.Username,
		Action:    "delete",
		IpAddress: ipAddress,
		UserAgent: userAgent,
		Status:    "success",
		Message:   "User deleted successfully",
	})

	return nil
}

// GetAll - get all users with filtering and pagination
func (u *UserService) GetAll(req model.UserListRequest) ([]model.UserResponse, int64, map[string]interface{}, error) {
	// Set defaults for pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Get users from repository
	users, total, err := u.userRepo.GetAll(req)
	if err != nil {
		return nil, 0, nil, err
	}

	// Convert to response format (exclude sensitive data like password)
	var userResponses []model.UserResponse
	for _, user := range users {
		userResponse := model.UserResponse{
			Id:          user.Id,
			Username:    user.Username,
			Level:       user.Level,
			LoginCount:  user.LoginCount,
			LastLoginAt: user.LastLoginAt,
			LastLoginIp: user.LastLoginIp,
		}
		userResponses = append(userResponses, userResponse)
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit)) // Ceiling division
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	// Prepare metadata
	meta := map[string]interface{}{
		"page":        req.Page,
		"limit":       req.Limit,
		"total_count": total,
		"total_pages": totalPages,
		"has_next":    hasNext,
		"has_prev":    hasPrev,
	}

	return userResponses, total, meta, nil
}

// GetProfile - get user profile by id
func (u *UserService) GetProfile(id string) (model.UserResponse, error) {
	// Get user from repository
	user, err := u.userRepo.GetProfile(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.UserResponse{}, errors.New("user not found")
		}
		return model.UserResponse{}, err
	}

	// Convert to response format (exclude password)
	userResponse := model.UserResponse{
		Id:          user.Id,
		Username:    user.Username,
		Level:       user.Level,
		LoginCount:  user.LoginCount,
		LastLoginAt: user.LastLoginAt,
		LastLoginIp: user.LastLoginIp,
	}

	return userResponse, nil
}

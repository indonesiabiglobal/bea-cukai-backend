package userService

import (
	"Bea-Cukai/helper"
	"Bea-Cukai/model"
	"Bea-Cukai/repo/userRepository"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *userRepository.UserRepository
}

func NewUserService(userRepository *userRepository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepository,
	}
}

// CreateUser implements UserService
func (u *UserService) CreateUser(userRequest model.UserRequest) (model.UserResponse, error) {
	// validate id_user
	_, err := u.userRepo.GetUserByIdUser(userRequest.IdUser)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, errors.New("id_user already exists")
	}

	// validate nm_user
	_, err = u.userRepo.GetUserByNmUser(userRequest.NmUser)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, errors.New("nm_user already exists")
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
		return model.UserResponse{}, err
	}

	var userResponse model.UserResponse
	err = copier.Copy(&userResponse, &createdUser)
	if err != nil {
		return model.UserResponse{}, err
	}

	return userResponse, nil
}

// LoginUser implements UserService
func (u *UserService) LoginUser(userLogin model.UserLoginRequest) (string, error) {
	// call repository to get user
	user, err := u.userRepo.LoginUser(userLogin)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("username or password is incorrect")
		}
		return "", err
	}

	// match := helper.CheckPasswordHash(userLogin.Password, user.Password)
	match := userLogin.Password == user.Password
	if !match {
		return "", errors.New("username or password is incorrect")
	}

	token, err := helper.GenerateToken(user.IdUser, user.NmUser)
	if err != nil {
		return "", err
	}

	return token, nil
}

// update user
func (u *UserService) UpdateUser(userRequest model.UserUpdateRequest, idUser string) (model.UserResponse, error) {
	// validate nm_user
	user, err := u.userRepo.GetUserByNmUser(userRequest.NmUser)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound && user.IdUser != idUser {
		return model.UserResponse{}, errors.New("nm_user already exists")
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
	updatedUser, err := u.userRepo.UpdateUser(userRequest, idUser)
	if err != nil {
		return model.UserResponse{}, err
	}

	var userResponse model.UserResponse
	err = copier.Copy(&userResponse, &updatedUser)
	if err != nil {
		return model.UserResponse{}, err
	}

	return userResponse, nil
}

// delete user
func (u *UserService) DeleteUser(idUser string) error {
	// call repository to delete user
	err := u.userRepo.DeleteUser(idUser)
	if err != nil {
		return err
	}

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
			IdUser: user.IdUser,
			NmUser: user.NmUser,
			Level:  user.Level,
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

// GetProfile - get user profile by id_user
func (u *UserService) GetProfile(idUser string) (model.UserResponse, error) {
	// Get user from repository
	user, err := u.userRepo.GetProfile(idUser)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.UserResponse{}, errors.New("user not found")
		}
		return model.UserResponse{}, err
	}

	// Convert to response format (exclude password)
	userResponse := model.UserResponse{
		IdUser: user.IdUser,
		NmUser: user.NmUser,
		Level:  user.Level,
	}

	return userResponse, nil
}

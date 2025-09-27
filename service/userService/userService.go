package userService

import (
	"Dashboard-TRDP/helper"
	"Dashboard-TRDP/model"
	"Dashboard-TRDP/repo/userRepository"
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
	// validate email
	_, err := u.userRepo.GetUserByEmail(userRequest.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, errors.New("email already exists")
	}

	// validate username
	_, err = u.userRepo.GetUserByUsername(userRequest.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound {
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
			return "", errors.New("email or password is incorrect")
		}
		return "", err
	}

	match := helper.CheckPasswordHash(userLogin.Password, user.Password)
	if !match {
		return "", errors.New("email or password is incorrect")
	}

	token, err := helper.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", err
	}

	return token, nil
}

// update user
func (u *UserService) UpdateUser(userRequest model.UserUpdateRequest, userID uint) (model.UserResponse, error) {
	// validate email
	user, err := u.userRepo.GetUserByEmail(userRequest.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound && user.ID != userID {
		return model.UserResponse{}, errors.New("email already exists")
	}

	// validate username
	user, err = u.userRepo.GetUserByUsername(userRequest.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return model.UserResponse{}, err
	}
	if err != gorm.ErrRecordNotFound && user.ID != userID {
		return model.UserResponse{}, errors.New("username already exists")
	}

	// call repository to update user
	updatedUser, err := u.userRepo.UpdateUser(userRequest, userID)
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
func (u *UserService) DeleteUser(userID uint) error {
	// call repository to delete user
	err := u.userRepo.DeleteUser(userID)
	if err != nil {
		return err
	}

	return nil
}

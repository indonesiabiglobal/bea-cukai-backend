package userRepository

import (
	"Dashboard-TRDP/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// CreateUser implements UserRepository
func (u *UserRepository) CreateUser(user model.UserRequest) (model.User, error) {
	userModel := model.User{
		Username:        user.Username,
		Email:           user.Email,
		Password:        user.Password,
		Age:             user.Age,
		ProfileImageUrl: user.ProfileImageUrl,
	}
	err := u.db.Create(&userModel).Error
	if err != nil {
		return model.User{}, err
	}
	return userModel, nil
}

// LoginUser implements UserRepository
func (u *UserRepository) LoginUser(userLogin model.UserLoginRequest) (model.User, error) {
	user := model.User{}
	err := u.db.Where("email = ?", userLogin.Email).First(&user).Error
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

// UpdateUser implements UserRepository
func (u *UserRepository) UpdateUser(userRequest model.UserUpdateRequest, userID uint) (model.User, error) {
	var user model.User

	// Mencari pengguna dengan userID yang diberikan
	err := u.db.First(&user, userID).Error
	if err != nil {
		return model.User{}, err
	}

	user.Username = userRequest.Username
	user.Email = userRequest.Email
	user.Age = userRequest.Age
	user.ProfileImageUrl = userRequest.ProfileImageUrl

	err = u.db.Save(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// DeleteUser implements UserRepository
func (u *UserRepository) DeleteUser(userID uint) error {
	err := u.db.Unscoped().Delete(&model.User{}, userID)
	if err != nil {
		if err.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return err.Error
	}
	return nil
}

// Get user by email
func (u *UserRepository) GetUserByEmail(email string) (model.User, error) {
	var user model.User
	err := u.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// get user by username
func (u *UserRepository) GetUserByUsername(username string) (model.User, error) {
	var user model.User
	err := u.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

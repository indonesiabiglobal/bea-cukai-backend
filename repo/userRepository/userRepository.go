package userRepository

import (
	"Bea-Cukai/model"

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
		Id:   user.Id,
		Username:   user.Username,
		Password: user.Password,
		Level:    user.Level,
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
	err := u.db.Where("username = ?", userLogin.Username).First(&user).Error
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

// UpdateUser implements UserRepository
func (u *UserRepository) UpdateUser(userRequest model.UserUpdateRequest, id string) (model.User, error) {
	var user model.User

	// Mencari pengguna dengan id yang diberikan
	err := u.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return model.User{}, err
	}

	user.Username = userRequest.Username
	if userRequest.Password != "" {
		user.Password = userRequest.Password
	}
	user.Level = userRequest.Level

	err = u.db.Save(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// DeleteUser implements UserRepository
func (u *UserRepository) DeleteUser(id string) error {
	result := u.db.Where("id = ?", id).Delete(&model.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetUserByUsername - get user by username (username)
func (u *UserRepository) GetUserByUsername(username string) (model.User, error) {
	var user model.User
	err := u.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// GetUserById - get user by id
func (u *UserRepository) GetUserById(id string) (model.User, error) {
	var user model.User
	err := u.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// GetProfile - get user profile by id (alias for GetUserById for clarity)
func (u *UserRepository) GetProfile(id string) (model.User, error) {
	return u.GetUserById(id)
}

// GetAll - get all users with filtering and pagination
func (u *UserRepository) GetAll(req model.UserListRequest) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// Build the base query
	query := u.db.Model(&model.User{})

	// Apply filters
	if req.Id != "" {
		query = query.Where("id LIKE ?", "%"+req.Id+"%")
	}
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Level != "" {
		query = query.Where("level = ?", req.Level)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Page > 0 && req.Limit > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Execute query with ordering
	if err := query.Order("id ASC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateLoginInfo - update login count, last login time and IP
func (u *UserRepository) UpdateLoginInfo(id string, ipAddress string) error {
	updates := map[string]interface{}{
		"login_count":   gorm.Expr("login_count + 1"),
		"last_login_at": gorm.Expr("NOW()"),
		"last_login_ip": ipAddress,
	}

	result := u.db.Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

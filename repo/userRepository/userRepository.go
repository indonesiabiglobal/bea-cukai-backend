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
		IdUser:   user.IdUser,
		NmUser:   user.NmUser,
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
	err := u.db.Where("nm_user = ?", userLogin.NmUser).First(&user).Error
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

// UpdateUser implements UserRepository
func (u *UserRepository) UpdateUser(userRequest model.UserUpdateRequest, idUser string) (model.User, error) {
	var user model.User

	// Mencari pengguna dengan idUser yang diberikan
	err := u.db.Where("id_user = ?", idUser).First(&user).Error
	if err != nil {
		return model.User{}, err
	}

	user.NmUser = userRequest.NmUser
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
func (u *UserRepository) DeleteUser(idUser string) error {
	result := u.db.Where("id_user = ?", idUser).Delete(&model.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetUserByNmUser - get user by nm_user (username)
func (u *UserRepository) GetUserByNmUser(nmUser string) (model.User, error) {
	var user model.User
	err := u.db.Where("nm_user = ?", nmUser).First(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// GetUserByIdUser - get user by id_user
func (u *UserRepository) GetUserByIdUser(idUser string) (model.User, error) {
	var user model.User
	err := u.db.Where("id_user = ?", idUser).First(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// GetProfile - get user profile by id_user (alias for GetUserByIdUser for clarity)
func (u *UserRepository) GetProfile(idUser string) (model.User, error) {
	return u.GetUserByIdUser(idUser)
}

// GetAll - get all users with filtering and pagination
func (u *UserRepository) GetAll(req model.UserListRequest) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// Build the base query
	query := u.db.Model(&model.User{})

	// Apply filters
	if req.IdUser != "" {
		query = query.Where("id_user LIKE ?", "%"+req.IdUser+"%")
	}
	if req.NmUser != "" {
		query = query.Where("nm_user LIKE ?", "%"+req.NmUser+"%")
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
	if err := query.Order("id_user ASC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

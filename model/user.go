package model

import "time"

type User struct {
	Id            string     `json:"id" gorm:"primaryKey;column:id"`
	Username      string     `json:"username" gorm:"column:username;not null"`
	Password      string     `json:"password" gorm:"column:password;not null"`
	Level         string     `json:"level" gorm:"column:level;not null"`
	LoginCount    int        `json:"login_count" gorm:"column:login_count;default:0"`
	LastLoginAt   *time.Time `json:"last_login_at" gorm:"column:last_login_at"`
	LastLoginIp   string     `json:"last_login_ip" gorm:"column:last_login_ip"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "user"
}

type UserRequest struct {
	Id       string `json:"id" form:"id" validate:"required"`
	Username string `json:"username" form:"username" validate:"required"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
	Level    string `json:"level" form:"level" validate:"required"`
}

type UserResponse struct {
	Id          string     `json:"id"`
	Username    string     `json:"username"`
	Level       string     `json:"level"`
	LoginCount  int        `json:"login_count"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIp string     `json:"last_login_ip"`
}

type UserResponseAssociation struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Level    string `json:"level"`
}
type UserLoginRequest struct {
	Username string `json:"username" form:"username" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
}

type UserUpdateRequest struct {
	Username string `json:"username" form:"username" validate:"required"`
	Password string `json:"password" form:"password" validate:"omitempty,min=6"`
	Level    string `json:"level" form:"level" validate:"required"`
}

type UserListRequest struct {
	Id       string `json:"id" form:"id"`
	Username string `json:"username" form:"username"`
	Level    string `json:"level" form:"level"`
	Page     int    `json:"page" form:"page"`
	Limit    int    `json:"limit" form:"limit"`
}

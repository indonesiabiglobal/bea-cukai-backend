package model

type User struct {
	IdUser   string `json:"id_user" gorm:"primaryKey;column:id_user"`
	NmUser   string `json:"nm_user" gorm:"column:nm_user;not null"`
	Password string `json:"password" gorm:"column:password;not null"`
	Level    string `json:"level" gorm:"column:level;not null"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "user"
}

type UserRequest struct {
	IdUser   string `json:"id_user" form:"id_user" validate:"required"`
	NmUser   string `json:"nm_user" form:"nm_user" validate:"required"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
	Level    string `json:"level" form:"level" validate:"required"`
}

type UserResponse struct {
	IdUser string `json:"id_user"`
	NmUser string `json:"nm_user"`
	Level  string `json:"level"`
}

type UserResponseAssociation struct {
	IdUser string `json:"id_user"`
	NmUser string `json:"nm_user"`
	Level  string `json:"level"`
}
type UserLoginRequest struct {
	NmUser   string `json:"nm_user" form:"nm_user" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
}

type UserUpdateRequest struct {
	NmUser   string `json:"nm_user" form:"nm_user" validate:"required"`
	Password string `json:"password" form:"password" validate:"omitempty,min=6"`
	Level    string `json:"level" form:"level" validate:"required"`
}

type UserListRequest struct {
	IdUser string `json:"id_user" form:"id_user"`
	NmUser string `json:"nm_user" form:"nm_user"`
	Level  string `json:"level" form:"level"`
	Page   int    `json:"page" form:"page"`
	Limit  int    `json:"limit" form:"limit"`
}

package model

type ItemGroup struct {
	Idx       int    `json:"idx" gorm:"primaryKey;autoIncrement"`
	ItemGroup string `json:"item_group" gorm:"type:varchar(255);not null"`
}

// TableName overrides the default table name.
func (ItemGroup) TableName() string { return "sys_item_group" }

type ItemGroupRequest struct {
	ItemGroup string `json:"item_group" validate:"required"`
}

type ItemGroupResponse struct {
	Idx       int    `json:"idx"`
	ItemGroup string `json:"item_group"`
}

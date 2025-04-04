package models

type ManagementUser struct {
	UserBaseModel
	IsAdmin bool `gorm:"column:is_admin"`
}

func (ManagementUser) TableName() string {
	return "management_users"
}

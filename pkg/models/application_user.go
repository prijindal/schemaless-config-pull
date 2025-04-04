package models

type ApplicationUser struct {
	UserBaseModel
	ApplicationID string `gorm:"column:application_id"`
	Application   Application
}

func (ApplicationUser) TableName() string {
	return "application_users"
}

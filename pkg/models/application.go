package models

type Application struct {
	BaseModel
	Name    string `gorm:"column:name"`
	OwnerID string `gorm:"column:owner_id"`
	Owner   ManagementUser
}

func (Application) TableName() string {
	return "applications"
}

package models

import "encoding/json"

type EntityHistory struct {
	BaseModel
	AppUserID     string `gorm:"column:app_user_id"`
	AppUser       ApplicationUser
	ApplicationID string `gorm:"column:application_id"`
	Application   Application
	EntityName    string          `gorm:"column:entity_name"`
	HostID        string          `gorm:"column:host_id"`
	EntityID      string          `gorm:"column:entity_id"`
	Action        string          `gorm:"column:action"`
	Payload       json.RawMessage `gorm:"column:payload;type:jsonb"`
}

func (EntityHistory) TableName() string {
	return "entity_history"
}

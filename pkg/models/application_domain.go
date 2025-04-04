package models

import "time"

type ApplicationDomain struct {
	ID            string    `gorm:"column:id;type:uuid;default:uuid_generate_v4()"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamptz"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamptz"`
	DomainName    string    `gorm:"column:domain_name"`
	ApplicationId string    `gorm:"column:application_id"`
	OwnerId       string    `gorm:"column:owner_id"`
	SoaEmail      string    `gorm:"column:soa_email"`
	Status        string    `gorm:"column:status"`
}

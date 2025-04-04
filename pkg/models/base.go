package models

import (
	"time"

	"gorm.io/gorm"
)

type UserStatus string

const (
	UserActivated   UserStatus = "ACTIVATED"
	UserDeactivated UserStatus = "DEACTIVATED"
	UserUnverified  UserStatus = "UNVERIFIED"
)

type BaseModel struct {
	gorm.Model
	ID        string    `gorm:"column:id;type:uuid"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz"`
}

type UserBaseModel struct {
	BaseModel
	Email      string     `gorm:"column:email"`
	BcryptHash string     `gorm:"column:bcrypt_hash"`
	Token      string     `gorm:"column:token,type:uuid"`
	Status     UserStatus `gorm:"column:status"`
}

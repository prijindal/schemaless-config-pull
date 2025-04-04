// repository will include all APIs related to management User
package repository

import (
	"errors"
	"schemaless/config-pull/pkg/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ManagementUserLoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ManagementUserLoginResponse struct {
	ID      string `json:"id"`
	IsAdmin bool   `json:"is_admin"`
}

func IsInitialized(db *gorm.DB) (bool, error) {
	var users []models.ManagementUser
	tx := db.Where(&models.ManagementUser{IsAdmin: true}).Limit(1).Find(&users)
	if tx.Error != nil {
		return false, tx.Error
	}
	return len(users) == 1, nil
}

func InitailizeWithUser(db *gorm.DB, input ManagementUserLoginBody) (*ManagementUserLoginResponse, error) {
	initialized, err := IsInitialized(db)
	if err != nil {
		return nil, err
	}
	if initialized {
		return nil, errors.New("admin user already exists")
	}
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, err
	}
	user := models.ManagementUser{
		UserBaseModel: models.UserBaseModel{
			BaseModel: models.BaseModel{
				ID: uuid.New().String(),
			},
			Email:      input.Email,
			BcryptHash: string(bcryptHash),
			Token:      uuid.New().String(),
			Status:     models.UserActivated,
		},
		IsAdmin: true,
	}
	tx := db.Create(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &ManagementUserLoginResponse{
		ID:      user.ID,
		IsAdmin: user.IsAdmin,
	}, nil
}

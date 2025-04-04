// repository will include all APIs related to management User
package repository

import (
	"errors"
	"schemaless/config-pull/pkg/database"
	"schemaless/config-pull/pkg/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ManagementUserRepository struct {
	database.DatabaseManager
}

type ManagementUserLoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ManagementUserRegisterResponse struct {
	ID      string `json:"id"`
	IsAdmin bool   `json:"is_admin"`
}

type ManagementUserLoginResponse struct {
	Token string `json:"token"`
}

func (r ManagementUserRepository) IsInitialized() (bool, error) {
	var users []models.ManagementUser
	tx := r.DB.Where(&models.ManagementUser{IsAdmin: true}).Limit(1).Find(&users)
	if tx.Error != nil {
		return false, tx.Error
	}
	return len(users) == 1, nil
}

func (r ManagementUserRepository) InitailizeWithUser(input ManagementUserLoginBody) (*ManagementUserRegisterResponse, error) {
	initialized, err := r.IsInitialized()
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
	tx := r.DB.Create(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &ManagementUserRegisterResponse{
		ID:      user.ID,
		IsAdmin: user.IsAdmin,
	}, nil
}

func (r ManagementUserRepository) GetUserWithEmail(email string) (*models.ManagementUser, bool, error) {
	var users []models.ManagementUser
	tx := r.DB.Where(&models.ManagementUser{UserBaseModel: models.UserBaseModel{Email: email}}).Limit(1).Find(&users)
	if tx.Error != nil {
		return nil, false, tx.Error
	}
	if len(users) == 0 {
		return nil, false, nil
	}
	return &users[0], true, nil
}

func (r ManagementUserRepository) RegisterUser(input ManagementUserLoginBody) (*ManagementUserRegisterResponse, error) {
	initialized, err := r.IsInitialized()
	if err != nil {
		return nil, err
	}
	if !initialized {
		return nil, errors.New("admin user does not exist")
	}
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, err
	}
	_, exists, err := r.GetUserWithEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user already exists")
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
		IsAdmin: false,
	}
	tx := r.DB.Create(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &ManagementUserRegisterResponse{
		ID:      user.ID,
		IsAdmin: user.IsAdmin,
	}, nil
}

func (r ManagementUserRepository) LoginUser(input ManagementUserLoginBody) (string, error) {
	user, exists, err := r.GetUserWithEmail(input.Email)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("uusername or password is incorrect")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.BcryptHash), []byte(input.Password))
	if err != nil {
		return "", err
	}
	if user.Status != models.UserActivated {
		return "", errors.New("user is deactivated")
	}
	// TODO: Generate token
	return "", nil
}

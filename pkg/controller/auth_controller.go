package controller

import (
	"context"
	"schemaless/config-pull/pkg/repository"

	"github.com/swaggest/rest/web"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

type AuthController struct {
	repository.ManagementUserRepository
}

func (c AuthController) RegisterRoutes(s *web.Service) {
	s.Get("/api/auth/initialized", c.IsInitialized())
	s.Post("/api/auth/initialize", c.InitializeAdmin())
	s.Post("/api/auth/register", c.RegisterUser())
	s.Post("/api/auth/login", c.LoginUser())
}

func (c AuthController) IsInitialized() usecase.Interactor {
	type isInitializedOutput struct {
		Initialized bool `json:"initialized"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input EmptyInput, output *isInitializedOutput) error {
		initialized, err := c.ManagementUserRepository.IsInitialized()
		if err != nil {
			return status.Wrap(err, status.NotFound)
		}
		output.Initialized = initialized
		return nil
	})
	u.SetTags("Management Auth")
	return u
}

func (c AuthController) InitializeAdmin() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input repository.ManagementUserLoginBody, output *repository.ManagementUserRegisterResponse) error {
		body, err := c.ManagementUserRepository.InitailizeWithUser(input)
		if err != nil {
			return status.Wrap(err, status.NotFound)
		}
		output.ID = body.ID
		output.IsAdmin = body.IsAdmin
		return nil
	})
	u.SetTags("Management Auth")
	return u
}

func (c AuthController) RegisterUser() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input repository.ManagementUserLoginBody, output *repository.ManagementUserRegisterResponse) error {
		body, err := c.ManagementUserRepository.RegisterUser(input)
		if err != nil {
			return status.Wrap(err, status.NotFound)
		}
		output.ID = body.ID
		output.IsAdmin = body.IsAdmin
		return nil
	})
	u.SetTags("Management Auth")
	return u
}

func (c AuthController) LoginUser() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input repository.ManagementUserLoginBody, output *repository.ManagementUserLoginResponse) error {
		body, err := c.ManagementUserRepository.LoginUser(input)
		if err != nil {
			return status.Wrap(err, status.NotFound)
		}
		output.Token = body
		return nil
	})
	u.SetTags("Management Auth")
	return u
}

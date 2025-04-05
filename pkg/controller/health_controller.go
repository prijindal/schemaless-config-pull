package controller

import (
	"context"
	"schemaless/config-pull/pkg/database"

	"github.com/swaggest/rest/web"
	"github.com/swaggest/usecase"
)

type HealthController struct {
	database.DatabaseManager
}

func (c HealthController) RegisterRoutes(s *web.Service) {
	s.Get("/api/health", c.GetHealth())
}

func (c HealthController) GetHealth() usecase.Interactor {
	type healthOutput struct {
		Healthy bool `json:"healthy"`
	}
	u := usecase.NewInteractor(func(ctx context.Context, input EmptyInput, output *healthOutput) error {
		db, err := c.DB.DB()
		if err != nil {
			output.Healthy = false
		} else {
			err = db.Ping()
			if err != nil {
				output.Healthy = false
			} else {
				output.Healthy = true
			}
		}
		return nil
	})
	u.SetTags("health")
	return u
}

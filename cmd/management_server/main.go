package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"schemaless/config-pull/pkg/database"
	"schemaless/config-pull/pkg/repository"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/swaggest/openapi-go/openapi31"
	"github.com/swaggest/rest/response/gzip"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v5emb"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

type emptyInput struct{}

func isInitialized(db *gorm.DB) usecase.Interactor {
	type isInitializedOutput struct {
		Initialized bool `json:"initialized"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input emptyInput, output *isInitializedOutput) error {
		initialized, err := repository.IsInitialized(db)
		if err != nil {
			return status.Wrap(err, status.NotFound)
		}
		output.Initialized = initialized
		return nil
	})
	u.SetTags("Management Auth")
	return u
}

func initializeAdmin(db *gorm.DB) usecase.Interactor {

	u := usecase.NewInteractor(func(ctx context.Context, input repository.ManagementUserLoginBody, output *repository.ManagementUserLoginResponse) error {
		body, err := repository.InitailizeWithUser(db, input)
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

func getHealth() usecase.Interactor {

	type healthOutput struct {
		Healthy bool `json:"healthy"`
	}
	u := usecase.NewInteractor(func(ctx context.Context, input emptyInput, output *healthOutput) error {
		output.Healthy = true
		return nil
	})
	u.SetTags("health")
	return u
}

func getHello() usecase.Interactor {
	// Declare input port type.
	type helloInput struct {
		Locale string `query:"locale" default:"en-US" pattern:"^[a-z]{2}-[A-Z]{2}$" enum:"ru-RU,en-US"`
		Name   string `path:"name" minLength:"3"` // Field tags define parameter location and JSON schema constraints.

		// Field tags of unnamed fields are applied to parent schema.
		// they are optional and can be used to disallow unknown parameters.
		// For non-body params, name tag must be provided explicitly.
		// E.g. here no unknown `query` and `cookie` parameters allowed,
		// unknown `header` params are ok.
		_ struct{} `query:"_" cookie:"_" additionalProperties:"false"`
	}

	// Declare output port type.
	type helloOutput struct {
		Now     time.Time `header:"X-Now" json:"-"`
		Message string    `json:"message"`
	}

	messages := map[string]string{
		"en-US": "Hello, %s!",
		"ru-RU": "Привет, %s!",
	}

	u := usecase.NewInteractor(func(ctx context.Context, input helloInput, output *helloOutput) error {
		msg, available := messages[input.Locale]
		if !available {
			return status.Wrap(errors.New("unknown locale"), status.InvalidArgument)
		}

		output.Message = fmt.Sprintf(msg, input.Name)
		output.Now = time.Now()

		return nil
	})
	u.SetTags("test")
	return u
}

func main() {
	db, err := database.NewGormConnectionFromString()
	if err != nil {
		panic(err)
	}
	err = database.PerformMigration(db)
	if err != nil {
		log.Error(err)
	}

	s := web.NewService(openapi31.NewReflector())

	// Init API documentation schema.
	s.OpenAPISchema().SetTitle("Basic Example")
	s.OpenAPISchema().SetDescription("This app showcases a trivial REST API.")
	s.OpenAPISchema().SetVersion("v1.2.3")

	// Setup middlewares.
	s.Wrap(
		gzip.Middleware, // Response compression with support for direct gzip pass through.
	)

	s.Get("/api/health", getHealth())

	s.Get("/api/auth/initialized", isInitialized(db))
	s.Post("/api/auth/initialize", initializeAdmin(db))

	// Add use case handler to router.
	s.Get("/hello/{name}", getHello())

	// Swagger UI endpoint at /docs.
	s.Docs("/docs", swgui.New)

	// Start server.
	log.Println("http://localhost:8011/docs")
	if err := http.ListenAndServe("localhost:8011", s); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"net/http"
	"schemaless/config-pull/pkg/controller"
	"schemaless/config-pull/pkg/database"
	"schemaless/config-pull/pkg/repository"

	log "github.com/sirupsen/logrus"

	"github.com/swaggest/openapi-go/openapi31"
	"github.com/swaggest/rest/response/gzip"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v5emb"
)

func main() {
	db_manager := database.DatabaseManager{}
	err := db_manager.NewGormConnectionFromString()
	if err != nil {
		panic(err)
	}
	err = db_manager.PerformMigration()
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

	managementUserRepository := repository.ManagementUserRepository{DatabaseManager: db_manager}
	authController := controller.AuthController{ManagementUserRepository: managementUserRepository}
	healthController := controller.HealthController{DatabaseManager: db_manager}
	authController.RegisterRoutes(s)
	healthController.RegisterRoutes(s)

	// Swagger UI endpoint at /docs.
	s.Docs("/docs", swgui.New)

	// Start server.
	log.Println("http://localhost:8011/docs")
	if err := http.ListenAndServe("localhost:8011", s); err != nil {
		log.Fatal(err)
	}
}

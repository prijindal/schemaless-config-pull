package database

import (
	"schemaless/config-pull/pkg/models"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func PerformMigration(db *gorm.DB) error {
	log.Info("Performing DB Migration")
	err := db.AutoMigrate(
		&models.ManagementUser{},
		&models.Application{},
		&models.ApplicationUser{},
		&models.ApplicationDomain{},
		&models.EntityHistory{},
	)
	return err
}

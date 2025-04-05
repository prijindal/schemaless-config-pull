package database

import (
	"database/sql"
	"schemaless/config-pull/pkg/config"
	"schemaless/config-pull/pkg/models"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseManager struct {
	DB *gorm.DB
}

func (m *DatabaseManager) NewGormConnectionFromString() error {
	postgresUri := config.Cfg.PostgresUri
	sqlDB, err := sql.Open("pgx", postgresUri)
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	newLogger := logger.New(
		log.StandardLogger(), // io writer
		logger.Config{SlowThreshold: time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
	)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 newLogger,
	})
	if err != nil {
		return err
	}
	m.DB = gormDB
	return nil

}

func (m *DatabaseManager) PerformMigration() error {
	log.Info("Performing DB Migration")
	err := m.DB.AutoMigrate(
		&models.ManagementUser{},
		&models.Application{},
		&models.ApplicationUser{},
		&models.ApplicationDomain{},
		&models.EntityHistory{},
	)
	return err
}

func (m *DatabaseManager) Close() error {
	db, err := m.DB.DB()
	if err != nil {
		return err
	}
	err = db.Close()
	if err != nil {
		return err
	}
	return nil
}

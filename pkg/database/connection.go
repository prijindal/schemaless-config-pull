package database

import (
	"database/sql"
	"schemaless/config-pull/pkg/config"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewGormConnectionFromString() (*gorm.DB, error) {
	postgresUri := config.Cfg.PostgresUri
	sqlDB, err := sql.Open("pgx", postgresUri)
	if err != nil {
		return nil, err
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
	return gormDB, err
}

package repository

import (
	"schemaless/config-pull/pkg/database"
	"schemaless/config-pull/pkg/models"
)

type ApplicationDomainRepository struct {
	database.DatabaseManager
}

func (r ApplicationDomainRepository) ListValidApplicationDomains() ([]models.ApplicationDomain, error) {

	var results []models.ApplicationDomain
	tx := r.DB.Table("application_domains").Where("status = ?", "ACTIVATED").Find(&results)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return results, nil
}

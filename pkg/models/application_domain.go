package models

type DomainStatus string

const (
	DomainActivated   DomainStatus = "ACTIVATED"
	DomainDeactivated DomainStatus = "DEACTIVATED"
	DomainUnverified  DomainStatus = "UNVERIFIED"
)

type ApplicationDomain struct {
	BaseModel
	DomainName    string `gorm:"column:domain_name"`
	ApplicationID string `gorm:"column:application_id"`
	Application   Application
	OwnerID       string `gorm:"column:owner_id"`
	Owner         ManagementUser
	SoaEmail      string       `gorm:"column:soa_email"`
	Status        DomainStatus `gorm:"column:status"`
	TxtRecord     string       `gorm:"column:txt_record"`
}

func (ApplicationDomain) TableName() string {
	return "application_domains"
}

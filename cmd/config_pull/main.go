package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"text/template"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

type ApplicationDomain struct {
	ID            string    `gorm:"column:id;type:uuid;default:uuid_generate_v4()"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamptz"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamptz"`
	DomainName    string    `gorm:"column:domain_name"`
	ApplicationId string    `gorm:"column:application_id"`
	OwnerId       string    `gorm:"column:owner_id"`
	SoaEmail      string    `gorm:"column:soa_email"`
	Status        string    `gorm:"column:status"`
}

func newGormConnectionFromString() (*gorm.DB, error) {
	postgresUri := os.Getenv("POSTGRES_URI")
	if postgresUri == "" {
		return nil, errors.New("POSTGRES_URI is not set")
	}
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

func loadEnv() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load() // The Original .env
	return env
}

func readBaseCaddyFile() (string, error) {
	// Read base caddyfile
	baseCaddyFilePath := os.Getenv("BASE_CADDY_FILE_PATH")
	if baseCaddyFilePath == "" {
		baseCaddyFilePath = "/etc/caddy/Caddyfile"
	}
	baseCaddyFile, err := os.ReadFile(baseCaddyFilePath)
	if err != nil {
		return "", err
	}
	baseCaddy := string(baseCaddyFile)
	return baseCaddy, nil
}

func loadConfigToCaddy() {

}

func reload() (*int, error) {
	log.Print("Reloading config")
	// Connect to postgres using env variable (POSTGRES_HOST can be a domain and will be resolved)
	gormdb, err := newGormConnectionFromString()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// Fetch list of domains
	var results []ApplicationDomain
	tx := gormdb.Table("application_domains").Where("status = ?", "ACTIVATED").Find(&results)
	if tx.Error != nil {
		log.Error(tx.Error)
		return nil, tx.Error
	}
	baseCadddy, err := readBaseCaddyFile()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// Create configuration file from domains and base caddyfile in caddyfile format
	templateString := `{{.DomainName}} {
    tls "{{.SoaEmail}}"
    reverse_proxy {
      import {{.PROXY_SNIPPET}}
      header_up Host {{.ApplicationId}}.{{.APPS_DOMAIN_NAME}}
    }
  }`
	tmpl, err := template.New("domaincaddy").Parse(templateString)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, result := range results {
		var buf bytes.Buffer
		dataMap := map[string]interface{}{
			"PROXY_SNIPPET":    os.Getenv("PROXY_SNIPPET"),
			"APPS_DOMAIN_NAME": os.Getenv("APPS_DOMAIN_NAME"),
			"ApplicationId":    result.ApplicationId,
			"DomainName":       result.DomainName,
			"SoaEmail":         result.SoaEmail,
		}
		err = tmpl.Execute(&buf, dataMap)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		baseCadddy += buf.String()
	}
	url := os.Getenv("CADDY_ADMIN_URL")
	if url == "" {
		url = "http://localhost:2019"
	}
	loadUrl := fmt.Sprintf("%s/load", url)
	log.Print(loadUrl)
	formattedCaddyfile := caddyfile.Format([]byte(baseCadddy))
	log.Print(string(formattedCaddyfile))
	resp, err := http.Post(loadUrl, "text/caddyfile", bytes.NewBuffer(formattedCaddyfile))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer resp.Body.Close() // Ensure the body is closed
	// Read the response body.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(respBody))
	// Call /load/ on caddy, using env variable ${CADDY_ADMIN_URL}/load/
	// Disconnect from postgres
	return nil, nil
}

func main() {
	// Load env variables
	loadEnv()
	s := gocron.NewScheduler(time.UTC)
	reloadInterval := os.Getenv("RELOAD_INTERVAL")
	if reloadInterval == "" {
		reloadInterval = "1m"
	}
	log.Printf("Configuring reload every %s", reloadInterval)
	s.Every(reloadInterval).Do(reload)
	s.StartAsync()
	// Call reload after every 1 minute

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)
	<-signals // wait for signal
	go func() {
		s.Clear()
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()
}

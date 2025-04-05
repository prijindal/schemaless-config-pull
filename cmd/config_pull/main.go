package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"schemaless/config-pull/pkg/config"
	"schemaless/config-pull/pkg/database"
	"schemaless/config-pull/pkg/models"
	"schemaless/config-pull/pkg/repository"
	"syscall"
	"time"

	"text/template"

	"github.com/go-co-op/gocron"

	log "github.com/sirupsen/logrus"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func readBaseCaddyFile() (string, error) {
	// Read base caddyfile
	baseCaddyFilePath := config.Cfg.BaseCaddyFilePath
	baseCaddyFile, err := os.ReadFile(baseCaddyFilePath)
	if err != nil {
		return "", err
	}
	baseCaddy := string(baseCaddyFile)
	return baseCaddy, nil
}

func loadConfigToCaddy(caddyContent string) error {
	url := config.Cfg.CaddyAdminUrl
	loadUrl := fmt.Sprintf("%s/load", url)
	log.Print(loadUrl)
	formattedCaddyfile := caddyfile.Format([]byte(caddyContent))
	log.Print(string(formattedCaddyfile))
	resp, err := http.Post(loadUrl, "text/caddyfile", bytes.NewBuffer(formattedCaddyfile))
	if err != nil {
		log.Error(err)
		return err
	}
	defer resp.Body.Close() // Ensure the body is closed
	// Read the response body.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
		return err
	}
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(respBody))
	return nil
}

func loadApplications() ([]models.ApplicationDomain, error) {
	// Connect to postgres using env variable (POSTGRES_HOST can be a domain and will be resolved)
	db_manager := database.DatabaseManager{}
	err := db_manager.NewGormConnectionFromString()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer db_manager.Close()
	applicationDomainRepository := repository.ApplicationDomainRepository{
		DatabaseManager: db_manager,
	}
	results, err := applicationDomainRepository.ListValidApplicationDomains()
	// Fetch list of domains
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return results, nil
}

func formApplicationsCaddy(applications []models.ApplicationDomain, baseCadddy string) (string, error) {
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
		return "", err
	}
	for _, result := range applications {
		var buf bytes.Buffer
		dataMap := map[string]interface{}{
			"PROXY_SNIPPET":    config.Cfg.ProxySnippet,
			"APPS_DOMAIN_NAME": config.Cfg.AppsDomainName,
			"ApplicationId":    result.ApplicationID,
			"DomainName":       result.DomainName,
			"SoaEmail":         result.SoaEmail,
		}
		err = tmpl.Execute(&buf, dataMap)
		if err != nil {
			log.Error(err)
			return "", err
		}
		baseCadddy += buf.String()
	}
	return baseCadddy, nil
}

func reload() error {
	log.Print("Reloading config")
	results, err := loadApplications()
	if err != nil {
		return err
	}
	baseCadddy, err := readBaseCaddyFile()
	if err != nil {
		log.Error(err)
		return err
	}
	changedCaddy, err := formApplicationsCaddy(results, baseCadddy)
	if err != nil {
		return err
	}
	err = loadConfigToCaddy(changedCaddy)
	if err != nil {
		return err
	}
	// Call /load/ on caddy, using env variable ${CADDY_ADMIN_URL}/load/
	// Disconnect from postgres
	return nil
}

func main() {
	// Load env variables
	s := gocron.NewScheduler(time.UTC)
	reloadInterval := config.Cfg.ReloadInterval
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

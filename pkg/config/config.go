package config

import (
	"os"
	"schemaless/config-pull/pkg/utils"

	log "github.com/sirupsen/logrus"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ReloadInterval    string `yaml:"reload_interval" env:"RELOAD_INTERVAL" env-default:"1m"`
	PostgresUri       string `yaml:"postgres_uri" env:"POSTGRES_URI" env-required:"true"`
	BaseCaddyFilePath string `yaml:"base_caddy_file_path" env:"BASE_CADDY_FILE_PATH" env-default:"/etc/caddy/Caddyfile"`
	CaddyAdminUrl     string `yaml:"caddy_admin_url" env:"CADDY_ADMIN_URL" env-default:"http://localhost:2019"`
	ProxySnippet      string `yaml:"proxy_snippet" env:"PROXY_SNIPPET" env-default:"schemaless-reverse-proxy"`
	AppsDomainName    string `yaml:"apps_domain_name" env:"APPS_DOMAIN_NAME" env-default:"apps.local.schemaless.click"`
}

var Cfg Config

func init() {
	configFileName := os.Getenv("CONFIG_PATH")
	if configFileName == "" {
		configFileName = "config.yaml"
	}
	if utils.CheckIfFileExists(configFileName) {
		log.Info("Loading config from config file")
		err := cleanenv.ReadConfig(configFileName, &Cfg)
		if err != nil {
			panic(err)
		}
	} else {
		log.Info("Loading config from env")
		err := cleanenv.ReadEnv(&Cfg)
		if err != nil {
			panic(err)
		}
	}
}

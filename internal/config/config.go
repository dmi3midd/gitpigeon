package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort    int    `mapstructure:"APP_PORT"`
	DBPath     string `mapstructure:"DB_PATH"`
	ApiKey     string `mapstructure:"API_KEY"`
	AppBaseURL string `mapstructure:"APP_BASE_URL"`

	GitHubToken  string     `mapstructure:"GITHUB_TOKEN"`
	ScanInterval int        `mapstructure:"SCAN_INTERVAL"` // in minutes
	SMTP         SMTPConfig `mapstructure:",squash"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     int    `mapstructure:"SMTP_PORT"`
	User     string `mapstructure:"SMTP_USER"`
	Password string `mapstructure:"SMTP_PASS"`
	From     string `mapstructure:"SMTP_FROM"`
}

func LoadConfig() *Config {
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_PATH", "./sql.db")
	viper.SetDefault("APP_BASE_URL", "http://localhost:8080")
	viper.SetDefault("SCAN_INTERVAL", 15)
	viper.SetDefault("SMTP_PORT", 587)

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, using environment variables and defaults")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	return &config
}

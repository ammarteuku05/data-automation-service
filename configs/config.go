package configs

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" required:"true" default:"app_user"`
	Password string `envconfig:"DB_PASSWORD" required:"true" default:"app_password"`
	Name     string `envconfig:"DB_NAME" required:"true" default:"app_db"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`

	URL   string `envconfig:"BI_API_URL"`
	Token string `envconfig:"BI_API_TOKEN"`

	MaxOpenConns int `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns int `envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
}

func (db Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode)
}

type APIConfig struct {
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to process envconfig: %v", err)
	}
	return &cfg
}

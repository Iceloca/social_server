package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string      `yaml:"env" env:"ENV" env-default:"local"`
	Postgres   PostgresCfg `yaml:"postgres" env-required:"true"`
	HTTPServer HTTPServer  `yaml:"http_server" env-required:"true"`
}

type PostgresCfg struct {
	Host     string `yaml:"host" env:"PG_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"PG_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"PG_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"PG_PASSWORD" env-default:"postgres"`
	DBName   string `yaml:"dbname" env:"PG_DBNAME" env-default:"postgres"`
	SSLMode  string `yaml:"sslmode" env:"PG_SSLMODE" env-default:"disable"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_ADDRESS" env-default:"0.0.0.0:8082"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file %s does not exist", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	return &cfg
}

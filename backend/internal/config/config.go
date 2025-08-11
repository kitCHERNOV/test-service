package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Config структура
type Config struct {
	HTTPServer         `yaml:"http_server"`
	PostgresConnection `yaml:"db_path"`
}

type PostgresConnection struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port" env:"DB_NAME" env-required:"true"`
	DBName   string `yaml:"database" env:"DB_NAME" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode" env:"SSL_MODE"`
	Username string `yaml:"username" env:"DB_USER" env-required:"true"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"localhost:8081"`
}

func MustLoad() *Config {
	const op = "config.config.MustLoad"
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Loc: %s, Err: %v", op, err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatalf("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

func (p *PostgresConnection) DataBasePath() (storagePath string) {
	// Собираем путь подключения
	storagePath = fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		p.Host,
		p.Username,
		p.Password,
		p.DBName,
		p.Port,
		p.SSLMode,
	)

	return storagePath
}

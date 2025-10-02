package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

// Определяем структуру Config для хранения конфигурационных данных

type Config struct {
	DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string

	JWTSecret string
}

func LoadConfig(l zerolog.Logger) *Config {
	// Загружаем .env файл (игнорируем ошибку если файла нет)
	_ = godotenv.Load(".env") 

	// Создаем экземпляр конфига
	cfg := &Config{}

	// Инициализируем данные из env файла для основной DB
	cfg.DBHost = "localhost"
	cfg.DBPort = getEnv("POSTGRES_PORT", "5432")
	cfg.DBUser = getEnv("POSTGRES_USER", "postgres")
	cfg.DBPassword = getEnv("POSTGRES_PASSWORD", "postgres")
	cfg.DBName = getEnv("POSTGRES_DB", "app")
	cfg.JWTSecret = getEnv("JWT_SECRET", "replace-this-secret")

	// Для запуска через Docker
	if getEnv("IS_DOCKER", "") == "true" {
		cfg.DBHost = getEnv("POSTGRES_HOST", "postgres")
	}

	l.Trace().Str("DBHost", cfg.DBHost).Str("DBPort", cfg.DBPort).Msg("Postgres config")

	// Возвращаем экземпляр конфига
	return cfg
}

func (c *Config) DBConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
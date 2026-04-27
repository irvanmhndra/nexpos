package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	JWT      JWTConfig
	Mongo    MongoConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type PostgresConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DB          string
	SSLMode     string
	DSNOverride string // used in tests to pass a testcontainer DSN directly
}

func (p PostgresConfig) DSN() string {
	if p.DSNOverride != "" {
		return p.DSNOverride
	}
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port + "/" + p.DB + "?sslmode=" + p.SSLMode
}

type MongoConfig struct {
	URI      string
	Database string
}

type JWTConfig struct {
	Secret             string
	AccessExpiresHours int
	RefreshExpiresDays int
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

func Load() *Config {
	accessHours := getEnvInt("JWT_ACCESS_EXPIRES_HOURS", 2)
	refreshDays := getEnvInt("JWT_REFRESH_EXPIRES_DAYS", 7)

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("SERVER_ENV", "development"),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "pos_user"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DB:       getEnv("POSTGRES_DB", "pos_db"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Mongo: MongoConfig{
			URI:      getEnv("MONGO_URI", ""),
			Database: getEnv("MONGO_DB", "nexpos"),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "secret"),
			AccessExpiresHours: accessHours,
			RefreshExpiresDays: refreshDays,
			AccessTokenExpiry:  time.Duration(accessHours) * time.Hour,
			RefreshTokenExpiry: time.Duration(refreshDays) * 24 * time.Hour,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

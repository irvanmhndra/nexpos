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
	R2       R2Config
}

type R2Config struct {
	AccountID     string
	AccessKeyID   string
	SecretKey     string
	Bucket        string
	Endpoint      string // https://<account>.r2.cloudflarestorage.com
	PublicBaseURL string // https://cdn.nexpos.irvanmahendra.com
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
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "secret"),
			AccessExpiresHours: accessHours,
			RefreshExpiresDays: refreshDays,
			AccessTokenExpiry:  time.Duration(accessHours) * time.Hour,
			RefreshTokenExpiry: time.Duration(refreshDays) * 24 * time.Hour,
		},
		R2: R2Config{
			AccountID:     getEnv("R2_ACCOUNT_ID", ""),
			AccessKeyID:   getEnv("R2_ACCESS_KEY_ID", ""),
			SecretKey:     getEnv("R2_SECRET_ACCESS_KEY", ""),
			Bucket:        getEnv("R2_BUCKET", ""),
			Endpoint:      getEnv("R2_ENDPOINT", ""),
			PublicBaseURL: getEnv("R2_PUBLIC_BASE_URL", ""),
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

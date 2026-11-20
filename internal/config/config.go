package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port string
	Env  string
}

func (s *ServerConfig) GetServerAddress() string {
	return ":" + s.Port
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Kuala_Lumpur",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret         string
	ExpiryHours    int
	RefreshExpDays int
}

func (j *JWTConfig) GetJWTSecret() string { return j.Secret }
func (j *JWTConfig) GetTokenExpiry() time.Duration {
	return time.Duration(j.ExpiryHours) * time.Hour
}
func (j *JWTConfig) GetRefreshExpiry() time.Duration {
	return time.Duration(j.RefreshExpDays) * 24 * time.Hour
}

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string
	AppURL   string
	Enabled  bool
}

type GeminiConfig struct {
	APIKey string
	Model  string
}

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Prefix    string
	UseSSL    bool
	PublicURL string
}

type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	Currency      string
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	S3       S3Config
	Stripe   StripeConfig
	Email    EmailConfig
	Gemini   GeminiConfig
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	jwtRefreshDays, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRATION_DAYS", "30"))
	s3UseSSL, _ := strconv.ParseBool(getEnv("S3_USE_SSL", "true"))
	smtpHost := getEnv("SMTP_HOST", "")

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("SERVER_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "ag_ecommerce_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "ascendgroup-secret-change-in-production"),
			ExpiryHours:    jwtExpiryHours,
			RefreshExpDays: jwtRefreshDays,
		},
		S3: S3Config{
			Endpoint:  getEnv("S3_ENDPOINT", ""),
			AccessKey: getEnv("S3_ACCESS_KEY", ""),
			SecretKey: getEnv("S3_SECRET_KEY", ""),
			Bucket:    getEnv("S3_BUCKET", "ag-ecommerce"),
			Prefix:    getEnv("S3_PREFIX", ""),
			UseSSL:    s3UseSSL,
			PublicURL: getEnv("S3_PUBLIC_URL", ""),
		},
		Stripe: StripeConfig{
			SecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
			Currency:      getEnv("STRIPE_CURRENCY", "myr"),
		},
		Email: EmailConfig{
			Host:     smtpHost,
			Port:     getEnv("SMTP_PORT", "465"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "no-reply@ascendgroup.my"),
			FromName: getEnv("SMTP_FROM_NAME", "Ascend Group"),
			AppURL:   getEnv("APP_URL", "http://localhost:5173"),
			Enabled:  smtpHost != "",
		},
		Gemini: GeminiConfig{
			APIKey: getEnv("GEMINI_API_KEY", ""),
			Model:  getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppURL          string
	BaseDomain      string
	DBPath          string
	UploadDir       string
	MaxUploadSize   int64
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPPass        string
	SMTPFrom        string
	SessionSecret   string
	ProvisionScript string
	TenantDir       string
	HMSSource       string
	WorkerCount     int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, reading from environment")
	}

	workerCount := 3
	if v := os.Getenv("WORKER_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			workerCount = n
		}
	}

	return &Config{
		AppURL:          getEnv("APP_URL", "https://master.hms.co.tz"),
		BaseDomain:      getEnv("BASE_DOMAIN", "hms.co.tz"),
		DBPath:          getEnv("DB_PATH", "./data/hms_master.db"),
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSize:   10 << 20, // 10MB
		SMTPHost:        getEnv("SMTP_HOST", ""),
		SMTPPort:        getEnv("SMTP_PORT", "587"),
		SMTPUser:        getEnv("SMTP_USER", ""),
		SMTPPass:        getEnv("SMTP_PASS", ""),
		SMTPFrom:        getEnv("SMTP_FROM", "noreply@hms.co.tz"),
		SessionSecret:   getEnv("SESSION_SECRET", "change-me-32-chars-minimum"),
		ProvisionScript: getEnv("PROVISION_SCRIPT", "/opt/hms-control/scripts/provision.sh"),
		TenantDir:       getEnv("TENANT_DIR", "/opt/tenants"),
		HMSSource:       getEnv("HMS_SOURCE", "/opt/hms-source"),
		WorkerCount:     workerCount,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
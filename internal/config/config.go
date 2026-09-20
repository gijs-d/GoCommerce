package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	AppEnv      string
	BaseURL     string
	JWTSecret   string
	DatabaseURL string

	// Cloudflare R2
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2PublicURL       string
}

func Load() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_insecure_development_secret_key_change_me_now!"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://shop_user:shop_password@localhost:5433/webshop_db?sslmode=disable"
	}

	return &Config{
		Port:              port,
		AppEnv:            appEnv,
		BaseURL:           baseURL,
		JWTSecret:         jwtSecret,
		DatabaseURL:       dbURL,
		R2AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      os.Getenv("R2_BUCKET_NAME"),
		R2PublicURL:       os.Getenv("R2_PUBLIC_URL"),
	}
}
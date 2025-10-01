package config

import (
	"os"
	"strings"
)

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

func NewMinioConfig() *MinioConfig {
	return &MinioConfig{
		Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9001"),
		AccessKey: getEnv("MINIO_ACCESS_KEY", "minio"),
		SecretKey: getEnv("MINIO_SECRET_KEY", "minio124"),
		UseSSL:    getEnvAsBool("MINIO_USE_SSL", false),
		Bucket:    getEnv("MINIO_BUCKET", "routes"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}
